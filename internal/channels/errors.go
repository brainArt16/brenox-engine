package channels

import "errors"

var (
	ErrChannelNotFound  = errors.New("channel not found")
	ErrAlreadyMember    = errors.New("already a channel member")
	ErrNotMember        = errors.New("not a channel member")
	ErrOwnerCannotLeave = errors.New("channel owner cannot leave; transfer ownership first")
)

// Broadcaster publishes membership change events to connected clients.
type Broadcaster interface {
	BroadcastMemberJoined(channelID, userID int64)
	BroadcastMemberLeft(channelID, userID int64)
}
