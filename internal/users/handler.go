package users

import (
	"errors"
	"net/http"

	"github.com/brainart16/brenox/internal/httperr"
	"github.com/brainart16/brenox/internal/platformadmin"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	service       *Service
	platformAdmin *platformadmin.Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) SetPlatformAdminService(platformAdmin *platformadmin.Service) {
	h.platformAdmin = platformAdmin
}

func (h *Handler) GetMe(c *gin.Context) {
	userID := c.MustGet("user_id").(int64)

	if h.platformAdmin != nil {
		if profile, err := h.service.GetProfile(c.Request.Context(), userID); err == nil {
			_ = h.platformAdmin.SyncBootstrapAdmin(c.Request.Context(), profile.Email)
		}
	}

	profile, err := h.service.GetProfile(c.Request.Context(), userID)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, profile)
}

func (h *Handler) UpdateMe(c *gin.Context) {
	var req UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	userID := c.MustGet("user_id").(int64)
	profile, err := h.service.UpdateProfile(c.Request.Context(), userID, req)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, profile)
}

func (h *Handler) ChangePassword(c *gin.Context) {
	var req ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	userID := c.MustGet("user_id").(int64)
	if err := h.service.ChangePassword(c.Request.Context(), userID, req); err != nil {
		writeError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"changed": true})
}

func writeError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, ErrNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": httperr.Sanitize(err.Error())})
	case errors.Is(err, ErrUsernameRequired):
		c.JSON(http.StatusBadRequest, gin.H{"error": httperr.Sanitize(err.Error())})
	case errors.Is(err, ErrUsernameTaken):
		c.JSON(http.StatusConflict, gin.H{"error": httperr.Sanitize(err.Error())})
	case errors.Is(err, ErrInvalidPassword):
		c.JSON(http.StatusUnauthorized, gin.H{"error": httperr.Sanitize(err.Error())})
	case errors.Is(err, ErrPasswordRequired), errors.Is(err, ErrPasswordTooShort):
		c.JSON(http.StatusBadRequest, gin.H{"error": httperr.Sanitize(err.Error())})
	default:
		httperr.WriteInternal(c, err)
	}
}
