package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

type AccessTokenValidator interface {
	ValidateAccessToken(ctx context.Context, tokenString string) (userID int64, appID int64, keyEnv string, err error)
}

// AuthMiddleware protects routes. Accepts JWT from Authorization header or ?token= query param.
func AuthMiddleware(validator AccessTokenValidator) gin.HandlerFunc {
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

		userID, appID, keyEnv, err := validator.ValidateAccessToken(c.Request.Context(), tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			c.Abort()
			return
		}

		c.Set("user_id", userID)
		if appID > 0 {
			c.Set("app_id", appID)
			c.Set("key_env", keyEnv)
		}
		c.Next()
	}
}
