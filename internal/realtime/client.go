package realtime

import (
	"log"

	"github.com/gorilla/websocket"
)

// Client represents one connected user
type Client struct {
	// WebSocket connection.
	conn *websocket.Conn

	// User identity
	userID int64

	// Reference to the hub.
	hub *Hub

	// outgoing message queue: This is Go channel
	send chan Message
}

// Create new hub instance.
func NewHub() *Hub {
	return &Hub{

		clients:    make(map[*Client]bool),

		register:   make(chan *Client),

		unregister: make(chan *Client),

		broadcast:  make(chan Message),

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
			h.clients[client] = true

		// When a client unregisters, remove it from the clients map and close its send channel.
		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}

		// When a message is broadcast, send it to all connected clients.
		case message := <-h.broadcast:
			for client := range h.clients {
				client.send <- message
			}
		}
	}

}



// Read messages from websocket.
func (c *Client) readPump() {

	defer func() {

		// Remove client on disconnect.
		c.hub.unregister <- c

		c.conn.Close()
	}()

	for {

		var message Message

		// Read JSON websocket message. 
		// The message is expected to have the structure defined by the Message struct.
		err := c.conn.ReadJSON(
			&message,
		)

		if err != nil {

			log.Println(err)

			break
		}

		// Broadcast to hub.
		c.hub.broadcast <- message
	}
}


// Write outbound messages to the WebSocket connection. This runs in a separate goroutine for each client.
func (c *Client) writePump() {
	defer c.conn.Close()
	
	for {
		// wait for outbound message
		message, ok := <-c.send

		if !ok {
			// Hub closed the channel, close WebSocket connection.
			log.Println("Hub closed the channel, closing WebSocket connection.")
			return
		}

		err := c.conn.WriteJSON(message)

		if err != nil {
			log.Printf("Error writing message to WebSocket: %v", err)
			return
		}
	}
}