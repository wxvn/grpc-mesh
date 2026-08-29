package service

import (
	"context"
	"fmt"
	"time"

	"github.com/wxvn/grpc-mesh/internal/auth/domain"
)

type RefreshTokensResult struct {
	AccessToken  string
	RefreshToken string
}

func (s *AuthService) RefreshTokens(ctx context.Context, refreshToken string) (RefreshTokensResult, error) {
	if refreshToken == "" {
		return RefreshTokensResult{}, fmt.Errorf("refresh token is required")
	}

	refreshTokenHash := s.tokenManager.HashRefreshToken(refreshToken)

	storedToken, err := s.tokens.GetByHash(ctx, refreshTokenHash)
	if err != nil {
		return RefreshTokensResult{}, fmt.Errorf("invalid refresh token")
	}

	if storedToken == nil {
		return RefreshTokensResult{}, fmt.Errorf("invalid refresh token")
	}

	if time.Now().After(storedToken.ExpiresAt) {
		_ = s.tokens.DeleteByHash(ctx, refreshTokenHash)
		return RefreshTokensResult{}, fmt.Errorf("refresh token expired")
	}

	newRefreshToken, newRefreshTokenHash, expiresAt, err := s.tokenManager.CreateRefreshToken()
	if err != nil {
		return RefreshTokensResult{}, fmt.Errorf("create refresh token: %w", err)
	}

	accessToken, err := s.tokenManager.CreateAccessToken(storedToken.UserID)
	if err != nil {
		return RefreshTokensResult{}, fmt.Errorf("create access token: %w", err)
	}

	if err := s.tokens.Rotate(ctx, refreshTokenHash, &domain.RefreshToken{
		UserID:    storedToken.UserID,
		TokenHash: newRefreshTokenHash,
		ExpiresAt: expiresAt,
	}); err != nil {
		return RefreshTokensResult{}, fmt.Errorf("rotate refresh token: %w", err)
	}

	return RefreshTokensResult{
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
	}, nil
}
