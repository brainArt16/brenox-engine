# WebSocket Events

Event envelope:

```json
{
  "type": "event.type",
  "channel_id": 1,
  "payload": {}
}
```

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

## Server → Client

### `message.new`

New message saved to the database.

```json
{
  "type": "message.new",
  "channel_id": 1,
  "payload": {
    "id": 1,
    "sender_id": 2,
    "content": "hello",
    "created_at": "2026-06-06T12:00:00Z"
  }
}
```

### `presence.online` / `presence.offline`

User connection count crossed 0 ↔ 1 globally.

```json
{
  "type": "presence.online",
  "channel_id": 1,
  "payload": { "user_id": 2 }
}
```

### `member.joined` / `member.left`

Channel membership changed via REST join/leave.

```json
{
  "type": "member.joined",
  "channel_id": 1,
  "payload": { "user_id": 2 }
}
```

### `error`

Sent to the client that triggered the failure.

```json
{
  "type": "error",
  "channel_id": 1,
  "payload": { "message": "invalid message payload" }
}
```
