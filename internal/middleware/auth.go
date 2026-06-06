package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	jwtutil "github.com/brainart16/brenox/pkg/jwt"
)

// AuthMiddleware protects routes. Accepts JWT from Authorization header or ?token= query param.
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString, ok := bearerToken(c)
		if !ok {
			tokenString = strings.TrimSpace(c.Query("token"))
		}

		if tokenString == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "missing auth token"})
			c.Abort()
			return
		}

		claims, err := jwtutil.ValidateToken(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			c.Abort()
			return
		}

		c.Set("user_id", claims.UserID)
		c.Next()
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
