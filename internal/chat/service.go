package chat

import (
	"context"
	"errors"
	"strings"
	"time"

	db "github.com/brainart16/brenox/internal/db"
	"github.com/jackc/pgx/v5"
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

func (s *Service) assertChannelAccess(
	ctx context.Context,
	workspaceID int64,
	channelID int64,
	userID int64,
) error {
	isWorkspaceMember, err := s.queries.IsWorkspaceMember(ctx, db.IsWorkspaceMemberParams{
		WorkspaceID: workspaceID,
		UserID:      userID,
	})
	if err != nil {
		return err
	}
	if !isWorkspaceMember {
		return ErrNotWorkspaceMember
	}

	_, err = s.queries.GetChannelInWorkspace(ctx, db.GetChannelInWorkspaceParams{
		ID:          channelID,
		WorkspaceID: workspaceID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrChannelNotFound
		}
		return err
	}

	isChannelMember, err := s.queries.IsChannelMember(ctx, db.IsChannelMemberParams{
		ChannelID: channelID,
		UserID:    userID,
	})
	if err != nil {
		return err
	}
	if !isChannelMember {
		return ErrNotMember
	}

	return nil
}

// SendMessage validates content, workspace access, and persists the message.
func (s *Service) SendMessage(
	ctx context.Context,
	workspaceID int64,
	channelID int64,
	senderID int64,
	content string,
) (*db.Message, error) {
	normalized, err := normalizeContent(content)
	if err != nil {
		return nil, err
	}

	if err := s.assertChannelAccess(ctx, workspaceID, channelID, senderID); err != nil {
		return nil, err
	}

	return s.saveMessage(ctx, channelID, senderID, normalized)
}

// ListMessages returns paginated channel history for workspace channel members.
func (s *Service) ListMessages(
	ctx context.Context,
	workspaceID int64,
	channelID int64,
	userID int64,
	limit int32,
	offset int32,
) ([]db.GetChannelMessagesRow, error) {
	if err := s.assertChannelAccess(ctx, workspaceID, channelID, userID); err != nil {
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
