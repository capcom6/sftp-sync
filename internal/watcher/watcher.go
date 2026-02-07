package watcher

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/capcom6/logutils"
	"github.com/fsnotify/fsnotify"
)

type Watcher struct {
	rootPath string
	excludes []string

	absRootPath string
	absExcludes []string
	fswatcher   *fsnotify.Watcher
	events      chan Event
}

func New(rootPath string, excludes []string) *Watcher {
	return &Watcher{
		rootPath: rootPath,
		excludes: excludes,

		absRootPath: "",
		absExcludes: nil,
		fswatcher:   nil,
		events:      nil,
	}
}

func (w *Watcher) Watch(ctx context.Context, wg *sync.WaitGroup) (EventsChannel, error) {
	if w.events != nil {
		return w.events, nil
	}

	if err := w.prepareRoot(); err != nil {
		return nil, fmt.Errorf("prepareRoot: %w", err)
	}

	w.absExcludes = make([]string, 0, len(w.excludes))
	for _, exclude := range w.excludes {
		w.absExcludes = append(w.absExcludes, filepath.Join(w.absRootPath, exclude))
	}

	var err error
	w.fswatcher, err = fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("fsnotify.NewWatcher: %w", err)
	}

	if addErr := w.addRecursive(w.absRootPath); addErr != nil {
		_ = w.fswatcher.Close()
		w.fswatcher = nil
		return nil, fmt.Errorf("addRecursive: %w", addErr)
	}

	w.events = make(chan Event)

	wg.Add(1)
	go func() {
		defer wg.Done()
		w.runWatcher(ctx)
	}()

	return w.events, nil
}

func (w *Watcher) runWatcher(ctx context.Context) {
	defer func() {
		_ = w.fswatcher.Close()
		close(w.events)
		w.fswatcher = nil
		w.absExcludes = nil
		w.events = nil
	}()

	for {
		select {
		case event, ok := <-w.fswatcher.Events:
			if !ok {
				return
			}

			if w.isExcluded(event.Name) {
				continue
			}

			logutils.Debug("event:", event)
			if prErr := w.processEvent(ctx, event); prErr != nil {
				logutils.Error(prErr)
			}

		case watchErr, ok := <-w.fswatcher.Errors:
			if !ok {
				return
			}
			logutils.Error(watchErr)
		case <-ctx.Done():
			return
		}
	}
}

func (w *Watcher) processEvent(ctx context.Context, source fsnotify.Event) error {
	if source.Op == fsnotify.Chmod {
		return nil
	}
	if source.Name == "" || source.Name == "." {
		return nil
	}

	proceed, err := w.updateObservers(source)
	if err != nil {
		return fmt.Errorf("updateObservers: %w", err)
	}
	if !proceed {
		return nil
	}

	var eventType EventType
	switch {
	case source.Has(fsnotify.Remove), source.Has(fsnotify.Rename):
		eventType = EventRemoved
	case source.Has(fsnotify.Create):
		eventType = EventCreated
	case source.Has(fsnotify.Write):
		eventType = EventModified
	default:
		return nil
	}

	relPath, err := filepath.Rel(w.absRootPath, source.Name)
	if err != nil {
		return fmt.Errorf("filepath.Rel: %w", err)
	}

	event := Event{
		AbsPath: source.Name,
		RelPath: relPath,
		Type:    eventType,
	}

	select {
	case w.events <- event:
	case <-ctx.Done():
		return nil
	}

	return nil
}

// updateObservers handles fsnotify events for directories.
// When fsnotify.Remove or fsnotify.Rename event is received, it removes
// corresponding watcher from the list of watched paths.
// When fsnotify.Create event is received, it adds the watched path
// recursively.
// When fsnotify.Write event is received, it skips recursive sync.
// The function returns true if the event needs to be processed
// further and false otherwise.
func (w *Watcher) updateObservers(source fsnotify.Event) (bool, error) {
	if source.Has(fsnotify.Remove) || source.Has(fsnotify.Rename) {
		wl := w.fswatcher.WatchList()
		for _, entry := range wl {
			if entry == source.Name || strings.HasPrefix(entry, source.Name+string(filepath.Separator)) {
				_ = w.fswatcher.Remove(entry)
			}
		}

		return true, nil
	}

	fullpath := source.Name
	isDir, err := w.isDir(fullpath)
	if err != nil {
		return false, fmt.Errorf("isDir: %w", err)
	}

	if !isDir {
		return true, nil
	}

	if source.Op.Has(fsnotify.Create) {
		if addErr := w.addRecursive(fullpath); addErr != nil {
			return false, fmt.Errorf("addRecursive: %w", addErr)
		}
	} else if source.Op.Has(fsnotify.Write) {
		// when creating a file on windows we have two events:
		// fsnotify.Create for file and fsnotify.Write for parent directory,
		// so we need to ignore fsnotify.Write to skip recursive sync
		return false, nil
	}

	return true, nil
}

func (w *Watcher) isDir(fullpath string) (bool, error) {
	info, err := os.Stat(fullpath)
	if err != nil {
		return false, fmt.Errorf("os.Stat: %w", err)
	}

	return info.IsDir(), nil
}

func (w *Watcher) prepareRoot() error {
	rootPath, err := filepath.Abs(w.rootPath)
	if err != nil {
		return fmt.Errorf("filepath.Abs: %w", err)
	}

	if ok, dirErr := w.isDir(rootPath); dirErr != nil {
		return fmt.Errorf("isDir: %w", dirErr)
	} else if !ok {
		return fmt.Errorf("%w: %s", ErrIsNotDir, rootPath)
	}

	w.absRootPath = rootPath

	return nil
}

func (w *Watcher) addRecursive(path string) error {
	if w.isExcluded(path) {
		return nil
	}

	err := w.fswatcher.Add(path)
	if err != nil {
		return fmt.Errorf("fswatcher.Add: %w", err)
	}

	entries, err := os.ReadDir(path)
	if err != nil {
		return fmt.Errorf("os.ReadDir: %w", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		entryPath := filepath.Join(path, entry.Name())
		if addErr := w.addRecursive(entryPath); addErr != nil {
			return addErr
		}
	}

	return nil
}

func (w *Watcher) isExcluded(fullpath string) bool {
	for _, exclude := range w.absExcludes {
		if fullpath == exclude || strings.HasPrefix(fullpath, exclude+string(filepath.Separator)) {
			return true
		}
	}
	return false
}
