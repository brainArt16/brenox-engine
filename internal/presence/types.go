package presence

import "errors"

const (
	StatusOnline  = "online"
	StatusAway    = "away"
	StatusOffline = "offline"
)

var (
	ErrInvalidStatus   = errors.New("invalid status")
	ErrNotWorkspaceMember = errors.New("not a workspace member")
)

type UserPresence struct {
	UserID          int64  `json:"user_id"`
	Status          string `json:"status"`
	ConnectionCount int64  `json:"connection_count"`
	LastSeen        string `json:"last_seen"`
}

type ChannelRef struct {
	WorkspaceID int64
	ChannelID   int64
}

type ConnectResult struct {
	BecameOnline  bool
	GlobalCount   int64
}

type DisconnectResult struct {
	BecameOffline bool
	GlobalCount   int64
}

type Broadcaster interface {
	PublishPresenceOnline(workspaceID, channelID, userID int64)
	PublishPresenceOffline(workspaceID, channelID, userID int64)
	PublishPresenceStatus(workspaceID, channelID, userID int64, status, lastSeen string)
}

func ValidStatus(status string) bool {
	switch status {
	case StatusOnline, StatusAway, StatusOffline:
		return true
	default:
		return false
	}
}
