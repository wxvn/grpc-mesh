package storage

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/wxvn/grpc-mesh/internal/auth/domain"
)

func (s *Storage) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	query := `
		SELECT
			id,
			email,
			password_hash,
			created_at
		FROM users
		WHERE email = $1
	`

	var user domain.User

	err := s.pool.QueryRow(ctx, query, email).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.CreatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("get user by email: %w", err)
	}

	return &user, nil
}

func (s *Storage) CreateUser(ctx context.Context, user *domain.User) error {
	query := `
		INSERT INTO users (
			id,
			email,
			password_hash
		)
		VALUES ($1, $2, $3)
	`

	_, err := s.pool.Exec(ctx, query, user.ID, user.Email, user.PasswordHash)
	if err != nil {
		return fmt.Errorf("create user: %w", err)
	}

	return nil
}
