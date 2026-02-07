package client

import "errors"

var (
	ErrUnsupportedScheme = errors.New("unsupported scheme")
	ErrClientIsNil       = errors.New("client is nil")
)
