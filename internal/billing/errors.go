package billing

import "errors"

var (
	ErrNotFound           = errors.New("billing record not found")
	ErrForbidden          = errors.New("forbidden")
	ErrInvalidPlan        = errors.New("invalid plan")
	ErrStripeNotConfigured = errors.New("billing is not configured")
	ErrMessageLimit       = errors.New("monthly message limit reached")
	ErrSubscriptionInactive = errors.New("subscription inactive")
	ErrWebhooksNotAllowed = errors.New("webhooks not included in current plan")
	ErrVideoNotAllowed    = errors.New("video calls not included in current plan")
	ErrInvalidRequest     = errors.New("invalid request")
	ErrPlanInUse          = errors.New("plan has active subscriptions")
	ErrSlugTaken          = errors.New("plan slug already taken")
	ErrInvalidSlug        = errors.New("invalid plan slug")
)

const (
	StatusIncomplete = "incomplete"
	StatusTrialing   = "trialing"
	StatusActive     = "active"
	StatusPastDue    = "past_due"
	StatusCanceled   = "canceled"
)

func statusAllowsUsage(status string) bool {
	switch status {
	case StatusActive, StatusTrialing, StatusIncomplete:
		return true
	default:
		return false
	}
}
