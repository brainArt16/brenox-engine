package realtime

import "github.com/brainart16/brenox/internal/notifications"

type notificationDeliverer struct {
	hub *Hub
}

func NewNotificationDeliverer(hub *Hub) notifications.RealtimeDeliverer {
	return &notificationDeliverer{hub: hub}
}

func (d *notificationDeliverer) DeliverNotification(userID int64, payload map[string]any) {
	event := NewOutboundEvent("notification.new", 0, 0, payload)
	if d.hub.broker != nil {
		d.hub.broker.PublishToUser(userID, event)
		return
	}
	d.hub.notifyUserLocal(userID, event)
}
