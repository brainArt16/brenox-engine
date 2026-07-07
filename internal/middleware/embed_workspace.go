package middleware

import (
	"net/http"
	"strconv"

	"github.com/brainart16/brenox/internal/apps"
	db "github.com/brainart16/brenox/internal/db"
	"github.com/gin-gonic/gin"
)

// EmbedWorkspaceGuard blocks embed JWTs from accessing the wrong app workspace lane.
func EmbedWorkspaceGuard(queries *db.Queries) gin.HandlerFunc {
	return func(c *gin.Context) {
		appIDRaw, ok := c.Get("app_id")
		if !ok {
			c.Next()
			return
		}
		appID, ok := appIDRaw.(int64)
		if !ok || appID == 0 {
			c.Next()
			return
		}

		workspaceID, err := strconv.ParseInt(c.Param("workspace_id"), 10, 64)
		if err != nil || workspaceID <= 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid workspace_id"})
			c.Abort()
			return
		}

		keyEnv := apps.EnvLive
		if raw, ok := c.Get("key_env"); ok {
			if env, ok := raw.(string); ok && env != "" {
				keyEnv = env
			}
		}

		app, err := queries.GetAppByID(c.Request.Context(), appID)
		if err != nil {
			c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
			c.Abort()
			return
		}

		if workspaceID != apps.WorkspaceForEnvironment(app, keyEnv) {
			c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
			c.Abort()
			return
		}

		c.Next()
	}
}

// EmbedWorkspaceQueryGuard is the WebSocket variant (workspace_id query param).
func EmbedWorkspaceQueryGuard(queries *db.Queries) gin.HandlerFunc {
	return func(c *gin.Context) {
		appIDRaw, ok := c.Get("app_id")
		if !ok {
			c.Next()
			return
		}
		appID, ok := appIDRaw.(int64)
		if !ok || appID == 0 {
			c.Next()
			return
		}

		workspaceID, err := strconv.ParseInt(c.Query("workspace_id"), 10, 64)
		if err != nil || workspaceID <= 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid workspace_id"})
			c.Abort()
			return
		}

		keyEnv := apps.EnvLive
		if raw, ok := c.Get("key_env"); ok {
			if env, ok := raw.(string); ok && env != "" {
				keyEnv = env
			}
		}

		app, err := queries.GetAppByID(c.Request.Context(), appID)
		if err != nil {
			c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
			c.Abort()
			return
		}

		if workspaceID != apps.WorkspaceForEnvironment(app, keyEnv) {
			c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
			c.Abort()
			return
		}

		c.Next()
	}
}
