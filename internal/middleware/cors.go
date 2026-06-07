package middleware

import (
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
)

type CORSConfig struct {
	AllowedOrigins []string
}

func LoadCORSConfig() CORSConfig {
	raw := strings.TrimSpace(os.Getenv("CORS_ALLOWED_ORIGINS"))
	if raw == "" {
		raw = strings.TrimSpace(os.Getenv("WS_ALLOWED_ORIGINS"))
	}
	return CORSConfig{AllowedOrigins: parseOrigins(raw)}
}

func CORSMiddleware(cfg CORSConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		if origin != "" && originAllowed(cfg.AllowedOrigins, origin) {
			c.Header("Access-Control-Allow-Origin", origin)
			c.Header("Vary", "Origin")
			c.Header("Access-Control-Allow-Credentials", "true")
			c.Header("Access-Control-Allow-Headers", "Authorization, Content-Type, Idempotency-Key, X-API-Key")
			c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
			c.Header("Access-Control-Max-Age", "86400")
		}

		if c.Request.Method == http.MethodOptions {
			if origin != "" && originAllowed(cfg.AllowedOrigins, origin) {
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

func parseOrigins(raw string) []string {
	raw = strings.TrimSpace(raw)
	if raw == "" || raw == "*" {
		return nil
	}
	parts := strings.Split(raw, ",")
	origins := make([]string, 0, len(parts))
	for _, part := range parts {
		origin := strings.TrimSpace(part)
		if origin != "" {
			origins = append(origins, origin)
		}
	}
	return origins
}

func originAllowed(allowed []string, origin string) bool {
	if len(allowed) == 0 {
		return true
	}
	for _, item := range allowed {
		if item == origin {
			return true
		}
	}
	return false
}
