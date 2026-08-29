package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/wxvn/grpc-mesh/internal/url/domain"
)

func (s *Service) GetURLStats(ctx context.Context, userID uuid.UUID, id uuid.UUID) (*domain.URL, error) {
	url, err := s.storage.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get url: %w", err)
	}

	if url == nil {
		return nil, fmt.Errorf("url not found")
	}

	if url.UserID != userID {
		return nil, fmt.Errorf("access denied")
	}

	return url, nil
}
