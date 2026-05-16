package auth

import (
	"context"
	"errors"

	db "github.com/brainart16/brenox/internal/db"
	"github.com/brainart16/brenox/pkg/jwt"
)

/*
	Service layer contains:
	- business rules
	- validations
	- orchestration

	NOT HTTP logic.
	NOT raw DB logic.
*/

type Service struct {

	// queries provides access to generated sqlc methods.
	queries *db.Queries
}


// Constructor function.
func NewService(
	queries *db.Queries,
) *Service {

	return &Service{
		queries: queries,
	}
}


// 	 Register creates new user
// 	 Register method belongs to Service. (s *Service)
// 	 Method operates on pointer. (*Service)

func (s *Service) Register(
	ctx context.Context,
	req RegisterRequest,
) (*db.User, error) {


	// Check if user already exists.
	_, err := s.queries.GetUserByEmail(
		ctx,
		req.Email,
	)

	// 	If no error: user already exists.
	if err == nil {
		return nil, errors.New("email already exists")
	}

	// Hash password securely.
	hashedPassword, err := HashPassword(
		req.Password,
	)

	if err != nil {
		return nil, err
	}

	// Create user in database.
	user, err := s.queries.CreateUser(
		ctx,
		db.CreateUserParams{
			Email:        req.Email,
			Username:     req.Username,
			PasswordHash: hashedPassword,
		},
	)

	if err != nil {
		return nil, err
	}

	return &user, nil
}



// Login authenticates user.
func (s *Service) Login(
	ctx context.Context,
	req LoginRequest,
) (string, error) {

	// Find user by email.
	user, err := s.queries.GetUserByEmail(
		ctx,
		req.Email,
	)

	if err != nil {
		return "", errors.New(
			"invalid credentials",
		)
	}

	// Compare password hash.
	err = CheckPassword(
		req.Password,
		user.PasswordHash,
	)

	if err != nil {
		return "", errors.New(
			"invalid credentials",
		)
	}

	// Generate JWT token.
	token, err := jwt.GenerateToken(
		user.ID,
	)

	if err != nil {
		return "", err
	}

	return token, nil
}