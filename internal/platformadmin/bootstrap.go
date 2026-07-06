package platformadmin

import (
	"os"
	"strings"
)

func AdminEmailsFromEnv() []string {
	raw := strings.TrimSpace(os.Getenv("PLATFORM_ADMIN_EMAILS"))
	if raw == "" {
		return nil
	}

	parts := strings.Split(raw, ",")
	emails := make([]string, 0, len(parts))
	for _, part := range parts {
		email := strings.ToLower(strings.TrimSpace(part))
		if email != "" {
			emails = append(emails, email)
		}
	}
	return emails
}

func IsBootstrapAdminEmail(email string) bool {
	target := strings.ToLower(strings.TrimSpace(email))
	if target == "" {
		return false
	}
	for _, adminEmail := range AdminEmailsFromEnv() {
		if adminEmail == target {
			return true
		}
	}
	return false
}
