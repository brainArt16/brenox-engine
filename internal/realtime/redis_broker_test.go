package realtime

import (
	"os"
	"testing"
	"time"

	goredis "github.com/redis/go-redis/v9"
)

func TestRedisBrokerCrossInstance(t *testing.T) {
	url := os.Getenv("REDIS_URL")
	if url == "" {
		t.Skip("REDIS_URL not set")
	}

	opts, err := goredis.ParseURL(url)
	if err != nil {
		t.Fatalf("parse REDIS_URL: %v", err)
	}

	client := goredis.NewClient(opts)
	t.Cleanup(func() { _ = client.Close() })

	if err := client.Ping(t.Context()).Err(); err != nil {
		t.Skipf("redis unavailable: %v", err)
	}

	hubA := NewHub(Config{})
	hubB := NewHub(Config{})
	go hubA.Run()
	go hubB.Run()
	t.Cleanup(func() {
		hubA.Shutdown()
		hubB.Shutdown()
	})

	brokerA := NewRedisBroker(client, hubA)
	brokerB := NewRedisBroker(client, hubB)
	hubA.SetBroker(brokerA)
	hubB.SetBroker(brokerB)
	brokerA.Start()
	brokerB.Start()
	t.Cleanup(func() {
		brokerA.Close()
		brokerB.Close()
	})

	const workspaceID int64 = 9001
	const channelID int64 = 9002

	brokerA.EnsureSubscribed(workspaceID, channelID)
	brokerB.EnsureSubscribed(workspaceID, channelID)

	received := make(chan Event, 1)
	clientB := &Client{
		workspaceID: workspaceID,
		channelID:   channelID,
		userID:      42,
		send:        received,
	}
	hubB.register <- clientB
	time.Sleep(100 * time.Millisecond)

	event := NewOutboundEvent("message.new", workspaceID, channelID, map[string]any{
		"content": "cross-node",
	})
	brokerA.Publish(event)

	deadline := time.After(3 * time.Second)
	for {
		select {
		case got := <-received:
			if got.Type != "message.new" {
				continue
			}
			payload, ok := got.Payload.(map[string]any)
			if !ok || payload["content"] != "cross-node" {
				t.Fatalf("unexpected payload: %#v", got.Payload)
			}
			return
		case <-deadline:
			t.Fatal("timed out waiting for cross-instance event")
		}
	}
}

func TestChannelTopic(t *testing.T) {
	if got := ChannelTopic(1, 2); got != "workspace:1:channel:2" {
		t.Fatalf("unexpected topic: %s", got)
	}
}

func TestUserTopic(t *testing.T) {
	if got := UserTopic(42); got != "user:42:notifications" {
		t.Fatalf("unexpected topic: %s", got)
	}
	userID, ok := parseUserTopic("user:42:notifications")
	if !ok || userID != 42 {
		t.Fatalf("unexpected parse result: %d %v", userID, ok)
	}
}
