package realtime

import (
	"testing"
	"time"
)

func TestDeliverCallSignalTargetsUser(t *testing.T) {
	hub := NewHub(Config{})
	go hub.Run()
	t.Cleanup(func() { hub.Shutdown() })

	receiver := &Client{
		userID:    2,
		channelID: 1,
		send:      make(chan Event, 1),
	}
	sender := &Client{
		userID:    1,
		channelID: 1,
		send:      make(chan Event, 1),
	}
	hub.register <- sender
	hub.register <- receiver
	time.Sleep(50 * time.Millisecond)

	event := Event{
		Type:      "call.offer",
		ChannelID: 1,
		Payload: map[string]any{
			"call_id":      float64(10),
			"from_user_id": float64(1),
			"to_user_id":   float64(2),
			"sdp":          "v=0",
		},
	}

	hub.mu.Lock()
	targets := make([]*Client, 0, len(hub.channels[1]))
	for client := range hub.channels[1] {
		targets = append(targets, client)
	}
	hub.mu.Unlock()

	hub.deliverCallSignal(event, targets)

	select {
	case got := <-receiver.send:
		if got.Type != "call.offer" {
			t.Fatalf("expected call.offer, got %s", got.Type)
		}
	case <-sender.send:
		t.Fatal("sender should not receive targeted offer")
	default:
		t.Fatal("receiver did not get offer")
	}
}

func TestPayloadInt64(t *testing.T) {
	payload := map[string]any{"call_id": float64(42)}
	if got := payloadInt64(payload, "call_id"); got != 42 {
		t.Fatalf("expected 42, got %d", got)
	}
}
