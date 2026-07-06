package notifications

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/brainart16/brenox/internal/httperr"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) List(c *gin.Context) {
	userID := c.MustGet("user_id").(int64)
	limit, _ := strconv.ParseInt(c.DefaultQuery("limit", "50"), 10, 32)
	offset, _ := strconv.ParseInt(c.DefaultQuery("offset", "0"), 10, 32)

	items, err := h.service.List(c.Request.Context(), userID, int32(limit), int32(offset))
	if err != nil {
		httperr.WriteInternal(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"notifications": items})
}

func (h *Handler) MarkRead(c *gin.Context) {
	userID := c.MustGet("user_id").(int64)
	notificationID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid notification id"})
		return
	}

	item, err := h.service.MarkRead(c.Request.Context(), userID, notificationID)
	if err != nil {
		writeError(c, err)
		return
	}

	c.JSON(http.StatusOK, item)
}

func (h *Handler) MarkAllRead(c *gin.Context) {
	userID := c.MustGet("user_id").(int64)

	count, err := h.service.MarkAllRead(c.Request.Context(), userID)
	if err != nil {
		httperr.WriteInternal(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"marked_read": count})
}

func writeError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, ErrNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": httperr.Sanitize(err.Error())})
	case errors.Is(err, ErrInvalidType):
		c.JSON(http.StatusBadRequest, gin.H{"error": httperr.Sanitize(err.Error())})
	default:
		httperr.WriteInternal(c, err)
	}
}
