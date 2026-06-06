package realtime

import "fmt"

// ChannelTopic returns the Redis pub/sub channel for a workspace channel.
func ChannelTopic(workspaceID, channelID int64) string {
	return fmt.Sprintf("workspace:%d:channel:%d", workspaceID, channelID)
}
