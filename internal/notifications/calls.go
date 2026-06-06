package notifications

import "context"

func (s *Service) NotifyCallInvite(
	ctx context.Context,
	workspaceID, channelID, callID, initiatorID, targetUserID int64,
	initiatorUsername string,
) error {
	_, err := s.Create(ctx, CreateInput{
		UserID: targetUserID,
		Type:   TypeCallInvite,
		Title:  "Incoming voice call",
		Body:   initiatorUsername + " started a voice call",
		Data: map[string]any{
			"workspace_id": workspaceID,
			"channel_id":   channelID,
			"call_id":      callID,
			"initiator_id": initiatorID,
		},
	})
	return err
}
