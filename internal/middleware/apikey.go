package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/brainart16/brenox/internal/apps"
	"github.com/gin-gonic/gin"
)

type APIKeyAuthenticator interface {
	AuthenticateAPIKey(ctx context.Context, plainKey string) (apps.AuthenticatedApp, error)
}

// APIKeyMiddleware authenticates developer API requests via Bearer API key or X-API-Key.
func APIKeyMiddleware(auth APIKeyAuthenticator) gin.HandlerFunc {
	return func(c *gin.Context) {
		key, ok := extractAPIKey(c)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "missing api key"})
			c.Abort()
			return
		}

		authCtx, err := auth.AuthenticateAPIKey(c.Request.Context(), key)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid api key"})
			c.Abort()
			return
		}

		sandbox := authCtx.APIKey.IsSandbox
		workspaceID := apps.WorkspaceForSandbox(authCtx.App, sandbox)
		effectiveApp := apps.EffectiveApp(authCtx.App, sandbox)

		c.Set("app", effectiveApp)
		c.Set("app_id", authCtx.App.ID)
		c.Set("workspace_id", workspaceID)
		c.Set("api_key_id", authCtx.APIKey.ID)
		c.Set("is_sandbox", sandbox)
		c.Set("key_env", apps.EnvironmentFromSandbox(sandbox))
		c.Next()
	}
}

func extractAPIKey(c *gin.Context) (string, bool) {
	if header := strings.TrimSpace(c.GetHeader("X-API-Key")); header != "" {
		return header, true
	}

	token, ok := bearerToken(c)
	if !ok {
		return "", false
	}
	if apps.IsAPIKeyToken(token) {
		return token, true
	}
	return "", false
}
