package realtime

import (
	"log"
	"time"

	"github.com/gorilla/websocket"
)

// Client represents one connected user
type Client struct {
	// WebSocket connection.
	conn *websocket.Conn

	// User identity
	userID int64

	// Channel the client is subscribed to. This allows us to manage clients by channel and broadcast messages to specific channels.
	channelID int64

	// Reference to the hub.
	hub *Hub

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
					Type:      "presence.online",
					ChannelID: client.channelID,
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
					Type:      "presence.offline",
					ChannelID: client.channelID,
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

		// Ensure the event has the correct channel ID before broadcasting.
		event.ChannelID = c.channelID

		// Attach authenticated sender
		if payloadMap, ok := event.Payload.(map[string]any); ok {
			payloadMap["sender_id"] = c.userID
			event.Payload = payloadMap
		}

		// Broadcast to hub.
		c.hub.broadcast <- event
	}
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