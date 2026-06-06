package channels

import (
	"context"
	"errors"

	db "github.com/brainart16/brenox/internal/db"
	"github.com/jackc/pgx/v5"
)

func (s *Service) JoinChannel(
	ctx context.Context,
	workspaceID int64,
	channelID int64,
	userID int64,
) error {
	if err := s.assertWorkspaceMember(ctx, workspaceID, userID); err != nil {
		return err
	}

	if _, err := s.getChannelInWorkspace(ctx, workspaceID, channelID); err != nil {
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
	workspaceID int64,
	channelID int64,
	userID int64,
) error {
	if err := s.assertWorkspaceMember(ctx, workspaceID, userID); err != nil {
		return err
	}

	channel, err := s.getChannelInWorkspace(ctx, workspaceID, channelID)
	if err != nil {
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
