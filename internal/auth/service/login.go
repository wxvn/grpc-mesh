package service

import (
	"context"
	"fmt"
)

type LoginResult struct {
	AccessToken  string
	RefreshToken string
}

func (s *AuthService) Login(ctx context.Context, email string, password string) (LoginResult, error) {
	user, err := s.users.GetByEmail(ctx, email)
	if err != nil || user == nil {
		return LoginResult{}, fmt.Errorf("invalid credentials")
	}

	if err := s.passwordHash.Compare(user.PasswordHash, password); err != nil {
		return LoginResult{}, fmt.Errorf("invalid credentials")
	}

	accessToken, refreshToken, err := s.issueTokens(ctx, user.ID)
	if err != nil {
		return LoginResult{}, fmt.Errorf("issue tokens: %w", err)
	}

	return LoginResult{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}
