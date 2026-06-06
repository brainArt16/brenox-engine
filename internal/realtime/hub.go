package realtime

import (
	"context"
	"log/slog"
	"sync"
)

const (
	broadcastBufferSize = 256
	clientSendBuffer    = 16
)

// PresenceTracker updates distributed presence on connect/disconnect.
type PresenceTracker interface {
	Connect(ctx context.Context, userID, workspaceID, channelID int64) error
	Disconnect(ctx context.Context, userID, workspaceID, channelID int64) error
	Touch(ctx context.Context, userID int64) error
}

// Hub co-ordinates all connected clients and manages message broadcasting.
type Hub struct {
	mu sync.RWMutex
	cfg Config
	broker EventBroker
	presence PresenceTracker

	channels map[int64]map[*Client]bool

	register   chan *Client
	unregister chan *Client
	broadcast  chan Event
	done       chan struct{}

	userConnections map[int64]int
	ipConnections   map[string]int
}

func NewHub(cfg Config) *Hub {
	return &Hub{
		cfg:             cfg,
		channels:        make(map[int64]map[*Client]bool),
		register:        make(chan *Client),
		unregister:      make(chan *Client),
		broadcast:       make(chan Event, broadcastBufferSize),
		done:            make(chan struct{}),
		userConnections: make(map[int64]int),
		ipConnections:   make(map[string]int),
	}
}

// SetPresenceTracker attaches the distributed presence service.
func (h *Hub) SetPresenceTracker(tracker PresenceTracker) {
	h.presence = tracker
}

// SetBroker attaches the cross-node event broker (Redis or local-only).
func (h *Hub) SetBroker(broker EventBroker) {
	h.broker = broker
}

func (h *Hub) CanConnect(userID int64, remoteIP string) bool {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if h.cfg.MaxConnectionsPerUser > 0 && h.userConnections[userID] >= h.cfg.MaxConnectionsPerUser {
		return false
	}
	if h.cfg.MaxConnectionsPerIP > 0 && h.ipConnections[remoteIP] >= h.cfg.MaxConnectionsPerIP {
		return false
	}
	return true
}

// Publish sends an outbound event to all nodes via the broker.
func (h *Hub) Publish(event Event) {
	if event.EventID == "" {
		event = NewOutboundEvent(event.Type, event.WorkspaceID, event.ChannelID, event.Payload)
	}

	if h.broker != nil {
		h.broker.Publish(event)
		return
	}

	h.enqueueBroadcast(event)
}

func (h *Hub) deliverLocal(event Event) {
	h.enqueueBroadcast(event)
}

func (h *Hub) enqueueBroadcast(event Event) {
	go func() {
		select {
		case h.broadcast <- event:
		case <-h.done:
		default:
			slog.Warn("broadcast queue full, dropping event", "type", event.Type, "channel_id", event.ChannelID)
		}
	}()
}

func (h *Hub) Run() {
	for {
		select {
		case <-h.done:
			return

		case client := <-h.register:
			h.handleRegister(client)

		case client := <-h.unregister:
			h.handleUnregister(client)

		case event := <-h.broadcast:
			h.deliver(event)
		}
	}
}

func (h *Hub) handleRegister(client *Client) {
	h.mu.Lock()
	if h.channels[client.channelID] == nil {
		h.channels[client.channelID] = make(map[*Client]bool)
	}
	firstOnChannel := len(h.channels[client.channelID]) == 0
	h.channels[client.channelID][client] = true

	h.userConnections[client.userID]++
	h.ipConnections[client.remoteIP]++
	h.mu.Unlock()

	if firstOnChannel && h.broker != nil {
		h.broker.EnsureSubscribed(client.workspaceID, client.channelID)
	}

	if h.presence != nil {
		if err := h.presence.Connect(context.Background(), client.userID, client.workspaceID, client.channelID); err != nil {
			slog.Error("presence connect failed", "user_id", client.userID, "error", err)
		}
	}

	slog.Info("websocket connected",
		"user_id", client.userID,
		"workspace_id", client.workspaceID,
		"channel_id", client.channelID,
		"remote_ip", client.remoteIP,
	)
}

func (h *Hub) handleUnregister(client *Client) {
	h.mu.Lock()
	_, ok := h.channels[client.channelID][client]
	if !ok {
		h.mu.Unlock()
		return
	}

	delete(h.channels[client.channelID], client)
	lastOnChannel := len(h.channels[client.channelID]) == 0

	h.userConnections[client.userID]--
	h.ipConnections[client.remoteIP]--
	if h.userConnections[client.userID] <= 0 {
		delete(h.userConnections, client.userID)
	}
	if h.ipConnections[client.remoteIP] <= 0 {
		delete(h.ipConnections, client.remoteIP)
	}
	h.mu.Unlock()

	if lastOnChannel && h.broker != nil {
		h.broker.MaybeUnsubscribe(client.workspaceID, client.channelID)
	}

	close(client.send)

	if h.presence != nil {
		if err := h.presence.Disconnect(context.Background(), client.userID, client.workspaceID, client.channelID); err != nil {
			slog.Error("presence disconnect failed", "user_id", client.userID, "error", err)
		}
	}

	slog.Info("websocket disconnected",
		"user_id", client.userID,
		"workspace_id", client.workspaceID,
		"channel_id", client.channelID,
	)
}

func (h *Hub) deliver(event Event) {
	h.mu.RLock()
	clients := h.channels[event.ChannelID]
	targets := make([]*Client, 0, len(clients))
	for client := range clients {
		targets = append(targets, client)
	}
	h.mu.RUnlock()

	for _, client := range targets {
		select {
		case client.send <- event:
		default:
			slog.Warn("client send buffer full, disconnecting slow client",
				"user_id", client.userID,
				"channel_id", client.channelID,
			)
			go func(c *Client) {
				h.unregister <- c
			}(client)
		}
	}
}

// Shutdown stops the hub loop and closes active connections.
func (h *Hub) Shutdown() {
	select {
	case <-h.done:
		return
	default:
		close(h.done)
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	for channelID, clients := range h.channels {
		for client := range clients {
			if h.presence != nil {
				_ = h.presence.Disconnect(context.Background(), client.userID, client.workspaceID, client.channelID)
			}
			close(client.send)
			if client.conn != nil {
				_ = client.conn.Close()
			}
			delete(clients, client)
		}
		delete(h.channels, channelID)
	}
}

func (h *Hub) touchPresence(userID int64) {
	if h.presence == nil {
		return
	}
	if err := h.presence.Touch(context.Background(), userID); err != nil {
		slog.Warn("presence touch failed", "user_id", userID, "error", err)
	}
}
