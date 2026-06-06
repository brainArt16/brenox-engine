package calls

import "context"

type Broadcaster interface {
	PublishCallEvent(eventType string, workspaceID, channelID int64, payload map[string]any)
}

type InviteNotifier interface {
	NotifyCallInvite(
		ctx context.Context,
		workspaceID, channelID, callID, initiatorID int64,
		targetUserID int64,
		initiatorUsername string,
	) error
}

type ChannelAccessChecker interface {
	AssertChannelMember(ctx context.Context, workspaceID, channelID, userID int64) error
}
