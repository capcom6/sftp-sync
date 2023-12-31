package config

import (
	"flag"
	"fmt"
	"os"
	"path"
)

type Config struct {
	WatchPath    string
	ExcludePaths []string
	Dest         string
	Debug        bool
}

func (c *Config) validate() error {
	if c.Dest == "" {
		return fmt.Errorf("destination server is required")
	}

	if c.WatchPath == "" {
		return fmt.Errorf("source directory is required")
	}

	return nil
}

func Parse(args []string) (Config, error) {
	cfg := Config{}

	dest := ""
	exclude := make(arrayValue, 0, 1)

	flagSet := flag.NewFlagSet("", flag.ExitOnError)
	flagSet.SetOutput(os.Stdout)
	flagSet.StringVar(&dest, "dest", "", "destination server")
	flagSet.Var(&exclude, "exclude", "exclude paths")
	flagSet.BoolVar(&cfg.Debug, "debug", false, "debug mode")

	flagSet.Usage = func() {
		fmt.Println("(S)FTP Syncer")
		printVersion()
		fmt.Printf("Usage: %s [flags]\n", path.Base(os.Args[0]))
		flagSet.PrintDefaults()
	}

	if err := flagSet.Parse(args); err != nil {
		return cfg, err
	}

	cfg.Dest = dest
	cfg.ExcludePaths = exclude
	cfg.WatchPath = flagSet.Arg(0)

	if err := cfg.validate(); err != nil {
		return cfg, err
	}

	return cfg, nil
}
