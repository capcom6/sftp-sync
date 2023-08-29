package watcher

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/fsnotify/fsnotify"
)

type Watcher struct {
	RootPath string
	Excludes []string
}

func New(rootPath string, excludes []string) *Watcher {
	return &Watcher{
		RootPath: rootPath,
		Excludes: excludes,
	}
}

func (w *Watcher) Watch(ctx context.Context) (EventsChannel, error) {
	rootPath, err := w.prepareRoot()
	if err != nil {
		return nil, fmt.Errorf("prepareRoot: %w", err)
	}

	fswatcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("fsnotify.NewWatcher: %w", err)
	}

	if err := w.addRecursive(fswatcher, rootPath); err != nil {
		return nil, fmt.Errorf("addRecursive: %w", err)
	}

	ch := make(chan Event)

	go func() {
		defer func() {
			fswatcher.Close()
			close(ch)
		}()

		for {
			select {
			case event, ok := <-fswatcher.Events:
				if !ok {
					return
				}
				log.Println("event:", event)
			case err, ok := <-fswatcher.Errors:
				if !ok {
					return
				}
				log.Println("error:", err)
			case <-ctx.Done():
				return
			}
		}
	}()

	return ch, nil
}

func (w *Watcher) prepareRoot() (string, error) {
	rootPath, err := filepath.Abs(w.RootPath)
	if err != nil {
		return "", fmt.Errorf("filepath.Abs: %w", err)
	}

	info, err := os.Stat(rootPath)
	if err != nil {
		return "", fmt.Errorf("os.Stat: %w", err)
	}

	if !info.IsDir() {
		return "", fmt.Errorf("rootPath is not a directory")
	}

	return rootPath, nil
}

func (w *Watcher) addRecursive(fswatcher *fsnotify.Watcher, path string) error {
	err := fswatcher.Add(path)
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
		err := w.addRecursive(fswatcher, path)
		if err != nil {
			return err
		}
	}

	return nil
}
