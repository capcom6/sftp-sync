package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"sync"

	"github.com/capcom6/logutils"
	"github.com/capcom6/sftp-sync/internal/client"
	"github.com/capcom6/sftp-sync/internal/config"
	"github.com/capcom6/sftp-sync/internal/syncer"
	"github.com/capcom6/sftp-sync/internal/watcher"
)

func main() {
	cfg, err := config.Parse(os.Args[1:])
	if err != nil {
		log.Fatalln(err)
	}
	setUpLogging(cfg)

	wg := &sync.WaitGroup{}
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	remoteClient, err := client.New(cfg.Dest)
	if err != nil {
		logutils.Fatalln(err)
	}

	watch := watcher.New(cfg.WatchPath, cfg.ExcludePaths)
	syncer := syncer.New(cfg.WatchPath, remoteClient)

	ch, err := watch.Watch(ctx, wg)
	if err != nil {
		logutils.Fatalln(err)
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case event, ok := <-ch:
				if !ok {
					logutils.Errorln("watcher channel closed")
					cancel()
					return
				}
				logutils.Debug("event:", event)
				if syncErr := syncer.Sync(ctx, event.AbsPath); syncErr != nil {
					logutils.Errorln(syncErr)
				}
			case <-ctx.Done():
				return
			}
		}
	}()

	logutils.Println("Watching...")
	wg.Wait()

	logutils.Println("Bye!")
}

func setUpLogging(cfg config.Config) {
	logLevel := "INFO"
	if cfg.Debug {
		logLevel = "DEBUG"
	}

	filter := logutils.LevelFilter{
		Levels:   []logutils.LogLevel{"DEBUG", "INFO", "WARN", "ERROR"},
		MinLevel: logutils.LogLevel(logLevel),
		Writer:   os.Stdout,
	}

	log.SetOutput(&filter)
}
