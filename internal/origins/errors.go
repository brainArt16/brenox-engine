package origins

import "errors"

var (
	ErrOriginRequired  = errors.New("origin is required")
	ErrInvalidOrigin   = errors.New("origin must be HTTPS with a hostname, except localhost/loopback origins for local development")
	ErrTooManyOrigins  = errors.New("too many allowed origins")
)
