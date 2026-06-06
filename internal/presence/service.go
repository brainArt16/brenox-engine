package presence

import (
	"context"
	"log/slog"

	goredis "github.com/redis/go-redis/v9"

	db "github.com/brainart16/brenox/internal/db"
)

type Service struct {
	store       Store
	queries     *db.Queries
	broadcaster Broadcaster
}

func NewService(client *goredis.Client, queries *db.Queries, broadcaster Broadcaster) *Service {
	cfg := LoadConfig()
	var store Store
	if client != nil {
		store = NewRedisStore(client, cfg)
		slog.Info("using redis presence store")
	} else {
		store = NewMemoryStore(cfg)
		slog.Info("using in-memory presence store")
	}

	return &Service{
		store:       store,
		queries:     queries,
		broadcaster: broadcaster,
	}
}

func (s *Service) Connect(ctx context.Context, userID, workspaceID, channelID int64) error {
	result, err := s.store.Connect(ctx, userID, workspaceID, channelID)
	if err != nil {
		return err
	}
	if result.BecameOnline && s.broadcaster != nil {
		s.broadcaster.PublishPresenceOnline(workspaceID, channelID, userID)
	}
	return nil
}

func (s *Service) Disconnect(ctx context.Context, userID, workspaceID, channelID int64) error {
	result, err := s.store.Disconnect(ctx, userID, workspaceID, channelID)
	if err != nil {
		return err
	}
	if result.BecameOffline && s.broadcaster != nil {
		s.broadcaster.PublishPresenceOffline(workspaceID, channelID, userID)
	}
	return nil
}

func (s *Service) Touch(ctx context.Context, userID int64) error {
	return s.store.Touch(ctx, userID)
}

func (s *Service) UpdateStatus(ctx context.Context, userID int64, status string) (UserPresence, error) {
	if !ValidStatus(status) {
		return UserPresence{}, ErrInvalidStatus
	}

	presence, err := s.store.SetStatus(ctx, userID, status)
	if err != nil {
		return UserPresence{}, err
	}

	if s.broadcaster != nil {
		channels, err := s.store.ActiveChannels(ctx, userID)
		if err != nil {
			return presence, err
		}
		for _, ch := range channels {
			s.broadcaster.PublishPresenceStatus(ch.WorkspaceID, ch.ChannelID, userID, presence.Status, presence.LastSeen)
		}
	}

	return presence, nil
}

func (s *Service) ListOnline(ctx context.Context) ([]UserPresence, error) {
	return s.store.ListOnline(ctx)
}

func (s *Service) ListWorkspacePresence(ctx context.Context, workspaceID, requesterID int64) ([]UserPresence, error) {
	isMember, err := s.queries.IsWorkspaceMember(ctx, db.IsWorkspaceMemberParams{
		WorkspaceID: workspaceID,
		UserID:      requesterID,
	})
	if err != nil {
		return nil, err
	}
	if !isMember {
		return nil, ErrNotWorkspaceMember
	}

	userIDs, err := s.store.ListWorkspaceOnlineUserIDs(ctx, workspaceID)
	if err != nil {
		return nil, err
	}

	items := make([]UserPresence, 0, len(userIDs))
	for _, userID := range userIDs {
		presence, err := s.store.Get(ctx, userID)
		if err != nil {
			continue
		}
		if presence.ConnectionCount > 0 {
			items = append(items, presence)
		}
	}
	return items, nil
}

func (s *Service) Get(ctx context.Context, userID int64) (UserPresence, error) {
	return s.store.Get(ctx, userID)
}
