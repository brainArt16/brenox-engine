package origins

import "errors"

var (
	ErrOriginRequired  = errors.New("origin is required")
	ErrInvalidOrigin   = errors.New("origin must be a full URL like https://app.example.com")
	ErrTooManyOrigins  = errors.New("too many allowed origins")
)
