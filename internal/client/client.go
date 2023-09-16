package client

import (
	"context"
	"fmt"
	"net/url"
)

type Client interface {
	MakeDir(ctx context.Context, remotePath string) error
	RemoveDir(ctx context.Context, remotePath string) error

	UploadFile(ctx context.Context, remotePath string, localPath string) error
	RemoveFile(ctx context.Context, remotePath string) error

	Remove(ctx context.Context, remotePath string) error
}

func New(address string) (Client, error) {
	u, err := url.Parse(address)
	if err != nil {
		return nil, err
	}

	if u.Scheme == "ftp" {
		return NewFtpClient(address), nil
	}

	return nil, fmt.Errorf("unsupported scheme: %s", u.Scheme)
}
