package webhooks

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	db "github.com/brainart16/brenox/internal/db"
)

type Dispatcher struct {
	queries *db.Queries
	client  *http.Client
}

func NewDispatcher(queries *db.Queries) *Dispatcher {
	return &Dispatcher{
		queries: queries,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (d *Dispatcher) Dispatch(ctx context.Context, appID int64, event string, payload any) {
	if d == nil || d.queries == nil {
		return
	}

	hooks, err := d.queries.ListWebhooksByApp(ctx, appID)
	if err != nil || len(hooks) == 0 {
		return
	}

	body, err := json.Marshal(map[string]any{
		"event":     event,
		"timestamp": time.Now().UTC().Format(time.RFC3339Nano),
		"data":      payload,
	})
	if err != nil {
		return
	}

	for _, hook := range hooks {
		if !hookSubscribed(hook.Events, event) {
			continue
		}
		go d.deliver(hook, body)
	}
}

func hookSubscribed(events []string, event string) bool {
	if len(events) == 0 {
		return true
	}
	for _, subscribed := range events {
		if subscribed == event {
			return true
		}
	}
	return false
}

func (d *Dispatcher) deliver(hook db.Webhook, body []byte) {
	req, err := http.NewRequest(http.MethodPost, hook.Url, bytes.NewReader(body))
	if err != nil {
		return
	}

	signature := sign(body, hook.Secret)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Brenox-Signature", signature)

	resp, err := d.client.Do(req)
	if err != nil {
		slog.Warn("webhook delivery failed", "url", hook.Url, "error", err)
		return
	}
	_ = resp.Body.Close()

	if resp.StatusCode >= 300 {
		slog.Warn("webhook delivery non-success", "url", hook.Url, "status", resp.StatusCode)
	}
}

func sign(body []byte, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write(body)
	return hex.EncodeToString(mac.Sum(nil))
}
