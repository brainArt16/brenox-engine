package realtime

import (
	"context"
	"fmt"
	"sync"

	goredis "github.com/redis/go-redis/v9"
)

// Sequencer assigns monotonic per-channel sequence numbers for gap detection.
type Sequencer interface {
	Next(ctx context.Context, channelID int64) int64
}

type memorySequencer struct {
	mu    sync.Mutex
	seq   map[int64]int64
}

func NewMemorySequencer() Sequencer {
	return &memorySequencer{seq: make(map[int64]int64)}
}

func (s *memorySequencer) Next(_ context.Context, channelID int64) int64 {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.seq[channelID]++
	return s.seq[channelID]
}

type redisSequencer struct {
	client *goredis.Client
}

func NewRedisSequencer(client *goredis.Client) Sequencer {
	return &redisSequencer{client: client}
}

func (s *redisSequencer) Next(ctx context.Context, channelID int64) int64 {
	key := fmt.Sprintf("channel:%d:seq", channelID)
	value, err := s.client.Incr(ctx, key).Result()
	if err != nil {
		return 0
	}
	return value
}

func NewSequencer(client *goredis.Client) Sequencer {
	if client != nil {
		return NewRedisSequencer(client)
	}
	return NewMemorySequencer()
}
