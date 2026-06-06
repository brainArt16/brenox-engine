package realtime

import (
	"os"
	"strconv"
	"strings"
)

type Config struct {
	AllowedOrigins        []string
	MaxConnectionsPerUser int
	MaxConnectionsPerIP   int
}

func LoadConfig() Config {
	cfg := Config{
		AllowedOrigins:        parseOrigins(os.Getenv("WS_ALLOWED_ORIGINS")),
		MaxConnectionsPerUser: envInt("WS_MAX_CONNECTIONS_PER_USER", 5),
		MaxConnectionsPerIP:   envInt("WS_MAX_CONNECTIONS_PER_IP", 20),
	}
	return cfg
}

func parseOrigins(raw string) []string {
	raw = strings.TrimSpace(raw)
	if raw == "" || raw == "*" {
		return nil
	}

	parts := strings.Split(raw, ",")
	origins := make([]string, 0, len(parts))
	for _, part := range parts {
		origin := strings.TrimSpace(part)
		if origin != "" {
			origins = append(origins, origin)
		}
	}
	return origins
}

func envInt(key string, fallback int) int {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return fallback
	}
	value, err := strconv.Atoi(raw)
	if err != nil || value <= 0 {
		return fallback
	}
	return value
}

func (c Config) originAllowed(origin string) bool {
	if len(c.AllowedOrigins) == 0 {
		return true
	}
	for _, allowed := range c.AllowedOrigins {
		if allowed == origin {
			return true
		}
	}
	return false
}
