package auth

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/brainart16/brenox/internal/httperr"
	"github.com/brainart16/brenox/pkg/jwt"
	"github.com/gin-gonic/gin"
)

type AdminBootstrapper interface {
	SyncBootstrapAdmin(ctx context.Context, email string) error
}

type Handler struct {
	service   *Service
	bootstrap AdminBootstrapper
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) SetAdminBootstrap(bootstrap AdminBootstrapper) {
	h.bootstrap = bootstrap
}

func (h *Handler) syncAdmin(ctx context.Context, email string) {
	if h.bootstrap == nil {
		return
	}
	_ = h.bootstrap.SyncBootstrapAdmin(ctx, email)
}

// Register endpoint handler.
func (h *Handler) Register(
	c *gin.Context,
) {

	var req RegisterRequest

	// Bind incoming JSON body into Go struct.
	err := c.ShouldBindJSON(&req)

	if err != nil {

		httperr.WriteJSON(
			c,
			http.StatusBadRequest,
			"invalid request body",
		)

		return
	}

	// Call business logic layer.

	user, err := h.service.Register(
		c.Request.Context(),
		req,
	)

	if err != nil {
		writeAuthError(c, err)
		return
	}

	h.syncAdmin(c.Request.Context(), user.Email)

	c.JSON(
		http.StatusCreated,
		gin.H{
			"id":       user.ID,
			"email":    user.Email,
			"username": user.Username,
		},
	)
}

// Login endpoint handler.
func (h *Handler) Login(
	c *gin.Context,
) {

	var req LoginRequest

	err := c.ShouldBindJSON(&req)

	if err != nil {

		httperr.WriteJSON(
			c,
			http.StatusBadRequest,
			"invalid request",
		)

		return
	}

	token, err := h.service.Login(
		c.Request.Context(),
		req,
	)

	if err != nil {
		writeAuthError(c, err)
		return
	}

	h.syncAdmin(c.Request.Context(), req.Email)

	c.JSON(
		http.StatusOK,
		gin.H{
			"token": token,
		},
	)
}

func (h *Handler) Refresh(c *gin.Context) {
	tokenString, ok := bearerToken(c)
	if !ok {
		var req refreshRequest
		if err := c.ShouldBindJSON(&req); err != nil || strings.TrimSpace(req.Token) == "" {
			httperr.WriteJSON(c, http.StatusBadRequest, "token required")
			return
		}
		tokenString = req.Token
	}

	token, err := h.service.Refresh(c.Request.Context(), tokenString)
	if err != nil {
		writeAuthError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": token})
}

func writeAuthError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, ErrEmailExists):
		httperr.WriteJSON(c, http.StatusBadRequest, httperr.ClientMessage(err, ErrEmailExists))
	case errors.Is(err, ErrRegistrationFailed):
		httperr.WriteJSON(c, http.StatusBadRequest, httperr.ClientMessage(err, ErrRegistrationFailed))
	case errors.Is(err, ErrInvalidCredentials):
		httperr.WriteJSON(c, http.StatusUnauthorized, httperr.ClientMessage(err, ErrInvalidCredentials))
	case errors.Is(err, ErrAccountSuspended):
		httperr.WriteJSON(c, http.StatusForbidden, httperr.ClientMessage(err, ErrAccountSuspended))
	case errors.Is(err, ErrInvalidToken), errors.Is(err, jwt.ErrTokenInvalid), errors.Is(err, jwt.ErrTokenRevoked):
		httperr.WriteJSON(c, http.StatusUnauthorized, httperr.ClientMessage(err, ErrInvalidToken, jwt.ErrTokenInvalid, jwt.ErrTokenRevoked))
	default:
		httperr.WriteInternal(c, err)
	}
}

func bearerToken(c *gin.Context) (string, bool) {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		return "", false
	}
	parts := strings.SplitN(authHeader, "Bearer ", 2)
	if len(parts) != 2 || parts[1] == "" {
		return "", false
	}
	return parts[1], true
}
