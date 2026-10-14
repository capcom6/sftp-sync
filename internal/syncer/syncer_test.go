package syncer_test

import (
	"bytes"
	"context"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"testing"

	"github.com/capcom6/sftp-sync/internal/client"
	"github.com/capcom6/sftp-sync/internal/exclude"
	"github.com/capcom6/sftp-sync/internal/syncer"
	logger "github.com/go-core-fx/cli-logger"
)

// fakeClient records the operations requested against the remote.
type fakeClient struct {
	mu      sync.Mutex
	dirs    []string
	uploads []string
	removes []string
}

func (c *fakeClient) MakeDir(_ context.Context, remotePath string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.dirs = append(c.dirs, remotePath)

	return nil
}

func (c *fakeClient) RemoveDir(_ context.Context, remotePath string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.removes = append(c.removes, remotePath)

	return nil
}

func (c *fakeClient) UploadFile(_ context.Context, remotePath string, _ string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.uploads = append(c.uploads, remotePath)

	return nil
}

func (c *fakeClient) RemoveFile(_ context.Context, remotePath string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.removes = append(c.removes, remotePath)

	return nil
}

func (c *fakeClient) Remove(_ context.Context, remotePath string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.removes = append(c.removes, remotePath)

	return nil
}

func testLogger(t *testing.T) logger.Logger {
	t.Helper()

	return newLogger(t, io.Discard)
}

func newLogger(t *testing.T, output io.Writer) logger.Logger {
	t.Helper()

	log, err := logger.New(logger.Config{
		Level:        logger.LogLevelDebug,
		Format:       logger.FormatHuman,
		Output:       output,
		EnableColors: false,
		TimeFormat:   "2006-01-02",
	})
	if err != nil {
		t.Fatalf("logger.New() error = %v", err)
	}

	return log
}

// writeTree creates files under root using slash-separated relative paths.
func writeTree(t *testing.T, root string, files ...string) {
	t.Helper()

	for _, file := range files {
		full := filepath.Join(root, filepath.FromSlash(file))
		if err := os.MkdirAll(filepath.Dir(full), 0o750); err != nil {
			t.Fatalf("os.MkdirAll() error = %v", err)
		}
		if err := os.WriteFile(full, []byte("data"), 0o600); err != nil {
			t.Fatalf("os.WriteFile() error = %v", err)
		}
	}
}

func newTestSyncer(t *testing.T, root string, excludes []string) (*syncer.Syncer, *fakeClient) {
	t.Helper()

	matcher, err := exclude.New(excludes, root)
	if err != nil {
		t.Fatalf("exclude.New() error = %v", err)
	}

	fake := &fakeClient{}

	return syncer.New(root, fake, matcher, testLogger(t)), fake
}

func TestSyncAllUploadsWholeTree(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	writeTree(t, root,
		"a.txt",
		"docs/readme.md",
		"docs/img/logo.png",
		"empty-dir/nested/deep/file.txt",
	)

	s, fake := newTestSyncer(t, root, nil)

	if err := s.SyncAll(context.Background()); err != nil {
		t.Fatalf("SyncAll() error = %v", err)
	}

	wantUploads := []string{"a.txt", "docs/img/logo.png", "docs/readme.md", "empty-dir/nested/deep/file.txt"}
	slices.Sort(fake.uploads)
	if !slices.Equal(fake.uploads, wantUploads) {
		t.Errorf("uploads = %v, want %v", fake.uploads, wantUploads)
	}

	if len(fake.removes) != 0 {
		t.Errorf("SyncAll must not remove anything, got %v", fake.removes)
	}
}

func TestSyncAllCreatesDirectories(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	writeTree(t, root, "docs/img/logo.png")

	s, fake := newTestSyncer(t, root, nil)

	if err := s.SyncAll(context.Background()); err != nil {
		t.Fatalf("SyncAll() error = %v", err)
	}

	wantDirs := []string{"docs", "docs/img"}
	slices.Sort(fake.dirs)
	if !slices.Equal(fake.dirs, wantDirs) {
		t.Errorf("dirs = %v, want %v", fake.dirs, wantDirs)
	}
}

func TestSyncAllSkipsExcludedPaths(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	writeTree(t, root,
		"keep.txt",
		"debug.log",
		"node_modules/pkg/index.js",
		"src/main.go",
	)

	s, fake := newTestSyncer(t, root, []string{"*.log", "node_modules/**"})

	if err := s.SyncAll(context.Background()); err != nil {
		t.Fatalf("SyncAll() error = %v", err)
	}

	wantUploads := []string{"keep.txt", "src/main.go"}
	slices.Sort(fake.uploads)
	if !slices.Equal(fake.uploads, wantUploads) {
		t.Errorf("uploads = %v, want %v", fake.uploads, wantUploads)
	}

	for _, dir := range fake.dirs {
		if dir == "node_modules" {
			t.Errorf("excluded directory must not be created, got dirs %v", fake.dirs)
		}
	}
}

func TestSyncAllEmptyRoot(t *testing.T) {
	t.Parallel()

	root := t.TempDir()

	s, fake := newTestSyncer(t, root, nil)

	if err := s.SyncAll(context.Background()); err != nil {
		t.Fatalf("SyncAll() error = %v", err)
	}

	if len(fake.uploads) != 0 || len(fake.dirs) != 0 {
		t.Errorf("empty source must produce no operations, got uploads %v dirs %v", fake.uploads, fake.dirs)
	}
}

func TestSyncAllStopsOnCancelledContext(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	writeTree(t, root, "a.txt", "b.txt", "c.txt")

	s, fake := newTestSyncer(t, root, nil)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if err := s.SyncAll(ctx); err != nil {
		t.Fatalf("SyncAll() error = %v", err)
	}

	if len(fake.uploads) != 0 {
		t.Errorf("cancelled context must stop the walk, got uploads %v", fake.uploads)
	}
}

// compile-time check that fakeClient satisfies the client interface.
var _ client.Client = (*fakeClient)(nil)

func TestSyncRecursesIntoDirectory(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	writeTree(t, root,
		"docs/readme.md",
		"docs/img/logo.png",
	)

	s, fake := newTestSyncer(t, root, nil)

	if err := s.Sync(context.Background(), filepath.Join(root, "docs")); err != nil {
		t.Fatalf("Sync() error = %v", err)
	}

	wantUploads := []string{"docs/img/logo.png", "docs/readme.md"}
	slices.Sort(fake.uploads)
	if !slices.Equal(fake.uploads, wantUploads) {
		t.Errorf("uploads = %v, want %v", fake.uploads, wantUploads)
	}

	wantDirs := []string{"docs", "docs/img"}
	slices.Sort(fake.dirs)
	if !slices.Equal(fake.dirs, wantDirs) {
		t.Errorf("dirs = %v, want %v", fake.dirs, wantDirs)
	}
}

func TestSyncRemovesDeletedPath(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	s, fake := newTestSyncer(t, root, nil)

	if err := s.Sync(context.Background(), filepath.Join(root, "gone.txt")); err != nil {
		t.Fatalf("Sync() error = %v", err)
	}

	if !slices.Equal(fake.removes, []string{"gone.txt"}) {
		t.Errorf("removes = %v, want [gone.txt]", fake.removes)
	}
}

func TestPreviewAllLogsWithoutTouchingRemote(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	writeTree(t, root,
		"a.txt",
		"docs/readme.md",
		"debug.log",
	)

	var out bytes.Buffer

	matcher, err := exclude.New([]string{"*.log"}, root)
	if err != nil {
		t.Fatalf("exclude.New() error = %v", err)
	}

	fake := &fakeClient{}
	s := syncer.New(root, fake, matcher, newLogger(t, &out))

	s.PreviewAll(context.Background())

	if len(fake.uploads) != 0 || len(fake.dirs) != 0 || len(fake.removes) != 0 {
		t.Errorf("PreviewAll must not call the client, got uploads %v dirs %v removes %v",
			fake.uploads, fake.dirs, fake.removes)
	}

	logs := out.String()
	for _, want := range []string{
		`Would create path="docs"`,
		`Would upload path="a.txt"`,
		`Would upload path="docs/readme.md"`,
	} {
		if !strings.Contains(logs, want) {
			t.Errorf("logs missing %q, got:\n%s", want, logs)
		}
	}

	if strings.Contains(logs, `Would upload path="debug.log"`) {
		t.Errorf("excluded file must not be previewed, got:\n%s", logs)
	}
}
