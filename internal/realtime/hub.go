package realtime



// Hub co-ordinates all connected clients and manages message broadcasting.
type Hub struct {
	// Connected clients registry. Channel Id 
	channels map[int64]map[*Client]bool

	// Register new clients. Add new clients to the registry.
	register chan *Client

	// Unregister clients. Remove disconnected clients from the registry.
	unregister chan *Client

	// Global broadcast channel. Event (Messages) sent here are broadcast to all clients.
	broadcast chan Event

	// Track online users. Use int instead of bool, one user can have multiple connections (e.g. multiple browser tabs).
	onlineUsers map[int64]int
}