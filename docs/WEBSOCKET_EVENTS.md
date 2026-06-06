# WebSocket Events

Connect: `GET /api/ws?workspace_id=&channel_id=` with JWT via `Authorization: Bearer …` or `?token=`.

All outbound events use a standard envelope:

```json
{
  "type": "event.type",
  "workspace_id": 1,
  "channel_id": 1,
  "event_id": "1717670400000000000-1",
  "timestamp": "2026-06-06T12:00:00.123456789Z",
  "payload": {}
}
```

| Field | Description |
|-------|-------------|
| `type` | Event name (see catalog below) |
| `workspace_id` | Workspace scope |
| `channel_id` | Channel scope for delivery |
| `event_id` | Unique ID for deduplication / tracing |
| `timestamp` | UTC RFC3339Nano |
| `payload` | Event-specific body |

Inbound client events only require `type` and `payload`; the server sets scope from the connection.

## Client → Server

### `message.send`

Send a chat message. Persisted and broadcast as `message.new`.

```json
{
  "type": "message.send",
  "payload": { "content": "hello" }
}
```

Errors return `type: "error"` to the sender only.

### `typing.start` / `typing.stop`

Ephemeral typing indicators (not stored). Broadcast to other channel members.

```json
{ "type": "typing.start" }
```

```json
{ "type": "typing.stop" }
```

## Server → Client

### `message.new`

New message saved to the database.

```json
{
  "type": "message.new",
  "workspace_id": 1,
  "channel_id": 1,
  "event_id": "…",
  "timestamp": "2026-06-06T12:00:00Z",
  "payload": {
    "id": 1,
    "sender_id": 2,
    "content": "hello",
    "created_at": "2026-06-06T12:00:00Z"
  }
}
```

### `message.updated`

Message metadata changed (attachments added).

```json
{
  "type": "message.updated",
  "workspace_id": 1,
  "channel_id": 1,
  "payload": {
    "id": 1,
    "channel_id": 1,
    "sender_id": 2,
    "content": "see attached",
    "created_at": "2026-06-06T12:00:00Z",
    "attachments": [
      {
        "id": 1,
        "file_name": "doc.pdf",
        "mime_type": "application/pdf",
        "size_bytes": 1024,
        "url": "https://…",
        "created_at": "2026-06-06T12:00:01Z"
      }
    ]
  }
}
```

### `typing.start` / `typing.stop`

Another member started or stopped typing.

```json
{
  "type": "typing.start",
  "workspace_id": 1,
  "channel_id": 1,
  "payload": { "user_id": 2 }
}
```

### `presence.online` / `presence.offline`

User's global connection count crossed 0 ↔ 1 (tracked in Redis across all nodes).

```json
{
  "type": "presence.online",
  "workspace_id": 1,
  "channel_id": 1,
  "payload": { "user_id": 2 }
}
```

### `presence.status`

User changed status via `PATCH /api/users/me/status` (`online`, `away`, `offline`).

```json
{
  "type": "presence.status",
  "workspace_id": 1,
  "channel_id": 1,
  "payload": {
    "user_id": 2,
    "status": "away",
    "last_seen": "2026-06-06T12:00:00Z"
  }
}
```

### `member.joined` / `member.left`

Channel membership changed via REST join/leave.

```json
{
  "type": "member.joined",
  "workspace_id": 1,
  "channel_id": 1,
  "payload": { "user_id": 2 }
}
```

### `notification.new`

Delivered to the target user's WebSocket connections only (via Redis `user:{id}:notifications` when scaled).

```json
{
  "type": "notification.new",
  "event_id": "…",
  "timestamp": "2026-06-06T12:00:00Z",
  "payload": {
    "id": 1,
    "type": "mention",
    "title": "You were mentioned",
    "body": "alice mentioned you in a message",
    "data": {
      "workspace_id": 1,
      "channel_id": 2,
      "message_id": 10,
      "sender_id": 3
    },
    "read": false,
    "created_at": "2026-06-06T12:00:00Z"
  }
}
```

Notification types: `mention`, `reply`, `channel_invite`, `workspace_invite`, `call_invite`.

### `error`

Sent to the client that triggered the failure.

```json
{
  "type": "error",
  "workspace_id": 1,
  "channel_id": 1,
  "payload": { "message": "invalid message payload" }
}
```

## Presence API

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/presence` | All globally online users (with status, connection_count, last_seen) |
| GET | `/api/workspaces/:workspace_id/presence` | Online members in a workspace |
| PATCH | `/api/users/me/status` | Set status: `online`, `away`, or `offline` |

Presence state is stored in Redis when `REDIS_URL` is set; otherwise in-memory on a single node. Keys expire after `PRESENCE_TTL_SECONDS` unless refreshed by WebSocket heartbeat pings.

## Connection limits

If per-user or per-IP limits are exceeded, the HTTP upgrade returns `429 Too Many Requests`. Configure via `WS_MAX_CONNECTIONS_PER_USER` and `WS_MAX_CONNECTIONS_PER_IP`.
