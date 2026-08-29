package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

func (s *Service) DeleteURL(ctx context.Context, userID uuid.UUID, id uuid.UUID) error {
	if err := s.storage.Delete(ctx, id, userID); err != nil {
		return fmt.Errorf("delete url: %w", err)
	}

	return nil
}
