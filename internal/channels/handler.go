package channels

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service *Service
}

func NewHandler(
	service *Service,
) *Handler {

	return &Handler{
		service: service,
	}
}

// Create channel endpoint.
func (h *Handler) CreateChannel(
	c *gin.Context,
) {

	var req CreateChannelRequest

	err := c.ShouldBindJSON(&req)

	if err != nil {

		c.JSON(
			http.StatusBadRequest,
			gin.H{
				"error": "invalid request",
			},
		)

		return
	}

	// Get authenticated user ID from middleware context.
	userID := c.MustGet("user_id").(int64)

	channel, err := h.service.CreateChannel(
		c.Request.Context(),
		userID,
		req,
	)

	if err != nil {

		c.JSON(
			http.StatusInternalServerError,
			gin.H{
				"error": err.Error(),
			},
		)

		return
	}

	c.JSON(
		http.StatusCreated,
		channel,
	)
}

// List user channels.
func (h *Handler) GetChannels(
	c *gin.Context,
) {

	userID := c.MustGet("user_id").(int64)

	channels, err := h.service.GetChannels(
		c.Request.Context(),
		userID,
	)

	if err != nil {

		c.JSON(
			http.StatusInternalServerError,
			gin.H{
				"error": err.Error(),
			},
		)

		return
	}

	c.JSON(
		http.StatusOK,
		channels,
	)
}