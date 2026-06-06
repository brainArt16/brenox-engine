package workspaces

import (
	"context"
	"errors"
	"strings"

	db "github.com/brainart16/brenox/internal/db"
	"github.com/brainart16/brenox/internal/authz"
	"github.com/jackc/pgx/v5"
)

func (s *Service) ListMembers(
	ctx context.Context,
	workspaceID int64,
	userID int64,
) ([]MemberResponse, error) {
	if err := s.assertMember(ctx, workspaceID, userID); err != nil {
		return nil, err
	}

	rows, err := s.queries.ListWorkspaceMembers(ctx, workspaceID)
	if err != nil {
		return nil, err
	}

	items := make([]MemberResponse, 0, len(rows))
	for _, row := range rows {
		items = append(items, ToMemberResponse(row))
	}
	return items, nil
}

func (s *Service) AddMember(
	ctx context.Context,
	workspaceID int64,
	actorID int64,
	req AddMemberRequest,
) (*MemberResponse, error) {
	if err := s.authz.Can(ctx, workspaceID, actorID, authz.ActionInviteMember, authz.Options{}); err != nil {
		return nil, mapAuthzError(err)
	}

	targetRole, ok := authz.ParseRole(strings.TrimSpace(req.Role))
	if !ok || targetRole == authz.RoleOwner {
		return nil, ErrInvalidRole
	}

	actorRole, err := s.authz.WorkspaceRole(ctx, workspaceID, actorID)
	if err != nil {
		return nil, mapAuthzError(err)
	}
	if !s.authz.CanAssignRole(actorRole, targetRole) {
		return nil, ErrForbidden
	}

	email := strings.TrimSpace(strings.ToLower(req.Email))
	if email == "" {
		return nil, errors.New("email is required")
	}

	user, err := s.queries.GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	if _, err := s.queries.GetWorkspaceMember(ctx, db.GetWorkspaceMemberParams{
		WorkspaceID: workspaceID,
		UserID:      user.ID,
	}); err == nil {
		return nil, ErrAlreadyMember
	} else if !errors.Is(err, pgx.ErrNoRows) {
		return nil, err
	}

	err = s.queries.AddWorkspaceMember(ctx, db.AddWorkspaceMemberParams{
		WorkspaceID: workspaceID,
		UserID:      user.ID,
		Role:        string(targetRole),
	})
	if err != nil {
		return nil, err
	}

	member, err := s.queries.GetWorkspaceMember(ctx, db.GetWorkspaceMemberParams{
		WorkspaceID: workspaceID,
		UserID:      user.ID,
	})
	if err != nil {
		return nil, err
	}

	return &MemberResponse{
		UserID:    user.ID,
		Username:  user.Username,
		Email:     user.Email,
		Role:      member.Role,
		CreatedAt: formatTime(member.CreatedAt),
	}, nil
}

func (s *Service) RemoveMember(
	ctx context.Context,
	workspaceID int64,
	actorID int64,
	targetUserID int64,
) error {
	if err := s.authz.Can(ctx, workspaceID, actorID, authz.ActionRemoveMember, authz.Options{}); err != nil {
		return mapAuthzError(err)
	}

	target, err := s.queries.GetWorkspaceMember(ctx, db.GetWorkspaceMemberParams{
		WorkspaceID: workspaceID,
		UserID:      targetUserID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrNotMember
		}
		return err
	}

	if target.Role == string(authz.RoleOwner) {
		return ErrCannotModifyOwner
	}

	return s.queries.RemoveWorkspaceMember(ctx, db.RemoveWorkspaceMemberParams{
		WorkspaceID: workspaceID,
		UserID:      targetUserID,
	})
}

func (s *Service) UpdateMemberRole(
	ctx context.Context,
	workspaceID int64,
	actorID int64,
	targetUserID int64,
	req UpdateMemberRoleRequest,
) (*MemberResponse, error) {
	if err := s.authz.Can(ctx, workspaceID, actorID, authz.ActionChangeMemberRole, authz.Options{}); err != nil {
		return nil, mapAuthzError(err)
	}

	newRole, ok := authz.ParseRole(strings.TrimSpace(req.Role))
	if !ok || newRole == authz.RoleOwner {
		return nil, ErrInvalidRole
	}

	actorRole, err := s.authz.WorkspaceRole(ctx, workspaceID, actorID)
	if err != nil {
		return nil, mapAuthzError(err)
	}
	if !s.authz.CanAssignRole(actorRole, newRole) {
		return nil, ErrForbidden
	}

	target, err := s.queries.GetWorkspaceMember(ctx, db.GetWorkspaceMemberParams{
		WorkspaceID: workspaceID,
		UserID:      targetUserID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotMember
		}
		return nil, err
	}

	if target.Role == string(authz.RoleOwner) {
		return nil, ErrCannotModifyOwner
	}

	err = s.queries.UpdateWorkspaceMemberRole(ctx, db.UpdateWorkspaceMemberRoleParams{
		WorkspaceID: workspaceID,
		UserID:      targetUserID,
		Role:        string(newRole),
	})
	if err != nil {
		return nil, err
	}

	user, err := s.queries.GetUserByID(ctx, targetUserID)
	if err != nil {
		return nil, err
	}

	updated, err := s.queries.GetWorkspaceMember(ctx, db.GetWorkspaceMemberParams{
		WorkspaceID: workspaceID,
		UserID:      targetUserID,
	})
	if err != nil {
		return nil, err
	}

	return &MemberResponse{
		UserID:    user.ID,
		Username:  user.Username,
		Email:     user.Email,
		Role:      updated.Role,
		CreatedAt: formatTime(updated.CreatedAt),
	}, nil
}

func ToMemberResponse(row db.ListWorkspaceMembersRow) MemberResponse {
	return MemberResponse{
		UserID:    row.UserID,
		Username:  row.Username,
		Email:     row.Email,
		Role:      row.Role,
		CreatedAt: formatTime(row.CreatedAt),
	}
}

func mapAuthzError(err error) error {
	if errors.Is(err, authz.ErrForbidden) {
		return ErrForbidden
	}
	return err
}
