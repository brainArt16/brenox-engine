package channels

import (
	"context"
	"errors"

	db "github.com/brainart16/brenox/internal/db"
	"github.com/brainart16/brenox/internal/calls"
	"github.com/jackc/pgx/v5"
)

func (s *Service) AssertChannelMember(ctx context.Context, workspaceID, channelID, userID int64) error {
	isWorkspaceMember, err := s.queries.IsWorkspaceMember(ctx, db.IsWorkspaceMemberParams{
		WorkspaceID: workspaceID,
		UserID:      userID,
	})
	if err != nil {
		return err
	}
	if !isWorkspaceMember {
		return calls.ErrNotMember
	}

	if _, err := s.queries.GetChannelInWorkspace(ctx, db.GetChannelInWorkspaceParams{
		ID:          channelID,
		WorkspaceID: workspaceID,
	}); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return calls.ErrChannelNotFound
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
	if !isMember {
		return calls.ErrNotMember
	}
	return nil
}
