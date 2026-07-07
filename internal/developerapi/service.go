package developerapi

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/brainart16/brenox/internal/auth"
	db "github.com/brainart16/brenox/internal/db"
	"github.com/brainart16/brenox/internal/webhooks"
	brenoxjwt "github.com/brainart16/brenox/pkg/jwt"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

var (
	ErrNotFound        = errors.New("resource not found")
	ErrInvalidRequest  = errors.New("invalid request")
	ErrExternalIDTaken = errors.New("external_id already exists")
	ErrUserNotFound    = errors.New("user not found")
	ErrChannelNotFound = errors.New("channel not found")
	ErrEmptyContent    = errors.New("message content is required")
)

type MessageBroadcaster interface {
	PublishMessageNew(workspaceID, channelID int64, message db.Message)
}

type Service struct {
	queries   *db.Queries
	broadcast MessageBroadcaster
	webhooks  *webhooks.Dispatcher
	billing   AppBilling
}

type AppBilling interface {
	CheckMessageQuota(ctx context.Context, appID int64) error
	RecordMessageByAppID(ctx context.Context, appID int64) error
}

func NewService(queries *db.Queries, broadcast MessageBroadcaster, dispatcher *webhooks.Dispatcher) *Service {
	return &Service{
		queries:   queries,
		broadcast: broadcast,
		webhooks:  dispatcher,
	}
}

func (s *Service) SetBilling(billing AppBilling) {
	s.billing = billing
}

func (s *Service) CreateSession(ctx context.Context, runtime RuntimeApp, req CreateSessionRequest) (SessionResponse, error) {
	externalID := strings.TrimSpace(req.ExternalID)
	if externalID == "" {
		return SessionResponse{}, ErrInvalidRequest
	}

	userID, err := s.resolveUserID(ctx, runtime, 0, externalID)
	if err != nil {
		return SessionResponse{}, err
	}

	user, err := s.queries.GetUserByID(ctx, userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return SessionResponse{}, ErrUserNotFound
		}
		return SessionResponse{}, err
	}

	token, err := brenoxjwt.GenerateSessionToken(userID, runtime.App.ID, runtime.Sandbox)
	if err != nil {
		return SessionResponse{}, err
	}

	resp := SessionResponse{
		Token:       token,
		WorkspaceID: runtime.WorkspaceID,
		Environment: runtime.Environment,
		User:        toUserResponse(user, externalID),
	}

	if req.ChannelID > 0 {
		if _, err := s.queries.GetChannelInWorkspace(ctx, db.GetChannelInWorkspaceParams{
			ID:          req.ChannelID,
			WorkspaceID: runtime.WorkspaceID,
		}); err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return SessionResponse{}, ErrChannelNotFound
			}
			return SessionResponse{}, err
		}
		if err := s.ensureChannelMember(ctx, req.ChannelID, userID); err != nil {
			return SessionResponse{}, err
		}
		resp.ChannelID = req.ChannelID
	}

	return resp, nil
}

func (s *Service) ProvisionUser(ctx context.Context, runtime RuntimeApp, req ProvisionUserRequest) (UserResponse, error) {
	externalID := strings.TrimSpace(req.ExternalID)
	if externalID == "" {
		return UserResponse{}, ErrInvalidRequest
	}

	if existing, err := s.queries.GetAppUserByExternalID(ctx, db.GetAppUserByExternalIDParams{
		AppID:       runtime.App.ID,
		ExternalID:  externalID,
		Environment: runtime.Environment,
	}); err == nil {
		user, err := s.queries.GetUserByID(ctx, existing.UserID)
		if err != nil {
			return UserResponse{}, err
		}
		return toUserResponse(user, externalID), nil
	} else if !errors.Is(err, pgx.ErrNoRows) {
		return UserResponse{}, err
	}

	email := strings.TrimSpace(req.Email)
	if email == "" {
		email = fmt.Sprintf("%s@%s.app.brenox", externalID, runtime.App.Slug)
	}
	username := strings.TrimSpace(req.Username)
	if username == "" {
		username = fmt.Sprintf("%s_%s", runtime.App.Slug, externalID)
	}

	password, err := randomPassword()
	if err != nil {
		return UserResponse{}, err
	}
	hashed, err := auth.HashPassword(password)
	if err != nil {
		return UserResponse{}, err
	}

	user, err := s.queries.CreateUser(ctx, db.CreateUserParams{
		Email:        email,
		Username:     username,
		PasswordHash: hashed,
	})
	if err != nil {
		return UserResponse{}, err
	}

	if err := s.queries.AddWorkspaceMember(ctx, db.AddWorkspaceMemberParams{
		WorkspaceID: runtime.WorkspaceID,
		UserID:      user.ID,
		Role:        "member",
	}); err != nil {
		return UserResponse{}, err
	}

	if _, err := s.queries.CreateAppUser(ctx, db.CreateAppUserParams{
		AppID:       runtime.App.ID,
		UserID:      user.ID,
		ExternalID:  externalID,
		Environment: runtime.Environment,
	}); err != nil {
		return UserResponse{}, err
	}

	resp := toUserResponse(user, externalID)
	s.dispatch(ctx, runtime, "user.provisioned", resp)
	return resp, nil
}

func (s *Service) CreateChannel(ctx context.Context, runtime RuntimeApp, req CreateChannelRequest) (ChannelResponse, error) {
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return ChannelResponse{}, ErrInvalidRequest
	}

	channel, err := s.queries.CreateChannel(ctx, db.CreateChannelParams{
		Name:        name,
		OwnerID:     runtime.App.OwnerID,
		WorkspaceID: runtime.WorkspaceID,
		IsReadOnly:  req.IsReadOnly,
	})
	if err != nil {
		if isDuplicateChannelName(err) {
			existing, lookupErr := s.queries.GetChannelByNameInWorkspace(ctx, db.GetChannelByNameInWorkspaceParams{
				WorkspaceID: runtime.WorkspaceID,
				Name:        name,
			})
			if lookupErr != nil {
				return ChannelResponse{}, fmt.Errorf("channel name already exists")
			}
			return ChannelResponse{
				ID:          existing.ID,
				Name:        existing.Name,
				WorkspaceID: existing.WorkspaceID,
				IsReadOnly:  existing.IsReadOnly,
			}, nil
		}
		return ChannelResponse{}, err
	}

	if err := s.queries.AddChannelMember(ctx, db.AddChannelMemberParams{
		ChannelID: channel.ID,
		UserID:    runtime.App.OwnerID,
	}); err != nil {
		return ChannelResponse{}, err
	}

	resp := toChannelResponse(channel)
	s.dispatch(ctx, runtime, "channel.created", resp)
	return resp, nil
}

func (s *Service) SendMessage(ctx context.Context, runtime RuntimeApp, req SendMessageRequest) (MessageResponse, error) {
	content := strings.TrimSpace(req.Content)
	if content == "" {
		return MessageResponse{}, ErrEmptyContent
	}

	userID, err := s.resolveUserID(ctx, runtime, req.UserID, req.ExternalID)
	if err != nil {
		return MessageResponse{}, err
	}

	channel, err := s.queries.GetChannelInWorkspace(ctx, db.GetChannelInWorkspaceParams{
		ID:          req.ChannelID,
		WorkspaceID: runtime.WorkspaceID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return MessageResponse{}, ErrChannelNotFound
		}
		return MessageResponse{}, err
	}

	if err := s.ensureChannelMember(ctx, channel.ID, userID); err != nil {
		return MessageResponse{}, err
	}

	if !runtime.Sandbox && s.billing != nil {
		if err := s.billing.CheckMessageQuota(ctx, runtime.App.ID); err != nil {
			return MessageResponse{}, err
		}
	}

	message, err := s.queries.CreateMessage(ctx, db.CreateMessageParams{
		ChannelID: channel.ID,
		SenderID:  userID,
		Content:   content,
	})
	if err != nil {
		return MessageResponse{}, err
	}

	if s.broadcast != nil {
		s.broadcast.PublishMessageNew(runtime.WorkspaceID, channel.ID, message)
	}

	if !runtime.Sandbox && s.billing != nil {
		_ = s.billing.RecordMessageByAppID(ctx, runtime.App.ID)
	}

	resp := toMessageResponse(message)
	s.dispatch(ctx, runtime, "message.created", resp)
	return resp, nil
}

func (s *Service) ListMessages(ctx context.Context, runtime RuntimeApp, channelID int64, limit, offset int32) ([]MessageListItem, error) {
	if _, err := s.queries.GetChannelInWorkspace(ctx, db.GetChannelInWorkspaceParams{
		ID:          channelID,
		WorkspaceID: runtime.WorkspaceID,
	}); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrChannelNotFound
		}
		return nil, err
	}

	if limit <= 0 {
		limit = 50
	}
	if limit > 100 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}

	rows, err := s.queries.GetChannelMessages(ctx, db.GetChannelMessagesParams{
		ChannelID: channelID,
		Limit:     limit,
		Offset:    offset,
	})
	if err != nil {
		return nil, err
	}

	items := make([]MessageListItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, toMessageListItem(row))
	}
	return items, nil
}

func (s *Service) resolveUserID(ctx context.Context, runtime RuntimeApp, userID int64, externalID string) (int64, error) {
	if userID > 0 {
		appUser, err := s.queries.GetAppUserByUserID(ctx, db.GetAppUserByUserIDParams{
			AppID:  runtime.App.ID,
			UserID: userID,
		})
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return 0, ErrUserNotFound
			}
			return 0, err
		}
		if appUser.Environment != runtime.Environment {
			return 0, ErrUserNotFound
		}
		return userID, nil
	}

	externalID = strings.TrimSpace(externalID)
	if externalID == "" {
		return 0, ErrInvalidRequest
	}

	appUser, err := s.queries.GetAppUserByExternalID(ctx, db.GetAppUserByExternalIDParams{
		AppID:       runtime.App.ID,
		ExternalID:  externalID,
		Environment: runtime.Environment,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, ErrUserNotFound
		}
		return 0, err
	}
	return appUser.UserID, nil
}

func (s *Service) ensureChannelMember(ctx context.Context, channelID, userID int64) error {
	isMember, err := s.queries.IsChannelMember(ctx, db.IsChannelMemberParams{
		ChannelID: channelID,
		UserID:    userID,
	})
	if err != nil {
		return err
	}
	if isMember {
		return nil
	}
	return s.queries.AddChannelMember(ctx, db.AddChannelMemberParams{
		ChannelID: channelID,
		UserID:    userID,
	})
}

func (s *Service) dispatch(ctx context.Context, runtime RuntimeApp, event string, payload any) {
	if s.webhooks == nil || runtime.Sandbox {
		return
	}
	s.webhooks.Dispatch(ctx, runtime.App.ID, event, payload)
}

func randomPassword() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

func isDuplicateChannelName(err error) bool {
	return err != nil && strings.Contains(err.Error(), "channels_workspace_name_unique")
}

func (s *Service) GetIdempotency(ctx context.Context, appID int64, key string) ([]byte, int, bool, error) {
	row, err := s.queries.GetIdempotencyKey(ctx, db.GetIdempotencyKeyParams{
		AppID:          appID,
		IdempotencyKey: key,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, 0, false, nil
		}
		return nil, 0, false, err
	}
	return row.ResponseBody, int(row.StatusCode), true, nil
}

func (s *Service) SaveIdempotency(ctx context.Context, appID int64, key, endpoint string, statusCode int, body []byte) error {
	_, err := s.queries.CreateIdempotencyKey(ctx, db.CreateIdempotencyKeyParams{
		AppID:          appID,
		IdempotencyKey: key,
		Endpoint:       endpoint,
		StatusCode:     int32(statusCode),
		ResponseBody:   body,
	})
	if err != nil && strings.Contains(err.Error(), "idempotency_keys_app_id_idempotency_key_key") {
		return nil
	}
	return err
}

func MarshalResponse(v any) ([]byte, error) {
	return json.Marshal(v)
}

func toUserResponse(user db.User, externalID string) UserResponse {
	return UserResponse{
		ID:         user.ID,
		ExternalID: externalID,
		Email:      user.Email,
		Username:   user.Username,
	}
}

func toChannelResponse(channel db.Channel) ChannelResponse {
	return ChannelResponse{
		ID:          channel.ID,
		Name:        channel.Name,
		WorkspaceID: channel.WorkspaceID,
		IsReadOnly:  channel.IsReadOnly,
	}
}

func toMessageResponse(message db.Message) MessageResponse {
	return MessageResponse{
		ID:        message.ID,
		ChannelID: message.ChannelID,
		SenderID:  message.SenderID,
		Content:   message.Content,
		CreatedAt: formatTime(message.CreatedAt),
	}
}

func toMessageListItem(row db.GetChannelMessagesRow) MessageListItem {
	return MessageListItem{
		ID:        row.ID,
		ChannelID: row.ChannelID,
		SenderID:  row.SenderID,
		Username:  row.Username,
		Content:   row.Content,
		CreatedAt: formatTime(row.CreatedAt),
	}
}

func formatTime(ts pgtype.Timestamptz) string {
	if !ts.Valid {
		return ""
	}
	return ts.Time.UTC().Format(time.RFC3339Nano)
}
