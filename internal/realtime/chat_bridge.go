package realtime

import (
	db "github.com/brainart16/brenox/internal/db"
	"github.com/brainart16/brenox/internal/chat"
)

type ChatBroadcaster struct {
	hub *Hub
}

func NewChatBroadcaster(hub *Hub) *ChatBroadcaster {
	return &ChatBroadcaster{hub: hub}
}

func (b *ChatBroadcaster) PublishMessageNew(workspaceID, channelID int64, message db.Message) {
	b.hub.Publish(NewOutboundEvent("message.new", workspaceID, channelID, chat.MessageNewPayload(message)))
}
