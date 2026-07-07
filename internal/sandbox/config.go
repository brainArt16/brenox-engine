package sandbox

import (
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	defaultMaxUsers            = 100
	defaultMaxChannels         = 20
	defaultMaxMessages         = 1000
	defaultAPIKeyTTLDays       = 90
	defaultDataTTLDays         = 30
	defaultCleanupIntervalHour = 24
)

type Config struct {
	MaxUsers        int64
	MaxChannels     int64
	MaxMessages     int64
	APIKeyTTL       time.Duration
	DataTTL         time.Duration
	CleanupInterval time.Duration
}

func LoadConfig() Config {
	return Config{
		MaxUsers:        int64(envInt("SANDBOX_MAX_USERS", defaultMaxUsers)),
		MaxChannels:     int64(envInt("SANDBOX_MAX_CHANNELS", defaultMaxChannels)),
		MaxMessages:     int64(envInt("SANDBOX_MAX_MESSAGES", defaultMaxMessages)),
		APIKeyTTL:       days("SANDBOX_API_KEY_TTL_DAYS", defaultAPIKeyTTLDays),
		DataTTL:         days("SANDBOX_DATA_TTL_DAYS", defaultDataTTLDays),
		CleanupInterval: hours("SANDBOX_CLEANUP_INTERVAL_HOURS", defaultCleanupIntervalHour),
	}
}

func envInt(key string, fallback int) int {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return fallback
	}
	value, err := strconv.Atoi(raw)
	if err != nil || value < 0 {
		return fallback
	}
	return value
}

func days(key string, fallback int) time.Duration {
	value := envInt(key, fallback)
	if value <= 0 {
		return 0
	}
	return time.Duration(value) * 24 * time.Hour
}

func hours(key string, fallback int) time.Duration {
	value := envInt(key, fallback)
	if value <= 0 {
		return 0
	}
	return time.Duration(value) * time.Hour
}
