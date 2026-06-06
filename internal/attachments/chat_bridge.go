package attachments

import (
	"context"

	db "github.com/brainart16/brenox/internal/db"
	chat "github.com/brainart16/brenox/internal/chat"
)

type ChatAttacher struct {
	service *Service
}

func NewChatAttacher(service *Service) *ChatAttacher {
	return &ChatAttacher{service: service}
}

func (a *ChatAttacher) AttachOnMessageCreate(
	ctx context.Context,
	workspaceID, channelID int64,
	message db.Message,
	userID int64,
	inputs []chat.AttachmentInput,
) (any, error) {
	if len(inputs) == 0 {
		return nil, nil
	}

	converted := make([]AttachmentInput, 0, len(inputs))
	for _, input := range inputs {
		converted = append(converted, AttachmentInput{
			ObjectKey: input.ObjectKey,
			FileName:  input.FileName,
			MimeType:  input.MimeType,
			SizeBytes: input.SizeBytes,
		})
	}

	return a.service.AttachOnMessageCreate(ctx, workspaceID, channelID, message, userID, converted)
}
