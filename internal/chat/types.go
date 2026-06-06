package chat

import "errors"

const (
	// MaxMessageLength is the maximum allowed message body size in characters.
	MaxMessageLength = 4000

	defaultMessageLimit = 50
	maxMessageLimit     = 100
)

var (
	ErrNotMember         = errors.New("not a channel member")
	ErrNotWorkspaceMember = errors.New("not a workspace member")
	ErrChannelNotFound   = errors.New("channel not found")
	ErrForbidden         = errors.New("permission denied")
	ErrEmptyContent      = errors.New("message content is required")
	ErrMessageTooLong    = errors.New("message exceeds maximum length")
)

type CreateMessageRequest struct {
	Content          string            `json:"content"`
	ReplyToMessageID *int64            `json:"reply_to_message_id,omitempty"`
	Attachments      []AttachmentInput `json:"attachments,omitempty"`
}

type AttachmentInput struct {
	ObjectKey string `json:"object_key"`
	FileName  string `json:"file_name"`
	MimeType  string `json:"mime_type"`
	SizeBytes int64  `json:"size_bytes"`
}

type MessageResponse struct {
	ID        int64  `json:"id"`
	ChannelID int64  `json:"channel_id"`
	SenderID  int64  `json:"sender_id"`
	Content   string `json:"content"`
	CreatedAt string `json:"created_at"`
}

type MessageListItem struct {
	ID        int64  `json:"id"`
	ChannelID int64  `json:"channel_id"`
	SenderID  int64  `json:"sender_id"`
	Username  string `json:"username"`
	Content   string `json:"content"`
	CreatedAt string `json:"created_at"`
}
