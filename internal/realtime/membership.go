package realtime

// BroadcastMemberJoined notifies channel subscribers that a user joined.
func (h *Hub) BroadcastMemberJoined(workspaceID, channelID, userID int64) {
	h.Publish(NewOutboundEvent("member.joined", workspaceID, channelID, map[string]any{
		"user_id": userID,
	}))
}

// BroadcastMemberLeft notifies channel subscribers that a user left.
func (h *Hub) BroadcastMemberLeft(workspaceID, channelID, userID int64) {
	h.Publish(NewOutboundEvent("member.left", workspaceID, channelID, map[string]any{
		"user_id": userID,
	}))
}
