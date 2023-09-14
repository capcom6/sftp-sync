package syncer

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/textproto"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"time"

	"github.com/jlaffaye/ftp"
)

type Syncer struct {
	RootPath  string
	RemoteURL string
}

func New(rootPath, remoteUrl string) *Syncer {
	return &Syncer{
		RootPath:  rootPath,
		RemoteURL: remoteUrl,
	}
}

func (s *Syncer) Sync(ctx context.Context, absPath string) error {
	u, err := url.Parse(s.RemoteURL)
	if err != nil {
		return fmt.Errorf("url.Parse: %w", err)
	}

	if u.Scheme != "ftp" {
		return fmt.Errorf("unsupported scheme: %s", u.Scheme)
	}

	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	c, err := ftp.Dial(u.Host, ftp.DialWithContext(ctx))
	if err != nil {
		return fmt.Errorf("ftp.Dial: %w", err)
	}
	defer c.Quit()

	password, ok := u.User.Password()
	if !ok {
		password = ""
	}
	if err := c.Login(u.User.Username(), password); err != nil {
		return fmt.Errorf("c.Login: %w", err)
	}

	if err := c.ChangeDir(u.Path); err != nil {
		return fmt.Errorf("c.ChangeDir: %w", err)
	}

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
		dir, name := path.Split(relPath)
		entries, err := c.List(dir)
		if err != nil {
			err, ok := err.(*textproto.Error)
			if !ok || err.Code != 550 {
				return fmt.Errorf("c.List: %w", err)
			}
			return nil
		}

		for _, v := range entries {
			if v.Name == name {
				if v.Type == ftp.EntryTypeFolder {
					if err := c.RemoveDirRecur(relPath); err != nil {
						return fmt.Errorf("c.RemoveDirRecur: %w", err)
					}
					log.Printf("Removed directory: %s\n", relPath)
				} else if v.Type == ftp.EntryTypeFile {
					if err := c.Delete(relPath); err != nil {
						return fmt.Errorf("c.Remove: %w", err)
					}
					log.Printf("Removed file: %s\n", relPath)
				}
			}
		}
	} else {
		for _, dir := range dirs(relPath) {
			if err := c.MakeDir(dir); err != nil {
				err, ok := err.(*textproto.Error)
				if !ok || err.Code != 550 {
					return fmt.Errorf("c.MakeDir: %w", err)
				}
			}
		}

		if isDir {
			s.syncDir(c, absPath, relPath)
		} else {
			s.syncFile(c, absPath, relPath)
		}
	}

	return nil
}

func (s *Syncer) syncFile(c *ftp.ServerConn, absPath, relPath string) error {
	h, err := os.Open(absPath)
	if err != nil {
		return fmt.Errorf("os.Open: %w", err)
	}
	defer h.Close()

	if err := c.Stor(relPath, h); err != nil {
		return fmt.Errorf("c.Stor: %w", err)
	}

	log.Printf("Uploaded file: %s\n", relPath)

	return nil
}

func (s *Syncer) syncDir(c *ftp.ServerConn, absPath, relPath string) error {
	if err := c.MakeDir(relPath); err != nil {
		return fmt.Errorf("c.MakeDir: %w", err)
	}
	log.Printf("Created directory: %s\n", relPath)

	files, err := os.ReadDir(absPath)
	if err != nil {
		return fmt.Errorf("os.ReadDir: %w", err)
	}

	for _, file := range files {
		if file.IsDir() {
			if err := s.syncDir(c, path.Join(absPath, file.Name()), path.Join(relPath, file.Name())); err != nil {
				return err
			}
		} else {
			if err := s.syncFile(c, path.Join(absPath, file.Name()), path.Join(relPath, file.Name())); err != nil {
				return err
			}
		}
	}

	return nil
}

func dirs(dir string) []string {
	entries := make([]string, 0, 4)

	for {
		dir = path.Dir(dir)
		if dir == "." {
			break
		}
		entries = append(entries, dir)
	}

	for i := 0; i < len(entries)/2; i++ {
		entries[i], entries[len(entries)-i-1] = entries[len(entries)-i-1], entries[i]
	}

	return entries
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
