package auth

import (
	"context"
	"time"

	db "github.com/brainart16/brenox/internal/db"
	"github.com/brainart16/brenox/pkg/jwt"
	"github.com/jackc/pgx/v5/pgtype"
)

func (s *Service) ValidateAccessToken(ctx context.Context, tokenString string) (int64, int64, string, error) {
	claims, err := jwt.ValidateToken(tokenString)
	if err != nil {
		return 0, 0, "", err
	}

	if claims.ID != "" {
		revoked, err := s.queries.IsTokenRevoked(ctx, claims.ID)
		if err != nil {
			return 0, 0, "", err
		}
		if revoked {
			return 0, 0, "", jwt.ErrTokenRevoked
		}
	}

	state, err := s.queries.GetUserAuthState(ctx, claims.UserID)
	if err != nil {
		return 0, 0, "", err
	}

	if state.SuspendedAt.Valid {
		return 0, 0, "", ErrAccountSuspended
	}

	if state.TokensInvalidatedAt.Valid && claims.IssuedAt != nil {
		if claims.IssuedAt.Time.Before(state.TokensInvalidatedAt.Time) {
			return 0, 0, "", jwt.ErrTokenRevoked
		}
	}

	keyEnv := claims.KeyEnv
	if claims.AppID > 0 && keyEnv == "" {
		keyEnv = "live"
	}

	return claims.UserID, claims.AppID, keyEnv, nil
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

	state, err := s.queries.GetUserAuthState(ctx, claims.UserID)
	if err != nil {
		return "", ErrInvalidToken
	}

	if state.SuspendedAt.Valid {
		return "", ErrAccountSuspended
	}

	if state.TokensInvalidatedAt.Valid && claims.IssuedAt != nil {
		if claims.IssuedAt.Time.Before(state.TokensInvalidatedAt.Time) {
			return "", ErrInvalidToken
		}
	}

	sandbox := claims.KeyEnv == "sandbox"
	newToken, err := jwt.GenerateSessionToken(claims.UserID, claims.AppID, sandbox)
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
