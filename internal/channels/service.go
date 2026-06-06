package channels

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

// CreateChannel creates new channel and automatically adds owner as member.
func (s *Service) CreateChannel(
	ctx context.Context,
	userID int64,
	req CreateChannelRequest,
) (*db.Channel, error) {

	// Create channel.
	channel, err := s.queries.CreateChannel(
		ctx,
		db.CreateChannelParams{
			Name:    req.Name,
			OwnerID: userID,
		},
	)

	if err != nil {
		return nil, err
	}

	// Add creator as member. Ownership and membership are NOT automatically same thing.
	err = s.queries.AddChannelMember(
		ctx,
		db.AddChannelMemberParams{
			ChannelID: channel.ID,
			UserID:    userID,
		},
	)

	if err != nil {
		return nil, err
	}

	return &channel, nil
}

// 	GetChannels returns channels user belongs to.
func (s *Service) GetChannels(
	ctx context.Context,
	userID int64,
) ([]db.GetChannelsByUserRow, error) {

	return s.queries.GetChannelsByUser(
		ctx,
		userID,
	)
}

// IsMember reports whether userID belongs to channelID.
func (s *Service) IsMember(
	ctx context.Context,
	channelID int64,
	userID int64,
) (bool, error) {
	return s.queries.IsChannelMember(ctx, db.IsChannelMemberParams{
		ChannelID: channelID,
		UserID:    userID,
	})
}