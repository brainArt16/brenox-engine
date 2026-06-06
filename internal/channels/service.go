package channels

import (
	"context"

	db "github.com/brainart16/brenox/internal/db"
	"github.com/brainart16/brenox/internal/authz"
)

type Service struct {
	queries *db.Queries
	authz   *authz.Service
}

func NewService(queries *db.Queries, authzService *authz.Service) *Service {
	return &Service{
		queries: queries,
		authz:   authzService,
	}
}

// CreateChannel creates a channel inside a workspace and adds the creator as a member.
func (s *Service) CreateChannel(
	ctx context.Context,
	workspaceID int64,
	userID int64,
	req CreateChannelRequest,
) (*db.Channel, error) {
	if err := s.assertWorkspaceMember(ctx, workspaceID, userID); err != nil {
		return nil, err
	}

	if err := s.authz.Can(ctx, workspaceID, userID, authz.ActionCreateChannel, authz.Options{}); err != nil {
		return nil, mapAuthzError(err)
	}

	isReadOnly := req.IsReadOnly
	if isReadOnly {
		role, err := s.authz.WorkspaceRole(ctx, workspaceID, userID)
		if err != nil {
			return nil, mapAuthzError(err)
		}
		if role != authz.RoleOwner && role != authz.RoleAdmin {
			return nil, ErrForbidden
		}
	}

	channel, err := s.queries.CreateChannel(ctx, db.CreateChannelParams{
		Name:        req.Name,
		OwnerID:     userID,
		WorkspaceID: workspaceID,
		IsReadOnly:  isReadOnly,
	})
	if err != nil {
		if isDuplicateChannelName(err) {
			return nil, ErrDuplicateChannelName
		}
		return nil, err
	}

	err = s.queries.AddChannelMember(ctx, db.AddChannelMemberParams{
		ChannelID: channel.ID,
		UserID:    userID,
	})
	if err != nil {
		return nil, err
	}

	return &channel, nil
}

// GetChannels returns channels in a workspace that the user belongs to.
func (s *Service) GetChannels(
	ctx context.Context,
	workspaceID int64,
	userID int64,
) ([]db.GetChannelsByWorkspaceAndUserRow, error) {
	if err := s.assertWorkspaceMember(ctx, workspaceID, userID); err != nil {
		return nil, err
	}

	return s.queries.GetChannelsByWorkspaceAndUser(ctx, db.GetChannelsByWorkspaceAndUserParams{
		UserID:      userID,
		WorkspaceID: workspaceID,
	})
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

// IsWorkspaceMember reports whether userID belongs to workspaceID.
func (s *Service) IsWorkspaceMember(
	ctx context.Context,
	workspaceID int64,
	userID int64,
) (bool, error) {
	return s.queries.IsWorkspaceMember(ctx, db.IsWorkspaceMemberParams{
		WorkspaceID: workspaceID,
		UserID:      userID,
	})
}

// GetChannelInWorkspace returns a channel scoped to a workspace.
func (s *Service) GetChannelInWorkspace(
	ctx context.Context,
	workspaceID int64,
	channelID int64,
) (db.GetChannelInWorkspaceRow, error) {
	return s.getChannelInWorkspace(ctx, workspaceID, channelID)
}

func mapAuthzError(err error) error {
	if err == authz.ErrForbidden {
		return ErrForbidden
	}
	return err
}
