package syncer

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/url"
	"os"
	"path"
	"time"

	"github.com/jlaffaye/ftp"
)

type Syncer struct {
	RemoteURL string
}

func New(remoteUrl string) *Syncer {
	return &Syncer{
		RemoteURL: remoteUrl,
	}
}

func (s *Syncer) Sync(ctx context.Context, absPath, relPath string) error {
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

	if !exists {
		dir, name := path.Split(relPath)
		entries, err := c.List(dir)
		if err != nil {
			return fmt.Errorf("c.List: %w", err)
		}

		for _, v := range entries {
			if v.Name == name {
				if v.Type == ftp.EntryTypeFolder {
					if err := c.RemoveDirRecur(relPath); err != nil {
						return fmt.Errorf("c.RemoveDirRecur: %w", err)
					}
				} else if v.Type == ftp.EntryTypeFile {
					if err := c.Delete(relPath); err != nil {
						return fmt.Errorf("c.Remove: %w", err)
					}
				}
			}
		}
	} else {
		for _, dir := range dirs(relPath) {
			if err := c.MakeDir(dir); err != nil {
				log.Printf("error: %s", err)
			}
		}

		if isDir {
			if err := c.MakeDir(relPath); err != nil {
				return fmt.Errorf("c.MakeDir: %+w", err)
			}
		} else {
			h, err := os.Open(absPath)
			if err != nil {
				return fmt.Errorf("os.Open: %w", err)
			}
			defer h.Close()

			if err := c.Stor(relPath, h); err != nil {
				return fmt.Errorf("c.Stor: %w", err)
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
