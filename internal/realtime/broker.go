package realtime

// EventBroker publishes outbound realtime events and manages cross-node subscriptions.
type EventBroker interface {
	Publish(event Event)
	EnsureSubscribed(workspaceID, channelID int64)
	MaybeUnsubscribe(workspaceID, channelID int64)
	Start()
	Close()
}

type localBroker struct {
	hub *Hub
}

func NewLocalBroker(hub *Hub) EventBroker {
	return &localBroker{hub: hub}
}

func (b *localBroker) Publish(event Event) {
	b.hub.enqueueBroadcast(event)
}

func (b *localBroker) EnsureSubscribed(_, _ int64) {}

func (b *localBroker) MaybeUnsubscribe(_, _ int64) {}

func (b *localBroker) Start() {}

func (b *localBroker) Close() {}
