package middleware

import (
	"context"
	"log/slog"
	"strings"

	db "github.com/brainart16/brenox/internal/db"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgtype"
)

type AuditRecorder interface {
	RecordAudit(ctx context.Context, entry AuditEntry) error
}

type AuditEntry struct {
	UserID     *int64
	AppID      *int64
	Action     string
	Method     string
	Path       string
	IPAddress  string
	StatusCode int
}

type auditRecorder struct {
	queries *db.Queries
}

func NewAuditRecorder(queries *db.Queries) AuditRecorder {
	return &auditRecorder{queries: queries}
}

func (r *auditRecorder) RecordAudit(ctx context.Context, entry AuditEntry) error {
	params := db.CreateAuditLogParams{
		Action:     entry.Action,
		Method:     entry.Method,
		Path:       entry.Path,
		IpAddress:  pgtype.Text{String: entry.IPAddress, Valid: entry.IPAddress != ""},
		StatusCode: pgtype.Int4{Int32: int32(entry.StatusCode), Valid: entry.StatusCode > 0},
	}
	if entry.UserID != nil {
		params.UserID = pgtype.Int8{Int64: *entry.UserID, Valid: true}
	}
	if entry.AppID != nil {
		params.AppID = pgtype.Int8{Int64: *entry.AppID, Valid: true}
	}
	return r.queries.CreateAuditLog(ctx, params)
}

func AuditMiddleware(recorder AuditRecorder) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		if c.Request.Method == "GET" || c.Request.Method == "OPTIONS" || c.Request.Method == "HEAD" {
			return
		}
		if !shouldAudit(c.FullPath(), c.Request.URL.Path) {
			return
		}

		entry := AuditEntry{
			Action:     auditAction(c),
			Method:     c.Request.Method,
			Path:       c.Request.URL.Path,
			IPAddress:  c.ClientIP(),
			StatusCode: c.Writer.Status(),
		}
		if userID, ok := c.Get("user_id"); ok {
			id := userID.(int64)
			entry.UserID = &id
		}
		if appID, ok := c.Get("app_id"); ok {
			id := appID.(int64)
			entry.AppID = &id
		}

		go func() {
			if err := recorder.RecordAudit(context.Background(), entry); err != nil {
				slog.Warn("audit log failed", "path", entry.Path, "error", err)
			}
		}()
	}
}

func shouldAudit(fullPath, rawPath string) bool {
	path := fullPath
	if path == "" {
		path = rawPath
	}
	if strings.HasPrefix(path, "/health") || strings.HasPrefix(path, "/metrics") {
		return false
	}
	return strings.HasPrefix(path, "/auth") ||
		strings.HasPrefix(path, "/api") ||
		strings.HasPrefix(path, "/v1")
}

func auditAction(c *gin.Context) string {
	switch {
	case strings.Contains(c.FullPath(), "/keys"):
		return "api_key.mutate"
	case strings.Contains(c.FullPath(), "/webhooks"):
		return "webhook.mutate"
	case strings.Contains(c.FullPath(), "/auth/login"):
		return "auth.login"
	case strings.Contains(c.FullPath(), "/auth/refresh"):
		return "auth.refresh"
	case strings.Contains(c.FullPath(), "/auth/register"):
		return "auth.register"
	default:
		return "http.mutate"
	}
}
