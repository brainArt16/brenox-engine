package chat

import (
	"context"

	db "github.com/brainart16/brenox/internal/db"
)

type Notifier interface {
	HandleMessageCreated(
		ctx context.Context,
		workspaceID, channelID, senderID int64,
		message db.Message,
		replyToMessageID *int64,
		senderUsername string,
	) error
}
