package apps

import "errors"

var (
	ErrNotFound            = errors.New("app not found")
	ErrForbidden           = errors.New("forbidden")
	ErrSlugTaken           = errors.New("app slug already taken")
	ErrInvalidSlug         = errors.New("invalid app slug")
	ErrNameRequired        = errors.New("app name is required")
	ErrInvalidKey          = errors.New("invalid api key")
	ErrKeyNotFound         = errors.New("api key not found")
	ErrWebhookNotFound     = errors.New("webhook not found")
	ErrWebhookURLRequired  = errors.New("webhook url is required")
	ErrInvalidOrigin       = errors.New("invalid allowed origin")
	ErrTooManyOrigins      = errors.New("too many allowed origins")
)

type CreateAppRequest struct {
	Name     string `json:"name"`
	Slug     string `json:"slug"`
	PlanSlug string `json:"plan_slug"`
}

type CreateAPIKeyRequest struct {
	Name    string `json:"name"`
	Sandbox bool   `json:"sandbox"`
}

type CreateWebhookRequest struct {
	URL    string   `json:"url"`
	Events []string `json:"events"`
}

type AppResponse struct {
	ID              int64    `json:"id"`
	Name            string   `json:"name"`
	Slug            string   `json:"slug"`
	WorkspaceID     int64    `json:"workspace_id"`
	OwnerID         int64    `json:"owner_id"`
	CreatedAt       string   `json:"created_at"`
	AllowedOrigins  []string `json:"allowed_origins"`
}

type UpdateAllowedOriginsRequest struct {
	AllowedOrigins []string `json:"allowed_origins"`
}

type APIKeyResponse struct {
	ID        int64  `json:"id"`
	AppID     int64  `json:"app_id"`
	Name      string `json:"name"`
	KeyPrefix string `json:"key_prefix"`
	IsSandbox bool   `json:"is_sandbox"`
	CreatedAt string `json:"created_at"`
	RevokedAt string `json:"revoked_at,omitempty"`
	LastUsed  string `json:"last_used_at,omitempty"`
}

type APIKeyCreatedResponse struct {
	APIKeyResponse
	Secret string `json:"secret"`
}

type WebhookResponse struct {
	ID        int64    `json:"id"`
	AppID     int64    `json:"app_id"`
	URL       string   `json:"url"`
	Events    []string `json:"events"`
	CreatedAt string   `json:"created_at"`
	Secret    string   `json:"secret,omitempty"`
}
