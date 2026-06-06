package realtime

// BroadcastMemberJoined notifies channel subscribers that a user joined.
func (h *Hub) BroadcastMemberJoined(workspaceID, channelID, userID int64) {
	h.broadcast <- Event{
		Type:        "member.joined",
		WorkspaceID: workspaceID,
		ChannelID:   channelID,
		Payload: map[string]any{
			"user_id": userID,
		},
	}
}

// BroadcastMemberLeft notifies channel subscribers that a user left.
func (h *Hub) BroadcastMemberLeft(workspaceID, channelID, userID int64) {
	h.broadcast <- Event{
		Type:        "member.left",
		WorkspaceID: workspaceID,
		ChannelID:   channelID,
		Payload: map[string]any{
			"user_id": userID,
		},
	}
}
