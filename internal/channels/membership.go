package channels

import (
	"context"
	"errors"

	db "github.com/brainart16/brenox/internal/db"
	"github.com/jackc/pgx/v5"
)

func (s *Service) JoinChannel(
	ctx context.Context,
	channelID int64,
	userID int64,
) error {
	if _, err := s.queries.GetChannelByID(ctx, channelID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrChannelNotFound
		}
		return err
	}

	isMember, err := s.queries.IsChannelMember(ctx, db.IsChannelMemberParams{
		ChannelID: channelID,
		UserID:    userID,
	})
	if err != nil {
		return err
	}
	if isMember {
		return ErrAlreadyMember
	}

	return s.queries.AddChannelMember(ctx, db.AddChannelMemberParams{
		ChannelID: channelID,
		UserID:    userID,
	})
}

func (s *Service) LeaveChannel(
	ctx context.Context,
	channelID int64,
	userID int64,
) error {
	channel, err := s.queries.GetChannelByID(ctx, channelID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrChannelNotFound
		}
		return err
	}

	if channel.OwnerID == userID {
		return ErrOwnerCannotLeave
	}

	_, err = s.queries.GetChannelMember(ctx, db.GetChannelMemberParams{
		ChannelID: channelID,
		UserID:    userID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrNotMember
		}
		return err
	}

	return s.queries.RemoveChannelMember(ctx, db.RemoveChannelMemberParams{
		ChannelID: channelID,
		UserID:    userID,
	})
}
