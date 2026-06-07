package middleware

import (
	"net/http"

	"github.com/brainart16/brenox/internal/ratelimit"
	"github.com/gin-gonic/gin"
)

func RateLimitMiddleware(limiter *ratelimit.Limiter) gin.HandlerFunc {
	return func(c *gin.Context) {
		if limiter == nil {
			c.Next()
			return
		}

		appID := c.MustGet("app_id").(int64)
		apiKeyID := c.MustGet("api_key_id").(int64)
		key := ratelimit.Key(appID, apiKeyID)

		if !limiter.Allow(c.Request.Context(), key) {
			c.JSON(http.StatusTooManyRequests, gin.H{"error": "rate limit exceeded"})
			c.Abort()
			return
		}

		c.Next()
	}
}
