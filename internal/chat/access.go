package chat

import "context"

func (s *Service) AssertChannelAccess(ctx context.Context, workspaceID, channelID, userID int64) error {
	return s.assertChannelAccess(ctx, workspaceID, channelID, userID)
}
