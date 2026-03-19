package client

import (
	"context"
	"fmt"
	"net/url"

	logger "github.com/go-core-fx/cli-logger"
)

type Client interface {
	MakeDir(ctx context.Context, remotePath string) error
	RemoveDir(ctx context.Context, remotePath string) error

	UploadFile(ctx context.Context, remotePath string, localPath string) error
	RemoveFile(ctx context.Context, remotePath string) error

	Remove(ctx context.Context, remotePath string) error
}

func New(address string, log logger.Logger) (Client, error) {
	u, err := url.Parse(address)
	if err != nil {
		return nil, fmt.Errorf("failed to parse URL: %w", err)
	}

	if u.Scheme == "ftp" {
		return NewFtpClient(address, log.WithContext("client", "")), nil
	}

	return nil, fmt.Errorf("%w: %s", ErrUnsupportedScheme, u.Scheme)
}
