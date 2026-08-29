package service

import (
	"context"

	"github.com/google/uuid"
)

func (s *AuthService) ValidateToken(ctx context.Context, accessToken string) (uuid.UUID, error) {
	return s.tokenManager.ParseAccessToken(accessToken)
}
