package realtime

import "github.com/gin-gonic/gin"


// 	Return currently online users.
func (h *Handler) GetPresence(
	c *gin.Context,
) {

	onlineUsers := []int64{}

	for userID := range h.hub.onlineUsers {

		onlineUsers = append(
			onlineUsers,
			userID,
		)
	}

	c.JSON(
		200,
		gin.H{
			"online_users": onlineUsers,
		},
	)
}