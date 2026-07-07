package billing

type PlanResponse struct {
	Slug              string `json:"slug"`
	Name              string `json:"name"`
	PriceCents        int32  `json:"price_cents"`
	PriceDisplay      string `json:"price_display"`
	MessagesLimit     int32  `json:"messages_limit"`
	ConnectionsLimit  int32  `json:"connections_limit"`
	RetentionDays     int32  `json:"retention_days"`
	WebhooksEnabled   bool   `json:"webhooks_enabled"`
	VideoCallsEnabled bool   `json:"video_calls_enabled"`
	ModerationEnabled bool   `json:"moderation_enabled"`
}

type UsageResponse struct {
	MessagesThisMonth int64 `json:"messages_this_month"`
	MessagesLimit     int32 `json:"messages_limit"`
}

type SubscriptionResponse struct {
	PlanSlug          string  `json:"plan_slug"`
	PlanName          string  `json:"plan_name"`
	Status            string  `json:"status"`
	PriceCents        int32   `json:"price_cents"`
	MessagesLimit     int32   `json:"messages_limit"`
	ConnectionsLimit  int32   `json:"connections_limit"`
	RetentionDays     int32   `json:"retention_days"`
	WebhooksEnabled   bool    `json:"webhooks_enabled"`
	VideoCallsEnabled bool    `json:"video_calls_enabled"`
	ModerationEnabled bool    `json:"moderation_enabled"`
	CurrentPeriodEnd  string  `json:"current_period_end,omitempty"`
	NeedsPayment      bool    `json:"needs_payment"`
}

type AppBillingResponse struct {
	AppID        int64                `json:"app_id"`
	Subscription SubscriptionResponse `json:"subscription"`
	Usage        UsageResponse        `json:"usage"`
}

type CheckoutRequest struct {
	PlanSlug string `json:"plan_slug"`
}

type CheckoutResponse struct {
	URL string `json:"url"`
}

type PlatformStatusResponse struct {
	MaintenanceMode    bool   `json:"maintenance_mode"`
	MaintenanceMessage string `json:"maintenance_message,omitempty"`
}

type PlatformSettingsResponse struct {
	MaintenanceMode    bool   `json:"maintenance_mode"`
	MaintenanceMessage string `json:"maintenance_message"`
}

type UpdatePlatformSettingsRequest struct {
	MaintenanceMode    *bool   `json:"maintenance_mode"`
	MaintenanceMessage *string `json:"maintenance_message"`
}

type AdminUpdateSubscriptionRequest struct {
	PlanSlug *string `json:"plan_slug"`
	Status   *string `json:"status"`
}

type AdminSubscriptionListItem struct {
	AppID              int64  `json:"app_id"`
	AppName            string `json:"app_name"`
	AppSlug            string `json:"app_slug"`
	PlanSlug           string `json:"plan_slug"`
	PlanName           string `json:"plan_name"`
	Status             string `json:"status"`
	CurrentPeriodEnd   string `json:"current_period_end,omitempty"`
	MessagesThisMonth  int64  `json:"messages_this_month"`
}

type AdminBillingOverview struct {
	ActiveSubscriptions int64 `json:"active_subscriptions"`
}

type AdminAppBillingResponse struct {
	AppBillingResponse
}
