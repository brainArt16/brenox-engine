package realtime

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/brainart16/brenox/internal/chat"
	"github.com/gorilla/websocket"
)

// Client represents one connected user
type Client struct {
	// WebSocket connection.
	conn *websocket.Conn

	// User identity
	userID int64

	// Channel the client is subscribed to.
	channelID int64

	workspaceID int64

	// Reference to the hub.
	hub *Hub

	chat *chat.Service

	// outgoing message queue: This is Go channel
	send chan Event
}

// Create new hub instance.
func NewHub() *Hub {
	return &Hub{

		channels:    make(map[int64]map[*Client]bool),

		register:   make(chan *Client),

		unregister: make(chan *Client),

		broadcast:  make(chan Event),

		onlineUsers: make(map[int64]int),

	}
}


// Run starts hub event loop. It listens for register, unregister, and broadcast events and processes them accordingly.
func (h *Hub) Run() {

	// infinity loop to keep the hub running
	for {

		// Select statement to handle different events: register, unregister, and broadcast.
		select {

		// When a new client registers, add it to the clients map.
		case client := <-h.register:

			h.mu.Lock()
			if h.channels[client.channelID] == nil {
				h.channels[client.channelID] = make(map[*Client]bool)
			}
			h.channels[client.channelID][client] = true

			h.onlineUsers[client.userID]++
			firstConnection := h.onlineUsers[client.userID] == 1
			h.mu.Unlock()

			// First connection only — additional tabs do not re-emit online.
			if firstConnection {
				h.broadcast <- Event{
					Type:        "presence.online",
					WorkspaceID: client.workspaceID,
					ChannelID:   client.channelID,
					Payload: map[string]any{
						"user_id": client.userID,
					},
				}
			}

		// When a client unregisters, remove it from the clients map and close its send channel.
		case client := <-h.unregister:

			h.mu.Lock()
			_, ok := h.channels[client.channelID][client]
			if !ok {
				h.mu.Unlock()
				continue
			}

			delete(h.channels[client.channelID], client)

			h.onlineUsers[client.userID]--
			lastConnection := h.onlineUsers[client.userID] <= 0
			if lastConnection {
				delete(h.onlineUsers, client.userID)
			}
			h.mu.Unlock()

			close(client.send)

			// Last connection only — other tabs stay online.
			if lastConnection {
				h.broadcast <- Event{
					Type:        "presence.offline",
					WorkspaceID: client.workspaceID,
					ChannelID:   client.channelID,
					Payload: map[string]any{
						"user_id": client.userID,
					},
				}
			}

		// When a message is broadcast, send it to all clients subscribed to the relevant channel.
		case event := <-h.broadcast:

			h.mu.RLock()
			clients := h.channels[event.ChannelID]
			targets := make([]*Client, 0, len(clients))
			for client := range clients {
				targets = append(targets, client)
			}
			h.mu.RUnlock()

			for _, client := range targets {
				client.send <- event
			}
		}
	}

}


// Read event (messages) from websocket.
func (c *Client) readPump() {

	defer func() {

		// Remove client on disconnect.
		c.hub.unregister <- c

		c.conn.Close()
	}()

	for {

		var event Event

		// Read JSON websocket event. 
		// The event is expected to have the structure defined by the Event struct.
		err := c.conn.ReadJSON(
			&event,
		)

		if err != nil {
			log.Println(err)
			break
		}

		event.ChannelID = c.channelID

		switch event.Type {
		case "message.send":
			c.handleMessageSend(event)
		default:
			// Ignore unknown inbound event types from clients.
			log.Printf("ignored websocket event type: %s", event.Type)
		}
	}
}

func (c *Client) handleMessageSend(event Event) {
	content, ok := parseMessageContent(event.Payload)
	if !ok {
		c.sendClientError("invalid message payload")
		return
	}

	message, err := c.chat.SendMessage(
		context.Background(),
		c.workspaceID,
		c.channelID,
		c.userID,
		content,
	)
	if err != nil {
		switch {
		case errors.Is(err, chat.ErrNotMember):
			c.sendClientError("not a channel member")
		case errors.Is(err, chat.ErrNotWorkspaceMember):
			c.sendClientError("not a workspace member")
		case errors.Is(err, chat.ErrChannelNotFound):
			c.sendClientError("channel not found")
		case errors.Is(err, chat.ErrForbidden):
			c.sendClientError("permission denied")
		case errors.Is(err, chat.ErrEmptyContent), errors.Is(err, chat.ErrMessageTooLong):
			c.sendClientError(err.Error())
		default:
			log.Printf("message.send failed: %v", err)
			c.sendClientError("failed to send message")
		}
		return
	}

	c.hub.broadcast <- Event{
		Type:        "message.new",
		WorkspaceID: c.workspaceID,
		ChannelID:   c.channelID,
		Payload:     chat.MessageNewPayload(*message),
	}
}

func (c *Client) sendClientError(message string) {
	c.send <- Event{
		Type:        "error",
		WorkspaceID: c.workspaceID,
		ChannelID:   c.channelID,
		Payload: map[string]any{
			"message": message,
		},
	}
}

func parseMessageContent(payload any) (string, bool) {
	payloadMap, ok := payload.(map[string]any)
	if !ok {
		return "", false
	}

	content, ok := payloadMap["content"].(string)
	if !ok {
		return "", false
	}

	return content, true
}


// Write outbound messages to the WebSocket connection. This runs in a separate goroutine for each client.
func (c *Client) writePump() {
	defer c.conn.Close()

	ticker := time.NewTicker(pingPeriod)
	defer ticker.Stop()
	
	for {

		select {
		// wait for outbound message
		case event, ok := <-c.send:

			if !ok {
				// Hub closed the channel, close WebSocket connection.
				log.Println("Hub closed the channel, closing WebSocket connection.")
				return
			}

			err := c.conn.WriteJSON(event)

			if err != nil {
				log.Printf("Error writing event to WebSocket: %v", err)
				return
			}
		
		case <-ticker.C:
			
			// Send ping to keep the connection alive.
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				log.Printf("Error sending ping: %v", err)
				return
			}
		}
	}
}