package workspaces

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

func (h *Handler) CreateWorkspace(c *gin.Context) {
	var req CreateWorkspaceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	userID := c.MustGet("user_id").(int64)

	workspace, err := h.service.CreateWorkspace(c.Request.Context(), userID, req)
	if err != nil {
		writeWorkspaceError(c, err)
		return
	}

	c.JSON(http.StatusCreated, ToWorkspaceResponse(*workspace))
}

func (h *Handler) ListWorkspaces(c *gin.Context) {
	userID := c.MustGet("user_id").(int64)

	rows, err := h.service.ListWorkspaces(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	items := make([]WorkspaceResponse, 0, len(rows))
	for _, row := range rows {
		items = append(items, ToWorkspaceListItem(row))
	}

	c.JSON(http.StatusOK, gin.H{"workspaces": items})
}

func (h *Handler) GetWorkspace(c *gin.Context) {
	workspaceID, err := parseWorkspaceID(c)
	if err != nil {
		return
	}

	userID := c.MustGet("user_id").(int64)

	workspace, err := h.service.GetWorkspace(c.Request.Context(), workspaceID, userID)
	if err != nil {
		writeWorkspaceError(c, err)
		return
	}

	resp := ToWorkspaceResponse(*workspace)
	if role, err := h.service.MemberRole(c.Request.Context(), workspaceID, userID); err == nil {
		resp.Role = role
	}

	c.JSON(http.StatusOK, resp)
}

func (h *Handler) ListMembers(c *gin.Context) {
	workspaceID, err := parseWorkspaceID(c)
	if err != nil {
		return
	}

	userID := c.MustGet("user_id").(int64)

	members, err := h.service.ListMembers(c.Request.Context(), workspaceID, userID)
	if err != nil {
		writeWorkspaceError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"members": members})
}

func (h *Handler) AddMember(c *gin.Context) {
	workspaceID, err := parseWorkspaceID(c)
	if err != nil {
		return
	}

	var req AddMemberRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	actorID := c.MustGet("user_id").(int64)

	member, err := h.service.AddMember(c.Request.Context(), workspaceID, actorID, req)
	if err != nil {
		writeWorkspaceError(c, err)
		return
	}

	c.JSON(http.StatusCreated, member)
}

func (h *Handler) RemoveMember(c *gin.Context) {
	workspaceID, err := parseWorkspaceID(c)
	if err != nil {
		return
	}

	targetUserID, err := strconv.ParseInt(c.Param("user_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}

	actorID := c.MustGet("user_id").(int64)

	if err := h.service.RemoveMember(c.Request.Context(), workspaceID, actorID, targetUserID); err != nil {
		writeWorkspaceError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *Handler) UpdateMemberRole(c *gin.Context) {
	workspaceID, err := parseWorkspaceID(c)
	if err != nil {
		return
	}

	targetUserID, err := strconv.ParseInt(c.Param("user_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}

	var req UpdateMemberRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	actorID := c.MustGet("user_id").(int64)

	member, err := h.service.UpdateMemberRole(c.Request.Context(), workspaceID, actorID, targetUserID, req)
	if err != nil {
		writeWorkspaceError(c, err)
		return
	}

	c.JSON(http.StatusOK, member)
}

func parseWorkspaceID(c *gin.Context) (int64, error) {
	workspaceID, err := strconv.ParseInt(c.Param("workspace_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid workspace id"})
		return 0, err
	}
	return workspaceID, nil
}

func writeWorkspaceError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, ErrNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
	case errors.Is(err, ErrNotMember):
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
	case errors.Is(err, ErrForbidden):
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
	case errors.Is(err, ErrInvalidRole), errors.Is(err, ErrCannotModifyOwner):
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
	case errors.Is(err, ErrUserNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
	case errors.Is(err, ErrAlreadyMember):
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
	case errors.Is(err, ErrSlugTaken), errors.Is(err, ErrInvalidSlug):
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
	default:
		if err.Error() == "workspace name is required" {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if err.Error() == "email is required" {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
	}
}
