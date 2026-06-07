package jwt

import (
	"os"
	"testing"
	"time"

	jwtlib "github.com/golang-jwt/jwt/v5"
)

func TestValidateTokenForRefreshAllowsRecentlyExpired(t *testing.T) {
	os.Setenv("JWT_SECRET", "test-secret")
	os.Setenv("JWT_REFRESH_GRACE_HOURS", "24")

	claims := CustomClaims{
		UserID: 42,
		RegisteredClaims: jwtlib.RegisteredClaims{
			ExpiresAt: jwtlib.NewNumericDate(time.Now().Add(-time.Hour)),
			IssuedAt:  jwtlib.NewNumericDate(time.Now().Add(-2 * time.Hour)),
		},
	}
	token := jwtlib.NewWithClaims(jwtlib.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte("test-secret"))
	if err != nil {
		t.Fatal(err)
	}

	got, err := ValidateTokenForRefresh(signed)
	if err != nil {
		t.Fatalf("expected refresh allowed, got %v", err)
	}
	if got.UserID != 42 {
		t.Fatalf("expected user 42, got %d", got.UserID)
	}
}

func TestValidateTokenRejectsExpired(t *testing.T) {
	os.Setenv("JWT_SECRET", "test-secret")

	claims := CustomClaims{
		UserID: 1,
		RegisteredClaims: jwtlib.RegisteredClaims{
			ExpiresAt: jwtlib.NewNumericDate(time.Now().Add(-time.Hour)),
		},
	}
	token := jwtlib.NewWithClaims(jwtlib.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte("test-secret"))
	if err != nil {
		t.Fatal(err)
	}

	if _, err := ValidateToken(signed); err == nil {
		t.Fatal("expected expired token to be rejected")
	}
}
