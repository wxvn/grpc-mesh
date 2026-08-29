package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/wxvn/grpc-mesh/internal/url/domain"
)

func (s *Service) CreateShortURL(ctx context.Context, userID uuid.UUID, originalURL string) (*domain.URL, error) {
	shortCode, err := generateShortCode()
	if err != nil {
		return nil, fmt.Errorf("generate short code: %w", err)
	}

	url := &domain.URL{
		ID:          uuid.New(),
		UserID:      userID,
		ShortCode:   shortCode,
		OriginalURL: originalURL,
		Clicks:      0,
	}

	createdURL, err := s.storage.Create(ctx, url)
	if err != nil {
		return nil, fmt.Errorf("create url: %w", err)
	}

	return createdURL, nil
}
