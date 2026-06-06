package realtime

import "github.com/brainart16/brenox/internal/attachments"

type messageBroadcaster struct {
	hub *Hub
}

func NewMessageBroadcaster(hub *Hub) attachments.MessageBroadcaster {
	return &messageBroadcaster{hub: hub}
}

func (b *messageBroadcaster) PublishMessageUpdated(workspaceID, channelID int64, payload map[string]any) {
	b.hub.Publish(NewOutboundEvent("message.updated", workspaceID, channelID, payload))
}
