package auth

import (
	"context"
	"time"

	db "github.com/brainart16/brenox/internal/db"
	"github.com/brainart16/brenox/pkg/jwt"
	"github.com/jackc/pgx/v5/pgtype"
)

func (s *Service) ValidateAccessToken(ctx context.Context, tokenString string) (int64, error) {
	claims, err := jwt.ValidateToken(tokenString)
	if err != nil {
		return 0, err
	}

	if claims.ID != "" {
		revoked, err := s.queries.IsTokenRevoked(ctx, claims.ID)
		if err != nil {
			return 0, err
		}
		if revoked {
			return 0, jwt.ErrTokenRevoked
		}
	}

	return claims.UserID, nil
}

func (s *Service) Refresh(ctx context.Context, tokenString string) (string, error) {
	claims, err := jwt.ValidateTokenForRefresh(tokenString)
	if err != nil {
		return "", ErrInvalidToken
	}

	if claims.ID != "" {
		revoked, err := s.queries.IsTokenRevoked(ctx, claims.ID)
		if err != nil {
			return "", ErrInvalidToken
		}
		if revoked {
			return "", ErrInvalidToken
		}
	}

	if _, err := s.queries.GetUserByID(ctx, claims.UserID); err != nil {
		return "", ErrInvalidToken
	}

	newToken, err := jwt.GenerateToken(claims.UserID)
	if err != nil {
		return "", ErrInvalidToken
	}

	if claims.ID != "" {
		expiresAt := time.Now().Add(24 * time.Hour)
		if claims.ExpiresAt != nil {
			expiresAt = claims.ExpiresAt.Time
		}
		_ = s.queries.RevokeToken(ctx, db.RevokeTokenParams{
			Jti:       claims.ID,
			UserID:    claims.UserID,
			ExpiresAt: pgtype.Timestamptz{Time: expiresAt, Valid: true},
		})
	}

	return newToken, nil
}
