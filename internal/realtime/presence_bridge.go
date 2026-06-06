package realtime

import "github.com/brainart16/brenox/internal/presence"

type hubBroadcaster struct {
	hub *Hub
}

func NewHubBroadcaster(hub *Hub) presence.Broadcaster {
	return &hubBroadcaster{hub: hub}
}

func (b *hubBroadcaster) PublishPresenceOnline(workspaceID, channelID, userID int64) {
	b.hub.Publish(NewOutboundEvent("presence.online", workspaceID, channelID, map[string]any{
		"user_id": userID,
	}))
}

func (b *hubBroadcaster) PublishPresenceOffline(workspaceID, channelID, userID int64) {
	b.hub.Publish(NewOutboundEvent("presence.offline", workspaceID, channelID, map[string]any{
		"user_id": userID,
	}))
}

func (b *hubBroadcaster) PublishPresenceStatus(workspaceID, channelID, userID int64, status, lastSeen string) {
	b.hub.Publish(NewOutboundEvent("presence.status", workspaceID, channelID, map[string]any{
		"user_id":   userID,
		"status":    status,
		"last_seen": lastSeen,
	}))
}
