package jwt

import (
	"os"
	"time"

	jwtlib "github.com/golang-jwt/jwt/v5"
)

// CustomClaims defines data stored inside JWT token.
type CustomClaims struct {

	// User ID embedded in token.
	UserID int64 `json:"user_id"`

	/*
		RegisteredClaims are standard JWT fields:
		- expiration
		- issued_at
		- issuer
	*/

	jwtlib.RegisteredClaims
}


// GenerateToken creates signed JWT token.
func GenerateToken(
	userID int64,
) (string, error) {


	// Create claims payload.
	claims := CustomClaims{
		UserID: userID,

		RegisteredClaims: jwtlib.RegisteredClaims{

			// 	Token expiration. 24 Hours
			ExpiresAt: jwtlib.NewNumericDate(
				time.Now().Add(24 * time.Hour),
			),

			// 	Token creation timestamp.
			IssuedAt: jwtlib.NewNumericDate(
				time.Now(),
			),
		},
	}

	// Create token object.
	token := jwtlib.NewWithClaims(
		jwtlib.SigningMethodHS256,
		claims,
	)

	// Sign token with secret key.
	return token.SignedString(
		[]byte(os.Getenv("JWT_SECRET")),
	)
}


// ValidateToken parses and validates token.
func ValidateToken(
	tokenString string,
) (*CustomClaims, error) {

	token, err := jwtlib.ParseWithClaims(
		tokenString,
		&CustomClaims{},

		func(token *jwtlib.Token) (interface{}, error) {

			return []byte(
				os.Getenv("JWT_SECRET"),
			), nil
		},
	)

	if err != nil {
		return nil, err
	}

	// Type assertion.Convert generic interface into CustomClaims type.
	claims, ok := token.Claims.(*CustomClaims)

	if !ok || !token.Valid {
		return nil, err
	}

	return claims, nil
}