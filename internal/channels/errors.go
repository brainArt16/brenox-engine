package channels

import "errors"

var (
	ErrChannelNotFound     = errors.New("channel not found")
	ErrAlreadyMember       = errors.New("already a channel member")
	ErrNotMember           = errors.New("not a channel member")
	ErrOwnerCannotLeave    = errors.New("channel owner cannot leave; transfer ownership first")
	ErrNotWorkspaceMember  = errors.New("not a workspace member")
	ErrDuplicateChannelName = errors.New("channel name already exists in workspace")
	ErrForbidden           = errors.New("permission denied")
)

// Broadcaster publishes membership change events to connected clients.
type Broadcaster interface {
	BroadcastMemberJoined(workspaceID, channelID, userID int64)
	BroadcastMemberLeft(workspaceID, channelID, userID int64)
}
