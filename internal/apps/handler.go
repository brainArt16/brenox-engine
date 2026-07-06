package apps

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

func (h *Handler) CreateApp(c *gin.Context) {
	var req CreateAppRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	userID := c.MustGet("user_id").(int64)
	app, err := h.service.CreateApp(c.Request.Context(), userID, req)
	if err != nil {
		writeError(c, err)
		return
	}

	c.JSON(http.StatusCreated, app)
}

func (h *Handler) ListApps(c *gin.Context) {
	userID := c.MustGet("user_id").(int64)
	apps, err := h.service.ListApps(c.Request.Context(), userID)
	if err != nil {
		writeError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"apps": apps})
}

func (h *Handler) GetApp(c *gin.Context) {
	appID, err := parseAppID(c)
	if err != nil {
		return
	}

	userID := c.MustGet("user_id").(int64)
	app, err := h.service.GetAppForOwner(c.Request.Context(), appID, userID)
	if err != nil {
		writeError(c, err)
		return
	}

	c.JSON(http.StatusOK, app)
}

func (h *Handler) CreateAPIKey(c *gin.Context) {
	appID, err := parseAppID(c)
	if err != nil {
		return
	}

	var req CreateAPIKeyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	userID := c.MustGet("user_id").(int64)
	key, err := h.service.CreateAPIKey(c.Request.Context(), appID, userID, req)
	if err != nil {
		writeError(c, err)
		return
	}

	c.JSON(http.StatusCreated, key)
}

func (h *Handler) ListAPIKeys(c *gin.Context) {
	appID, err := parseAppID(c)
	if err != nil {
		return
	}

	userID := c.MustGet("user_id").(int64)
	keys, err := h.service.ListAPIKeys(c.Request.Context(), appID, userID)
	if err != nil {
		writeError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"api_keys": keys})
}

func (h *Handler) RevokeAPIKey(c *gin.Context) {
	appID, err := parseAppID(c)
	if err != nil {
		return
	}

	keyID, err := parseKeyID(c)
	if err != nil {
		return
	}

	userID := c.MustGet("user_id").(int64)
	if err := h.service.RevokeAPIKey(c.Request.Context(), appID, keyID, userID); err != nil {
		writeError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"revoked": true})
}

func (h *Handler) CreateWebhook(c *gin.Context) {
	appID, err := parseAppID(c)
	if err != nil {
		return
	}

	var req CreateWebhookRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	userID := c.MustGet("user_id").(int64)
	webhook, err := h.service.CreateWebhook(c.Request.Context(), appID, userID, req)
	if err != nil {
		writeError(c, err)
		return
	}

	c.JSON(http.StatusCreated, webhook)
}

func (h *Handler) ListWebhooks(c *gin.Context) {
	appID, err := parseAppID(c)
	if err != nil {
		return
	}

	userID := c.MustGet("user_id").(int64)
	webhooks, err := h.service.ListWebhooks(c.Request.Context(), appID, userID)
	if err != nil {
		writeError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"webhooks": webhooks})
}

func (h *Handler) DeleteWebhook(c *gin.Context) {
	appID, err := parseAppID(c)
	if err != nil {
		return
	}

	webhookID, err := parseWebhookID(c)
	if err != nil {
		return
	}

	userID := c.MustGet("user_id").(int64)
	if err := h.service.DisableWebhook(c.Request.Context(), appID, webhookID, userID); err != nil {
		writeError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"deleted": true})
}

func parseAppID(c *gin.Context) (int64, error) {
	appID, err := strconv.ParseInt(c.Param("app_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid app id"})
		return 0, err
	}
	return appID, nil
}

func parseKeyID(c *gin.Context) (int64, error) {
	keyID, err := strconv.ParseInt(c.Param("key_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid key id"})
		return 0, err
	}
	return keyID, nil
}

func parseWebhookID(c *gin.Context) (int64, error) {
	webhookID, err := strconv.ParseInt(c.Param("webhook_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid webhook id"})
		return 0, err
	}
	return webhookID, nil
}

func writeError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, ErrNotFound), errors.Is(err, ErrKeyNotFound), errors.Is(err, ErrWebhookNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": httperr.Sanitize(err.Error())})
	case errors.Is(err, ErrForbidden):
		c.JSON(http.StatusForbidden, gin.H{"error": httperr.Sanitize(err.Error())})
	case errors.Is(err, ErrSlugTaken):
		c.JSON(http.StatusConflict, gin.H{"error": httperr.Sanitize(err.Error())})
	case errors.Is(err, ErrNameRequired), errors.Is(err, ErrInvalidSlug), errors.Is(err, ErrWebhookURLRequired):
		c.JSON(http.StatusBadRequest, gin.H{"error": httperr.Sanitize(err.Error())})
	default:
		httperr.WriteInternal(c, err)
	}
}
