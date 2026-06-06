package calls

import "errors"

const (
	StatusRinging = "ringing"
	StatusActive  = "active"
	StatusEnded   = "ended"
)

var (
	ErrNotFound          = errors.New("call not found")
	ErrNotMember         = errors.New("not a channel member")
	ErrNotParticipant    = errors.New("not a call participant")
	ErrCallEnded         = errors.New("call has ended")
	ErrCallAlreadyActive = errors.New("channel already has an active call")
	ErrChannelNotFound   = errors.New("channel not found")
)

type CallResponse struct {
	ID          int64  `json:"id"`
	ChannelID   int64  `json:"channel_id"`
	WorkspaceID int64  `json:"workspace_id"`
	InitiatorID int64  `json:"initiator_id"`
	Status      string `json:"status"`
	CreatedAt   string `json:"created_at"`
	EndedAt     string `json:"ended_at,omitempty"`
}

type ParticipantResponse struct {
	UserID   int64  `json:"user_id"`
	JoinedAt string `json:"joined_at"`
}

type SignalContext struct {
	CallID      int64
	WorkspaceID int64
	ChannelID   int64
	UserID      int64
}
