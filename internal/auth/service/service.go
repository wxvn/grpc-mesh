package service

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/wxvn/grpc-mesh/internal/auth/domain"
)

type UserStorage interface {
	GetByEmail(ctx context.Context, email string) (*domain.User, error)
	CreateUser(ctx context.Context, user *domain.User) error
}

type TokenStorage interface {
	CreateToken(ctx context.Context, token *domain.RefreshToken) error
	GetByHash(ctx context.Context, hash string) (*domain.RefreshToken, error)
	DeleteByHash(ctx context.Context, hash string) error
	Rotate(ctx context.Context, oldHash string, newToken *domain.RefreshToken) error
}

type PasswordHasher interface {
	Hash(password string) (string, error)
	Compare(passwordHash, password string) error
}

type TokenManager interface {
	CreateAccessToken(userID uuid.UUID) (string, error)
	ParseAccessToken(token string) (uuid.UUID, error)

	CreateRefreshToken() (plain string, hash string, expiresAt time.Time, err error)
	HashRefreshToken(token string) string
}

type AuthService struct {
	users        UserStorage
	tokens       TokenStorage
	passwordHash PasswordHasher
	tokenManager TokenManager
}

func NewAuthService(users UserStorage, tokens TokenStorage, passwordHash PasswordHasher, tokenManager TokenManager) *AuthService {
	return &AuthService{
		users:        users,
		tokens:       tokens,
		passwordHash: passwordHash,
		tokenManager: tokenManager,
	}
}
