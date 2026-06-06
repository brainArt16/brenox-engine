package presence

import (
	"context"
	"sync"
	"time"
)

type memoryStore struct {
	mu sync.Mutex
	cfg Config

	users      map[int64]*memoryUser
	globalOnline map[int64]struct{}
	wsOnline   map[int64]map[int64]struct{} // workspace -> users
}

type memoryUser struct {
	status          string
	connectionCount int64
	lastSeen        time.Time
	channels        map[int64]map[int64]int // workspace -> channel -> count
}

func NewMemoryStore(cfg Config) Store {
	return &memoryStore{
		cfg:          cfg,
		users:        make(map[int64]*memoryUser),
		globalOnline: make(map[int64]struct{}),
		wsOnline:     make(map[int64]map[int64]struct{}),
	}
}

func (s *memoryStore) Connect(ctx context.Context, userID, workspaceID, channelID int64) (ConnectResult, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	user := s.ensureUser(userID)
	user.connectionCount++
	user.lastSeen = time.Now().UTC()
	if user.status == "" || user.status == StatusOffline {
		user.status = StatusOnline
	}

	if user.channels[workspaceID] == nil {
		user.channels[workspaceID] = make(map[int64]int)
	}
	user.channels[workspaceID][channelID]++

	if user.channels[workspaceID][channelID] == 1 {
		s.addChannelRef(userID, workspaceID, channelID)
	}

	if user.connectionCount == 1 {
		s.globalOnline[userID] = struct{}{}
	}
	s.markWorkspaceOnline(workspaceID, userID)

	return ConnectResult{
		BecameOnline: user.connectionCount == 1,
		GlobalCount:  user.connectionCount,
	}, nil
}

func (s *memoryStore) Disconnect(ctx context.Context, userID, workspaceID, channelID int64) (DisconnectResult, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	user, ok := s.users[userID]
	if !ok {
		return DisconnectResult{}, nil
	}

	user.connectionCount--
	if user.connectionCount < 0 {
		user.connectionCount = 0
	}
	user.lastSeen = time.Now().UTC()

	if chs, ok := user.channels[workspaceID]; ok {
		chs[channelID]--
		if chs[channelID] <= 0 {
			delete(chs, channelID)
			s.removeChannelRef(userID, workspaceID, channelID)
		}
		if len(chs) == 0 {
			delete(user.channels, workspaceID)
		}
	}

	s.unmarkWorkspaceOnline(workspaceID, userID, user)

	becameOffline := false
	if user.connectionCount == 0 {
		user.status = StatusOffline
		delete(s.globalOnline, userID)
		becameOffline = true
	}

	return DisconnectResult{
		BecameOffline: becameOffline,
		GlobalCount:   user.connectionCount,
	}, nil
}

func (s *memoryStore) Touch(ctx context.Context, userID int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	user, ok := s.users[userID]
	if !ok {
		return nil
	}
	user.lastSeen = time.Now().UTC()
	return nil
}

func (s *memoryStore) SetStatus(ctx context.Context, userID int64, status string) (UserPresence, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	user := s.ensureUser(userID)
	if user.connectionCount > 0 && status == StatusOffline {
		status = StatusAway
	}
	user.status = status
	user.lastSeen = time.Now().UTC()
	return s.toPresence(userID, user), nil
}

func (s *memoryStore) Get(ctx context.Context, userID int64) (UserPresence, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	user, ok := s.users[userID]
	if !ok {
		return UserPresence{
			UserID:          userID,
			Status:          StatusOffline,
			ConnectionCount: 0,
			LastSeen:        time.Time{}.Format(time.RFC3339Nano),
		}, nil
	}
	return s.toPresence(userID, user), nil
}

func (s *memoryStore) ListOnline(ctx context.Context) ([]UserPresence, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	items := make([]UserPresence, 0, len(s.globalOnline))
	for userID := range s.globalOnline {
		user := s.users[userID]
		if user == nil {
			continue
		}
		items = append(items, s.toPresence(userID, user))
	}
	return items, nil
}

func (s *memoryStore) ListWorkspaceOnlineUserIDs(ctx context.Context, workspaceID int64) ([]int64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	users := s.wsOnline[workspaceID]
	ids := make([]int64, 0, len(users))
	for userID := range users {
		ids = append(ids, userID)
	}
	return ids, nil
}

func (s *memoryStore) ActiveChannels(ctx context.Context, userID int64) ([]ChannelRef, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	user, ok := s.users[userID]
	if !ok {
		return nil, nil
	}

	refs := make([]ChannelRef, 0)
	for workspaceID, channels := range user.channels {
		for channelID, count := range channels {
			if count > 0 {
				refs = append(refs, ChannelRef{WorkspaceID: workspaceID, ChannelID: channelID})
			}
		}
	}
	return refs, nil
}

func (s *memoryStore) ensureUser(userID int64) *memoryUser {
	user, ok := s.users[userID]
	if !ok {
		user = &memoryUser{
			status:   StatusOffline,
			channels: make(map[int64]map[int64]int),
		}
		s.users[userID] = user
	}
	return user
}

func (s *memoryStore) markWorkspaceOnline(workspaceID, userID int64) {
	if s.wsOnline[workspaceID] == nil {
		s.wsOnline[workspaceID] = make(map[int64]struct{})
	}
	s.wsOnline[workspaceID][userID] = struct{}{}
}

func (s *memoryStore) unmarkWorkspaceOnline(workspaceID, userID int64, user *memoryUser) {
	if user.channels[workspaceID] != nil && len(user.channels[workspaceID]) > 0 {
		return
	}
	if users, ok := s.wsOnline[workspaceID]; ok {
		delete(users, userID)
	}
}

func (s *memoryStore) addChannelRef(userID, workspaceID, channelID int64) {
	_ = userID
	_ = workspaceID
	_ = channelID
}

func (s *memoryStore) removeChannelRef(userID, workspaceID, channelID int64) {
	_ = userID
	_ = workspaceID
	_ = channelID
}

func (s *memoryStore) toPresence(userID int64, user *memoryUser) UserPresence {
	status := user.status
	if user.connectionCount == 0 {
		status = StatusOffline
	}
	return UserPresence{
		UserID:          userID,
		Status:          status,
		ConnectionCount: user.connectionCount,
		LastSeen:        user.lastSeen.UTC().Format(time.RFC3339Nano),
	}
}
