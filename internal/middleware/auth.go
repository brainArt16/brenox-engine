package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	jwtutil "github.com/brainart16/brenox/pkg/jwt"
)


// AuthMiddleware protects routes.
func AuthMiddleware() gin.HandlerFunc {

	return func(c *gin.Context) {

		// Get Authorization header.
		authHeader := c.GetHeader(
			"Authorization",
		)

		if authHeader == "" {

			c.JSON(
				http.StatusUnauthorized,
				gin.H{
					"error": "missing auth header",
				},
			)

			c.Abort()

			return
		}

		// Expected format: Bearer token_here
		splitToken := strings.Split(
			authHeader,
			"Bearer ",
		)

		if len(splitToken) != 2 {

			c.JSON(
				http.StatusUnauthorized,
				gin.H{
					"error": "invalid auth header",
				},
			)

			c.Abort()

			return
		}

		tokenString := splitToken[1]

		// Validate JWT token.
		claims, err := jwtutil.ValidateToken(
			tokenString,
		)

		if err != nil {

			c.JSON(
				http.StatusUnauthorized,
				gin.H{
					"error": "invalid token",
				},
			)

			// Stop processing request pipeline immediately
			c.Abort()

			return
		}

		// Store authenticated user inside request context.
		c.Set("user_id", claims.UserID)

		// Continue processing request pipeline.
		c.Next()
	}
}