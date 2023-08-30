package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"sync"

	"github.com/capcom6/sftp-sync/internal/config"
	"github.com/capcom6/sftp-sync/internal/watcher"
)

func main() {
	cfg, err := config.Parse(os.Args[1:])
	if err != nil {
		log.Fatalln(err)
	}

	wg := &sync.WaitGroup{}
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	watch := watcher.New(cfg.WatchPath, cfg.ExcludePaths)

	ch, err := watch.Watch(ctx, wg)
	if err != nil {
		log.Fatalln(err)
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case event, ok := <-ch:
				if !ok {
					return
				}
				log.Println("event:", event)
			case <-ctx.Done():
				return
			}
		}
	}()

	log.Println("Watching...")
	wg.Wait()

	log.Println("Bye!")
}
