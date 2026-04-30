package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"syscall"

	"github.com/capcom6/sftp-sync/internal/cli/codes"
	"github.com/capcom6/sftp-sync/internal/cli/commands/sync"
	logger "github.com/go-core-fx/cli-logger"
	"github.com/joho/godotenv"
	"github.com/samber/lo"
	"github.com/urfave/cli/v3"
)

//nolint:gochecknoglobals // build metadata
var (
	appVersion   = "dev"
	appBuildDate = "unknown"
	appGitCommit = "unknown"
	appGoVersion = runtime.Version()
)

func main() {
	log := logger.NewDefault()

	rootCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)

	ctx := logger.WithLogger(rootCtx, log)

	// Load environment variables
	if err := godotenv.Load(); err != nil && !os.IsNotExist(err) {
		log.Error(ctx, "Failed to load .env file", err)
	}

	//nolint:reassign // urfave/cli specific
	cli.VersionPrinter = func(cmd *cli.Command) {
		fmt.Fprintf(cmd.Root().Writer, "Version:    %s\n", appVersion)
		fmt.Fprintf(cmd.Root().Writer, "Build Date: %s\n", appBuildDate)
		fmt.Fprintf(cmd.Root().Writer, "Git Commit: %s\n", appGitCommit)
		fmt.Fprintf(cmd.Root().Writer, "Go Version: %s\n", appGoVersion)
	}

	//nolint:reassign // urfave/cli specific
	cli.VersionFlag = &cli.BoolFlag{
		Name:        "version",
		Usage:       "print the version",
		HideDefault: true,
		Local:       true,
	}

	app := &cli.Command{
		Name:      "sftp-sync",
		Usage:     "a command-line utility for syncing a local folder with a remote FTP server on every change of files or directories.",
		Version:   appVersion,
		ArgsUsage: "[source]",
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
			&cli.BoolFlag{
				Name:    "debug",
				Usage:   "enable debug mode",
				Sources: cli.EnvVars("DEBUG"),
			},

			&cli.StringFlag{
				Name:     "dest",
				Usage:    "destination FTP server URL",
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
		},
		Before: func(ctx context.Context, cmd *cli.Command) (context.Context, error) {
			if cmd.Bool("debug") {
				log.SetLevel(logger.LogLevelDebug)
			}

			return sync.Before(ctx, cmd)
		},
		Action: sync.Action,
		Authors: []any{
			"Aleksandr Soloshenko <i@capcom.me>",
		},
		Copyright: "License: Apache-2.0",
	}

	exitCode := codes.Success
	if err := app.Run(ctx, os.Args); err != nil {
		log.Error(ctx, "Application failed", err)
		exitCode = codes.InternalError
		if exitErr, ok := lo.ErrorsAs[cli.ExitCoder](err); ok {
			exitCode = exitErr.ExitCode()
		}
	}

	if err := log.Close(); err != nil {
		fmt.Fprintf(os.Stderr, "failed to close logger: %v\n", err)
	}

	stop()
	os.Exit(exitCode)
}
