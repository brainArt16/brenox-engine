package platformadmin

import (
	"context"
	"errors"
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

type OverviewResponse struct {
	Users      int64 `json:"users"`
	Workspaces int64 `json:"workspaces"`
	Apps       int64 `json:"apps"`
}

type UserListItem struct {
	ID           int64  `json:"id"`
	Email        string `json:"email"`
	Username     string `json:"username"`
	PlatformRole string `json:"platform_role"`
	Suspended    bool   `json:"suspended"`
	CreatedAt    string `json:"created_at"`
}

type UserDetailResponse struct {
	UserListItem
	WorkspaceCount int64 `json:"workspace_count"`
	AppCount       int64 `json:"app_count"`
}

type WorkspaceListItem struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	Slug      string `json:"slug"`
	OwnerID   int64  `json:"owner_id"`
	CreatedAt string `json:"created_at"`
}

type WorkspaceDetailResponse struct {
	WorkspaceListItem
	MemberCount  int64 `json:"member_count"`
	ChannelCount int64 `json:"channel_count"`
}

type AppListItem struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Slug        string `json:"slug"`
	WorkspaceID int64  `json:"workspace_id"`
	OwnerID     int64  `json:"owner_id"`
	OwnerEmail  string `json:"owner_email"`
	CreatedAt   string `json:"created_at"`
}

type AuditLogItem struct {
	ID         int64   `json:"id"`
	UserID     *int64  `json:"user_id,omitempty"`
	AppID      *int64  `json:"app_id,omitempty"`
	Action     string  `json:"action"`
	Method     string  `json:"method"`
	Path       string  `json:"path"`
	IPAddress  *string `json:"ip_address,omitempty"`
	StatusCode *int32  `json:"status_code,omitempty"`
	CreatedAt  string  `json:"created_at"`
}

type UpdateUserRequest struct {
	PlatformRole *string `json:"platform_role"`
	Suspended    *bool   `json:"suspended"`
}

func (s *Service) SyncBootstrapAdmin(ctx context.Context, email string) error {
	if !IsBootstrapAdminEmail(email) {
		return nil
	}
	_, err := s.queries.PromoteUserToAdminByEmail(ctx, email)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil
	}
	return err
}

func (s *Service) SyncBootstrapAdminsFromEnv(ctx context.Context) {
	for _, email := range AdminEmailsFromEnv() {
		_ = s.SyncBootstrapAdmin(ctx, email)
	}
}

func (s *Service) GetOverview(ctx context.Context) (OverviewResponse, error) {
	users, err := s.queries.AdminCountUsers(ctx)
	if err != nil {
		return OverviewResponse{}, err
	}
	workspaces, err := s.queries.AdminCountWorkspaces(ctx)
	if err != nil {
		return OverviewResponse{}, err
	}
	apps, err := s.queries.AdminCountApps(ctx)
	if err != nil {
		return OverviewResponse{}, err
	}
	return OverviewResponse{
		Users:      users,
		Workspaces: workspaces,
		Apps:       apps,
	}, nil
}

func (s *Service) ListUsers(ctx context.Context, search string, limit, offset int32) ([]UserListItem, error) {
	var rows []db.ListUsersAdminRow
	var err error

	search = strings.TrimSpace(search)
	if search == "" {
		rows, err = s.queries.ListUsersAdmin(ctx, db.ListUsersAdminParams{
			Limit:  limit,
			Offset: offset,
		})
		if err != nil {
			return nil, err
		}
	} else {
		searchRows, searchErr := s.queries.SearchUsersAdmin(ctx, db.SearchUsersAdminParams{
			Search: search,
			Limit:  limit,
			Offset: offset,
		})
		if searchErr != nil {
			return nil, searchErr
		}
		rows = make([]db.ListUsersAdminRow, len(searchRows))
		for i, row := range searchRows {
			rows[i] = db.ListUsersAdminRow{
				ID:           row.ID,
				Email:        row.Email,
				Username:     row.Username,
				PlatformRole: row.PlatformRole,
				SuspendedAt:  row.SuspendedAt,
				CreatedAt:    row.CreatedAt,
			}
		}
	}

	items := make([]UserListItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, toUserListItem(row.ID, row.Email, row.Username, row.PlatformRole, row.SuspendedAt, row.CreatedAt))
	}
	return items, nil
}

func (s *Service) GetUser(ctx context.Context, userID int64) (UserDetailResponse, error) {
	row, err := s.queries.GetUserAdmin(ctx, userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return UserDetailResponse{}, ErrNotFound
		}
		return UserDetailResponse{}, err
	}

	workspaceCount, err := s.queries.CountUserWorkspaces(ctx, userID)
	if err != nil {
		return UserDetailResponse{}, err
	}
	appCount, err := s.queries.CountUserApps(ctx, userID)
	if err != nil {
		return UserDetailResponse{}, err
	}

	return UserDetailResponse{
		UserListItem:   toUserListItem(row.ID, row.Email, row.Username, row.PlatformRole, row.SuspendedAt, row.CreatedAt),
		WorkspaceCount: workspaceCount,
		AppCount:       appCount,
	}, nil
}

func (s *Service) UpdateUser(ctx context.Context, actorID, userID int64, req UpdateUserRequest) (UserDetailResponse, error) {
	if req.PlatformRole == nil && req.Suspended == nil {
		return UserDetailResponse{}, ErrInvalidRequest
	}

	if _, err := s.queries.GetUserAdmin(ctx, userID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return UserDetailResponse{}, ErrNotFound
		}
		return UserDetailResponse{}, err
	}

	if req.PlatformRole != nil {
		role := strings.TrimSpace(*req.PlatformRole)
		if !IsValidRole(role) {
			return UserDetailResponse{}, ErrInvalidRole
		}
		if userID == actorID && role != RoleAdmin {
			return UserDetailResponse{}, ErrSelfDemotion
		}
		if _, err := s.queries.UpdateUserPlatformRole(ctx, db.UpdateUserPlatformRoleParams{
			ID:           userID,
			PlatformRole: role,
		}); err != nil {
			return UserDetailResponse{}, err
		}
	}

	if req.Suspended != nil {
		if userID == actorID && *req.Suspended {
			return UserDetailResponse{}, ErrSelfSuspend
		}
		if *req.Suspended {
			if _, err := s.queries.SuspendUser(ctx, userID); err != nil {
				return UserDetailResponse{}, err
			}
		} else if _, err := s.queries.UnsuspendUser(ctx, userID); err != nil {
			return UserDetailResponse{}, err
		}
	}

	return s.GetUser(ctx, userID)
}

func (s *Service) ListWorkspaces(ctx context.Context, limit, offset int32) ([]WorkspaceListItem, error) {
	rows, err := s.queries.ListWorkspacesAdmin(ctx, db.ListWorkspacesAdminParams{
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		return nil, err
	}

	items := make([]WorkspaceListItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, WorkspaceListItem{
			ID:        row.ID,
			Name:      row.Name,
			Slug:      row.Slug,
			OwnerID:   row.OwnerID,
			CreatedAt: formatTime(row.CreatedAt),
		})
	}
	return items, nil
}

func (s *Service) GetWorkspace(ctx context.Context, workspaceID int64) (WorkspaceDetailResponse, error) {
	row, err := s.queries.GetWorkspaceAdmin(ctx, workspaceID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return WorkspaceDetailResponse{}, ErrNotFound
		}
		return WorkspaceDetailResponse{}, err
	}

	memberCount, err := s.queries.CountWorkspaceMembers(ctx, workspaceID)
	if err != nil {
		return WorkspaceDetailResponse{}, err
	}
	channelCount, err := s.queries.CountWorkspaceChannels(ctx, workspaceID)
	if err != nil {
		return WorkspaceDetailResponse{}, err
	}

	return WorkspaceDetailResponse{
		WorkspaceListItem: WorkspaceListItem{
			ID:        row.ID,
			Name:      row.Name,
			Slug:      row.Slug,
			OwnerID:   row.OwnerID,
			CreatedAt: formatTime(row.CreatedAt),
		},
		MemberCount:  memberCount,
		ChannelCount: channelCount,
	}, nil
}

func (s *Service) ListApps(ctx context.Context, limit, offset int32) ([]AppListItem, error) {
	rows, err := s.queries.ListAppsAdmin(ctx, db.ListAppsAdminParams{
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		return nil, err
	}

	items := make([]AppListItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, AppListItem{
			ID:          row.ID,
			Name:        row.Name,
			Slug:        row.Slug,
			WorkspaceID: row.WorkspaceID,
			OwnerID:     row.OwnerID,
			OwnerEmail:  row.OwnerEmail,
			CreatedAt:   formatTime(row.CreatedAt),
		})
	}
	return items, nil
}

func (s *Service) ListAuditLogs(ctx context.Context, limit, offset int32) ([]AuditLogItem, error) {
	rows, err := s.queries.ListAuditLogsAdmin(ctx, db.ListAuditLogsAdminParams{
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		return nil, err
	}

	items := make([]AuditLogItem, 0, len(rows))
	for _, row := range rows {
		item := AuditLogItem{
			ID:        row.ID,
			Action:    row.Action,
			Method:    row.Method,
			Path:      row.Path,
			CreatedAt: formatTime(row.CreatedAt),
		}
		if row.UserID.Valid {
			item.UserID = &row.UserID.Int64
		}
		if row.AppID.Valid {
			item.AppID = &row.AppID.Int64
		}
		if row.IpAddress.Valid {
			item.IPAddress = &row.IpAddress.String
		}
		if row.StatusCode.Valid {
			item.StatusCode = &row.StatusCode.Int32
		}
		items = append(items, item)
	}
	return items, nil
}

func toUserListItem(id int64, email, username, role string, suspendedAt, createdAt pgtype.Timestamptz) UserListItem {
	return UserListItem{
		ID:           id,
		Email:        email,
		Username:     username,
		PlatformRole: role,
		Suspended:    suspendedAt.Valid,
		CreatedAt:    formatTime(createdAt),
	}
}

func formatTime(ts pgtype.Timestamptz) string {
	if !ts.Valid {
		return ""
	}
	return ts.Time.UTC().Format(time.RFC3339Nano)
}
