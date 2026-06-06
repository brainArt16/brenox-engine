package realtime

// Event is the WebSocket envelope for all realtime messages.
type Event struct {
	Type        string `json:"type"`
	WorkspaceID int64  `json:"workspace_id,omitempty"`
	ChannelID   int64  `json:"channel_id"`
	Payload     any    `json:"payload"`
}
