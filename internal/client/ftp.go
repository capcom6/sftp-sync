package client

import (
	"context"
	"fmt"
	"log"
	"net/textproto"
	"net/url"
	"os"
	"path"
	"sync"

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
		return c.ping(ctx)
	}

	u, err := url.Parse(c.URL)
	if err != nil {
		return fmt.Errorf("url.Parse: %w", err)
	}

	if u.Scheme != "ftp" {
		return fmt.Errorf("unsupported scheme: %s", u.Scheme)
	}

	c.client, err = ftp.Dial(u.Host, ftp.DialWithContext(ctx))
	if err != nil {
		return fmt.Errorf("ftp.Dial: %w", err)
	}

	password, ok := u.User.Password()
	if !ok {
		password = ""
	}

	if err := c.client.Login(u.User.Username(), password); err != nil {
		return fmt.Errorf("c.Login: %w", err)
	}

	if err := c.client.ChangeDir(u.Path); err != nil {
		return fmt.Errorf("c.ChangeDir: %w", err)
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

	dirs := splitPath(remotePath)
	// dirs = append(dirs, remotePath)

	for _, dir := range dirs {
		if err := c.client.MakeDir(dir); err != nil && !isIgnorableError(err) {
			return fmt.Errorf("c.MakeDir: %w", err)
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
		return fmt.Errorf("os.Open: %w", err)
	}
	defer h.Close()

	if err := c.client.Stor(remotePath, h); err != nil {
		return fmt.Errorf("c.Stor: %w", err)
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
		return fmt.Errorf("c.List: %w", err)
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
		log.Printf("ignore %s", err)
		return true
	}
	return false
}

func splitPath(dir string) []string {
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
