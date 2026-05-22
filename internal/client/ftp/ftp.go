package ftp

import (
	"context"
	"fmt"
	"net"
	"net/textproto"
	"net/url"
	"os"
	"path"
	"strings"
	"sync"

	"github.com/capcom6/sftp-sync/internal/client/types"
	logger "github.com/go-core-fx/cli-logger"
	"github.com/jlaffaye/ftp"
	"github.com/samber/lo"
)

type Client struct {
	url string

	logger logger.Logger

	client *ftp.ServerConn
	lock   sync.Mutex
}

func NewClient(url string, logger logger.Logger) *Client {
	return &Client{
		url: url,

		logger: logger,

		client: nil,
		lock:   sync.Mutex{},
	}
}

func (c *Client) init(ctx context.Context) error {
	c.lock.Lock()
	defer c.lock.Unlock()

	if c.client != nil {
		var err error
		if err = c.ping(ctx); err == nil {
			return nil
		}

		c.logger.Warn(ctx, "Reconnecting because of error", logger.Fields{
			"error": err,
		})

		_ = c.client.Quit()
		c.client = nil
	}

	u, err := url.Parse(c.url)
	if err != nil {
		return fmt.Errorf("can't parse URL: %w", err)
	}

	if u.Scheme != "ftp" {
		return fmt.Errorf("%w: %s", types.ErrUnsupportedScheme, u.Scheme)
	}

	host := u.Host
	if u.Port() == "" {
		host = net.JoinHostPort(u.Hostname(), "21")
	}

	conn, err := ftp.Dial(host, ftp.DialWithContext(ctx))
	if err != nil {
		return fmt.Errorf("can't connect to %s: %w", u.Host, err)
	}

	user := u.User
	if user == nil || user.Username() == "" {
		_ = conn.Quit()
		return fmt.Errorf("%w: missing FTP username in URL", types.ErrInvalidParams)
	}
	password, _ := user.Password()
	if loginErr := conn.Login(user.Username(), password); loginErr != nil {
		_ = conn.Quit()
		return fmt.Errorf("can't login as %s: %w", user.Username(), loginErr)
	}

	if u.Path != "" && u.Path != "/" {
		if chErr := conn.ChangeDir(u.Path); chErr != nil {
			_ = conn.Quit()
			return fmt.Errorf(
				"%w: remote path %s does not exist or is not accessible: %w",
				types.ErrInvalidPath,
				u.Path,
				chErr,
			)
		}
	}

	c.client = conn

	return nil
}

func (c *Client) ping(_ context.Context) error {
	if c.client == nil {
		return types.ErrClientIsNil
	}

	if err := c.client.NoOp(); err != nil {
		return fmt.Errorf("failed to ping: %w", err)
	}

	return nil
}

func (c *Client) MakeDir(ctx context.Context, remotePath string) error {
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

func (c *Client) RemoveDir(ctx context.Context, remotePath string) error {
	if err := c.init(ctx); err != nil {
		return err
	}

	err := c.client.RemoveDirRecur(remotePath)
	if err != nil {
		if err, ok := lo.ErrorsAs[*textproto.Error](err); ok && err.Code == 550 {
			return nil
		}
		return fmt.Errorf("can't remove directory %s: %w", remotePath, err)
	}

	return nil
}

func (c *Client) UploadFile(ctx context.Context, remotePath string, localPath string) error {
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

	if stErr := c.client.Stor(remotePath, h); stErr != nil {
		return fmt.Errorf("can't upload file to %s: %w", remotePath, stErr)
	}

	return nil
}

func (c *Client) RemoveFile(ctx context.Context, remotePath string) error {
	if err := c.init(ctx); err != nil {
		return err
	}

	err := c.client.Delete(remotePath)
	if err != nil && !isIgnorableError(err) {
		return fmt.Errorf("failed to remove file %s: %w", remotePath, err)
	}

	return nil
}

func (c *Client) Remove(ctx context.Context, remotePath string) error {
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
	if err, ok := lo.ErrorsAs[*textproto.Error](err); ok && err.Code == 550 {
		return true
	}
	return false
}

func splitPath(dir string) []string {
	entries := make([]string, 0, strings.Count(dir, "/"))

	dir = path.Clean(dir)

	for {
		dir = path.Dir(dir)
		if dir == "." || dir == "/" {
			break
		}
		entries = append(entries, dir)
	}

	for i := range len(entries) / 2 {
		entries[i], entries[len(entries)-i-1] = entries[len(entries)-i-1], entries[i]
	}

	return entries
}
