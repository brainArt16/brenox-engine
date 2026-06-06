package realtime

import (
	"fmt"
	"sync/atomic"
	"time"
)

var eventCounter atomic.Uint64

// Event is the WebSocket envelope for all realtime messages.
type Event struct {
	Type        string `json:"type"`
	WorkspaceID int64  `json:"workspace_id,omitempty"`
	ChannelID   int64  `json:"channel_id"`
	EventID     string `json:"event_id"`
	Timestamp   string `json:"timestamp"`
	Payload     any    `json:"payload"`
}

func NewOutboundEvent(eventType string, workspaceID, channelID int64, payload any) Event {
	return Event{
		Type:        eventType,
		WorkspaceID: workspaceID,
		ChannelID:   channelID,
		EventID:     newEventID(),
		Timestamp:   time.Now().UTC().Format(time.RFC3339Nano),
		Payload:     payload,
	}
}

func newEventID() string {
	return fmt.Sprintf("%d-%d", time.Now().UnixNano(), eventCounter.Add(1))
}
