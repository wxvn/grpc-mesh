package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/wxvn/grpc-mesh/internal/url/domain"
)

type Storage interface {
	Create(ctx context.Context, url *domain.URL) (*domain.URL, error)
	GetByID(ctx context.Context, id uuid.UUID) (*domain.URL, error)
	GetByShortCode(ctx context.Context, shortCode string) (*domain.URL, error)
	GetByUserID(ctx context.Context, userID uuid.UUID, limit int, offset int) ([]*domain.URL, error)
	IncrementClicks(ctx context.Context, id uuid.UUID) (*domain.URL, error)
	Delete(ctx context.Context, id uuid.UUID, userID uuid.UUID) error
}

type Service struct {
	storage Storage
}

func NewURLService(storage Storage) *Service {
	return &Service{
		storage: storage,
	}
}
