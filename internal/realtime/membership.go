package realtime

// BroadcastMemberJoined notifies channel subscribers that a user joined.
func (h *Hub) BroadcastMemberJoined(channelID, userID int64) {
	h.broadcast <- Event{
		Type:      "member.joined",
		ChannelID: channelID,
		Payload: map[string]any{
			"user_id": userID,
		},
	}
}

// BroadcastMemberLeft notifies channel subscribers that a user left.
func (h *Hub) BroadcastMemberLeft(channelID, userID int64) {
	h.broadcast <- Event{
		Type:      "member.left",
		ChannelID: channelID,
		Payload: map[string]any{
			"user_id": userID,
		},
	}
}
