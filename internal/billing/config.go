package billing

import (
	"os"
	"strings"
)

type Config struct {
	StripeSecretKey     string
	StripeWebhookSecret string
	CheckoutBaseURL     string
	PriceStarter        string
	PriceGrowth         string
	PriceScale          string
}

func LoadConfig() Config {
	return Config{
		StripeSecretKey:     strings.TrimSpace(os.Getenv("STRIPE_SECRET_KEY")),
		StripeWebhookSecret: strings.TrimSpace(os.Getenv("STRIPE_WEBHOOK_SECRET")),
		CheckoutBaseURL:     strings.TrimRight(strings.TrimSpace(os.Getenv("STRIPE_CHECKOUT_BASE_URL")), "/"),
		PriceStarter:        strings.TrimSpace(os.Getenv("STRIPE_PRICE_STARTER")),
		PriceGrowth:         strings.TrimSpace(os.Getenv("STRIPE_PRICE_GROWTH")),
		PriceScale:          strings.TrimSpace(os.Getenv("STRIPE_PRICE_SCALE")),
	}
}

func (c Config) Enabled() bool {
	return c.StripeSecretKey != "" && c.CheckoutBaseURL != ""
}

func (c Config) PriceIDForPlan(slug string) string {
	switch slug {
	case "starter":
		return c.PriceStarter
	case "growth":
		return c.PriceGrowth
	case "scale":
		return c.PriceScale
	default:
		return ""
	}
}
