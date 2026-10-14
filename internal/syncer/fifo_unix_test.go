//go:build unix

package syncer_test

import (
	"context"
	"path/filepath"
	"slices"
	"syscall"
	"testing"
	"time"
)

// TestSyncAllSkipsFIFO ensures the initial pass neither uploads a FIFO nor
// blocks on it: opening a FIFO without a writer would hang the walk forever
// and the watcher would never start.
func TestSyncAllSkipsFIFO(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	writeTree(t, root, "a.txt")
	if err := syscall.Mkfifo(filepath.Join(root, "pipe.fifo"), 0o600); err != nil {
		t.Fatalf("syscall.Mkfifo() error = %v", err)
	}

	s, fake := newTestSyncer(t, root, nil)

	done := make(chan error, 1)
	go func() {
		done <- s.SyncAll(context.Background())
	}()

	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("SyncAll() error = %v", err)
		}
	case <-time.After(10 * time.Second):
		t.Fatal("SyncAll() blocked on a FIFO instead of skipping it")
	}

	wantUploads := []string{"a.txt"}
	slices.Sort(fake.uploads)
	if !slices.Equal(fake.uploads, wantUploads) {
		t.Errorf("uploads = %v, want %v", fake.uploads, wantUploads)
	}
}
