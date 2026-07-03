package users

import (
	"context"
	"errors"
	"strings"
	"time"

	db "github.com/brainart16/brenox/internal/db"
	"github.com/brainart16/brenox/internal/auth"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

var (
	ErrNotFound              = errors.New("user not found")
	ErrUsernameRequired      = errors.New("username is required")
	ErrUsernameTaken         = errors.New("username already taken")
	ErrInvalidPassword       = errors.New("current password is incorrect")
	ErrPasswordRequired      = errors.New("password is required")
	ErrPasswordTooShort      = errors.New("password must be at least 8 characters")
)

type Service struct {
	queries *db.Queries
}

func NewService(queries *db.Queries) *Service {
	return &Service{queries: queries}
}

type ProfileResponse struct {
	ID        int64  `json:"id"`
	Email     string `json:"email"`
	Username  string `json:"username"`
	CreatedAt string `json:"created_at"`
}

type UpdateProfileRequest struct {
	Username string `json:"username"`
}

type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password"`
	NewPassword     string `json:"new_password"`
}

func (s *Service) GetProfile(ctx context.Context, userID int64) (ProfileResponse, error) {
	user, err := s.queries.GetUserByID(ctx, userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ProfileResponse{}, ErrNotFound
		}
		return ProfileResponse{}, err
	}
	return toProfileResponse(user), nil
}

func (s *Service) UpdateProfile(ctx context.Context, userID int64, req UpdateProfileRequest) (ProfileResponse, error) {
	username := strings.TrimSpace(req.Username)
	if username == "" {
		return ProfileResponse{}, ErrUsernameRequired
	}

	if existing, err := s.queries.GetUserByUsername(ctx, username); err == nil {
		if existing.ID != userID {
			return ProfileResponse{}, ErrUsernameTaken
		}
	} else if !errors.Is(err, pgx.ErrNoRows) {
		return ProfileResponse{}, err
	}

	user, err := s.queries.UpdateUserProfile(ctx, db.UpdateUserProfileParams{
		ID:       userID,
		Username: username,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ProfileResponse{}, ErrNotFound
		}
		return ProfileResponse{}, err
	}

	return toProfileResponse(user), nil
}

func (s *Service) ChangePassword(ctx context.Context, userID int64, req ChangePasswordRequest) error {
	current := strings.TrimSpace(req.CurrentPassword)
	next := strings.TrimSpace(req.NewPassword)

	if current == "" || next == "" {
		return ErrPasswordRequired
	}
	if len(next) < 8 {
		return ErrPasswordTooShort
	}

	user, err := s.queries.GetUserByID(ctx, userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrNotFound
		}
		return err
	}

	if err := auth.CheckPassword(current, user.PasswordHash); err != nil {
		return ErrInvalidPassword
	}

	hashed, err := auth.HashPassword(next)
	if err != nil {
		return err
	}

	return s.queries.UpdateUserPassword(ctx, db.UpdateUserPasswordParams{
		ID:           userID,
		PasswordHash: hashed,
	})
}

func toProfileResponse(user db.User) ProfileResponse {
	return ProfileResponse{
		ID:        user.ID,
		Email:     user.Email,
		Username:  user.Username,
		CreatedAt: formatTime(user.CreatedAt),
	}
}

func formatTime(ts pgtype.Timestamptz) string {
	if !ts.Valid {
		return ""
	}
	return ts.Time.UTC().Format(time.RFC3339Nano)
}
