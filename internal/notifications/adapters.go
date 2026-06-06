package notifications

import "context"

// RealtimeDeliverer pushes notification.new events to a user's WebSocket connections.
type RealtimeDeliverer interface {
	DeliverNotification(userID int64, payload map[string]any)
}

// PushSender sends mobile push notifications (stub).
type PushSender interface {
	SendPush(ctx context.Context, userID int64, title, body string, data map[string]any) error
}

// EmailSender sends email notifications (stub).
type EmailSender interface {
	SendEmail(ctx context.Context, userID int64, subject, body string) error
}

type noopPush struct{}

func (noopPush) SendPush(context.Context, int64, string, string, map[string]any) error { return nil }

type noopEmail struct{}

func (noopEmail) SendEmail(context.Context, int64, string, string) error { return nil }

func NewNoopPushSender() PushSender  { return noopPush{} }
func NewNoopEmailSender() EmailSender { return noopEmail{} }
