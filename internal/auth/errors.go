package auth

import "errors"

var (
	ErrEmailExists        = errors.New("email already exists")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrInvalidToken       = errors.New("invalid or expired token")
	ErrRegistrationFailed = errors.New("registration failed")
	ErrAccountSuspended   = errors.New("account suspended")
)
