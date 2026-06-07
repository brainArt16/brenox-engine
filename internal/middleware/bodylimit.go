package middleware

import (
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

func LoadMaxBodyBytes() int64 {
	raw := strings.TrimSpace(os.Getenv("MAX_REQUEST_BODY_BYTES"))
	if raw == "" {
		return 1 << 20
	}
	value, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || value <= 0 {
		return 1 << 20
	}
	return value
}

func RequestSizeLimitMiddleware(maxBytes int64) gin.HandlerFunc {
	return func(c *gin.Context) {
		if maxBytes <= 0 {
			c.Next()
			return
		}

		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxBytes)
		c.Next()

		if c.Errors.Last() != nil && strings.Contains(c.Errors.Last().Error(), "http: request body too large") {
			c.AbortWithStatusJSON(http.StatusRequestEntityTooLarge, gin.H{"error": "request body too large"})
		}
	}
}
