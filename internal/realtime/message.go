package realtime


// Realtime websocket event.
type Event struct {

	Type string `json:"type"`

	ChannelID int64 `json:"channel_id"`

	Payload any `json:"payload"`
}