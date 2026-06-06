package realtime

import (
	"context"
	"encoding/json"
	"log/slog"
	"sync"
	"time"

	goredis "github.com/redis/go-redis/v9"
)

type redisBroker struct {
	client *goredis.Client
	hub    *Hub

	mu     sync.Mutex
	subs   map[string]int
	pubsub *goredis.PubSub

	ctx    context.Context
	cancel context.CancelFunc
}

func NewRedisBroker(client *goredis.Client, hub *Hub) EventBroker {
	ctx, cancel := context.WithCancel(context.Background())
	return &redisBroker{
		client: client,
		hub:    hub,
		subs:   make(map[string]int),
		pubsub: client.Subscribe(ctx),
		ctx:    ctx,
		cancel: cancel,
	}
}

func (b *redisBroker) Start() {
	go b.readLoop()
}

func (b *redisBroker) Close() {
	b.cancel()
	_ = b.pubsub.Close()
}

func (b *redisBroker) Publish(event Event) {
	if event.EventID == "" {
		event = NewOutboundEvent(event.Type, event.WorkspaceID, event.ChannelID, event.Payload)
	}

	data, err := json.Marshal(event)
	if err != nil {
		slog.Error("marshal realtime event", "error", err)
		return
	}

	topic := ChannelTopic(event.WorkspaceID, event.ChannelID)
	if err := b.client.Publish(b.ctx, topic, data).Err(); err != nil {
		slog.Error("redis publish failed", "topic", topic, "error", err)
	}
}

func (b *redisBroker) EnsureSubscribed(workspaceID, channelID int64) {
	topic := ChannelTopic(workspaceID, channelID)

	b.mu.Lock()
	defer b.mu.Unlock()

	b.subs[topic]++
	if b.subs[topic] == 1 {
		if err := b.pubsub.Subscribe(b.ctx, topic); err != nil {
			slog.Error("redis subscribe failed", "topic", topic, "error", err)
			b.subs[topic]--
			if b.subs[topic] == 0 {
				delete(b.subs, topic)
			}
		}
	}
}

func (b *redisBroker) MaybeUnsubscribe(workspaceID, channelID int64) {
	topic := ChannelTopic(workspaceID, channelID)

	b.mu.Lock()
	defer b.mu.Unlock()

	_, ok := b.subs[topic]
	if !ok {
		return
	}

	b.subs[topic]--
	if b.subs[topic] > 0 {
		return
	}

	delete(b.subs, topic)
	if err := b.pubsub.Unsubscribe(b.ctx, topic); err != nil {
		slog.Error("redis unsubscribe failed", "topic", topic, "error", err)
	}
}

func (b *redisBroker) readLoop() {
	for {
		if b.ctx.Err() != nil {
			return
		}

		ch := b.pubsub.Channel()
		for msg := range ch {
			if msg == nil {
				break
			}

			var event Event
			if err := json.Unmarshal([]byte(msg.Payload), &event); err != nil {
				slog.Warn("invalid redis event payload", "error", err)
				continue
			}

			b.hub.deliverLocal(event)
		}

		if b.ctx.Err() != nil {
			return
		}

		slog.Warn("redis pubsub channel closed, reconnecting")
		b.reconnect()
		time.Sleep(time.Second)
	}
}

func (b *redisBroker) reconnect() {
	b.mu.Lock()
	defer b.mu.Unlock()

	_ = b.pubsub.Close()

	topics := make([]string, 0, len(b.subs))
	for topic, count := range b.subs {
		if count > 0 {
			topics = append(topics, topic)
		}
	}

	if len(topics) == 0 {
		b.pubsub = b.client.Subscribe(b.ctx)
		return
	}

	b.pubsub = b.client.Subscribe(b.ctx, topics...)
}
