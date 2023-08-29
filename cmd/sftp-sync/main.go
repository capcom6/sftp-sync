package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"sync"

	"github.com/fsnotify/fsnotify"
)

func isDir(path string) (bool, error) {
	info, err := os.Stat(path)
	if err == nil {
		return info.IsDir(), nil
	}
	return false, err
}

func main() {
	// isDebug := flag.Bool("debug", false, "enable debug logging")
	dest := flag.String("dest", "", "destination server")
	// exclude := flag.String("exclude", "", "exclude paths")
	flag.Parse()

	source := flag.Arg(0)

	if *dest == "" {
		log.Fatalln("destination server is required")
	}
	if source == "" {
		log.Fatalln("source directory is required")
	}

	if ok, err := isDir(source); !ok {
		if err != nil {
			log.Fatalln(err)
		}
		log.Fatalln("source is not a directory")
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	wg := &sync.WaitGroup{}

	if err := startWatch(ctx, wg, source); err != nil {
		log.Fatalln(err)
	}

	log.Println("Watching...")
	wg.Wait()

	log.Println("Bye!")
}

func startWatch(ctx context.Context, wg *sync.WaitGroup, source string) error {
	// Create new watcher.
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("fsnotify.NewWatcher: %w", err)
	}

	// Add a path.
	source, err = filepath.Abs(source)
	if err != nil {
		return fmt.Errorf("filepath.Abs: %w", err)
	}

	err = watcher.Add(source)
	if err != nil {
		return fmt.Errorf("watcher.Add: %w", err)
	}

	// Start listening for events.
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				log.Println("event:", event)
				if event.Has(fsnotify.Remove) {
					// ignore error
					watcher.Remove(event.Name)
					log.Printf("%+v", watcher.WatchList())
				} else if event.Has(fsnotify.Create) {
					if ok, _ := isDir(event.Name); ok {
						dirs, err := listRecursive(event.Name)
						if err != nil {
							log.Println(err)
							break
						}

						for _, dir := range dirs {
							watcher.Add(dir)
						}

					}
					log.Printf("%+v", watcher.WatchList())
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Println("error:", err)
			case <-ctx.Done():
				watcher.Close()
				return
			}
		}
	}()

	return nil
}

func listRecursive(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	dirs := make([]string, 0, len(entries))
	dirs = append(dirs, dir)
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		path := filepath.Join(dir, entry.Name())
		subdirs, err := listRecursive(path)
		if err != nil {
			return nil, err
		}
		dirs = append(dirs, subdirs...)
	}

	return dirs, nil
}
