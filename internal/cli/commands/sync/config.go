package sync

import "github.com/urfave/cli/v3"

type config struct {
	Source   string
	Dest     string
	Excludes []string
	DryRun   bool
}

func (c config) validate() error {
	if c.Source == "" {
		return cli.Exit("source directory is required", 1)
	}

	if c.Dest == "" {
		return cli.Exit("destination server is required", 1)
	}

	return nil
}

func parseConfig(cmd *cli.Command) (config, error) {
	cfg := config{
		Source:   "",
		Dest:     "",
		Excludes: nil,
		DryRun:   false,
	}

	cfg.Source = cmd.StringArg("source")
	cfg.Dest = cmd.String("dest")
	cfg.Excludes = cmd.StringSlice("exclude")
	cfg.DryRun = cmd.Bool("dry-run")

	return cfg, cfg.validate()
}
