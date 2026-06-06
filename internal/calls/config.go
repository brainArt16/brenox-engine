package calls

import (
	"os"
	"strconv"
	"strings"
)

type Config struct {
	MaxParticipants int
}

func LoadConfig() Config {
	cfg := Config{MaxParticipants: 25}
	raw := strings.TrimSpace(os.Getenv("CALL_MAX_PARTICIPANTS"))
	if raw == "" {
		return cfg
	}
	value, err := strconv.Atoi(raw)
	if err != nil || value <= 0 {
		return cfg
	}
	cfg.MaxParticipants = value
	return cfg
}
