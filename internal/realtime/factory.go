package realtime

import (
	"log/slog"

	goredis "github.com/redis/go-redis/v9"
)

// NewBroker returns a Redis-backed broker when client is non-nil, otherwise local-only delivery.
func NewBroker(client *goredis.Client, hub *Hub) EventBroker {
	if client == nil {
		slog.Info("redis not configured, using local-only realtime broker")
		return NewLocalBroker(hub)
	}

	slog.Info("using redis realtime broker")
	return NewRedisBroker(client, hub)
}
