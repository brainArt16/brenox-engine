package realtime

import "github.com/brainart16/brenox/internal/calls"

type callBroadcaster struct {
	hub *Hub
}

func NewCallBroadcaster(hub *Hub) calls.Broadcaster {
	return &callBroadcaster{hub: hub}
}

func (b *callBroadcaster) PublishCallEvent(eventType string, workspaceID, channelID int64, payload map[string]any) {
	b.hub.Publish(NewOutboundEvent(eventType, workspaceID, channelID, payload))
}
