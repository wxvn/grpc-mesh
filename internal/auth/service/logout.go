package service

import (
	"context"
	"fmt"
)

func (s *AuthService) Logout(ctx context.Context, refreshToken string) error {
	if refreshToken == "" {
		return fmt.Errorf("refresh token is required")
	}

	refreshTokenHash := s.tokenManager.HashRefreshToken(refreshToken)

	if err := s.tokens.DeleteByHash(ctx, refreshTokenHash); err != nil {
		return fmt.Errorf("logout: %w", err)
	}

	return nil
}
