package ratelimit

// IPConfigToConfig adapts IP rate limit settings to the shared limiter Config shape.
func IPConfigToConfig(cfg IPConfig) Config {
	return Config{RequestsPerMinute: cfg.RequestsPerMinute}
}
