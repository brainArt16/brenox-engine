package ratelimit

import (
	"os"
	"strconv"
	"strings"
)

type IPConfig struct {
	RequestsPerMinute int
}

func LoadIPConfig() IPConfig {
	cfg := IPConfig{RequestsPerMinute: 300}
	raw := strings.TrimSpace(os.Getenv("HTTP_RATE_LIMIT_PER_IP"))
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

func IPKey(ip string) string {
	return "ip:" + ip
}
