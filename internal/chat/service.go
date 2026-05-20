package chat

import (
	"context"

	db "github.com/brainart16/brenox/internal/db"
)

type Service struct {
	queries *db.Queries
}

func NewService(
	queries *db.Queries,
) *Service {

	return &Service{
		queries: queries,
	}
}


// Persist message into database.
func (s *Service) SaveMessage(
	ctx context.Context,
	channelID int64,
	senderID int64,
	content string,
) (*db.Message, error) {

	message, err := s.queries.CreateMessage(
		ctx,
		db.CreateMessageParams{
			ChannelID: channelID,
			SenderID:  senderID,
			Content:   content,
		},
	)

	if err != nil {
		return nil, err
	}

	return &message, nil
}