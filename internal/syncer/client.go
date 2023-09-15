package syncer

import (
	"fmt"
	"net/url"
)

func NewClient(address string) (RemoteClient, error) {
	u, err := url.Parse(address)
	if err != nil {
		return nil, err
	}

	if u.Scheme == "ftp" {
		return NewFtpClient(address), nil
	}

	return nil, fmt.Errorf("unsupported scheme: %s", u.Scheme)
}
