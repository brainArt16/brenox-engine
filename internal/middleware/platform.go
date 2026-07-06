package middleware

import (
	"context"
	"net/http"

	db "github.com/brainart16/brenox/internal/db"
	"github.com/brainart16/brenox/internal/platformadmin"
	"github.com/gin-gonic/gin"
)

type PlatformUserLoader interface {
	GetUserAuthState(ctx context.Context, id int64) (db.GetUserAuthStateRow, error)
}

func PlatformUserMiddleware(loader PlatformUserLoader) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.MustGet("user_id").(int64)

		state, err := loader.GetUserAuthState(c.Request.Context(), userID)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			c.Abort()
			return
		}

		if state.SuspendedAt.Valid {
			c.JSON(http.StatusForbidden, gin.H{"error": "account suspended"})
			c.Abort()
			return
		}

		c.Set("platform_role", state.PlatformRole)
		c.Next()
	}
}

func RequirePlatformRole(minRole string) gin.HandlerFunc {
	return func(c *gin.Context) {
		role, _ := c.Get("platform_role")
		roleStr, _ := role.(string)
		if platformadmin.RoleRank(roleStr) < platformadmin.RoleRank(minRole) {
			c.JSON(http.StatusForbidden, gin.H{"error": "platform admin access required"})
			c.Abort()
			return
		}
		c.Next()
	}
}

func RequirePlatformWrite() gin.HandlerFunc {
	return RequirePlatformRole(platformadmin.RoleAdmin)
}
