package watcher

import (
	"context"
	"fmt"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"

	"github.com/fsnotify/fsnotify"
)

type Watcher struct {
	RootPath string
	Excludes []string

	absExcludes []string
	fswatcher   *fsnotify.Watcher
	events      chan Event
}

func New(rootPath string, excludes []string) *Watcher {
	return &Watcher{
		RootPath: rootPath,
		Excludes: excludes,
	}
}

func (w *Watcher) Watch(ctx context.Context, wg *sync.WaitGroup) (EventsChannel, error) {
	if w.events != nil {
		return w.events, nil
	}

	rootPath, err := w.prepareRoot()
	if err != nil {
		return nil, fmt.Errorf("prepareRoot: %w", err)
	}

	w.absExcludes = make([]string, 0, len(w.Excludes))
	for _, exclude := range w.Excludes {
		w.absExcludes = append(w.absExcludes, path.Join(rootPath, exclude))
	}

	w.fswatcher, err = fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("fsnotify.NewWatcher: %w", err)
	}

	if err := w.addRecursive(rootPath); err != nil {
		return nil, fmt.Errorf("addRecursive: %w", err)
	}

	w.events = make(chan Event)

	wg.Add(1)
	go func() {
		defer func() {
			w.fswatcher.Close()
			close(w.events)
			w.fswatcher = nil
			w.absExcludes = nil
			w.events = nil
			wg.Done()
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

				// log.Println("event:", event)
				if err := w.processEvent(ctx, event); err != nil {
					log.Println("error:", err)
				}

			case err, ok := <-w.fswatcher.Errors:
				if !ok {
					return
				}
				log.Println("error:", err)
			case <-ctx.Done():
				return
			}
		}
	}()

	return w.events, nil
}

func (w *Watcher) processEvent(ctx context.Context, source fsnotify.Event) error {
	// defer func() {
	// 	fmt.Printf("%+v\n", w.fswatcher.WatchList())
	// }()

	if source.Op == fsnotify.Chmod {
		return nil
	}
	if source.Name == "" || source.Name == "." {
		return nil
	}

	if !source.Has(fsnotify.Rename) && !source.Has(fsnotify.Remove) {
		fullpath := source.Name
		isDir, err := w.isDir(fullpath)
		if err != nil {
			return fmt.Errorf("isDir: %w", err)
		}
		if isDir {
			if source.Op.Has(fsnotify.Create) {
				w.addRecursive(fullpath)
			}
		}
	} else if source.Has(fsnotify.Remove) || source.Has(fsnotify.Rename) {
		wl := w.fswatcher.WatchList()
		for _, entry := range wl {
			if strings.HasPrefix(entry, source.Name) {
				w.fswatcher.Remove(entry)
			}
		}
	}

	var eventType EventType
	if source.Has(fsnotify.Remove) || source.Has(fsnotify.Rename) {
		eventType = EventRemoved
	} else if source.Has(fsnotify.Create) {
		eventType = EventCreated
	} else if source.Has(fsnotify.Write) {
		eventType = EventModified
	} else {
		return nil
	}

	relPath, err := filepath.Rel(w.RootPath, source.Name)
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

func (w *Watcher) isDir(fullpath string) (bool, error) {
	info, err := os.Stat(fullpath)
	if err != nil {
		return false, fmt.Errorf("os.Stat: %w", err)
	}

	return info.IsDir(), nil
}

func (w *Watcher) prepareRoot() (string, error) {
	rootPath, err := filepath.Abs(w.RootPath)
	if err != nil {
		return rootPath, fmt.Errorf("filepath.Abs: %w", err)
	}

	if ok, err := w.isDir(rootPath); err != nil {
		return rootPath, fmt.Errorf("isDir: %w", err)
	} else if !ok {
		return rootPath, fmt.Errorf("%s is not a directory", rootPath)
	}

	return rootPath, nil
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

		path := filepath.Join(path, entry.Name())
		err := w.addRecursive(path)
		if err != nil {
			return err
		}
	}

	return nil
}

func (w *Watcher) isExcluded(fullpath string) bool {
	for _, exclude := range w.absExcludes {
		if strings.HasPrefix(fullpath, exclude) {
			return true
		}
	}
	return false
}
