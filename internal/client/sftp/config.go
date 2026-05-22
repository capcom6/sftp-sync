package sftp

import (
	"fmt"
	"net/url"
	"time"

	"github.com/capcom6/sftp-sync/internal/client/types"
	"github.com/samber/lo"
)

const defaultConnectTimeout = 30 * time.Second

type Config struct {
	Host           string
	Port           string
	Username       string
	Auth           AuthConfig
	ConnectTimeout time.Duration
}

func ParseConfigFromURL(u *url.URL) (Config, error) {
	hostname := u.Hostname()
	if hostname == "" {
		return Config{}, fmt.Errorf("%w: sftp URL must include host", types.ErrInvalidParams)
	}

	user := u.User
	if user == nil || user.Username() == "" {
		return Config{}, fmt.Errorf("%w: sftp URL must include username", types.ErrInvalidParams)
	}

	timeout := defaultConnectTimeout
	if timeoutStr := u.Query().Get("timeout"); timeoutStr != "" {
		var err error
		timeout, err = time.ParseDuration(timeoutStr)
		if err != nil {
			return Config{}, fmt.Errorf("failed to parse timeout: %w", err)
		}
	}

	return Config{
		Host:           hostname,
		Port:           lo.CoalesceOrEmpty(u.Port(), "22"),
		Username:       user.Username(),
		Auth:           ParseAuthFromURL(u),
		ConnectTimeout: timeout,
	}, nil
}
