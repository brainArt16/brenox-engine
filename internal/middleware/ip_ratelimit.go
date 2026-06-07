package middleware

import (
	"net/http"

	"github.com/brainart16/brenox/internal/ratelimit"
	"github.com/gin-gonic/gin"
)

func IPRateLimitMiddleware(limiter *ratelimit.Limiter) gin.HandlerFunc {
	return func(c *gin.Context) {
		if limiter == nil {
			c.Next()
			return
		}

		ip := c.ClientIP()
		if !limiter.Allow(c.Request.Context(), ratelimit.IPKey(ip)) {
			c.JSON(http.StatusTooManyRequests, gin.H{"error": "rate limit exceeded"})
			c.Abort()
			return
		}

		c.Next()
	}
}
