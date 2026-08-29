package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/wxvn/grpc-mesh/internal/auth/domain"
)

type mockUserStorage struct {
	user           *domain.User
	getByEmailErr  error
	createUserErr  error
	createUserCall int
}

func (m *mockUserStorage) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	return m.user, m.getByEmailErr
}

func (m *mockUserStorage) CreateUser(ctx context.Context, user *domain.User) error {
	m.createUserCall++
	if m.createUserErr != nil {
		return m.createUserErr
	}
	m.user = user
	return nil
}

type mockTokenStorage struct {
	token            *domain.RefreshToken
	getByHashToken   *domain.RefreshToken
	getByHashErr     error
	createTokenErr   error
	deleteByHashErr  error
	rotateErr        error
	createTokenCall  int
	deleteByHashCall int
	rotateCall       int
	rotatedOldHash   string
	rotatedNewToken  *domain.RefreshToken
}

func (m *mockTokenStorage) CreateToken(ctx context.Context, token *domain.RefreshToken) error {
	m.createTokenCall++
	if m.createTokenErr != nil {
		return m.createTokenErr
	}
	m.token = token
	return nil
}

func (m *mockTokenStorage) GetByHash(ctx context.Context, hash string) (*domain.RefreshToken, error) {
	return m.getByHashToken, m.getByHashErr
}

func (m *mockTokenStorage) DeleteByHash(ctx context.Context, hash string) error {
	m.deleteByHashCall++
	return m.deleteByHashErr
}

func (m *mockTokenStorage) Rotate(ctx context.Context, oldHash string, newToken *domain.RefreshToken) error {
	m.rotateCall++
	m.rotatedOldHash = oldHash
	m.rotatedNewToken = newToken
	return m.rotateErr
}

type mockPasswordHasher struct {
	hash        string
	hashErr     error
	compareErr  error
	hashCall    int
	compareCall int
}

func (m *mockPasswordHasher) Hash(password string) (string, error) {
	m.hashCall++
	return m.hash, m.hashErr
}

func (m *mockPasswordHasher) Compare(passwordHash, password string) error {
	m.compareCall++
	return m.compareErr
}

type mockTokenManager struct {
	accessToken         string
	accessTokenErr      error
	refreshToken        string
	refreshTokenHash    string
	refreshExpiresAt    time.Time
	refreshTokenErr     error
	hashRefreshToken    string
	parseUserID         uuid.UUID
	parseAccessTokenErr error

	createAccessTokenCall  int
	createRefreshTokenCall int
	parseAccessTokenCall   int
	hashRefreshTokenCall   int
}

func (m *mockTokenManager) CreateAccessToken(userID uuid.UUID) (string, error) {
	m.createAccessTokenCall++
	return m.accessToken, m.accessTokenErr
}

func (m *mockTokenManager) ParseAccessToken(token string) (uuid.UUID, error) {
	m.parseAccessTokenCall++
	return m.parseUserID, m.parseAccessTokenErr
}

func (m *mockTokenManager) CreateRefreshToken() (string, string, time.Time, error) {
	m.createRefreshTokenCall++
	return m.refreshToken, m.refreshTokenHash, m.refreshExpiresAt, m.refreshTokenErr
}

func (m *mockTokenManager) HashRefreshToken(token string) string {
	m.hashRefreshTokenCall++
	return m.hashRefreshToken
}

func newTestService() (
	*AuthService,
	*mockUserStorage,
	*mockTokenStorage,
	*mockPasswordHasher,
	*mockTokenManager,
) {
	users := &mockUserStorage{}
	tokens := &mockTokenStorage{}

	passwordHasher := &mockPasswordHasher{
		hash: "hashed-password",
	}

	tokenManager := &mockTokenManager{
		accessToken:      "access-token",
		refreshToken:     "refresh-token",
		refreshTokenHash: "refresh-hash",
		refreshExpiresAt: time.Now().Add(24 * time.Hour),
		hashRefreshToken: "refresh-hash",
	}

	service := NewAuthService(
		users,
		tokens,
		passwordHasher,
		tokenManager,
	)

	return service, users, tokens, passwordHasher, tokenManager
}

func TestAuthService_Register_Success(t *testing.T) {
	service, users, tokens, passwordHasher, tokenManager := newTestService()

	result, err := service.Register(
		context.Background(),
		"test@example.com",
		"password123",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.UserID == uuid.Nil {
		t.Fatal("expected user ID")
	}

	if result.AccessToken != "access-token" {
		t.Errorf("expected access token %q, got %q", "access-token", result.AccessToken)
	}

	if result.RefreshToken != "refresh-token" {
		t.Errorf("expected refresh token %q, got %q", "refresh-token", result.RefreshToken)
	}

	if users.user == nil {
		t.Fatal("expected user to be created")
	}

	if users.user.Email != "test@example.com" {
		t.Errorf("expected email %q, got %q", "test@example.com", users.user.Email)
	}

	if users.user.PasswordHash != "hashed-password" {
		t.Errorf("expected password hash %q, got %q", "hashed-password", users.user.PasswordHash)
	}

	if passwordHasher.hashCall != 1 {
		t.Errorf("expected Hash to be called once, got %d", passwordHasher.hashCall)
	}

	if tokens.createTokenCall != 1 {
		t.Errorf("expected CreateToken to be called once, got %d", tokens.createTokenCall)
	}

	if tokenManager.createAccessTokenCall != 1 {
		t.Errorf("expected CreateAccessToken to be called once, got %d", tokenManager.createAccessTokenCall)
	}

	if tokenManager.createRefreshTokenCall != 1 {
		t.Errorf("expected CreateRefreshToken to be called once, got %d", tokenManager.createRefreshTokenCall)
	}
}

func TestAuthService_Register_InvalidEmail(t *testing.T) {
	service, users, tokens, passwordHasher, _ := newTestService()

	tests := []struct {
		name  string
		email string
	}{
		{"empty email", ""},
		{"invalid email", "test"},
		{"invalid email", "test@"},
		{"invalid email", "@example.com"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := service.Register(
				context.Background(),
				tt.email,
				"password123",
			)

			if err == nil {
				t.Fatal("expected error")
			}

			if users.createUserCall != 0 {
				t.Fatal("CreateUser should not be called")
			}

			if passwordHasher.hashCall != 0 {
				t.Fatal("Hash should not be called")
			}

			if tokens.createTokenCall != 0 {
				t.Fatal("CreateToken should not be called")
			}
		})
	}
}

func TestAuthService_Register_EmptyPassword(t *testing.T) {
	service, users, tokens, passwordHasher, _ := newTestService()

	_, err := service.Register(
		context.Background(),
		"test@example.com",
		"",
	)
	if err == nil {
		t.Fatal("expected error")
	}

	if err.Error() != "password is required" {
		t.Errorf("expected %q, got %q", "password is required", err.Error())
	}

	if users.createUserCall != 0 {
		t.Fatal("CreateUser should not be called")
	}

	if passwordHasher.hashCall != 0 {
		t.Fatal("Hash should not be called")
	}

	if tokens.createTokenCall != 0 {
		t.Fatal("CreateToken should not be called")
	}
}

func TestAuthService_Register_EmailAlreadyExists(t *testing.T) {
	service, users, tokens, passwordHasher, _ := newTestService()

	users.user = &domain.User{
		ID:    uuid.New(),
		Email: "test@example.com",
	}

	_, err := service.Register(
		context.Background(),
		"test@example.com",
		"password123",
	)
	if err == nil {
		t.Fatal("expected error")
	}

	if err.Error() != "email already exists" {
		t.Errorf("expected %q, got %q", "email already exists", err.Error())
	}

	if passwordHasher.hashCall != 0 {
		t.Fatal("Hash should not be called")
	}

	if users.createUserCall != 0 {
		t.Fatal("CreateUser should not be called")
	}

	if tokens.createTokenCall != 0 {
		t.Fatal("CreateToken should not be called")
	}
}

func TestAuthService_Register_GetUserError(t *testing.T) {
	service, users, _, passwordHasher, _ := newTestService()

	users.getByEmailErr = errors.New("database error")

	_, err := service.Register(
		context.Background(),
		"test@example.com",
		"password123",
	)
	if err == nil {
		t.Fatal("expected error")
	}

	if passwordHasher.hashCall != 0 {
		t.Fatal("Hash should not be called")
	}
}

func TestAuthService_Register_HashError(t *testing.T) {
	service, users, tokens, passwordHasher, _ := newTestService()

	passwordHasher.hashErr = errors.New("hash error")

	_, err := service.Register(
		context.Background(),
		"test@example.com",
		"password123",
	)
	if err == nil {
		t.Fatal("expected error")
	}

	if users.createUserCall != 0 {
		t.Fatal("CreateUser should not be called")
	}

	if tokens.createTokenCall != 0 {
		t.Fatal("CreateToken should not be called")
	}
}

func TestAuthService_Register_CreateUserError(t *testing.T) {
	service, users, tokens, _, _ := newTestService()

	users.createUserErr = errors.New("database error")

	_, err := service.Register(
		context.Background(),
		"test@example.com",
		"password123",
	)
	if err == nil {
		t.Fatal("expected error")
	}

	if tokens.createTokenCall != 0 {
		t.Fatal("CreateToken should not be called")
	}
}

func TestAuthService_Register_CreateAccessTokenError(t *testing.T) {
	service, _, tokens, _, tokenManager := newTestService()

	tokenManager.accessTokenErr = errors.New("token error")

	_, err := service.Register(
		context.Background(),
		"test@example.com",
		"password123",
	)
	if err == nil {
		t.Fatal("expected error")
	}

	if tokens.createTokenCall != 0 {
		t.Fatal("CreateToken should not be called")
	}
}

func TestAuthService_Register_CreateRefreshTokenError(t *testing.T) {
	service, _, tokens, _, tokenManager := newTestService()

	tokenManager.refreshTokenErr = errors.New("token error")

	_, err := service.Register(
		context.Background(),
		"test@example.com",
		"password123",
	)
	if err == nil {
		t.Fatal("expected error")
	}

	if tokens.createTokenCall != 0 {
		t.Fatal("CreateToken should not be called")
	}
}

func TestAuthService_Register_CreateTokenError(t *testing.T) {
	service, _, tokens, _, _ := newTestService()

	tokens.createTokenErr = errors.New("database error")

	_, err := service.Register(
		context.Background(),
		"test@example.com",
		"password123",
	)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestAuthService_Login_Success(t *testing.T) {
	service, users, _, passwordHasher, tokenManager := newTestService()

	userID := uuid.New()

	users.user = &domain.User{
		ID:           userID,
		Email:        "test@example.com",
		PasswordHash: "hashed-password",
	}

	result, err := service.Login(
		context.Background(),
		"test@example.com",
		"password123",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.AccessToken != "access-token" {
		t.Errorf("expected access token %q, got %q", "access-token", result.AccessToken)
	}

	if result.RefreshToken != "refresh-token" {
		t.Errorf("expected refresh token %q, got %q", "refresh-token", result.RefreshToken)
	}

	if passwordHasher.compareCall != 1 {
		t.Errorf("expected Compare to be called once, got %d", passwordHasher.compareCall)
	}

	if tokenManager.createAccessTokenCall != 1 {
		t.Errorf("expected CreateAccessToken to be called once, got %d", tokenManager.createAccessTokenCall)
	}
}

func TestAuthService_Login_InvalidUser(t *testing.T) {
	service, _, _, passwordHasher, tokenManager := newTestService()

	result, err := service.Login(
		context.Background(),
		"test@example.com",
		"password123",
	)

	if err == nil {
		t.Fatal("expected error")
	}

	if result != (LoginResult{}) {
		t.Fatal("expected empty result")
	}

	if passwordHasher.compareCall != 0 {
		t.Fatal("Compare should not be called")
	}

	if tokenManager.createAccessTokenCall != 0 {
		t.Fatal("CreateAccessToken should not be called")
	}
}

func TestAuthService_Login_GetUserError(t *testing.T) {
	service, users, _, passwordHasher, _ := newTestService()

	users.getByEmailErr = errors.New("database error")

	_, err := service.Login(
		context.Background(),
		"test@example.com",
		"password123",
	)
	if err == nil {
		t.Fatal("expected error")
	}

	if err.Error() != "invalid credentials" {
		t.Errorf("expected invalid credentials, got %q", err.Error())
	}

	if passwordHasher.compareCall != 0 {
		t.Fatal("Compare should not be called")
	}
}

func TestAuthService_Login_InvalidPassword(t *testing.T) {
	service, users, _, passwordHasher, tokenManager := newTestService()

	users.user = &domain.User{
		ID:           uuid.New(),
		Email:        "test@example.com",
		PasswordHash: "hashed-password",
	}

	passwordHasher.compareErr = errors.New("invalid password")

	_, err := service.Login(
		context.Background(),
		"test@example.com",
		"wrong-password",
	)
	if err == nil {
		t.Fatal("expected error")
	}

	if err.Error() != "invalid credentials" {
		t.Errorf("expected invalid credentials, got %q", err.Error())
	}

	if tokenManager.createAccessTokenCall != 0 {
		t.Fatal("CreateAccessToken should not be called")
	}
}

func TestAuthService_Login_CreateTokenError(t *testing.T) {
	service, users, tokens, _, _ := newTestService()

	users.user = &domain.User{
		ID:           uuid.New(),
		Email:        "test@example.com",
		PasswordHash: "hashed-password",
	}

	tokens.createTokenErr = errors.New("database error")

	_, err := service.Login(
		context.Background(),
		"test@example.com",
		"password123",
	)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestAuthService_Logout_Success(t *testing.T) {
	service, _, tokens, _, tokenManager := newTestService()

	err := service.Logout(
		context.Background(),
		"refresh-token",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if tokenManager.hashRefreshTokenCall != 1 {
		t.Errorf("expected HashRefreshToken to be called once, got %d", tokenManager.hashRefreshTokenCall)
	}

	if tokens.deleteByHashCall != 1 {
		t.Errorf("expected DeleteByHash to be called once, got %d", tokens.deleteByHashCall)
	}
}

func TestAuthService_Logout_EmptyToken(t *testing.T) {
	service, _, tokens, _, tokenManager := newTestService()

	err := service.Logout(
		context.Background(),
		"",
	)
	if err == nil {
		t.Fatal("expected error")
	}

	if err.Error() != "refresh token is required" {
		t.Errorf("expected %q, got %q", "refresh token is required", err.Error())
	}

	if tokenManager.hashRefreshTokenCall != 0 {
		t.Fatal("HashRefreshToken should not be called")
	}

	if tokens.deleteByHashCall != 0 {
		t.Fatal("DeleteByHash should not be called")
	}
}

func TestAuthService_Logout_DeleteError(t *testing.T) {
	service, _, tokens, _, _ := newTestService()

	tokens.deleteByHashErr = errors.New("database error")

	err := service.Logout(
		context.Background(),
		"refresh-token",
	)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestAuthService_RefreshTokens_Success(t *testing.T) {
	service, _, tokens, _, tokenManager := newTestService()

	userID := uuid.New()

	tokens.getByHashToken = &domain.RefreshToken{
		ID:        uuid.New(),
		UserID:    userID,
		TokenHash: "refresh-hash",
		ExpiresAt: time.Now().Add(time.Hour),
	}

	result, err := service.RefreshTokens(
		context.Background(),
		"refresh-token",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.AccessToken != "access-token" {
		t.Errorf("expected access token %q, got %q", "access-token", result.AccessToken)
	}

	if result.RefreshToken != "refresh-token" {
		t.Errorf("expected refresh token %q, got %q", "refresh-token", result.RefreshToken)
	}

	if tokenManager.hashRefreshTokenCall != 1 {
		t.Errorf("expected HashRefreshToken to be called once, got %d", tokenManager.hashRefreshTokenCall)
	}

	if tokens.rotateCall != 1 {
		t.Errorf("expected Rotate to be called once, got %d", tokens.rotateCall)
	}

	if tokens.rotatedOldHash != "refresh-hash" {
		t.Errorf("expected old hash %q, got %q", "refresh-hash", tokens.rotatedOldHash)
	}

	if tokens.rotatedNewToken == nil {
		t.Fatal("expected new refresh token")
	}

	if tokens.rotatedNewToken.UserID != userID {
		t.Errorf("expected user ID %q, got %q", userID, tokens.rotatedNewToken.UserID)
	}
}

func TestAuthService_RefreshTokens_EmptyToken(t *testing.T) {
	service, _, tokens, _, tokenManager := newTestService()

	_, err := service.RefreshTokens(
		context.Background(),
		"",
	)
	if err == nil {
		t.Fatal("expected error")
	}

	if err.Error() != "refresh token is required" {
		t.Errorf("expected %q, got %q", "refresh token is required", err.Error())
	}

	if tokenManager.hashRefreshTokenCall != 0 {
		t.Fatal("HashRefreshToken should not be called")
	}

	if tokens.rotateCall != 0 {
		t.Fatal("Rotate should not be called")
	}
}

func TestAuthService_RefreshTokens_GetTokenError(t *testing.T) {
	service, _, tokens, _, _ := newTestService()

	tokens.getByHashErr = errors.New("database error")

	_, err := service.RefreshTokens(
		context.Background(),
		"refresh-token",
	)
	if err == nil {
		t.Fatal("expected error")
	}

	if err.Error() != "invalid refresh token" {
		t.Errorf("expected %q, got %q", "invalid refresh token", err.Error())
	}
}

func TestAuthService_RefreshTokens_TokenNotFound(t *testing.T) {
	service, _, _, _, _ := newTestService()

	_, err := service.RefreshTokens(
		context.Background(),
		"refresh-token",
	)
	if err == nil {
		t.Fatal("expected error")
	}

	if err.Error() != "invalid refresh token" {
		t.Errorf("expected %q, got %q", "invalid refresh token", err.Error())
	}
}

func TestAuthService_RefreshTokens_Expired(t *testing.T) {
	service, _, tokens, _, tokenManager := newTestService()

	tokens.getByHashToken = &domain.RefreshToken{
		ID:        uuid.New(),
		UserID:    uuid.New(),
		TokenHash: "refresh-hash",
		ExpiresAt: time.Now().Add(-time.Hour),
	}

	_, err := service.RefreshTokens(
		context.Background(),
		"refresh-token",
	)
	if err == nil {
		t.Fatal("expected error")
	}

	if err.Error() != "refresh token expired" {
		t.Errorf("expected %q, got %q", "refresh token expired", err.Error())
	}

	if tokens.deleteByHashCall != 1 {
		t.Errorf("expected DeleteByHash to be called once, got %d", tokens.deleteByHashCall)
	}

	if tokens.rotateCall != 0 {
		t.Fatal("Rotate should not be called")
	}

	if tokenManager.createAccessTokenCall != 0 {
		t.Fatal("CreateAccessToken should not be called")
	}
}

func TestAuthService_RefreshTokens_CreateRefreshTokenError(t *testing.T) {
	service, _, tokens, _, tokenManager := newTestService()

	tokens.getByHashToken = &domain.RefreshToken{
		ID:        uuid.New(),
		UserID:    uuid.New(),
		TokenHash: "refresh-hash",
		ExpiresAt: time.Now().Add(time.Hour),
	}

	tokenManager.refreshTokenErr = errors.New("token error")

	_, err := service.RefreshTokens(
		context.Background(),
		"refresh-token",
	)
	if err == nil {
		t.Fatal("expected error")
	}

	if tokens.rotateCall != 0 {
		t.Fatal("Rotate should not be called")
	}
}

func TestAuthService_RefreshTokens_CreateAccessTokenError(t *testing.T) {
	service, _, tokens, _, tokenManager := newTestService()

	tokens.getByHashToken = &domain.RefreshToken{
		ID:        uuid.New(),
		UserID:    uuid.New(),
		TokenHash: "refresh-hash",
		ExpiresAt: time.Now().Add(time.Hour),
	}

	tokenManager.accessTokenErr = errors.New("token error")

	_, err := service.RefreshTokens(
		context.Background(),
		"refresh-token",
	)
	if err == nil {
		t.Fatal("expected error")
	}

	if tokens.rotateCall != 0 {
		t.Fatal("Rotate should not be called")
	}
}

func TestAuthService_RefreshTokens_RotateError(t *testing.T) {
	service, _, tokens, _, _ := newTestService()

	tokens.getByHashToken = &domain.RefreshToken{
		ID:        uuid.New(),
		UserID:    uuid.New(),
		TokenHash: "refresh-hash",
		ExpiresAt: time.Now().Add(time.Hour),
	}

	tokens.rotateErr = errors.New("database error")

	_, err := service.RefreshTokens(
		context.Background(),
		"refresh-token",
	)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestAuthService_ValidateToken_Success(t *testing.T) {
	service, _, _, _, tokenManager := newTestService()

	userID := uuid.New()
	tokenManager.parseUserID = userID

	result, err := service.ValidateToken(
		context.Background(),
		"access-token",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result != userID {
		t.Errorf("expected user ID %q, got %q", userID, result)
	}

	if tokenManager.parseAccessTokenCall != 1 {
		t.Errorf("expected ParseAccessToken to be called once, got %d", tokenManager.parseAccessTokenCall)
	}
}

func TestAuthService_ValidateToken_Error(t *testing.T) {
	service, _, _, _, tokenManager := newTestService()

	tokenManager.parseAccessTokenErr = errors.New("invalid token")

	_, err := service.ValidateToken(
		context.Background(),
		"invalid-token",
	)
	if err == nil {
		t.Fatal("expected error")
	}
}
