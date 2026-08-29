package storage

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"

	"github.com/wxvn/grpc-mesh/internal/auth/domain"
)

func (s *Storage) CreateToken(ctx context.Context, token *domain.RefreshToken) error {
	query := `
		INSERT INTO refresh_tokens (
			user_id,
			token_hash,
			expires_at
		)
		VALUES ($1, $2, $3)
	`

	_, err := s.pool.Exec(ctx, query, token.UserID, token.TokenHash, token.ExpiresAt)
	if err != nil {
		return fmt.Errorf("create refresh token: %w", err)
	}

	return nil
}

func (s *Storage) GetByHash(ctx context.Context, hash string) (*domain.RefreshToken, error) {
	query := `
		SELECT
			id,
			user_id,
			token_hash,
			expires_at,
			created_at
		FROM refresh_tokens
		WHERE token_hash = $1
	`

	var token domain.RefreshToken

	err := s.pool.QueryRow(ctx, query, hash).Scan(
		&token.ID,
		&token.UserID,
		&token.TokenHash,
		&token.ExpiresAt,
		&token.CreatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("get refresh token: %w", err)
	}

	return &token, nil
}

func (s *Storage) DeleteByHash(ctx context.Context, hash string) error {
	query := `
		DELETE FROM refresh_tokens
		WHERE token_hash = $1
	`

	_, err := s.pool.Exec(ctx, query, hash)
	if err != nil {
		return fmt.Errorf("delete refresh token: %w", err)
	}

	return nil
}

func (s *Storage) Rotate(ctx context.Context, oldHash string, newToken *domain.RefreshToken) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	deleteQuery := `
		DELETE FROM refresh_tokens
		WHERE token_hash = $1
	`

	_, err = tx.Exec(ctx, deleteQuery, oldHash)
	if err != nil {
		return fmt.Errorf("delete old refresh token: %w", err)
	}

	insertQuery := `
		INSERT INTO refresh_tokens (
			user_id,
			token_hash,
			expires_at
		)
		VALUES ($1, $2, $3)
	`

	_, err = tx.Exec(ctx, insertQuery, newToken.UserID, newToken.TokenHash, newToken.ExpiresAt)
	if err != nil {
		return fmt.Errorf("insert new refresh token: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	return nil
}
