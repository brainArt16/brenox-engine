package presence

import (
	"context"
	"strconv"
	"time"

	goredis "github.com/redis/go-redis/v9"
)

const connectScript = `
local ws_key = KEYS[1]
local global_key = KEYS[2]
local ws_online_key = KEYS[3]
local global_online_key = KEYS[4]
local channel_key = KEYS[5]
local channels_key = KEYS[6]
local user_id = ARGV[1]
local channel_ref = ARGV[2]
local now = ARGV[3]
local ttl = ARGV[4]

local ws_count = redis.call('INCR', ws_key)
redis.call('EXPIRE', ws_key, ttl)
if tonumber(ws_count) == 1 then
  redis.call('SADD', ws_online_key, user_id)
end
redis.call('EXPIRE', ws_online_key, ttl)

local ch_count = redis.call('INCR', channel_key)
redis.call('EXPIRE', channel_key, ttl)
if tonumber(ch_count) == 1 then
  redis.call('SADD', channels_key, channel_ref)
end
redis.call('EXPIRE', channels_key, ttl)

local global_count = redis.call('HINCRBY', global_key, 'connection_count', 1)
redis.call('HSET', global_key, 'last_seen', now)
redis.call('EXPIRE', global_key, ttl)

local became_online = 0
if tonumber(global_count) == 1 then
  redis.call('HSET', global_key, 'status', 'online')
  redis.call('SADD', global_online_key, user_id)
  became_online = 1
end
redis.call('EXPIRE', global_online_key, ttl)

return {global_count, became_online}
`

const disconnectScript = `
local ws_key = KEYS[1]
local global_key = KEYS[2]
local ws_online_key = KEYS[3]
local global_online_key = KEYS[4]
local channel_key = KEYS[5]
local channels_key = KEYS[6]
local user_id = ARGV[1]
local channel_ref = ARGV[2]
local now = ARGV[3]
local ttl = ARGV[4]

local ws_count = redis.call('DECR', ws_key)
if tonumber(ws_count) <= 0 then
  redis.call('DEL', ws_key)
  redis.call('SREM', ws_online_key, user_id)
end

local ch_count = redis.call('DECR', channel_key)
if tonumber(ch_count) <= 0 then
  redis.call('DEL', channel_key)
  redis.call('SREM', channels_key, channel_ref)
end

local global_count = redis.call('HINCRBY', global_key, 'connection_count', -1)
redis.call('HSET', global_key, 'last_seen', now)
if tonumber(global_count) < 0 then
  redis.call('HSET', global_key, 'connection_count', 0)
  global_count = 0
end

local became_offline = 0
if tonumber(global_count) <= 0 then
  redis.call('HSET', global_key, 'status', 'offline', 'connection_count', 0)
  redis.call('SREM', global_online_key, user_id)
  became_offline = 1
end
redis.call('EXPIRE', global_key, ttl)

return {global_count, became_offline}
`

type redisStore struct {
	client *goredis.Client
	cfg    Config
}

func NewRedisStore(client *goredis.Client, cfg Config) Store {
	return &redisStore{client: client, cfg: cfg}
}

func (s *redisStore) Connect(ctx context.Context, userID, workspaceID, channelID int64) (ConnectResult, error) {
	now := time.Now().UTC().Format(time.RFC3339Nano)
	ttlSec := int(s.cfg.TTL.Seconds())

	result, err := s.client.Eval(ctx, connectScript, []string{
		workspaceConnKey(userID, workspaceID),
		userKey(userID),
		workspaceOnlineKey(workspaceID),
		globalOnlineKey(),
		channelConnKey(userID, workspaceID, channelID),
		activeChannelsKey(userID),
	},
		strconv.FormatInt(userID, 10),
		channelRef(workspaceID, channelID),
		now,
		strconv.Itoa(ttlSec),
	).Int64Slice()
	if err != nil {
		return ConnectResult{}, err
	}

	return ConnectResult{
		GlobalCount:  result[0],
		BecameOnline: result[1] == 1,
	}, nil
}

func (s *redisStore) Disconnect(ctx context.Context, userID, workspaceID, channelID int64) (DisconnectResult, error) {
	now := time.Now().UTC().Format(time.RFC3339Nano)
	ttlSec := int(s.cfg.TTL.Seconds())

	result, err := s.client.Eval(ctx, disconnectScript, []string{
		workspaceConnKey(userID, workspaceID),
		userKey(userID),
		workspaceOnlineKey(workspaceID),
		globalOnlineKey(),
		channelConnKey(userID, workspaceID, channelID),
		activeChannelsKey(userID),
	},
		strconv.FormatInt(userID, 10),
		channelRef(workspaceID, channelID),
		now,
		strconv.Itoa(ttlSec),
	).Int64Slice()
	if err != nil {
		return DisconnectResult{}, err
	}

	return DisconnectResult{
		GlobalCount:   maxInt64(result[0], 0),
		BecameOffline: result[1] == 1,
	}, nil
}

func (s *redisStore) Touch(ctx context.Context, userID int64) error {
	now := time.Now().UTC().Format(time.RFC3339Nano)
	pipe := s.client.Pipeline()
	pipe.HSet(ctx, userKey(userID), "last_seen", now)
	pipe.Expire(ctx, userKey(userID), s.cfg.TTL)
	pipe.Expire(ctx, globalOnlineKey(), s.cfg.TTL)
	pipe.Expire(ctx, activeChannelsKey(userID), s.cfg.TTL)
	_, err := pipe.Exec(ctx)
	return err
}

func (s *redisStore) SetStatus(ctx context.Context, userID int64, status string) (UserPresence, error) {
	now := time.Now().UTC().Format(time.RFC3339Nano)
	key := userKey(userID)

	count, err := s.client.HGet(ctx, key, "connection_count").Int64()
	if err != nil && err != goredis.Nil {
		return UserPresence{}, err
	}
	if count > 0 && status == StatusOffline {
		status = StatusAway
	}

	pipe := s.client.Pipeline()
	pipe.HSet(ctx, key, "status", status, "last_seen", now)
	pipe.Expire(ctx, key, s.cfg.TTL)
	_, err = pipe.Exec(ctx)
	if err != nil {
		return UserPresence{}, err
	}

	return s.Get(ctx, userID)
}

func (s *redisStore) Get(ctx context.Context, userID int64) (UserPresence, error) {
	values, err := s.client.HGetAll(ctx, userKey(userID)).Result()
	if err != nil {
		return UserPresence{}, err
	}
	if len(values) == 0 {
		return UserPresence{
			UserID:          userID,
			Status:          StatusOffline,
			ConnectionCount: 0,
			LastSeen:        time.Time{}.Format(time.RFC3339Nano),
		}, nil
	}

	count, _ := strconv.ParseInt(values["connection_count"], 10, 64)
	status := values["status"]
	if count <= 0 {
		status = StatusOffline
	}
	lastSeen := values["last_seen"]
	if lastSeen == "" {
		lastSeen = time.Time{}.Format(time.RFC3339Nano)
	}

	return UserPresence{
		UserID:          userID,
		Status:          status,
		ConnectionCount: count,
		LastSeen:        lastSeen,
	}, nil
}

func (s *redisStore) ListOnline(ctx context.Context) ([]UserPresence, error) {
	userIDs, err := s.client.SMembers(ctx, globalOnlineKey()).Result()
	if err != nil {
		return nil, err
	}

	items := make([]UserPresence, 0, len(userIDs))
	for _, rawID := range userIDs {
		userID, err := strconv.ParseInt(rawID, 10, 64)
		if err != nil {
			continue
		}
		presence, err := s.Get(ctx, userID)
		if err != nil {
			continue
		}
		if presence.ConnectionCount > 0 {
			items = append(items, presence)
		}
	}
	return items, nil
}

func (s *redisStore) ListWorkspaceOnlineUserIDs(ctx context.Context, workspaceID int64) ([]int64, error) {
	rawIDs, err := s.client.SMembers(ctx, workspaceOnlineKey(workspaceID)).Result()
	if err != nil {
		return nil, err
	}

	ids := make([]int64, 0, len(rawIDs))
	for _, rawID := range rawIDs {
		userID, err := strconv.ParseInt(rawID, 10, 64)
		if err != nil {
			continue
		}
		ids = append(ids, userID)
	}
	return ids, nil
}

func (s *redisStore) ActiveChannels(ctx context.Context, userID int64) ([]ChannelRef, error) {
	rawRefs, err := s.client.SMembers(ctx, activeChannelsKey(userID)).Result()
	if err != nil {
		return nil, err
	}

	refs := make([]ChannelRef, 0, len(rawRefs))
	for _, raw := range rawRefs {
		ref, ok := parseChannelRef(raw)
		if ok {
			refs = append(refs, ref)
		}
	}
	return refs, nil
}

func maxInt64(a, b int64) int64 {
	if a > b {
		return a
	}
	return b
}
