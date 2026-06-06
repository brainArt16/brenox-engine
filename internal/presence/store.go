package presence

import "context"

type Store interface {
	Connect(ctx context.Context, userID, workspaceID, channelID int64) (ConnectResult, error)
	Disconnect(ctx context.Context, userID, workspaceID, channelID int64) (DisconnectResult, error)
	Touch(ctx context.Context, userID int64) error
	SetStatus(ctx context.Context, userID int64, status string) (UserPresence, error)
	Get(ctx context.Context, userID int64) (UserPresence, error)
	ListOnline(ctx context.Context) ([]UserPresence, error)
	ListWorkspaceOnlineUserIDs(ctx context.Context, workspaceID int64) ([]int64, error)
	ActiveChannels(ctx context.Context, userID int64) ([]ChannelRef, error)
}
