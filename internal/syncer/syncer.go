package syncer

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"path"
	"path/filepath"
)

type RemoteClient interface {
	MakeDir(ctx context.Context, remotePath string) error
	RemoveDir(ctx context.Context, remotePath string) error

	UploadFile(ctx context.Context, remotePath string, localPath string) error
	RemoveFile(ctx context.Context, remotePath string) error

	Remove(ctx context.Context, remotePath string) error
}

type Syncer struct {
	RootPath string
	Client   RemoteClient
}

func New(rootPath string, client RemoteClient) *Syncer {
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
		if err := s.Client.Remove(ctx, relPath); err != nil {
			return fmt.Errorf("c.Remove: %w", err)
		}

		log.Printf("--- %s\n", relPath)

		return nil
	}

	if isDir {
		return s.syncDir(ctx, absPath, relPath)
	}

	return s.syncFile(ctx, absPath, relPath)
}

func (s *Syncer) syncFile(ctx context.Context, absPath, relPath string) error {
	if err := s.Client.UploadFile(ctx, relPath, absPath); err != nil {
		return fmt.Errorf("c.UploadFile: %w", err)
	}

	log.Printf("--> %s\n", relPath)

	return nil
}

func (s *Syncer) syncDir(ctx context.Context, absPath, relPath string) error {
	if err := s.Client.MakeDir(ctx, relPath); err != nil {
		return fmt.Errorf("c.MakeDir: %w", err)
	}
	log.Printf("+++ %s\n", relPath)

	files, err := os.ReadDir(absPath)
	if err != nil {
		return fmt.Errorf("os.ReadDir: %w", err)
	}

	for _, file := range files {
		if file.IsDir() {
			if err := s.syncDir(ctx, path.Join(absPath, file.Name()), path.Join(relPath, file.Name())); err != nil {
				return err
			}
		} else {
			if err := s.syncFile(ctx, path.Join(absPath, file.Name()), path.Join(relPath, file.Name())); err != nil {
				return err
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
