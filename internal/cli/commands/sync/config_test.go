package sync_test

import (
	"context"
	"testing"

	command "github.com/capcom6/sftp-sync/internal/cli/commands/sync"
	"github.com/urfave/cli/v3"
)

// runCommand executes the sync subcommand with the given args and returns the
// parsed config. It swaps in a no-op Action so parsing can be asserted without
// contacting a remote.
func runCommand(t *testing.T, args ...string) command.Config {
	t.Helper()

	var got command.Config

	cmd := command.Command()
	cmd.Action = func(_ context.Context, c *cli.Command) error {
		parsed, err := command.ParseConfig(c)
		if err != nil {
			t.Fatalf("ParseConfig() error = %v", err)
		}
		got = parsed

		return nil
	}

	if err := cmd.Run(t.Context(), args); err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	return got
}

func TestParseConfigInitialSyncDefault(t *testing.T) {
	t.Parallel()

	cfg := runCommand(t, "sync", "--dest=ftp://user:pass@host/path", "/local")

	if cfg.InitialSync {
		t.Error("InitialSync must be false by default")
	}
}

func TestParseConfigInitialSyncEnabled(t *testing.T) {
	t.Parallel()

	cfg := runCommand(t, "sync", "--dest=ftp://user:pass@host/path", "--initial-sync", "/local")

	if !cfg.InitialSync {
		t.Error("InitialSync must be true when --initial-sync is set")
	}
}
