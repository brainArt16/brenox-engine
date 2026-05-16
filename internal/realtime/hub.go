package realtime

// Hub co-ordinates all connected clients and manages message broadcasting.
type Hub struct {
	// Connected clients registry.
	clients map[*Client]bool

	// Register new clients. Add new clients to the registry.
	register chan *Client

	// Unregister clients. Remove disconnected clients from the registry.
	unregister chan *Client

	// Global broadcast channel. Messages sent here are broadcast to all clients.
	broadcast chan Message
}