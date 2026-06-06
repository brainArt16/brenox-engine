package chat

import (
	"context"
	"strings"
	"time"

	db "github.com/brainart16/brenox/internal/db"
	"github.com/jackc/pgx/v5/pgtype"
)

type Service struct {
	queries *db.Queries
}

func NewService(queries *db.Queries) *Service {
	return &Service{queries: queries}
}

func normalizeContent(content string) (string, error) {
	content = strings.TrimSpace(content)
	if content == "" {
		return "", ErrEmptyContent
	}
	if len(content) > MaxMessageLength {
		return "", ErrMessageTooLong
	}
	return content, nil
}

func (s *Service) assertMember(ctx context.Context, channelID, userID int64) error {
	isMember, err := s.queries.IsChannelMember(ctx, db.IsChannelMemberParams{
		ChannelID: channelID,
		UserID:    userID,
	})
	if err != nil {
		return err
	}
	if !isMember {
		return ErrNotMember
	}
	return nil
}

// SendMessage validates content, checks membership, and persists the message.
func (s *Service) SendMessage(
	ctx context.Context,
	channelID int64,
	senderID int64,
	content string,
) (*db.Message, error) {
	normalized, err := normalizeContent(content)
	if err != nil {
		return nil, err
	}

	if err := s.assertMember(ctx, channelID, senderID); err != nil {
		return nil, err
	}

	return s.saveMessage(ctx, channelID, senderID, normalized)
}

// ListMessages returns paginated channel history for members only.
func (s *Service) ListMessages(
	ctx context.Context,
	channelID int64,
	userID int64,
	limit int32,
	offset int32,
) ([]db.GetChannelMessagesRow, error) {
	if err := s.assertMember(ctx, channelID, userID); err != nil {
		return nil, err
	}

	if limit <= 0 {
		limit = defaultMessageLimit
	}
	if limit > maxMessageLimit {
		limit = maxMessageLimit
	}
	if offset < 0 {
		offset = 0
	}

	return s.queries.GetChannelMessages(ctx, db.GetChannelMessagesParams{
		ChannelID: channelID,
		Limit:     limit,
		Offset:    offset,
	})
}

func (s *Service) saveMessage(
	ctx context.Context,
	channelID int64,
	senderID int64,
	content string,
) (*db.Message, error) {
	message, err := s.queries.CreateMessage(ctx, db.CreateMessageParams{
		ChannelID: channelID,
		SenderID:  senderID,
		Content:   content,
	})
	if err != nil {
		return nil, err
	}
	return &message, nil
}

func ToMessageResponse(msg db.Message) MessageResponse {
	return MessageResponse{
		ID:        msg.ID,
		ChannelID: msg.ChannelID,
		SenderID:  msg.SenderID,
		Content:   msg.Content,
		CreatedAt: formatTime(msg.CreatedAt),
	}
}

func ToMessageListItem(row db.GetChannelMessagesRow) MessageListItem {
	return MessageListItem{
		ID:        row.ID,
		ChannelID: row.ChannelID,
		SenderID:  row.SenderID,
		Username:  row.Username,
		Content:   row.Content,
		CreatedAt: formatTime(row.CreatedAt),
	}
}

// MessageNewPayload builds the WebSocket payload for message.new events.
func MessageNewPayload(msg db.Message) map[string]any {
	return map[string]any{
		"id":         msg.ID,
		"sender_id":  msg.SenderID,
		"content":    msg.Content,
		"created_at": formatTime(msg.CreatedAt),
	}
}

func formatTime(ts pgtype.Timestamptz) string {
	if !ts.Valid {
		return ""
	}
	return ts.Time.Format(time.RFC3339)
}
