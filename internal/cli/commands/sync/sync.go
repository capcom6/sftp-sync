package sync

import (
	"context"
	"sync"

	"github.com/capcom6/sftp-sync/internal/cli/codes"
	"github.com/capcom6/sftp-sync/internal/client"
	"github.com/capcom6/sftp-sync/internal/syncer"
	"github.com/capcom6/sftp-sync/internal/watcher"
	logger "github.com/go-core-fx/cli-logger"
	"github.com/urfave/cli/v3"
)

func Command() *cli.Command {
	return &cli.Command{
		Name:  "sync",
		Usage: "watch a local folder for changes and sync them to a remote FTP server.",
		Arguments: []cli.Argument{
			&cli.StringArg{
				Name:      "source",
				UsageText: "local directory to watch for changes",
				Config: cli.StringConfig{
					TrimSpace: true,
				},
			},
		},
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "dest",
				Usage:    "destination FTP server URL",
				Required: true,
			},
			&cli.StringSliceFlag{
				Name:  "exclude",
				Usage: "paths or patterns to exclude from the synchronization process",
			},
		},
		ArgsUsage: "[source]",
		Before:    Before,
		Action:    Action,
	}
}

func Before(ctx context.Context, cmd *cli.Command) (context.Context, error) {
	if cmd.Args().Len() != 1 {
		return ctx, cli.Exit("exactly one argument is required", codes.ParamsError)
	}

	return ctx, nil
}

func Action(ctx context.Context, cmd *cli.Command) error {
	log := logger.GetLogger(ctx)
	if log == nil {
		return cli.Exit("failed to retrieve logger", codes.InternalError)
	}

	operationID := logger.GenerateOperationID("sync")
	log = log.WithContext("sync-cmd", operationID)

	log.Info(ctx, "Sync command initiated")

	cfg, err := parseConfig(cmd)
	if err != nil {
		log.Error(ctx, "Failed to parse config", err)
		return cli.Exit(err.Error(), codes.ParamsError)
	}

	remote, err := client.New(cfg.Dest, log)
	if err != nil {
		log.Error(ctx, "Failed to create remote client", err)
		return cli.Exit(err.Error(), codes.ClientError)
	}

	watcher := watcher.New(cfg.Source, cfg.Excludes, log)
	syncer := syncer.New(cfg.Source, remote, log)

	var wg sync.WaitGroup

	ch, err := watcher.Watch(ctx, &wg)
	if err != nil {
		log.Error(ctx, "Failed to start watcher", err)
		return cli.Exit(err.Error(), codes.InternalError)
	}

	wg.Add(1)
	go func() {
		defer wg.Done()

		for {
			select {
			case event, ok := <-ch:
				if !ok {
					log.Warn(ctx, "watcher channel closed")
					return
				}
				log.Debug(ctx, "Event received", logger.Fields{"event": event})
				if syncErr := syncer.Sync(ctx, event.AbsPath); syncErr != nil {
					log.Error(ctx, "Failed to sync", syncErr)
				}
			case <-ctx.Done():
				return
			}
		}
	}()

	log.Info(ctx, "Sync command started")

	wg.Wait()

	log.Info(ctx, "Sync command completed")
	return nil
}
