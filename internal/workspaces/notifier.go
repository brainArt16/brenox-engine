package workspaces

import "context"

type InviteNotifier interface {
	HandleWorkspaceInvite(
		ctx context.Context,
		workspaceID, actorID, targetUserID int64,
		workspaceName, actorUsername string,
	) error
}
