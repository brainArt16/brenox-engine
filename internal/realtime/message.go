package realtime


// WebSocket messages should have structured event format.
type Message struct {

	/*
		Event type.

		Examples:
		- message.send
		- typing.start
		- presence.update
	*/

	Type string `json:"type"`

	// Actual message payload.
	Payload any `json:"payload"`
}