package types

import "errors"

var (
	ErrUnsupportedScheme = errors.New("unsupported scheme")
	ErrClientIsNil       = errors.New("client is nil")
	ErrInvalidParams     = errors.New("invalid params")
	ErrInvalidPath       = errors.New("invalid path")
)
