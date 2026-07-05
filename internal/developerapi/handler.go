package developerapi

import (
	"errors"
	"net/http"
	"strconv"

	db "github.com/brainart16/brenox/internal/db"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) CreateSession(c *gin.Context) {
	var req CreateSessionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	app := appFromContext(c)
	session, err := h.service.CreateSession(c.Request.Context(), app, req)
	if err != nil {
		writeError(c, err)
		return
	}

	c.JSON(http.StatusOK, session)
}

func (h *Handler) ProvisionUser(c *gin.Context) {
	var req ProvisionUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	app := appFromContext(c)
	user, err := h.service.ProvisionUser(c.Request.Context(), app, req)
	if err != nil {
		writeError(c, err)
		return
	}

	c.JSON(http.StatusCreated, user)
}

func (h *Handler) CreateChannel(c *gin.Context) {
	var req CreateChannelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	app := appFromContext(c)
	channel, err := h.service.CreateChannel(c.Request.Context(), app, req)
	if err != nil {
		writeError(c, err)
		return
	}

	c.JSON(http.StatusCreated, channel)
}

func (h *Handler) SendMessage(c *gin.Context) {
	var req SendMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	app := appFromContext(c)
	message, err := h.service.SendMessage(c.Request.Context(), app, req)
	if err != nil {
		writeError(c, err)
		return
	}

	c.JSON(http.StatusCreated, message)
}

func (h *Handler) ListMessages(c *gin.Context) {
	channelID, err := strconv.ParseInt(c.Query("channel_id"), 10, 64)
	if err != nil || channelID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "channel_id is required"})
		return
	}

	limit, _ := strconv.ParseInt(c.DefaultQuery("limit", "50"), 10, 32)
	offset, _ := strconv.ParseInt(c.DefaultQuery("offset", "0"), 10, 32)

	app := appFromContext(c)
	messages, err := h.service.ListMessages(c.Request.Context(), app, channelID, int32(limit), int32(offset))
	if err != nil {
		writeError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"messages": messages})
}

func appFromContext(c *gin.Context) db.App {
	return c.MustGet("app").(db.App)
}

func writeError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, ErrChannelNotFound), errors.Is(err, ErrUserNotFound), errors.Is(err, ErrNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
	case errors.Is(err, ErrInvalidRequest), errors.Is(err, ErrEmptyContent):
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	default:
		if err != nil && err.Error() == "channel name already exists" {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
	}
}
