package ratelimit

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	goredis "github.com/redis/go-redis/v9"
)

type Limiter struct {
	redis  *goredis.Client
	limit  int
	window time.Duration

	mu    sync.Mutex
	count map[string][]time.Time
}

func LoadConfig() Config {
	cfg := Config{RequestsPerMinute: 120}
	raw := strings.TrimSpace(os.Getenv("API_RATE_LIMIT_PER_MINUTE"))
	if raw == "" {
		return cfg
	}
	value, err := strconv.Atoi(raw)
	if err != nil || value <= 0 {
		return cfg
	}
	cfg.RequestsPerMinute = value
	return cfg
}

func LoadSandboxConfig() Config {
	cfg := Config{RequestsPerMinute: 30}
	raw := strings.TrimSpace(os.Getenv("SANDBOX_API_RATE_LIMIT_PER_MINUTE"))
	if raw == "" {
		return cfg
	}
	value, err := strconv.Atoi(raw)
	if err != nil || value <= 0 {
		return cfg
	}
	cfg.RequestsPerMinute = value
	return cfg
}

type Config struct {
	RequestsPerMinute int
}

func NewLimiter(redis *goredis.Client, cfg Config) *Limiter {
	if cfg.RequestsPerMinute <= 0 {
		cfg.RequestsPerMinute = 120
	}
	return &Limiter{
		redis:  redis,
		limit:  cfg.RequestsPerMinute,
		window: time.Minute,
		count:  make(map[string][]time.Time),
	}
}

func (l *Limiter) Allow(ctx context.Context, key string) bool {
	if l.redis != nil {
		return l.allowRedis(ctx, key)
	}
	return l.allowMemory(key)
}

func (l *Limiter) allowRedis(ctx context.Context, key string) bool {
	redisKey := "ratelimit:" + key
	count, err := l.redis.Incr(ctx, redisKey).Result()
	if err != nil {
		return l.allowMemory(key)
	}
	if count == 1 {
		_ = l.redis.Expire(ctx, redisKey, l.window).Err()
	}
	return int(count) <= l.limit
}

func (l *Limiter) allowMemory(key string) bool {
	now := time.Now()
	cutoff := now.Add(-l.window)

	l.mu.Lock()
	defer l.mu.Unlock()

	times := l.count[key]
	filtered := times[:0]
	for _, ts := range times {
		if ts.After(cutoff) {
			filtered = append(filtered, ts)
		}
	}
	if len(filtered) >= l.limit {
		l.count[key] = filtered
		return false
	}
	filtered = append(filtered, now)
	l.count[key] = filtered
	return true
}

func Key(appID int64, apiKeyID int64) string {
	return fmt.Sprintf("app:%d:key:%d", appID, apiKeyID)
}
