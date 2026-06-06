package realtime

import (
	"net"
	"net/http"
	"strconv"

	"github.com/brainart16/brenox/internal/channels"
	"github.com/brainart16/brenox/internal/chat"
	"github.com/brainart16/brenox/internal/calls"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

type Handler struct {
	hub      *Hub
	chat     *chat.Service
	channels *channels.Service
	calls    *calls.Service
	cfg      Config
	upgrader websocket.Upgrader
}

func NewHandler(hub *Hub, chatService *chat.Service, channelsService *channels.Service, callsService *calls.Service, cfg Config) *Handler {
	return &Handler{
		hub:      hub,
		chat:     chatService,
		channels: channelsService,
		calls:    callsService,
		cfg:      cfg,
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return cfg.originAllowed(r.Header.Get("Origin"))
			},
		},
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
	remoteIP := clientIP(c)

	if !h.hub.CanConnect(userID, remoteIP) {
		c.JSON(http.StatusTooManyRequests, gin.H{"error": "connection limit exceeded"})
		return
	}

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

	conn, err := h.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}

	client := &Client{
		conn:        conn,
		userID:      userID,
		workspaceID: workspaceID,
		channelID:   channelID,
		remoteIP:    remoteIP,
		hub:         h.hub,
		chat:        h.chat,
		calls:       h.calls,
		send:        make(chan Event, clientSendBuffer),
	}

	h.hub.register <- client

	go client.writePump()
	go client.readPump()
}

func clientIP(c *gin.Context) string {
	ip := c.ClientIP()
	if host, _, err := net.SplitHostPort(ip); err == nil {
		return host
	}
	return ip
}
