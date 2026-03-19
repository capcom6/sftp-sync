package syncer

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"

	"github.com/capcom6/sftp-sync/internal/client"
	logger "github.com/go-core-fx/cli-logger"
)

type Syncer struct {
	rootPath string
	client   client.Client

	logger logger.Logger
}

func New(rootPath string, client client.Client, logger logger.Logger) *Syncer {
	return &Syncer{
		rootPath: rootPath,
		client:   client,

		logger: logger.WithContext("syncer", ""),
	}
}

func (s *Syncer) Sync(ctx context.Context, absPath string) error {
	exists, isDir, err := fsInfo(absPath)
	if err != nil {
		return fmt.Errorf("fsInfo: %w", err)
	}

	absRoot, err := filepath.Abs(s.rootPath)
	if err != nil {
		return fmt.Errorf("filepath.Abs: %w", err)
	}

	relPath, err := filepath.Rel(absRoot, absPath)
	if err != nil {
		return fmt.Errorf("filepath.Rel: %w", err)
	}

	if !exists {
		if rmErr := s.client.Remove(ctx, pathNormalize(relPath)); rmErr != nil {
			return fmt.Errorf("c.Remove: %w", rmErr)
		}

		s.logger.Info(ctx, "Removed", logger.Fields{
			"path": relPath,
		})

		return nil
	}

	if isDir {
		return s.syncDir(ctx, absPath, relPath)
	}

	return s.syncFile(ctx, absPath, relPath)
}

func (s *Syncer) syncFile(ctx context.Context, absPath, relPath string) error {
	if err := s.client.UploadFile(ctx, pathNormalize(relPath), pathNormalize(absPath)); err != nil {
		return fmt.Errorf("c.UploadFile: %w", err)
	}

	s.logger.Info(ctx, "Uploaded", logger.Fields{
		"path": relPath,
	})

	return nil
}

func (s *Syncer) syncDir(ctx context.Context, absPath, relPath string) error {
	if err := s.client.MakeDir(ctx, pathNormalize(relPath)); err != nil {
		return fmt.Errorf("c.MakeDir: %w", err)
	}
	s.logger.Info(ctx, "Created", logger.Fields{
		"path": relPath,
	})

	files, err := os.ReadDir(absPath)
	if err != nil {
		return fmt.Errorf("os.ReadDir: %w", err)
	}

	for _, file := range files {
		select {
		case <-ctx.Done():
			return nil
		default:
		}

		if file.IsDir() {
			if dErr := s.syncDir(ctx, path.Join(absPath, file.Name()), path.Join(relPath, file.Name())); dErr != nil {
				return dErr
			}
		} else {
			if fErr := s.syncFile(ctx, path.Join(absPath, file.Name()), path.Join(relPath, file.Name())); fErr != nil {
				return fErr
			}
		}
	}

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
