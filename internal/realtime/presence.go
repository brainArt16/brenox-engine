package realtime

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// GetPresence returns user IDs with at least one active WebSocket connection.
func (h *Handler) GetPresence(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"online_users": h.hub.OnlineUserIDs(),
	})
}
