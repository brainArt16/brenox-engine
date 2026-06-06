package channels

import (
	"context"
	"errors"

	db "github.com/brainart16/brenox/internal/db"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

func (s *Service) assertWorkspaceMember(
	ctx context.Context,
	workspaceID int64,
	userID int64,
) error {
	isMember, err := s.queries.IsWorkspaceMember(ctx, db.IsWorkspaceMemberParams{
		WorkspaceID: workspaceID,
		UserID:      userID,
	})
	if err != nil {
		return err
	}
	if !isMember {
		return ErrNotWorkspaceMember
	}
	return nil
}

func (s *Service) getChannelInWorkspace(
	ctx context.Context,
	workspaceID int64,
	channelID int64,
) (db.GetChannelInWorkspaceRow, error) {
	channel, err := s.queries.GetChannelInWorkspace(ctx, db.GetChannelInWorkspaceParams{
		ID:          channelID,
		WorkspaceID: workspaceID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return db.GetChannelInWorkspaceRow{}, ErrChannelNotFound
		}
		return db.GetChannelInWorkspaceRow{}, err
	}
	return channel, nil
}

func isDuplicateChannelName(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}
