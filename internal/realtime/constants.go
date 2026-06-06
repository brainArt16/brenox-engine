package realtime

import "time"

// WebSocket keepalive timings (gorilla/websocket defaults pattern).
const (
	writeWait  = 10 * time.Second
	pongWait   = 60 * time.Second
	pingPeriod = (pongWait * 9) / 10
)
