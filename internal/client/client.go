package client

import (
	"context"
	"fmt"
	"net/url"

	"github.com/capcom6/sftp-sync/internal/client/ftp"
	"github.com/capcom6/sftp-sync/internal/client/sftp"
	"github.com/capcom6/sftp-sync/internal/client/types"
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

	switch u.Scheme {
	case "ftp":
		return ftp.NewClient(address, log.WithContext("ftp", "")), nil
	case "sftp":
		return sftp.NewClient(address, log.WithContext("sftp", "")), nil
	}

	return nil, fmt.Errorf("%w: %s (supported: ftp, sftp)", types.ErrUnsupportedScheme, u.Scheme)
}
