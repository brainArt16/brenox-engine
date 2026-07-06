package chat

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

func (h *Handler) CreateMessage(c *gin.Context) {
	workspaceID, err := parseWorkspaceID(c)
	if err != nil {
		return
	}

	channelID, err := parseChannelID(c)
	if err != nil {
		return
	}

	var req CreateMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	userID := c.MustGet("user_id").(int64)

	message, err := h.service.SendMessage(
		c.Request.Context(),
		workspaceID,
		channelID,
		userID,
		req.Content,
		req.ReplyToMessageID,
		req.Attachments,
	)
	if err != nil {
		writeServiceError(c, err)
		return
	}

	c.JSON(http.StatusCreated, ToMessageResponse(*message))
}

func (h *Handler) GetMessages(c *gin.Context) {
	workspaceID, err := parseWorkspaceID(c)
	if err != nil {
		return
	}

	channelID, err := parseChannelID(c)
	if err != nil {
		return
	}

	limit, _ := strconv.ParseInt(c.DefaultQuery("limit", "50"), 10, 32)
	offset, _ := strconv.ParseInt(c.DefaultQuery("offset", "0"), 10, 32)

	userID := c.MustGet("user_id").(int64)

	rows, err := h.service.ListMessages(
		c.Request.Context(),
		workspaceID,
		channelID,
		userID,
		int32(limit),
		int32(offset),
	)
	if err != nil {
		writeServiceError(c, err)
		return
	}

	items := make([]MessageListItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, ToMessageListItem(row))
	}

	c.JSON(http.StatusOK, gin.H{"messages": items})
}

func parseWorkspaceID(c *gin.Context) (int64, error) {
	workspaceID, err := strconv.ParseInt(c.Param("workspace_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid workspace id"})
		return 0, err
	}
	return workspaceID, nil
}

func parseChannelID(c *gin.Context) (int64, error) {
	channelID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid channel id"})
		return 0, err
	}
	return channelID, nil
}

func writeServiceError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, ErrChannelNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": httperr.Sanitize(err.Error())})
	case errors.Is(err, ErrNotMember), errors.Is(err, ErrNotWorkspaceMember), errors.Is(err, ErrForbidden):
		c.JSON(http.StatusForbidden, gin.H{"error": httperr.Sanitize(err.Error())})
	case errors.Is(err, ErrEmptyContent), errors.Is(err, ErrMessageTooLong):
		c.JSON(http.StatusBadRequest, gin.H{"error": httperr.Sanitize(err.Error())})
	default:
		httperr.WriteInternal(c, err)
	}
}
