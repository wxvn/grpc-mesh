package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/wxvn/grpc-mesh/internal/url/domain"
)

func (s *Service) GetMyURLs(ctx context.Context, userID uuid.UUID, limit int, offset int) ([]*domain.URL, error) {
	urls, err := s.storage.GetByUserID(ctx, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("get my urls: %w", err)
	}

	return urls, nil
}
