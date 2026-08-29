package domain

import (
	"time"

	"github.com/google/uuid"
)

type URL struct {
	ID          uuid.UUID
	UserID      uuid.UUID
	ShortCode   string
	OriginalURL string
	Clicks      int64
	CreatedAt   time.Time
}
