package calls

import "errors"

const (
	StatusRinging = "ringing"
	StatusActive  = "active"
	StatusEnded   = "ended"

	ModeVoice = "voice"
	ModeVideo = "video"
)

var (
	ErrNotFound          = errors.New("call not found")
	ErrNotMember         = errors.New("not a channel member")
	ErrNotParticipant    = errors.New("not a call participant")
	ErrCallEnded         = errors.New("call has ended")
	ErrCallAlreadyActive = errors.New("channel already has an active call")
	ErrCallFull          = errors.New("call participant limit reached")
	ErrChannelNotFound   = errors.New("channel not found")
	ErrRecordingActive   = errors.New("call already has an active recording")
	ErrRecordingNotFound = errors.New("no active recording for call")
	ErrInvalidMode       = errors.New("invalid call mode")
)

type CallResponse struct {
	ID          int64  `json:"id"`
	ChannelID   int64  `json:"channel_id"`
	WorkspaceID int64  `json:"workspace_id"`
	InitiatorID int64  `json:"initiator_id"`
	Status      string `json:"status"`
	Mode        string `json:"mode"`
	CreatedAt   string `json:"created_at"`
	EndedAt     string `json:"ended_at,omitempty"`
}

type InitiateCallRequest struct {
	Mode string `json:"mode"`
}

type RecordingResponse struct {
	ID        int64  `json:"id"`
	CallID    int64  `json:"call_id"`
	StartedBy int64  `json:"started_by"`
	StartedAt string `json:"started_at"`
	EndedAt   string `json:"ended_at,omitempty"`
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
	Mode        string
}
