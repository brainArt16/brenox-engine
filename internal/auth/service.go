package auth

import (
	"context"
	"errors"

	db "github.com/brainart16/brenox/internal/db"
	"github.com/brainart16/brenox/pkg/jwt"
	"github.com/jackc/pgx/v5"
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

	if err == nil {
		return nil, ErrEmailExists
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrRegistrationFailed
	}

	// Hash password securely.
	hashedPassword, err := HashPassword(
		req.Password,
	)

	if err != nil {
		return nil, ErrRegistrationFailed
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
		return nil, ErrRegistrationFailed
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
		return "", ErrInvalidCredentials
	}

	err = CheckPassword(
		req.Password,
		user.PasswordHash,
	)

	if err != nil {
		return "", ErrInvalidCredentials
	}

	token, err := jwt.GenerateToken(
		user.ID,
	)

	if err != nil {
		return "", ErrInvalidCredentials
	}

	return token, nil
}
