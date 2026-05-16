package realtime

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

// WebSocket upgrader. Converts HTTP connection into WebSocket connection.
var upgrader = websocket.Upgrader{

	CheckOrigin: func(r *http.Request) bool {

		/*
			Allow all origins for now.

			Later:
			restrict domains properly.
		*/

		return true
	},
}

type Handler struct {
	hub *Hub
}

func NewHandler(
	hub *Hub,
) *Handler {

	return &Handler{
		hub: hub,
	}
}

// WebSocket endpoint.
func (h *Handler) HandleWebSocket(
	c *gin.Context,
) {

	// Upgrade HTTP → WebSocket
	conn, err := upgrader.Upgrade(
		c.Writer,
		c.Request,
		nil,
	)

	if err != nil {
		return
	}

	userID := c.MustGet("user_id").(int64)

	client := &Client{
		conn:   conn,
		userID: userID,
		hub:    h.hub,
		send:   make(chan Message),
	}

	// Register client.
	h.hub.register <- client

	// Start concurrent loops.go keyword starts goroutine.
	go client.writePump()

	go client.readPump()
}