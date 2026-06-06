# WebRTC Voice Calling

Brenox provides **signaling only** — audio media flows peer-to-peer (or via TURN). The backend never handles RTP/media.

## Flow

1. **Initiate** — `POST /api/workspaces/:workspace_id/channels/:id/calls`
2. **Join** — `POST /api/calls/:id/join` (channel members only)
3. **Signaling** — WebSocket events on the channel connection (`call.offer`, `call.answer`, `call.ice`)
4. **Leave** — `POST /api/calls/:id/leave`
5. **End** — automatic when last participant leaves; broadcasts `call.end`

## Call states

| Status | Meaning |
|--------|---------|
| `ringing` | Call created, waiting for participants |
| `active` | Two or more participants joined |
| `ended` | Call finished |

## WebSocket signaling events

### Client → Server

Send on the channel WebSocket (`?workspace_id=&channel_id=`). Must be an active call participant.

**`call.offer`**
```json
{
  "type": "call.offer",
  "payload": {
    "call_id": 1,
    "to_user_id": 2,
    "sdp": "v=0..."
  }
}
```

**`call.answer`**
```json
{
  "type": "call.answer",
  "payload": {
    "call_id": 1,
    "to_user_id": 1,
    "sdp": "v=0..."
  }
}
```

**`call.ice`**
```json
{
  "type": "call.ice",
  "payload": {
    "call_id": 1,
    "to_user_id": 2,
    "candidate": "{...}"
  }
}
```

The server adds `from_user_id` and relays via Redis to all channel subscribers. When `to_user_id` is set, only that user receives the event.

### Server → Client

**`call.join`** / **`call.leave`** / **`call.end`**
```json
{
  "type": "call.join",
  "workspace_id": 1,
  "channel_id": 1,
  "payload": {
    "call_id": 1,
    "user_id": 2,
    "status": "active"
  }
}
```

## TURN / STUN (client-side)

WebRTC requires STUN for NAT discovery and often TURN for relay in restrictive networks. Brenox does **not** host TURN — configure a external service in your client/SDK.

### Recommended setup

| Service | Use |
|---------|-----|
| [Google STUN](stun:stun.l.google.com:19302) | Dev / fallback STUN |
| [Twilio TURN](https://www.twilio.com/docs/stun-turn) | Managed TURN |
| [coturn](https://github.com/coturn/coturn) | Self-hosted TURN |

Example client `RTCPeerConnection` config:

```javascript
const pc = new RTCPeerConnection({
  iceServers: [
    { urls: "stun:stun.l.google.com:19302" },
    {
      urls: "turn:turn.example.com:3478",
      username: "<turn-user>",
      credential: "<turn-password>",
    },
  ],
});
```

Pass ICE candidates through `call.ice` WebSocket events.

## Notifications

Starting a call sends `call_invite` notifications to other channel members (see Phase 8 notifications API + `notification.new` WebSocket event).

## Permissions

Only **channel members** may initiate, join, or send signaling for a channel call.
