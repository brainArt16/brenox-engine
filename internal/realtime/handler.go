package realtime

import (
	"net/http"
	"strconv"

	"github.com/brainart16/brenox/internal/channels"
	"github.com/brainart16/brenox/internal/chat"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type Handler struct {
	hub      *Hub
	chat     *chat.Service
	channels *channels.Service
}

func NewHandler(hub *Hub, chatService *chat.Service, channelsService *channels.Service) *Handler {
	return &Handler{
		hub:      hub,
		chat:     chatService,
		channels: channelsService,
	}
}

func (h *Handler) HandleWebSocket(c *gin.Context) {
	workspaceID, err := strconv.ParseInt(c.Query("workspace_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid workspace_id"})
		return
	}

	channelID, err := strconv.ParseInt(c.Query("channel_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid channel_id"})
		return
	}

	userID := c.MustGet("user_id").(int64)

	isWorkspaceMember, err := h.channels.IsWorkspaceMember(c.Request.Context(), workspaceID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	if !isWorkspaceMember {
		c.JSON(http.StatusForbidden, gin.H{"error": "not a workspace member"})
		return
	}

	if _, err := h.channels.GetChannelInWorkspace(c.Request.Context(), workspaceID, channelID); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "channel not found"})
		return
	}

	isMember, err := h.channels.IsMember(c.Request.Context(), channelID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	if !isMember {
		c.JSON(http.StatusForbidden, gin.H{"error": "not a channel member"})
		return
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}

	client := &Client{
		conn:        conn,
		userID:      userID,
		workspaceID: workspaceID,
		channelID:   channelID,
		hub:         h.hub,
		chat:        h.chat,
		send:        make(chan Event, 16),
	}

	h.hub.register <- client

	go client.writePump()
	go client.readPump()
}
