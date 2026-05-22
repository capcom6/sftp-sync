package sftp

import (
	"context"
	"fmt"
	"net"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"

	"github.com/capcom6/sftp-sync/internal/client/types"
	logger "github.com/go-core-fx/cli-logger"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/knownhosts"
)

type Client struct {
	url      string
	basePath string

	logger logger.Logger

	conn *ssh.Client
	sftp *sftp.Client
	lock sync.Mutex
}

func NewClient(url string, logger logger.Logger) *Client {
	return &Client{
		url:      url,
		basePath: "",

		logger: logger,
		conn:   nil,
		sftp:   nil,
		lock:   sync.Mutex{},
	}
}

func (c *Client) init(ctx context.Context) error {
	c.lock.Lock()
	defer c.lock.Unlock()

	if c.sftp != nil {
		pingErr := c.ping(ctx)
		if pingErr == nil {
			return nil
		}

		c.logger.Warn(ctx, "Reconnecting because of error", logger.Fields{
			"error": pingErr,
		})

		_ = c.sftp.Close()
		_ = c.conn.Close()
		c.sftp = nil
		c.conn = nil
	}

	u, err := url.Parse(c.url)
	if err != nil {
		return fmt.Errorf("can't parse URL: %w", err)
	}

	if u.Scheme != "sftp" {
		return fmt.Errorf("%w: %s", types.ErrUnsupportedScheme, u.Scheme)
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("can't get home directory: %w", err)
	}

	hostKeyCallback, err := knownhosts.New(filepath.Join(homeDir, ".ssh", "known_hosts"))
	if err != nil {
		return fmt.Errorf("can't load known_hosts: %w", err)
	}

	config, err := ParseConfigFromURL(u)
	if err != nil {
		return err
	}

	authMethods, err := config.Auth.BuildAuthMethods()
	if err != nil {
		return fmt.Errorf("can't build auth methods: %w", err)
	}

	client, conn, err := c.connect(ctx, config, authMethods, hostKeyCallback)
	if err != nil {
		return err
	}

	if u.Path != "" && u.Path != "/" {
		info, statErr := client.Stat(u.Path)
		if statErr != nil {
			_ = client.Close()
			_ = conn.Close()
			return fmt.Errorf("%w: remote path %s does not exist: %w", types.ErrInvalidPath, u.Path, statErr)
		}
		if !info.IsDir() {
			_ = client.Close()
			_ = conn.Close()
			return fmt.Errorf("%w: remote path %s is not a directory", types.ErrInvalidPath, u.Path)
		}
		c.basePath = u.Path
	} else {
		c.basePath = ""
	}

	c.conn = conn
	c.sftp = client

	return nil
}

func (c *Client) connect(
	ctx context.Context,
	config Config,
	authMethods []ssh.AuthMethod,
	hostKeyCallback ssh.HostKeyCallback,
) (*sftp.Client, *ssh.Client, error) {
	sshConfig := &ssh.ClientConfig{
		User:            config.Username,
		Auth:            authMethods,
		HostKeyCallback: hostKeyCallback,
	}

	addr := net.JoinHostPort(config.Host, config.Port)
	dialer := &net.Dialer{Timeout: config.ConnectTimeout} //nolint:exhaustruct // only Timeout is relevant
	netConn, err := dialer.DialContext(ctx, "tcp", addr)
	if err != nil {
		return nil, nil, fmt.Errorf("can't connect to %s: %w", addr, err)
	}

	sshConn, chans, reqs, err := ssh.NewClientConn(netConn, addr, sshConfig)
	if err != nil {
		_ = netConn.Close()
		return nil, nil, fmt.Errorf("can't connect to %s: %w", addr, err)
	}

	conn := ssh.NewClient(sshConn, chans, reqs)

	client, err := sftp.NewClient(conn)
	if err != nil {
		_ = conn.Close()
		return nil, nil, fmt.Errorf("can't create SFTP client: %w", err)
	}

	return client, conn, nil
}

func (c *Client) ping(_ context.Context) error {
	if _, err := c.sftp.Getwd(); err != nil {
		return fmt.Errorf("failed to ping: %w", err)
	}

	return nil
}

func (c *Client) resolvePath(remotePath string) string {
	if c.basePath == "" {
		return remotePath
	}
	if remotePath == "" {
		return c.basePath
	}
	return path.Join(c.basePath, remotePath)
}

func (c *Client) MakeDir(ctx context.Context, remotePath string) error {
	if err := c.init(ctx); err != nil {
		return err
	}

	if remotePath == "" {
		return nil
	}

	fullPath := c.resolvePath(remotePath)
	dirs := splitPath(fullPath)
	dirs = append(dirs, fullPath)

	for _, dir := range dirs {
		c.logger.Debug(ctx, "Creating directory", logger.Fields{
			"path": dir,
		})

		if _, err := c.sftp.Stat(dir); err == nil {
			continue
		}

		if err := c.sftp.Mkdir(dir); err != nil {
			if isIgnorableError(err) {
				continue
			}
			return fmt.Errorf("can't make directory %s: %w", dir, err)
		}
	}

	return nil
}

func (c *Client) RemoveDir(ctx context.Context, remotePath string) error {
	if err := c.init(ctx); err != nil {
		return err
	}

	if err := c.removeDirRecur(ctx, c.resolvePath(remotePath)); err != nil {
		if isIgnorableError(err) {
			return nil
		}
		return fmt.Errorf("can't remove directory %s: %w", remotePath, err)
	}

	return nil
}

func (c *Client) removeDirRecur(ctx context.Context, p string) error {
	entries, err := c.sftp.ReadDirContext(ctx, p)
	if err != nil {
		if isIgnorableError(err) {
			return nil
		}
		return fmt.Errorf("can't read directory %s: %w", p, err)
	}

	for _, entry := range entries {
		entryPath := path.Join(p, entry.Name())

		if entry.IsDir() {
			if rmErr := c.removeDirRecur(ctx, entryPath); rmErr != nil {
				return fmt.Errorf("can't remove directory %s: %w", entryPath, rmErr)
			}

			continue
		}

		if rmErr := c.sftp.Remove(entryPath); rmErr != nil {
			if isIgnorableError(rmErr) {
				continue
			}
			return fmt.Errorf("can't remove file %s: %w", entryPath, rmErr)
		}
	}

	if rmErr := c.sftp.RemoveDirectory(p); rmErr != nil {
		if isIgnorableError(rmErr) {
			return nil
		}
		return fmt.Errorf("can't remove directory %s: %w", p, rmErr)
	}

	return nil
}

func (c *Client) UploadFile(ctx context.Context, remotePath string, localPath string) error {
	if err := c.init(ctx); err != nil {
		return err
	}

	fullPath := c.resolvePath(remotePath)
	dir, _ := path.Split(remotePath)
	if err := c.MakeDir(ctx, dir); err != nil {
		return err
	}

	localFile, err := os.Open(localPath)
	if err != nil {
		return fmt.Errorf("can't open local file %s: %w", localPath, err)
	}
	defer localFile.Close()

	remoteFile, err := c.sftp.Create(fullPath)
	if err != nil {
		return fmt.Errorf("can't create remote file %s: %w", fullPath, err)
	}
	defer remoteFile.Close()

	written, err := remoteFile.ReadFrom(localFile)
	if err != nil {
		return fmt.Errorf("can't upload file to %s: %w", fullPath, err)
	}

	c.logger.Debug(ctx, "File uploaded", logger.Fields{
		"path":  fullPath,
		"bytes": written,
		"local": localPath,
	})

	return nil
}

func (c *Client) RemoveFile(ctx context.Context, remotePath string) error {
	if err := c.init(ctx); err != nil {
		return err
	}

	if err := c.sftp.Remove(c.resolvePath(remotePath)); err != nil {
		if isIgnorableError(err) {
			return nil
		}
		return fmt.Errorf("can't remove file %s: %w", remotePath, err)
	}

	return nil
}

func (c *Client) Remove(ctx context.Context, remotePath string) error {
	if err := c.init(ctx); err != nil {
		return err
	}

	info, err := c.sftp.Stat(c.resolvePath(remotePath))
	if err != nil {
		if isIgnorableError(err) {
			return nil
		}
		return fmt.Errorf("can't stat %s: %w", remotePath, err)
	}

	if info.IsDir() {
		return c.RemoveDir(ctx, remotePath)
	}

	return c.RemoveFile(ctx, remotePath)
}

func isIgnorableError(err error) bool {
	if err == nil {
		return true
	}

	errStr := err.Error()
	notFoundErrors := []string{
		"does not exist",
		"not found",
		"No such file",
	}

	for _, pattern := range notFoundErrors {
		if strings.Contains(errStr, pattern) {
			return true
		}
	}

	return false
}

func splitPath(dir string) []string {
	if dir == "" || dir == "." {
		return nil
	}

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
