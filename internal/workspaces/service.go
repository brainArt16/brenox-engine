package workspaces

import (
	"context"
	"errors"
	"regexp"
	"strings"
	"time"

	db "github.com/brainart16/brenox/internal/db"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type Service struct {
	queries *db.Queries
}

func NewService(queries *db.Queries) *Service {
	return &Service{queries: queries}
}

func (s *Service) CreateWorkspace(
	ctx context.Context,
	userID int64,
	req CreateWorkspaceRequest,
) (*db.Workspace, error) {
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return nil, errors.New("workspace name is required")
	}

	slug, err := normalizeSlug(req.Slug, name)
	if err != nil {
		return nil, err
	}

	if _, err := s.queries.GetWorkspaceBySlug(ctx, slug); err == nil {
		return nil, ErrSlugTaken
	} else if !errors.Is(err, pgx.ErrNoRows) {
		return nil, err
	}

	workspace, err := s.queries.CreateWorkspace(ctx, db.CreateWorkspaceParams{
		Name:    name,
		Slug:    slug,
		OwnerID: userID,
	})
	if err != nil {
		return nil, err
	}

	err = s.queries.AddWorkspaceMember(ctx, db.AddWorkspaceMemberParams{
		WorkspaceID: workspace.ID,
		UserID:      userID,
		Role:        "owner",
	})
	if err != nil {
		return nil, err
	}

	return &workspace, nil
}

func (s *Service) ListWorkspaces(
	ctx context.Context,
	userID int64,
) ([]db.GetWorkspacesByUserRow, error) {
	return s.queries.GetWorkspacesByUser(ctx, userID)
}

func (s *Service) GetWorkspace(
	ctx context.Context,
	workspaceID int64,
	userID int64,
) (*db.Workspace, error) {
	if err := s.assertMember(ctx, workspaceID, userID); err != nil {
		return nil, err
	}

	workspace, err := s.queries.GetWorkspaceByID(ctx, workspaceID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	return &workspace, nil
}

func (s *Service) IsMember(
	ctx context.Context,
	workspaceID int64,
	userID int64,
) (bool, error) {
	return s.queries.IsWorkspaceMember(ctx, db.IsWorkspaceMemberParams{
		WorkspaceID: workspaceID,
		UserID:      userID,
	})
}

func (s *Service) assertMember(ctx context.Context, workspaceID, userID int64) error {
	isMember, err := s.IsMember(ctx, workspaceID, userID)
	if err != nil {
		return err
	}
	if !isMember {
		return ErrNotMember
	}
	return nil
}

var slugPattern = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)

func normalizeSlug(rawSlug, name string) (string, error) {
	slug := strings.TrimSpace(strings.ToLower(rawSlug))
	if slug == "" {
		slug = strings.ToLower(name)
		slug = strings.ReplaceAll(slug, " ", "-")
		slug = regexp.MustCompile(`[^a-z0-9-]+`).ReplaceAllString(slug, "")
		slug = regexp.MustCompile(`-+`).ReplaceAllString(slug, "-")
		slug = strings.Trim(slug, "-")
	}

	if slug == "" || !slugPattern.MatchString(slug) {
		return "", ErrInvalidSlug
	}

	return slug, nil
}

func ToWorkspaceResponse(workspace db.Workspace) WorkspaceResponse {
	return WorkspaceResponse{
		ID:        workspace.ID,
		Name:      workspace.Name,
		Slug:      workspace.Slug,
		OwnerID:   workspace.OwnerID,
		CreatedAt: formatTime(workspace.CreatedAt),
	}
}

func ToWorkspaceListItem(row db.GetWorkspacesByUserRow) WorkspaceResponse {
	return WorkspaceResponse{
		ID:        row.ID,
		Name:      row.Name,
		Slug:      row.Slug,
		OwnerID:   row.OwnerID,
		CreatedAt: formatTime(row.CreatedAt),
	}
}

func formatTime(ts pgtype.Timestamptz) string {
	if !ts.Valid {
		return ""
	}
	return ts.Time.Format(time.RFC3339)
}
