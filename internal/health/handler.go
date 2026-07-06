package health

import (
	"net/http"

	"github.com/brainart16/brenox/internal/database"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	goredis "github.com/redis/go-redis/v9"

	redisutil "github.com/brainart16/brenox/internal/redis"
)

type Handler struct {
	pool  *pgxpool.Pool
	redis *goredis.Client
}

func NewHandler(pool *pgxpool.Pool, redisClient *goredis.Client) *Handler {
	return &Handler{
		pool:  pool,
		redis: redisClient,
	}
}

func (h *Handler) Check(c *gin.Context) {
	ctx := c.Request.Context()

	dbOK := h.pool.Ping(ctx) == nil

	redisConfigured := redisutil.LoadURL() != ""
	redisOK := !redisConfigured
	if redisConfigured {
		redisOK = h.redis != nil && redisutil.Ping(ctx, h.redis) == nil
	}

	status := http.StatusOK
	if !dbOK || !redisOK {
		status = http.StatusServiceUnavailable
	}

	migrationStatus, _ := database.CheckMigrations(ctx, h.pool)
	if !migrationStatus.OK {
		status = http.StatusServiceUnavailable
	}

	c.JSON(status, gin.H{
		"status": statusText(status),
		"checks": gin.H{
			"database":   dbCheck(dbOK),
			"redis":      redisCheck(redisConfigured, redisOK),
			"migrations": migrationCheck(migrationStatus),
		},
	})
}

func migrationCheck(status database.MigrationStatus) gin.H {
	check := gin.H{
		"status":  "up",
		"version": status.Version,
	}
	if !status.OK {
		check["status"] = "pending"
		check["message"] = status.Message
	}
	if status.Dirty {
		check["dirty"] = true
	}
	return check
}

func statusText(code int) string {
	if code == http.StatusOK {
		return "ok"
	}
	return "degraded"
}

func dbCheck(ok bool) gin.H {
	status := "up"
	if !ok {
		status = "down"
	}
	return gin.H{"status": status}
}

func redisCheck(configured, ok bool) gin.H {
	if !configured {
		return gin.H{"status": "disabled"}
	}

	status := "up"
	if !ok {
		status = "down"
	}
	return gin.H{"status": status}
}
