package realtime

import (
	"strconv"
	"strings"
)

func parseUserTopic(topic string) (int64, bool) {
	if !strings.HasPrefix(topic, "user:") || !strings.HasSuffix(topic, ":notifications") {
		return 0, false
	}
	parts := strings.Split(topic, ":")
	if len(parts) != 3 {
		return 0, false
	}
	userID, err := strconv.ParseInt(parts[1], 10, 64)
	return userID, err == nil
}
