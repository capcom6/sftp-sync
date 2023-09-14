package syncer

import (
	"context"
	"fmt"
	"net/textproto"
	"net/url"
	"os"
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

	return c.client.MakeDir(remotePath)
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
	if err != nil {
		err, ok := err.(*textproto.Error)
		if ok && err.Code == 550 {
			return nil
		}
	}

	return err
}
