package channels

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/brainart16/brenox/internal/httperr"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	service     *Service
	broadcaster Broadcaster
}

func NewHandler(service *Service, broadcaster Broadcaster) *Handler {
	return &Handler{
		service:     service,
		broadcaster: broadcaster,
	}
}

func (h *Handler) CreateChannel(c *gin.Context) {
	workspaceID, err := parseWorkspaceID(c)
	if err != nil {
		return
	}

	var req CreateChannelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	userID := c.MustGet("user_id").(int64)

	channel, err := h.service.CreateChannel(c.Request.Context(), workspaceID, userID, req)
	if err != nil {
		writeChannelError(c, err)
		return
	}

	c.JSON(http.StatusCreated, channel)
}

func (h *Handler) GetChannels(c *gin.Context) {
	workspaceID, err := parseWorkspaceID(c)
	if err != nil {
		return
	}

	userID := c.MustGet("user_id").(int64)

	channels, err := h.service.GetChannels(c.Request.Context(), workspaceID, userID)
	if err != nil {
		writeChannelError(c, err)
		return
	}

	c.JSON(http.StatusOK, channels)
}

func (h *Handler) JoinChannel(c *gin.Context) {
	workspaceID, err := parseWorkspaceID(c)
	if err != nil {
		return
	}

	channelID, err := parseChannelID(c)
	if err != nil {
		return
	}

	userID := c.MustGet("user_id").(int64)

	if err := h.service.JoinChannel(c.Request.Context(), workspaceID, channelID, userID); err != nil {
		writeChannelError(c, err)
		return
	}

	if h.broadcaster != nil {
		h.broadcaster.BroadcastMemberJoined(workspaceID, channelID, userID)
	}

	c.JSON(http.StatusOK, gin.H{
		"workspace_id": workspaceID,
		"channel_id":   channelID,
		"user_id":      userID,
		"status":       "joined",
	})
}

func (h *Handler) LeaveChannel(c *gin.Context) {
	workspaceID, err := parseWorkspaceID(c)
	if err != nil {
		return
	}

	channelID, err := parseChannelID(c)
	if err != nil {
		return
	}

	userID := c.MustGet("user_id").(int64)

	if err := h.service.LeaveChannel(c.Request.Context(), workspaceID, channelID, userID); err != nil {
		writeChannelError(c, err)
		return
	}

	if h.broadcaster != nil {
		h.broadcaster.BroadcastMemberLeft(workspaceID, channelID, userID)
	}

	c.JSON(http.StatusOK, gin.H{
		"workspace_id": workspaceID,
		"channel_id":   channelID,
		"user_id":      userID,
		"status":       "left",
	})
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

func writeChannelError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, ErrChannelNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": httperr.Sanitize(err.Error())})
	case errors.Is(err, ErrAlreadyMember):
		c.JSON(http.StatusConflict, gin.H{"error": httperr.Sanitize(err.Error())})
	case errors.Is(err, ErrNotMember), errors.Is(err, ErrNotWorkspaceMember):
		c.JSON(http.StatusForbidden, gin.H{"error": httperr.Sanitize(err.Error())})
	case errors.Is(err, ErrOwnerCannotLeave), errors.Is(err, ErrDuplicateChannelName):
		c.JSON(http.StatusConflict, gin.H{"error": httperr.Sanitize(err.Error())})
	case errors.Is(err, ErrForbidden):
		c.JSON(http.StatusForbidden, gin.H{"error": httperr.Sanitize(err.Error())})
	default:
		httperr.WriteInternal(c, err)
	}
}
