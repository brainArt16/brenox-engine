package platformadmin

import "errors"

var (
	ErrForbidden     = errors.New("platform admin access required")
	ErrNotFound      = errors.New("resource not found")
	ErrInvalidRole   = errors.New("invalid platform role")
	ErrSelfDemotion  = errors.New("cannot change your own platform role")
	ErrSelfSuspend   = errors.New("cannot suspend your own account")
	ErrInvalidRequest = errors.New("invalid request")
	ErrKeyNotFound    = errors.New("api key not found")
)

const (
	RoleUser    = "user"
	RoleSupport = "support"
	RoleAdmin   = "admin"
)

func RoleRank(role string) int {
	switch role {
	case RoleAdmin:
		return 3
	case RoleSupport:
		return 2
	case RoleUser:
		return 1
	default:
		return 0
	}
}

func IsValidRole(role string) bool {
	return RoleRank(role) > 0
}

func CanAccess(role string) bool {
	return RoleRank(role) >= RoleRank(RoleSupport)
}

func CanWrite(role string) bool {
	return role == RoleAdmin
}
