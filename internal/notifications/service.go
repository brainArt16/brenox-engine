package notifications

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"time"

	db "github.com/brainart16/brenox/internal/db"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

const (
	defaultListLimit = 50
	maxListLimit     = 100
)

type Service struct {
	queries   *db.Queries
	realtime  RealtimeDeliverer
	push      PushSender
	email     EmailSender
}

func NewService(
	queries *db.Queries,
	realtime RealtimeDeliverer,
	push PushSender,
	email EmailSender,
) *Service {
	if push == nil {
		push = NewNoopPushSender()
	}
	if email == nil {
		email = NewNoopEmailSender()
	}
	return &Service{
		queries:  queries,
		realtime: realtime,
		push:     push,
		email:    email,
	}
}

func (s *Service) Create(ctx context.Context, input CreateInput) (*db.Notification, error) {
	if !ValidType(input.Type) {
		return nil, ErrInvalidType
	}

	dataBytes, err := json.Marshal(input.Data)
	if err != nil {
		return nil, err
	}
	if input.Data == nil {
		dataBytes = []byte("{}")
	}

	row, err := s.queries.CreateNotification(ctx, db.CreateNotificationParams{
		UserID: input.UserID,
		Type:   input.Type,
		Title:  input.Title,
		Body:   input.Body,
		Data:   dataBytes,
	})
	if err != nil {
		return nil, err
	}

	s.dispatch(ctx, row)
	return &row, nil
}

func (s *Service) List(ctx context.Context, userID int64, limit, offset int32) ([]NotificationResponse, error) {
	if limit <= 0 {
		limit = defaultListLimit
	}
	if limit > maxListLimit {
		limit = maxListLimit
	}
	if offset < 0 {
		offset = 0
	}

	rows, err := s.queries.ListNotificationsByUser(ctx, db.ListNotificationsByUserParams{
		UserID: userID,
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		return nil, err
	}

	items := make([]NotificationResponse, 0, len(rows))
	for _, row := range rows {
		items = append(items, ToResponse(row))
	}
	return items, nil
}

func (s *Service) MarkRead(ctx context.Context, userID, notificationID int64) (NotificationResponse, error) {
	row, err := s.queries.MarkNotificationRead(ctx, db.MarkNotificationReadParams{
		ID:     notificationID,
		UserID: userID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return NotificationResponse{}, ErrNotFound
		}
		return NotificationResponse{}, err
	}
	return ToResponse(row), nil
}

func (s *Service) MarkAllRead(ctx context.Context, userID int64) (int64, error) {
	return s.queries.MarkAllNotificationsRead(ctx, userID)
}

func (s *Service) HandleMessageCreated(
	ctx context.Context,
	workspaceID, channelID, senderID int64,
	message db.Message,
	replyToMessageID *int64,
	senderUsername string,
) error {
	if replyToMessageID != nil {
		if err := s.notifyReply(ctx, workspaceID, channelID, senderID, message, *replyToMessageID, senderUsername); err != nil {
			slog.Warn("reply notification failed", "error", err)
		}
	}

	return s.notifyMentions(ctx, workspaceID, channelID, senderID, message, senderUsername)
}

func (s *Service) HandleWorkspaceInvite(
	ctx context.Context,
	workspaceID, actorID, targetUserID int64,
	workspaceName, actorUsername string,
) error {
	if targetUserID == actorID {
		return nil
	}

	_, err := s.Create(ctx, CreateInput{
		UserID: targetUserID,
		Type:   TypeWorkspaceInvite,
		Title:  "Workspace invitation",
		Body:   actorUsername + " added you to " + workspaceName,
		Data: map[string]any{
			"workspace_id": workspaceID,
			"actor_id":     actorID,
		},
	})
	return err
}

func (s *Service) notifyMentions(
	ctx context.Context,
	workspaceID, channelID, senderID int64,
	message db.Message,
	senderUsername string,
) error {
	usernames := ParseMentions(message.Content)
	for _, username := range usernames {
		member, err := s.queries.GetWorkspaceMemberByUsername(ctx, db.GetWorkspaceMemberByUsernameParams{
			WorkspaceID: workspaceID,
			Lower:       username,
		})
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				continue
			}
			return err
		}
		if member.UserID == senderID {
			continue
		}

		_, err = s.Create(ctx, CreateInput{
			UserID: member.UserID,
			Type:   TypeMention,
			Title:  "You were mentioned",
			Body:   senderUsername + " mentioned you in a message",
			Data: map[string]any{
				"workspace_id": workspaceID,
				"channel_id":   channelID,
				"message_id":   message.ID,
				"sender_id":    senderID,
			},
		})
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) notifyReply(
	ctx context.Context,
	workspaceID, channelID, senderID int64,
	message db.Message,
	replyToMessageID int64,
	senderUsername string,
) error {
	parent, err := s.queries.GetMessageByID(ctx, replyToMessageID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil
		}
		return err
	}
	if parent.ChannelID != channelID || parent.SenderID == senderID {
		return nil
	}

	_, err = s.Create(ctx, CreateInput{
		UserID: parent.SenderID,
		Type:   TypeReply,
		Title:  "New reply",
		Body:   senderUsername + " replied to your message",
		Data: map[string]any{
			"workspace_id":        workspaceID,
			"channel_id":            channelID,
			"message_id":            message.ID,
			"reply_to_message_id":   replyToMessageID,
			"sender_id":             senderID,
		},
	})
	return err
}

func (s *Service) dispatch(ctx context.Context, row db.Notification) {
	payload := ToResponse(row)
	payloadMap := map[string]any{
		"id":         payload.ID,
		"type":       payload.Type,
		"title":      payload.Title,
		"body":       payload.Body,
		"data":       payload.Data,
		"read":       payload.Read,
		"created_at": payload.CreatedAt,
	}
	if payload.ReadAt != "" {
		payloadMap["read_at"] = payload.ReadAt
	}

	if s.realtime != nil {
		s.realtime.DeliverNotification(row.UserID, payloadMap)
	}

	if err := s.push.SendPush(ctx, row.UserID, row.Title, row.Body, payload.Data); err != nil {
		slog.Warn("push notification failed", "user_id", row.UserID, "error", err)
	}
	if err := s.email.SendEmail(ctx, row.UserID, row.Title, row.Body); err != nil {
		slog.Warn("email notification failed", "user_id", row.UserID, "error", err)
	}
}

func ToResponse(row db.Notification) NotificationResponse {
	data := map[string]any{}
	if len(row.Data) > 0 {
		_ = json.Unmarshal(row.Data, &data)
	}
	if data == nil {
		data = map[string]any{}
	}

	resp := NotificationResponse{
		ID:        row.ID,
		Type:      row.Type,
		Title:     row.Title,
		Body:      row.Body,
		Data:      data,
		Read:      row.ReadAt.Valid,
		CreatedAt: formatTime(row.CreatedAt),
	}
	if row.ReadAt.Valid {
		resp.ReadAt = formatTime(row.ReadAt)
	}
	return resp
}

func formatTime(ts pgtype.Timestamptz) string {
	if !ts.Valid {
		return ""
	}
	return ts.Time.UTC().Format(time.RFC3339Nano)
}
