package redis

import (
	"context"
	"fmt"
	"os"
	"strings"

	goredis "github.com/redis/go-redis/v9"
)

func LoadURL() string {
	return strings.TrimSpace(os.Getenv("REDIS_URL"))
}

func NewClient() (*goredis.Client, error) {
	url := LoadURL()
	if url == "" {
		return nil, fmt.Errorf("REDIS_URL is not set")
	}

	opts, err := goredis.ParseURL(url)
	if err != nil {
		return nil, fmt.Errorf("parse REDIS_URL: %w", err)
	}

	client := goredis.NewClient(opts)
	if err := client.Ping(context.Background()).Err(); err != nil {
		_ = client.Close()
		return nil, fmt.Errorf("redis ping: %w", err)
	}

	return client, nil
}

func Ping(ctx context.Context, client *goredis.Client) error {
	if client == nil {
		return fmt.Errorf("redis client is nil")
	}
	return client.Ping(ctx).Err()
}
