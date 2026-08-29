package service

import (
	"context"
	"fmt"
)

func (s *Service) Redirect(ctx context.Context, shortCode string) (string, error) {
	url, err := s.storage.GetByShortCode(ctx, shortCode)
	if err != nil {
		return "", fmt.Errorf("get url by short code: %w", err)
	}

	if url == nil {
		return "", fmt.Errorf("url not found")
	}

	if _, err := s.storage.IncrementClicks(ctx, url.ID); err != nil {
		return "", fmt.Errorf("increment clicks: %w", err)
	}

	return url.OriginalURL, nil
}
