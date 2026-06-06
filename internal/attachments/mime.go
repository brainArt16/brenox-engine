package attachments

import (
	"strings"
)

var allowedMIMETypes = map[string]struct{}{
	"image/jpeg":         {},
	"image/png":          {},
	"image/gif":          {},
	"image/webp":         {},
	"application/pdf":    {},
	"text/plain":         {},
	"text/markdown":      {},
	"application/json":   {},
	"application/zip":    {},
	"application/msword": {},
	"application/vnd.openxmlformats-officedocument.wordprocessingml.document": {},
	"application/vnd.ms-excel": {},
	"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet": {},
}

func validateUpload(fileName, mimeType string, sizeBytes, maxBytes int64) error {
	if strings.TrimSpace(fileName) == "" {
		return ErrInvalidFile
	}
	if sizeBytes <= 0 {
		return ErrInvalidFile
	}
	if sizeBytes > maxBytes {
		return ErrFileTooLarge
	}

	normalized := strings.ToLower(strings.TrimSpace(mimeType))
	if normalized == "" {
		return ErrMimeNotAllowed
	}
	if _, ok := allowedMIMETypes[normalized]; !ok {
		return ErrMimeNotAllowed
	}
	return nil
}
