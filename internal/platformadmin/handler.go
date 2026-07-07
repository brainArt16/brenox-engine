package platformadmin

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/brainart16/brenox/internal/httperr"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) GetOverview(c *gin.Context) {
	overview, err := h.service.GetOverview(c.Request.Context())
	if err != nil {
		httperr.WriteInternal(c, err)
		return
	}
	c.JSON(http.StatusOK, overview)
}

func (h *Handler) ListUsers(c *gin.Context) {
	limit, offset := pagination(c)
	search := c.Query("search")

	users, err := h.service.ListUsers(c.Request.Context(), search, limit, offset)
	if err != nil {
		httperr.WriteInternal(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"users": users})
}

func (h *Handler) GetUser(c *gin.Context) {
	userID, ok := parseID(c, "id")
	if !ok {
		return
	}

	user, err := h.service.GetUser(c.Request.Context(), userID)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, user)
}

func (h *Handler) UpdateUser(c *gin.Context) {
	userID, ok := parseID(c, "id")
	if !ok {
		return
	}

	var req UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httperr.WriteJSON(c, http.StatusBadRequest, "invalid request body")
		return
	}

	actorID := c.MustGet("user_id").(int64)
	user, err := h.service.UpdateUser(c.Request.Context(), actorID, userID, req)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, user)
}

func (h *Handler) ListWorkspaces(c *gin.Context) {
	limit, offset := pagination(c)

	workspaces, err := h.service.ListWorkspaces(c.Request.Context(), limit, offset)
	if err != nil {
		httperr.WriteInternal(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"workspaces": workspaces})
}

func (h *Handler) GetWorkspace(c *gin.Context) {
	workspaceID, ok := parseID(c, "id")
	if !ok {
		return
	}

	workspace, err := h.service.GetWorkspace(c.Request.Context(), workspaceID)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, workspace)
}

func (h *Handler) ListApps(c *gin.Context) {
	limit, offset := pagination(c)

	apps, err := h.service.ListApps(c.Request.Context(), limit, offset)
	if err != nil {
		httperr.WriteInternal(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"apps": apps})
}

func (h *Handler) GetApp(c *gin.Context) {
	appID, ok := parseID(c, "app_id")
	if !ok {
		return
	}

	app, err := h.service.GetApp(c.Request.Context(), appID)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, app)
}

func (h *Handler) ListAppKeys(c *gin.Context) {
	appID, ok := parseID(c, "app_id")
	if !ok {
		return
	}

	keys, err := h.service.ListAppKeys(c.Request.Context(), appID)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"keys": keys})
}

func (h *Handler) RevokeAppKey(c *gin.Context) {
	appID, ok := parseID(c, "app_id")
	if !ok {
		return
	}
	keyID, ok := parseID(c, "key_id")
	if !ok {
		return
	}

	if err := h.service.RevokeAppKey(c.Request.Context(), appID, keyID); err != nil {
		writeError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *Handler) ListWorkspaceMembers(c *gin.Context) {
	workspaceID, ok := parseID(c, "id")
	if !ok {
		return
	}

	members, err := h.service.ListWorkspaceMembers(c.Request.Context(), workspaceID)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"members": members})
}

func (h *Handler) ListAuditLogs(c *gin.Context) {
	limit, offset := pagination(c)

	var userID *int64
	if raw := strings.TrimSpace(c.Query("user_id")); raw != "" {
		parsed, err := strconv.ParseInt(raw, 10, 64)
		if err != nil {
			httperr.WriteJSON(c, http.StatusBadRequest, "invalid user_id")
			return
		}
		userID = &parsed
	}

	logs, err := h.service.ListAuditLogs(c.Request.Context(), userID, c.Query("action"), limit, offset)
	if err != nil {
		httperr.WriteInternal(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"audit_logs": logs})
}

func pagination(c *gin.Context) (int32, int32) {
	limit := int32(50)
	offset := int32(0)

	if raw := c.Query("limit"); raw != "" {
		if parsed, err := strconv.Atoi(raw); err == nil && parsed > 0 && parsed <= 100 {
			limit = int32(parsed)
		}
	}
	if raw := c.Query("offset"); raw != "" {
		if parsed, err := strconv.Atoi(raw); err == nil && parsed >= 0 {
			offset = int32(parsed)
		}
	}
	return limit, offset
}

func parseID(c *gin.Context, param string) (int64, bool) {
	id, err := strconv.ParseInt(c.Param(param), 10, 64)
	if err != nil {
		httperr.WriteJSON(c, http.StatusBadRequest, "invalid id")
		return 0, false
	}
	return id, true
}

func writeError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, ErrNotFound):
		httperr.WriteJSON(c, http.StatusNotFound, httperr.ClientMessage(err, ErrNotFound))
	case errors.Is(err, ErrForbidden):
		httperr.WriteJSON(c, http.StatusForbidden, httperr.ClientMessage(err, ErrForbidden))
	case errors.Is(err, ErrInvalidRole):
		httperr.WriteJSON(c, http.StatusBadRequest, httperr.ClientMessage(err, ErrInvalidRole))
	case errors.Is(err, ErrSelfDemotion), errors.Is(err, ErrSelfSuspend):
		httperr.WriteJSON(c, http.StatusConflict, httperr.ClientMessage(err, ErrSelfDemotion, ErrSelfSuspend))
	case errors.Is(err, ErrInvalidRequest):
		httperr.WriteJSON(c, http.StatusBadRequest, httperr.ClientMessage(err, ErrInvalidRequest))
	case errors.Is(err, ErrKeyNotFound):
		httperr.WriteJSON(c, http.StatusNotFound, httperr.ClientMessage(err, ErrKeyNotFound))
	default:
		httperr.WriteInternal(c, err)
	}
}
