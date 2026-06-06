package notifications

import "errors"

const (
	TypeMention         = "mention"
	TypeReply           = "reply"
	TypeChannelInvite   = "channel_invite"
	TypeWorkspaceInvite = "workspace_invite"
	TypeCallInvite      = "call_invite"
)

var (
	ErrNotFound    = errors.New("notification not found")
	ErrInvalidType = errors.New("invalid notification type")
)

func ValidType(value string) bool {
	switch value {
	case TypeMention, TypeReply, TypeChannelInvite, TypeWorkspaceInvite, TypeCallInvite:
		return true
	default:
		return false
	}
}

type NotificationResponse struct {
	ID        int64          `json:"id"`
	Type      string         `json:"type"`
	Title     string         `json:"title"`
	Body      string         `json:"body"`
	Data      map[string]any `json:"data"`
	Read      bool           `json:"read"`
	CreatedAt string         `json:"created_at"`
	ReadAt    string         `json:"read_at,omitempty"`
}

type CreateInput struct {
	UserID int64
	Type   string
	Title  string
	Body   string
	Data   map[string]any
}
