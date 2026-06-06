package authz

import (
	"context"
	"errors"

	db "github.com/brainart16/brenox/internal/db"
	"github.com/jackc/pgx/v5"
)

type Service struct {
	queries *db.Queries
}

func NewService(queries *db.Queries) *Service {
	return &Service{queries: queries}
}

func (s *Service) Can(
	ctx context.Context,
	workspaceID int64,
	userID int64,
	action Action,
	opts Options,
) error {
	role, err := s.WorkspaceRole(ctx, workspaceID, userID)
	if err != nil {
		return err
	}

	if opts.ReadOnlyChannel && action == ActionSendMessage && opts.ChannelRole == "" {
		channelRole, err := s.channelRole(ctx, opts.channelID(), userID)
		if err != nil && !errors.Is(err, pgx.ErrNoRows) {
			return err
		}
		if err == nil {
			if parsed, ok := ParseRole(channelRole); ok {
				opts.ChannelRole = parsed
			}
		}
	}

	if Allowed(role, action, opts) {
		return nil
	}
	return ErrForbidden
}

func (s *Service) WorkspaceRole(
	ctx context.Context,
	workspaceID int64,
	userID int64,
) (Role, error) {
	member, err := s.queries.GetWorkspaceMember(ctx, db.GetWorkspaceMemberParams{
		WorkspaceID: workspaceID,
		UserID:      userID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", ErrForbidden
		}
		return "", err
	}

	role, ok := ParseRole(member.Role)
	if !ok {
		return "", ErrForbidden
	}
	return role, nil
}

func (s *Service) CanAssignRole(actor, target Role) bool {
	return canAssignRole(actor, target)
}

func (s *Service) channelRole(ctx context.Context, channelID, userID int64) (string, error) {
	if channelID == 0 {
		return "", pgx.ErrNoRows
	}
	row, err := s.queries.GetChannelRole(ctx, db.GetChannelRoleParams{
		ChannelID: channelID,
		UserID:    userID,
	})
	if err != nil {
		return "", err
	}
	return row.Role, nil
}

// Options helper — channel ID stored via unexported field pattern.
func MessageOptions(channelID int64, readOnly bool) Options {
	return Options{
		ReadOnlyChannel: readOnly,
		channelIDValue:  channelID,
	}
}

func (o Options) channelID() int64 {
	return o.channelIDValue
}
