package service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/wxvn/grpc-mesh/internal/auth/domain"
)

func (s *AuthService) issueTokens(ctx context.Context, userID uuid.UUID) (accessToken string, refreshToken string, err error) {
	accessToken, err = s.tokenManager.CreateAccessToken(userID)
	if err != nil {
		return "", "", err
	}

	refreshToken, refreshTokenHash, expiresAt, err := s.tokenManager.CreateRefreshToken()
	if err != nil {
		return "", "", err
	}

	if err := s.tokens.CreateToken(ctx, &domain.RefreshToken{
		ID:        uuid.New(),
		UserID:    userID,
		TokenHash: refreshTokenHash,
		ExpiresAt: expiresAt,
		CreatedAt: time.Now(),
	}); err != nil {
		return "", "", err
	}

	return accessToken, refreshToken, nil
}
