package transport

import (
	"context"

	"github.com/google/uuid"

	"github.com/wxvn/grpc-mesh/internal/url/domain"
)

type URLService interface {
	CreateShortURL(ctx context.Context, userID uuid.UUID, originalURL string) (*domain.URL, error)
	GetMyURLs(ctx context.Context, userID uuid.UUID, limit int, offset int) ([]*domain.URL, error)
	GetURLStats(ctx context.Context, userID uuid.UUID, id uuid.UUID) (*domain.URL, error)
	DeleteURL(ctx context.Context, userID uuid.UUID, id uuid.UUID) error
	Redirect(ctx context.Context, shortCode string) (string, error)
}
