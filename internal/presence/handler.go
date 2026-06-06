package presence

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) GetGlobalPresence(c *gin.Context) {
	items, err := h.service.ListOnline(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"online_users": items})
}

func (h *Handler) GetWorkspacePresence(c *gin.Context) {
	workspaceID, err := parseWorkspaceID(c)
	if err != nil {
		return
	}

	requesterID := c.MustGet("user_id").(int64)
	items, err := h.service.ListWorkspacePresence(c.Request.Context(), workspaceID, requesterID)
	if err != nil {
		writePresenceError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"online_users": items})
}

type updateStatusRequest struct {
	Status string `json:"status" binding:"required"`
}

func (h *Handler) UpdateMyStatus(c *gin.Context) {
	var req updateStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	userID := c.MustGet("user_id").(int64)
	presence, err := h.service.UpdateStatus(c.Request.Context(), userID, req.Status)
	if err != nil {
		writePresenceError(c, err)
		return
	}

	c.JSON(http.StatusOK, presence)
}

func parseWorkspaceID(c *gin.Context) (int64, error) {
	workspaceID, err := strconv.ParseInt(c.Param("workspace_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid workspace id"})
		return 0, err
	}
	return workspaceID, nil
}

func writePresenceError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, ErrInvalidStatus):
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	case errors.Is(err, ErrNotWorkspaceMember):
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
	}
}
