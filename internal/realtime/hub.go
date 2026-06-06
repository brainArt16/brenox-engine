package realtime

import "sync"

// Hub co-ordinates all connected clients and manages message broadcasting.
type Hub struct {
	mu sync.RWMutex

	// Connected clients registry keyed by channel ID.
	channels map[int64]map[*Client]bool

	register   chan *Client
	unregister chan *Client
	broadcast  chan Event

	// Global connection count per user (supports multiple tabs/devices).
	onlineUsers map[int64]int
}

// OnlineUserIDs returns a snapshot of users with at least one active WebSocket connection.
func (h *Hub) OnlineUserIDs() []int64 {
	h.mu.RLock()
	defer h.mu.RUnlock()

	ids := make([]int64, 0, len(h.onlineUsers))
	for userID := range h.onlineUsers {
		ids = append(ids, userID)
	}
	return ids
}