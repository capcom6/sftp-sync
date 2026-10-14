package syncer

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/capcom6/sftp-sync/internal/client"
	"github.com/capcom6/sftp-sync/internal/exclude"
	logger "github.com/go-core-fx/cli-logger"
)

const (
	fieldPath = "path"
	fieldRule = "rule"
)

type Syncer struct {
	rootPath string
	client   client.Client
	matcher  *exclude.Matcher

	logger logger.Logger
}

func New(rootPath string, client client.Client, matcher *exclude.Matcher, logger logger.Logger) *Syncer {
	return &Syncer{
		rootPath: rootPath,
		client:   client,
		matcher:  matcher,

		logger: logger.WithContext("syncer", ""),
	}
}

// entry is a single filesystem entry visited during a tree walk.
type entry struct {
	absPath string
	relPath string
	isDir   bool
}

// visitFunc handles a single non-excluded entry of a tree walk. Returning an
// error aborts the walk and the error is propagated to the caller.
type visitFunc func(ctx context.Context, e entry) error

// Sync applies a single filesystem change reported by the watcher to the
// remote: it removes, creates, or uploads the path, recursing into
// directories. Remote-only files are not deleted.
func (s *Syncer) Sync(ctx context.Context, absPath string) error {
	absRoot, err := filepath.Abs(s.rootPath)
	if err != nil {
		return fmt.Errorf("filepath.Abs: %w", err)
	}

	relPath, err := filepath.Rel(absRoot, absPath)
	if err != nil {
		return fmt.Errorf("filepath.Rel: %w", err)
	}

	if matched, rule := s.isExcluded(absPath); matched {
		s.logger.Debug(ctx, "Excluded path skipped", logger.Fields{
			fieldPath: relPath,
			fieldRule: rule,
		})
		return nil
	}

	exists, _, err := fsInfo(absPath)
	if err != nil {
		return fmt.Errorf("fsInfo: %w", err)
	}

	if !exists {
		if rmErr := s.client.Remove(ctx, pathNormalize(relPath)); rmErr != nil {
			return fmt.Errorf("c.Remove: %w", rmErr)
		}

		s.logger.Info(ctx, "Removed", logger.Fields{
			fieldPath: relPath,
		})

		return nil
	}

	return s.walkEntry(ctx, absRoot, relPath, s.syncEntry)
}

// SyncAll mirrors the entire source tree to the remote by walking every
// top-level entry. Remote-only files are not deleted.
func (s *Syncer) SyncAll(ctx context.Context) error {
	return s.walkTree(ctx, s.syncEntry)
}

// PreviewAll walks the source tree the same way SyncAll does but only logs the
// operations that would be performed, without contacting the remote.
func (s *Syncer) PreviewAll(ctx context.Context) {
	preview := func(_ context.Context, e entry) error {
		action := "upload"
		if e.isDir {
			action = "create"
		}

		s.logger.Info(ctx, "Would "+action, logger.Fields{fieldPath: e.relPath})

		return nil
	}

	if err := s.walkTree(ctx, preview); err != nil {
		s.logger.Error(ctx, "Failed to preview initial sync", err)
	}
}

func (s *Syncer) syncEntry(ctx context.Context, e entry) error {
	if e.isDir {
		return s.createDir(ctx, e.relPath)
	}

	return s.syncFile(ctx, e.absPath, e.relPath)
}

// walkTree visits every top-level entry of the source root. The root itself is
// never visited: the remote destination is expected to exist already.
func (s *Syncer) walkTree(ctx context.Context, visit visitFunc) error {
	absRoot, err := filepath.Abs(s.rootPath)
	if err != nil {
		return fmt.Errorf("filepath.Abs: %w", err)
	}

	entries, err := os.ReadDir(absRoot)
	if err != nil {
		return fmt.Errorf("os.ReadDir: %w", err)
	}

	for _, item := range entries {
		if walkErr := s.walkEntry(ctx, absRoot, item.Name(), visit); walkErr != nil {
			return walkErr
		}
	}

	return nil
}

// walkEntry visits the entry at relPath and, when it is a directory,
// recursively every entry below it. Excluded paths are skipped entirely, so
// their subtrees are not traversed. A cancelled context stops the walk.
func (s *Syncer) walkEntry(ctx context.Context, absRoot, relPath string, visit visitFunc) error {
	select {
	case <-ctx.Done():
		return nil
	default:
	}

	absPath := filepath.Join(absRoot, relPath)

	if matched, rule := s.isExcluded(absPath); matched {
		s.logger.Debug(ctx, "Excluded path skipped", logger.Fields{
			fieldPath: relPath,
			fieldRule: rule,
		})
		return nil
	}

	exists, isDir, err := fsInfo(absPath)
	if err != nil {
		return fmt.Errorf("fsInfo: %w", err)
	}
	if !exists {
		return nil
	}

	e := entry{absPath: absPath, relPath: relPath, isDir: isDir}

	if visitErr := visit(ctx, e); visitErr != nil {
		return visitErr
	}

	if !isDir {
		return nil
	}

	entries, err := os.ReadDir(absPath)
	if err != nil {
		return fmt.Errorf("os.ReadDir: %w", err)
	}

	for _, item := range entries {
		if walkErr := s.walkEntry(ctx, absRoot, filepath.Join(relPath, item.Name()), visit); walkErr != nil {
			return walkErr
		}
	}

	return nil
}

func (s *Syncer) createDir(ctx context.Context, relPath string) error {
	if err := s.client.MakeDir(ctx, pathNormalize(relPath)); err != nil {
		return fmt.Errorf("c.MakeDir: %w", err)
	}
	s.logger.Info(ctx, "Created", logger.Fields{
		fieldPath: relPath,
	})

	return nil
}

func (s *Syncer) syncFile(ctx context.Context, absPath, relPath string) error {
	if err := s.client.UploadFile(ctx, pathNormalize(relPath), pathNormalize(absPath)); err != nil {
		return fmt.Errorf("c.UploadFile: %w", err)
	}

	fields := logger.Fields{
		fieldPath: relPath,
	}
	if info, err := os.Stat(absPath); err == nil {
		fields["size"] = info.Size()
	}
	s.logger.Info(ctx, "Uploaded", fields)

	return nil
}

func fsInfo(path string) (bool, bool, error) {
	fi, err := os.Stat(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return false, false, nil
		}
		return false, false, fmt.Errorf("os.Stat: %w", err)
	}

	return true, fi.IsDir(), nil
}

func pathNormalize(path string) string {
	return filepath.ToSlash(path)
}

func (s *Syncer) isExcluded(absPath string) (bool, string) {
	if s.matcher == nil {
		return false, ""
	}

	return s.matcher.MatchRule(absPath)
}
