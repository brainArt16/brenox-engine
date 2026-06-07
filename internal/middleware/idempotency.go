package middleware

import (
	"bytes"
	"context"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

type IdempotencyStore interface {
	GetIdempotency(ctx context.Context, appID int64, key string) ([]byte, int, bool, error)
	SaveIdempotency(ctx context.Context, appID int64, key, endpoint string, statusCode int, body []byte) error
}

func IdempotencyMiddleware(store IdempotencyStore) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method == http.MethodGet {
			c.Next()
			return
		}

		key := strings.TrimSpace(c.GetHeader("Idempotency-Key"))
		if key == "" {
			c.Next()
			return
		}

		appID := c.MustGet("app_id").(int64)
		endpoint := c.FullPath()

		if body, status, ok, err := store.GetIdempotency(c.Request.Context(), appID, key); err == nil && ok {
			c.Data(status, "application/json", body)
			c.Abort()
			return
		}

		writer := &captureWriter{ResponseWriter: c.Writer, status: http.StatusOK, body: &bytes.Buffer{}}
		c.Writer = writer
		c.Next()

		if writer.status >= 200 && writer.status < 300 && writer.body.Len() > 0 {
			_ = store.SaveIdempotency(c.Request.Context(), appID, key, endpoint, writer.status, writer.body.Bytes())
		}
	}
}

type captureWriter struct {
	gin.ResponseWriter
	status int
	body   *bytes.Buffer
}

func (w *captureWriter) WriteHeader(code int) {
	w.status = code
	w.ResponseWriter.WriteHeader(code)
}

func (w *captureWriter) Write(b []byte) (int, error) {
	if w.status == 0 {
		w.status = http.StatusOK
	}
	if len(b) > 0 && w.body != nil {
		_, _ = w.body.Write(b)
	}
	return w.ResponseWriter.Write(b)
}
