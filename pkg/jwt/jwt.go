package jwt

import (
	"errors"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	jwtlib "github.com/golang-jwt/jwt/v5"
)

var ErrTokenInvalid = errors.New("invalid token")
var ErrTokenRevoked = errors.New("token revoked")

// CustomClaims defines data stored inside JWT token.
type CustomClaims struct {
	UserID int64  `json:"user_id"`
	AppID  int64  `json:"app_id,omitempty"`
	KeyEnv string `json:"key_env,omitempty"`
	jwtlib.RegisteredClaims
}

func tokenSecret() []byte {
	return []byte(os.Getenv("JWT_SECRET"))
}

func accessTokenTTL() time.Duration {
	raw := strings.TrimSpace(os.Getenv("JWT_ACCESS_TTL_HOURS"))
	if raw == "" {
		return 24 * time.Hour
	}
	hours, err := strconv.Atoi(raw)
	if err != nil || hours <= 0 {
		return 24 * time.Hour
	}
	return time.Duration(hours) * time.Hour
}

func refreshGracePeriod() time.Duration {
	raw := strings.TrimSpace(os.Getenv("JWT_REFRESH_GRACE_HOURS"))
	if raw == "" {
		return 7 * 24 * time.Hour
	}
	hours, err := strconv.Atoi(raw)
	if err != nil || hours <= 0 {
		return 7 * 24 * time.Hour
	}
	return time.Duration(hours) * time.Hour
}

// GenerateToken creates a signed access JWT with a unique token ID (jti).
func GenerateToken(userID int64) (string, error) {
	return GenerateSessionToken(userID, 0, false)
}

// GenerateSessionToken creates a JWT for an end-user session. appID is set for embed clients.
// sandbox selects the sandbox data lane when appID is set.
func GenerateSessionToken(userID, appID int64, sandbox bool) (string, error) {
	now := time.Now()
	claims := CustomClaims{
		UserID: userID,
		AppID:  appID,
		RegisteredClaims: jwtlib.RegisteredClaims{
			ID:        uuid.NewString(),
			ExpiresAt: jwtlib.NewNumericDate(now.Add(accessTokenTTL())),
			IssuedAt:  jwtlib.NewNumericDate(now),
		},
	}
	if appID > 0 {
		claims.KeyEnv = "live"
		if sandbox {
			claims.KeyEnv = "sandbox"
		}
	}

	token := jwtlib.NewWithClaims(jwtlib.SigningMethodHS256, claims)
	return token.SignedString(tokenSecret())
}

// KeyEnv returns the embed environment claim, defaulting to live for legacy tokens.
func KeyEnv(tokenString string) string {
	claims, err := parseClaims(tokenString)
	if err != nil || claims.AppID == 0 {
		return ""
	}
	if claims.KeyEnv == "" {
		return "live"
	}
	return claims.KeyEnv
}
func TokenExpiry(tokenString string) (time.Time, bool) {
	claims, err := parseClaims(tokenString)
	if err != nil || claims.ExpiresAt == nil {
		return time.Time{}, false
	}
	return claims.ExpiresAt.Time, true
}

// TokenID returns the jti claim when present.
func TokenID(tokenString string) (string, error) {
	claims, err := parseClaims(tokenString)
	if err != nil {
		return "", err
	}
	return claims.ID, nil
}

// ValidateToken parses and validates a non-expired token.
func ValidateToken(tokenString string) (*CustomClaims, error) {
	claims, err := parseClaims(tokenString)
	if err != nil {
		return nil, err
	}
	if claims.ExpiresAt != nil && claims.ExpiresAt.Before(time.Now()) {
		return nil, jwtlib.ErrTokenExpired
	}
	return claims, nil
}

// ValidateTokenForRefresh accepts valid or recently expired tokens for refresh.
func ValidateTokenForRefresh(tokenString string) (*CustomClaims, error) {
	claims, err := parseClaims(tokenString)
	if err != nil {
		return nil, err
	}

	if claims.ExpiresAt == nil {
		return claims, nil
	}

	expiredFor := time.Since(claims.ExpiresAt.Time)
	if expiredFor <= 0 {
		return claims, nil
	}
	if expiredFor > refreshGracePeriod() {
		return nil, ErrTokenInvalid
	}
	return claims, nil
}

func parseClaims(tokenString string) (*CustomClaims, error) {
	claims := &CustomClaims{}
	token, err := jwtlib.ParseWithClaims(
		tokenString,
		claims,
		func(token *jwtlib.Token) (any, error) {
			return tokenSecret(), nil
		},
	)
	if err != nil {
		if errors.Is(err, jwtlib.ErrTokenExpired) {
			return claims, nil
		}
		return nil, err
	}
	if !token.Valid {
		return nil, ErrTokenInvalid
	}
	return claims, nil
}
