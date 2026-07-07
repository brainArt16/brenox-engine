package billing

import (
	"os"
	"strings"
)

type Config struct {
	StripeSecretKey     string
	StripeWebhookSecret string
	CheckoutBaseURL     string
}

func LoadConfig() Config {
	return Config{
		StripeSecretKey:     strings.TrimSpace(os.Getenv("STRIPE_SECRET_KEY")),
		StripeWebhookSecret: strings.TrimSpace(os.Getenv("STRIPE_WEBHOOK_SECRET")),
		CheckoutBaseURL:     strings.TrimRight(strings.TrimSpace(os.Getenv("STRIPE_CHECKOUT_BASE_URL")), "/"),
	}
}

func (c Config) Enabled() bool {
	return c.StripeSecretKey != "" && c.CheckoutBaseURL != ""
}
