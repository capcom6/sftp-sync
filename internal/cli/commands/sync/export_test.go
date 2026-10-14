package sync

import "github.com/urfave/cli/v3"

// Config is the exported view of the internal config struct for tests.
type Config = config

// ParseConfig is the exported view of parseConfig for tests.
func ParseConfig(c *cli.Command) (Config, error) {
	return parseConfig(c)
}
