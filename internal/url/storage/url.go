package storage

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/wxvn/grpc-mesh/internal/url/domain"
)

func (s *Storage) Create(ctx context.Context, url *domain.URL) (*domain.URL, error) {
	query := `
		INSERT INTO urls (
			id,
			user_id,
			short_code,
			original_url,
			clicks
		)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING
			id,
			user_id,
			short_code,
			original_url,
			clicks,
			created_at
	`

	var createdURL domain.URL

	err := s.pool.QueryRow(
		ctx,
		query,
		url.ID,
		url.UserID,
		url.ShortCode,
		url.OriginalURL,
		url.Clicks,
	).Scan(
		&createdURL.ID,
		&createdURL.UserID,
		&createdURL.ShortCode,
		&createdURL.OriginalURL,
		&createdURL.Clicks,
		&createdURL.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("create url: %w", err)
	}

	return &createdURL, nil
}

func (s *Storage) GetByID(ctx context.Context, id uuid.UUID) (*domain.URL, error) {
	query := `
		SELECT
			id,
			user_id,
			short_code,
			original_url,
			clicks,
			created_at
		FROM urls
		WHERE id = $1
	`

	var url domain.URL

	err := s.pool.QueryRow(
		ctx,
		query,
		id,
	).Scan(
		&url.ID,
		&url.UserID,
		&url.ShortCode,
		&url.OriginalURL,
		&url.Clicks,
		&url.CreatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}

		return nil, fmt.Errorf("get url by id: %w", err)
	}

	return &url, nil
}

func (s *Storage) GetByShortCode(ctx context.Context, shortCode string) (*domain.URL, error) {
	query := `
		SELECT
			id,
			user_id,
			short_code,
			original_url,
			clicks,
			created_at
		FROM urls
		WHERE short_code = $1
	`

	var url domain.URL

	err := s.pool.QueryRow(
		ctx,
		query,
		shortCode,
	).Scan(
		&url.ID,
		&url.UserID,
		&url.ShortCode,
		&url.OriginalURL,
		&url.Clicks,
		&url.CreatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}

		return nil, fmt.Errorf("get url by short code: %w", err)
	}

	return &url, nil
}

func (s *Storage) GetByUserID(ctx context.Context, userID uuid.UUID, limit int, offset int) ([]*domain.URL, error) {
	query := `
		SELECT
			id,
			user_id,
			short_code,
			original_url,
			clicks,
			created_at
		FROM urls
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2
		OFFSET $3
	`

	rows, err := s.pool.Query(
		ctx,
		query,
		userID,
		limit,
		offset,
	)
	if err != nil {
		return nil, fmt.Errorf("get urls by user id: %w", err)
	}
	defer rows.Close()

	urls := make([]*domain.URL, 0)

	for rows.Next() {
		var url domain.URL

		err := rows.Scan(
			&url.ID,
			&url.UserID,
			&url.ShortCode,
			&url.OriginalURL,
			&url.Clicks,
			&url.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan url: %w", err)
		}

		urls = append(urls, &url)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate urls: %w", err)
	}

	return urls, nil
}

func (s *Storage) IncrementClicks(ctx context.Context, id uuid.UUID) (*domain.URL, error) {
	query := `
		UPDATE urls
		SET clicks = clicks + 1
		WHERE id = $1
		RETURNING
			id,
			user_id,
			short_code,
			original_url,
			clicks,
			created_at
	`

	var url domain.URL

	err := s.pool.QueryRow(
		ctx,
		query,
		id,
	).Scan(
		&url.ID,
		&url.UserID,
		&url.ShortCode,
		&url.OriginalURL,
		&url.Clicks,
		&url.CreatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}

		return nil, fmt.Errorf("increment url clicks: %w", err)
	}

	return &url, nil
}

func (s *Storage) Delete(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {
	query := `
		DELETE FROM urls
		WHERE id = $1
		AND user_id = $2
	`

	_, err := s.pool.Exec(
		ctx,
		query,
		id,
		userID,
	)
	if err != nil {
		return fmt.Errorf("delete url: %w", err)
	}

	return nil
}
