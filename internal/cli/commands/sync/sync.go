package sync

import (
	"context"
	"sync"

	"github.com/capcom6/sftp-sync/internal/cli/codes"
	"github.com/capcom6/sftp-sync/internal/client"
	"github.com/capcom6/sftp-sync/internal/exclude"
	"github.com/capcom6/sftp-sync/internal/syncer"
	"github.com/capcom6/sftp-sync/internal/watcher"
	logger "github.com/go-core-fx/cli-logger"
	"github.com/urfave/cli/v3"
)

func Command() *cli.Command {
	return &cli.Command{
		Name:  "sync",
		Usage: "watch a local folder for changes and sync them to a remote FTP or SFTP server.",
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
				Usage:    "destination FTP/SFTP server URL (e.g., ftp://user:pass@host/path or sftp://user:pass@host:22/path)",
				Required: true,
			},
			&cli.StringSliceFlag{
				Name:  "exclude",
				Usage: "paths or glob patterns to exclude (supports *, **, ?)",
			},
			&cli.BoolFlag{
				Name:  "dry-run",
				Usage: "perform a dry run without actually syncing files",
			},
			&cli.BoolFlag{
				Name:  "initial-sync",
				Usage: "sync the whole local folder to the remote before starting to watch for changes",
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

	excludeMatcher, err := exclude.New(cfg.Excludes, cfg.Source)
	if err != nil {
		log.Error(ctx, "Failed to build exclude matcher", err)
		return cli.Exit(err.Error(), codes.ParamsError)
	}

	watcher := watcher.New(cfg.Source, excludeMatcher, log)
	syncer := syncer.New(cfg.Source, remote, excludeMatcher, log)

	if cfg.InitialSync {
		runInitialSync(ctx, log, syncer, cfg.DryRun)
	}

	var wg sync.WaitGroup

	ch, err := watcher.Watch(ctx, &wg)
	if err != nil {
		log.Error(ctx, "Failed to start watcher", err)
		return cli.Exit(err.Error(), codes.InternalError)
	}

	wg.Go(func() {
		for {
			select {
			case event, ok := <-ch:
				if !ok {
					log.Warn(ctx, "watcher channel closed")
					return
				}
				log.Debug(ctx, "Event received", logger.Fields{"event": event})
				if cfg.DryRun {
					dryRunLog(ctx, event)
				} else if syncErr := syncer.Sync(ctx, event.AbsPath); syncErr != nil {
					log.Error(ctx, "Failed to sync", syncErr)
				}
			case <-ctx.Done():
				return
			}
		}
	})

	log.Info(ctx, "Sync command started")

	wg.Wait()

	log.Info(ctx, "Sync command completed")
	return nil
}

// runInitialSync mirrors the whole source tree to the remote once, before the
// watcher starts. Failures are logged and do not stop the process, consistent
// with runtime sync failures, so watching still covers future changes.
func runInitialSync(ctx context.Context, log logger.Logger, s *syncer.Syncer, dryRun bool) {
	log.Info(ctx, "Initial sync started")

	if dryRun {
		s.PreviewAll(ctx)
	} else if syncErr := s.SyncAll(ctx); syncErr != nil {
		log.Error(ctx, "Initial sync failed", syncErr)
	}

	log.Info(ctx, "Initial sync completed")
}

func dryRunLog(ctx context.Context, event watcher.Event) {
	log := logger.GetLogger(ctx)

	var action string
	switch event.Type {
	case watcher.EventCreated:
		action = "create"
	case watcher.EventModified:
		action = "modify"
	case watcher.EventRemoved:
		action = "remove"
	default:
		return
	}

	log.Info(ctx, "Would "+action, logger.Fields{"path": event.RelPath})
}
