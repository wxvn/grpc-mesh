package service

import (
	"context"
	"fmt"
	"net/mail"
	"strings"

	"github.com/google/uuid"
	"github.com/wxvn/grpc-mesh/internal/auth/domain"
)

type RegisterResult struct {
	UserID       uuid.UUID
	AccessToken  string
	RefreshToken string
}

func (s *AuthService) Register(ctx context.Context, email string, password string) (RegisterResult, error) {
	if err := validateEmail(email); err != nil {
		return RegisterResult{}, err
	}

	if password == "" {
		return RegisterResult{}, fmt.Errorf("password is required")
	}

	user, err := s.users.GetByEmail(ctx, email)
	if err != nil {
		return RegisterResult{}, fmt.Errorf("get user by email: %w", err)
	}

	if user != nil {
		return RegisterResult{}, fmt.Errorf("email already exists")
	}

	passwordHash, err := s.passwordHash.Hash(password)
	if err != nil {
		return RegisterResult{}, fmt.Errorf("hash password: %w", err)
	}

	user = &domain.User{
		ID:           uuid.New(),
		Email:        email,
		PasswordHash: passwordHash,
	}

	if err := s.users.CreateUser(ctx, user); err != nil {
		return RegisterResult{}, fmt.Errorf("create user: %w", err)
	}

	accessToken, refreshToken, err := s.issueTokens(ctx, user.ID)
	if err != nil {
		return RegisterResult{}, fmt.Errorf("issue tokens: %w", err)
	}

	return RegisterResult{
		UserID:       user.ID,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func validateEmail(email string) error {
	email = strings.TrimSpace(email)

	if email == "" {
		return fmt.Errorf("email is required")
	}

	if len(email) > 254 {
		return fmt.Errorf("invalid email")
	}

	_, err := mail.ParseAddress(email)
	if err != nil {
		return fmt.Errorf("invalid email")
	}

	return nil
}
