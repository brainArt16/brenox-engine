package middleware

import (
	"net/http"

	"github.com/brainart16/brenox/internal/origins"
	"github.com/gin-gonic/gin"
)

// CORSMiddleware applies per-app and platform origin rules for browser requests.
func CORSMiddleware(checker *origins.Checker) gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := origins.Normalize(c.GetHeader("Origin"))
		appIDValue := origins.AppIDFromToken(c.GetHeader("Authorization"), c.Query("token"))
		if id, ok := c.Get("app_id"); ok {
			appIDValue = id.(int64)
		}

		hints := origins.HintsFromRequest(
			c.Request.Method,
			c.Request.URL.Path,
			c.Request.URL.RawQuery,
			appIDValue,
		)

		if origin != "" && checker.IsAllowed(c.Request.Context(), origin, hints) {
			c.Header("Access-Control-Allow-Origin", origin)
			c.Header("Vary", "Origin")
			c.Header("Access-Control-Allow-Credentials", "true")
			c.Header("Access-Control-Allow-Headers", "Authorization, Content-Type, Idempotency-Key, X-API-Key")
			c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
			c.Header("Access-Control-Max-Age", "86400")
		}

		if c.Request.Method == http.MethodOptions {
			if origin != "" && checker.IsAllowed(c.Request.Context(), origin, hints) {
				c.Status(http.StatusNoContent)
			} else {
				c.Status(http.StatusForbidden)
			}
			c.Abort()
			return
		}

		c.Next()
	}
}
