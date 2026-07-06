package httperr

import (
	"errors"
	"log/slog"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

const InternalMessage = "Something went wrong. Please try again later."

var sensitiveSubstrings = []string{
	"failed to connect",
	"database=",
	"hostname resolving",
	"lookup ",
	":53:",
	"server misbehaving",
	"connection refused",
	"dial tcp",
	"pq:",
	"pgx",
	"postgres",
	"sql:",
	"redis",
	"aws ",
	"s3://",
	"access denied for user",
	"password authentication failed",
	"no such host",
	"certificate",
	"tls:",
	"stack trace",
	".go:",
	"/internal/",
	"user=",
}

// IsSensitive reports whether a message looks like internal infrastructure detail.
func IsSensitive(message string) bool {
	lower := strings.ToLower(strings.TrimSpace(message))
	if lower == "" {
		return false
	}
	for _, needle := range sensitiveSubstrings {
		if strings.Contains(lower, needle) {
			return true
		}
	}
	return false
}

// Sanitize returns message when safe for clients, otherwise InternalMessage.
func Sanitize(message string) string {
	if IsSensitive(message) {
		return InternalMessage
	}
	return message
}

// ClientMessage returns err.Error() when err matches a known sentinel, otherwise InternalMessage.
func ClientMessage(err error, known ...error) string {
	if err == nil {
		return InternalMessage
	}
	for _, sentinel := range known {
		if errors.Is(err, sentinel) {
			return Sanitize(sentinel.Error())
		}
	}
	slog.Error("unhandled error exposed to handler", "error", err)
	return InternalMessage
}

// WriteJSON writes a sanitized error payload.
func WriteJSON(c *gin.Context, status int, message string) {
	c.JSON(status, gin.H{"error": Sanitize(message)})
}

// WriteInternal logs the underlying error and returns a generic 500 response.
func WriteInternal(c *gin.Context, err error) {
	if err != nil {
		slog.Error("internal server error", "error", err)
	}
	WriteJSON(c, http.StatusInternalServerError, InternalMessage)
}
