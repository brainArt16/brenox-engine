package integration

import (
	"context"
	"os"
	"testing"

	"github.com/brainart16/brenox/internal/auth"
	"github.com/brainart16/brenox/internal/database"
	db "github.com/brainart16/brenox/internal/db"
)

func TestAuthRegisterAndLogin(t *testing.T) {
	if os.Getenv("DB_HOST") == "" {
		t.Skip("DB_HOST not set")
	}

	pool, err := database.NewPostgresPool()
	if err != nil {
		t.Fatalf("connect postgres: %v", err)
	}
	t.Cleanup(func() { pool.Close() })

	queries := db.New(pool)
	service := auth.NewService(queries)
	ctx := context.Background()

	email := "integration-" + t.Name() + "@example.com"
	user, err := service.Register(ctx, auth.RegisterRequest{
		Email:    email,
		Username: "integration_user",
		Password: "secret123",
	})
	if err != nil {
		t.Fatalf("register: %v", err)
	}
	if user.ID == 0 {
		t.Fatal("expected user id")
	}

	token, err := service.Login(ctx, auth.LoginRequest{
		Email:    email,
		Password: "secret123",
	})
	if err != nil {
		t.Fatalf("login: %v", err)
	}
	if token == "" {
		t.Fatal("expected token")
	}

	userID, err := service.ValidateAccessToken(ctx, token)
	if err != nil {
		t.Fatalf("validate: %v", err)
	}
	if userID != user.ID {
		t.Fatalf("expected user %d, got %d", user.ID, userID)
	}

	refreshed, err := service.Refresh(ctx, token)
	if err != nil {
		t.Fatalf("refresh: %v", err)
	}
	if refreshed == token {
		t.Fatal("expected rotated token")
	}

	if _, err := service.ValidateAccessToken(ctx, token); err == nil {
		t.Fatal("old token should be revoked after refresh")
	}
}
