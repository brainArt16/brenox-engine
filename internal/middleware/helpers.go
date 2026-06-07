package middleware

import "github.com/gin-gonic/gin"

func bearerToken(c *gin.Context) (string, bool) {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		return "", false
	}

	parts := splitBearer(authHeader)
	if len(parts) != 2 || parts[1] == "" {
		return "", false
	}

	return parts[1], true
}

func splitBearer(authHeader string) []string {
	const prefix = "Bearer "
	if len(authHeader) > len(prefix) && authHeader[:len(prefix)] == prefix {
		return []string{prefix, authHeader[len(prefix):]}
	}
	return nil
}
