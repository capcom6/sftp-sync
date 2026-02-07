package syncer

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"

	"github.com/capcom6/logutils"
	"github.com/capcom6/sftp-sync/internal/client"
)

type Syncer struct {
	RootPath string
	Client   client.Client
}

func New(rootPath string, client client.Client) *Syncer {
	return &Syncer{
		RootPath: rootPath,
		Client:   client,
	}
}

func (s *Syncer) Sync(ctx context.Context, absPath string) error {
	exists, isDir, err := fsInfo(absPath)
	if err != nil {
		return fmt.Errorf("fsInfo: %w", err)
	}

	absRoot, err := filepath.Abs(s.RootPath)
	if err != nil {
		return fmt.Errorf("filepath.Abs: %w", err)
	}

	relPath, err := filepath.Rel(absRoot, absPath)
	if err != nil {
		return fmt.Errorf("filepath.Rel: %w", err)
	}

	if !exists {
		if rmErr := s.Client.Remove(ctx, pathNormalize(relPath)); rmErr != nil {
			return fmt.Errorf("c.Remove: %w", rmErr)
		}

		logutils.Printf("--- %s\n", relPath)

		return nil
	}

	if isDir {
		return s.syncDir(ctx, absPath, relPath)
	}

	return s.syncFile(ctx, absPath, relPath)
}

func (s *Syncer) syncFile(ctx context.Context, absPath, relPath string) error {
	if err := s.Client.UploadFile(ctx, pathNormalize(relPath), pathNormalize(absPath)); err != nil {
		return fmt.Errorf("c.UploadFile: %w", err)
	}

	logutils.Printf("--> %s\n", relPath)

	return nil
}

func (s *Syncer) syncDir(ctx context.Context, absPath, relPath string) error {
	if err := s.Client.MakeDir(ctx, pathNormalize(relPath)); err != nil {
		return fmt.Errorf("c.MakeDir: %w", err)
	}
	logutils.Printf("+++ %s\n", relPath)

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
