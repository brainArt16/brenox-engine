package calls

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

func (h *Handler) InitiateCall(c *gin.Context) {
	workspaceID, err := parseWorkspaceID(c)
	if err != nil {
		return
	}
	channelID, err := parseChannelID(c)
	if err != nil {
		return
	}

	userID := c.MustGet("user_id").(int64)

	var req InitiateCallRequest
	_ = c.ShouldBindJSON(&req)

	call, err := h.service.InitiateCall(c.Request.Context(), workspaceID, channelID, userID, req.Mode)
	if err != nil {
		writeError(c, err)
		return
	}

	c.JSON(http.StatusCreated, call)
}

func (h *Handler) JoinCall(c *gin.Context) {
	callID, err := parseCallID(c)
	if err != nil {
		return
	}

	userID := c.MustGet("user_id").(int64)
	call, err := h.service.JoinCall(c.Request.Context(), callID, userID)
	if err != nil {
		writeError(c, err)
		return
	}

	c.JSON(http.StatusOK, call)
}

func (h *Handler) LeaveCall(c *gin.Context) {
	callID, err := parseCallID(c)
	if err != nil {
		return
	}

	userID := c.MustGet("user_id").(int64)
	call, err := h.service.LeaveCall(c.Request.Context(), callID, userID)
	if err != nil {
		writeError(c, err)
		return
	}

	c.JSON(http.StatusOK, call)
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

func parseCallID(c *gin.Context) (int64, error) {
	callID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid call id"})
		return 0, err
	}
	return callID, nil
}

func writeError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, ErrNotFound), errors.Is(err, ErrChannelNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": httperr.Sanitize(err.Error())})
	case errors.Is(err, ErrNotMember), errors.Is(err, ErrNotParticipant), errors.Is(err, ErrCallEnded):
		c.JSON(http.StatusForbidden, gin.H{"error": httperr.Sanitize(err.Error())})
	case errors.Is(err, ErrCallAlreadyActive):
		c.JSON(http.StatusConflict, gin.H{"error": httperr.Sanitize(err.Error())})
	case errors.Is(err, ErrCallFull), errors.Is(err, ErrRecordingActive):
		c.JSON(http.StatusConflict, gin.H{"error": httperr.Sanitize(err.Error())})
	case errors.Is(err, ErrInvalidMode):
		c.JSON(http.StatusBadRequest, gin.H{"error": httperr.Sanitize(err.Error())})
	default:
		httperr.WriteInternal(c, err)
	}
}
