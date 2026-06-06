package presence

import (
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	TTL time.Duration
}

func LoadConfig() Config {
	raw := strings.TrimSpace(os.Getenv("PRESENCE_TTL_SECONDS"))
	seconds := 120
	if raw != "" {
		if value, err := strconv.Atoi(raw); err == nil && value > 0 {
			seconds = value
		}
	}
	return Config{TTL: time.Duration(seconds) * time.Second}
}

func userKey(userID int64) string {
	return "presence:" + strconv.FormatInt(userID, 10)
}

func globalOnlineKey() string {
	return "presence:online"
}

func workspaceOnlineKey(workspaceID int64) string {
	return "presence:workspace:" + strconv.FormatInt(workspaceID, 10) + ":online"
}

func workspaceConnKey(userID, workspaceID int64) string {
	return "presence:ws:" + strconv.FormatInt(userID, 10) + ":" + strconv.FormatInt(workspaceID, 10)
}

func channelConnKey(userID, workspaceID, channelID int64) string {
	return "presence:ch:" + strconv.FormatInt(userID, 10) + ":" + strconv.FormatInt(workspaceID, 10) + ":" + strconv.FormatInt(channelID, 10)
}

func activeChannelsKey(userID int64) string {
	return "presence:user:" + strconv.FormatInt(userID, 10) + ":channels"
}

func channelRef(workspaceID, channelID int64) string {
	return strconv.FormatInt(workspaceID, 10) + ":" + strconv.FormatInt(channelID, 10)
}

func parseChannelRef(raw string) (ChannelRef, bool) {
	parts := strings.SplitN(raw, ":", 2)
	if len(parts) != 2 {
		return ChannelRef{}, false
	}
	workspaceID, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return ChannelRef{}, false
	}
	channelID, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		return ChannelRef{}, false
	}
	return ChannelRef{WorkspaceID: workspaceID, ChannelID: channelID}, true
}
