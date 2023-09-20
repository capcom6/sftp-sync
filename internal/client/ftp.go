package client

import (
	"context"
	"fmt"
	"net/textproto"
	"net/url"
	"os"
	"path"
	"sync"

	"github.com/capcom6/logutils"
	"github.com/jlaffaye/ftp"
)

type FtpClient struct {
	URL string

	client *ftp.ServerConn
	lock   sync.Mutex
}

func NewFtpClient(url string) *FtpClient {
	return &FtpClient{
		URL: url,
	}
}

func (c *FtpClient) init(ctx context.Context) error {
	c.lock.Lock()
	defer c.lock.Unlock()

	if c.client != nil {
		if err := c.ping(ctx); err == nil {
			return nil
		} else {
			logutils.Debugln("Reconnecting because of error:", err)
		}

		c.client.Quit()
		c.client = nil
	}

	u, err := url.Parse(c.URL)
	if err != nil {
		return fmt.Errorf("can't parse URL: %w", err)
	}

	if u.Scheme != "ftp" {
		return fmt.Errorf("unsupported scheme: %s", u.Scheme)
	}

	c.client, err = ftp.Dial(u.Host, ftp.DialWithContext(ctx))
	if err != nil {
		return fmt.Errorf("can't connect to %s: %w", u.Host, err)
	}

	password, ok := u.User.Password()
	if !ok {
		password = ""
	}

	if err := c.client.Login(u.User.Username(), password); err != nil {
		return fmt.Errorf("can't login as %s: %w", u.User.Username(), err)
	}

	if err := c.client.ChangeDir(u.Path); err != nil {
		return fmt.Errorf("can't change directory to %s: %w", u.Path, err)
	}

	return nil
}

func (c *FtpClient) ping(ctx context.Context) error {
	if c.client == nil {
		return fmt.Errorf("client is nil")
	}

	return c.client.NoOp()
}

func (c *FtpClient) MakeDir(ctx context.Context, remotePath string) error {
	if err := c.init(ctx); err != nil {
		return err
	}

	if remotePath == "" {
		// root path
		return nil
	}

	dirs := splitPath(remotePath)
	dirs = append(dirs, remotePath)

	for _, dir := range dirs {
		if err := c.client.MakeDir(dir); err != nil && !isIgnorableError(err) {
			return fmt.Errorf("can't make directory %s: %w", dir, err)
		}
	}

	return nil
}

func (c *FtpClient) RemoveDir(ctx context.Context, remotePath string) error {
	if err := c.init(ctx); err != nil {
		return err
	}

	err := c.client.RemoveDirRecur(remotePath)
	if err != nil {
		err, ok := err.(*textproto.Error)
		if ok && err.Code == 550 {
			return nil
		}
	}

	return err
}

func (c *FtpClient) UploadFile(ctx context.Context, remotePath string, localPath string) error {
	if err := c.init(ctx); err != nil {
		return err
	}

	dir, _ := path.Split(remotePath)
	if err := c.MakeDir(ctx, dir); err != nil {
		return err
	}

	h, err := os.Open(localPath)
	if err != nil {
		return fmt.Errorf("can't open local file %s: %w", localPath, err)
	}
	defer h.Close()

	if err := c.client.Stor(remotePath, h); err != nil {
		return fmt.Errorf("can't upload file to %s: %w", remotePath, err)
	}

	return nil
}

func (c *FtpClient) RemoveFile(ctx context.Context, remotePath string) error {
	if err := c.init(ctx); err != nil {
		return err
	}

	err := c.client.Delete(remotePath)
	if err != nil && !isIgnorableError(err) {
		return err
	}

	return nil
}

func (c *FtpClient) Remove(ctx context.Context, remotePath string) error {
	if err := c.init(ctx); err != nil {
		return err
	}

	dir, name := path.Split(remotePath)
	entries, err := c.client.List(dir)
	if err != nil && !isIgnorableError(err) {
		return fmt.Errorf("can't list directory %s: %w", dir, err)
	}

	for _, entry := range entries {
		if entry.Name != name {
			continue
		}

		if entry.Type == ftp.EntryTypeFolder {
			return c.RemoveDir(ctx, remotePath)
		}
		if entry.Type == ftp.EntryTypeFile {
			return c.RemoveFile(ctx, remotePath)
		}
	}

	return nil
}

func isIgnorableError(err error) bool {
	if err, ok := err.(*textproto.Error); ok && err.Code == 550 {
		logutils.Debugf("ignore error %s", err)
		return true
	}
	return false
}

func splitPath(dir string) []string {
	entries := make([]string, 0, 4)

	dir = path.Clean(dir)

	for {
		dir = path.Dir(dir)
		if dir == "." || dir == "/" {
			break
		}
		entries = append(entries, dir)
	}

	for i := 0; i < len(entries)/2; i++ {
		entries[i], entries[len(entries)-i-1] = entries[len(entries)-i-1], entries[i]
	}

	return entries
}
