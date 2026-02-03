package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/dhruvsaxena1998/splitplus/internal/db/sqlc"
	"github.com/dhruvsaxena1998/splitplus/internal/service"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// Mock UserService
type MockUserService struct {
	mock.Mock
}

func (m *MockUserService) CreateUser(ctx context.Context, name, email, password string) (sqlc.User, error) {
	args := m.Called(ctx, name, email, password)
	return args.Get(0).(sqlc.User), args.Error(1)
}

func (m *MockUserService) AuthenticateUser(ctx context.Context, email, password string) (sqlc.User, error) {
	args := m.Called(ctx, email, password)
	return args.Get(0).(sqlc.User), args.Error(1)
}

func (m *MockUserService) GetUser(ctx context.Context, id pgtype.UUID) (sqlc.User, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(sqlc.User), args.Error(1)
}

// Mock SessionRepository
type MockSessionRepository struct {
	mock.Mock
}

func (m *MockSessionRepository) CreateSession(ctx context.Context, params sqlc.CreateSessionParams) (sqlc.Session, error) {
	args := m.Called(ctx, params)
	return args.Get(0).(sqlc.Session), args.Error(1)
}

func (m *MockSessionRepository) GetSessionByRefreshTokenHash(ctx context.Context, hash string) (sqlc.Session, error) {
	args := m.Called(ctx, hash)
	return args.Get(0).(sqlc.Session), args.Error(1)
}

func (m *MockSessionRepository) UpdateSessionLastUsed(ctx context.Context, sessionID pgtype.UUID) error {
	args := m.Called(ctx, sessionID)
	return args.Error(0)
}

func (m *MockSessionRepository) DeleteSession(ctx context.Context, refreshTokenHash string) error {
	args := m.Called(ctx, refreshTokenHash)
	return args.Error(0)
}

func (m *MockSessionRepository) DeleteAllUserSessions(ctx context.Context, userID pgtype.UUID) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

func (m *MockSessionRepository) DeleteExpiredSessions(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockSessionRepository) BlacklistToken(ctx context.Context, params sqlc.BlacklistTokenParams) error {
	args := m.Called(ctx, params)
	return args.Error(0)
}

func (m *MockSessionRepository) IsTokenBlacklisted(ctx context.Context, jti string) (bool, error) {
	args := m.Called(ctx, jti)
	return args.Bool(0), args.Error(1)
}

func (m *MockSessionRepository) DeleteExpiredBlacklistedTokens(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockSessionRepository) GetActiveSessionsByUserID(ctx context.Context, userID pgtype.UUID) ([]sqlc.Session, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]sqlc.Session), args.Error(1)
}

func TestAuthService_Login(t *testing.T) {
	mockUserService := new(MockUserService)
	mockSessionRepo := new(MockSessionRepository)
	jwtService := service.NewJWTService("test-secret-key-32-chars-long", 15*time.Minute, 24*time.Hour)
	authService := service.NewAuthService(mockUserService, mockSessionRepo, jwtService, 15*time.Minute, 24*time.Hour)

	ctx := context.Background()
	email := "test@example.com"
	password := "password123"
	userAgent := "TestAgent"
	ipAddress := "127.0.0.1"

	userID := pgtype.UUID{}
	_ = userID.Scan("550e8400-e29b-41d4-a716-446655440000")

	user := sqlc.User{
		ID:           userID,
		Email:        email,
		PasswordHash: "hashed_password",
	}

	t.Run("successful login", func(t *testing.T) {
		mockUserService.On("AuthenticateUser", ctx, email, password).Return(user, nil).Once()
		mockSessionRepo.On("CreateSession", ctx, mock.AnythingOfType("sqlc.CreateSessionParams")).Return(sqlc.Session{}, nil).Once()

		accessToken, refreshToken, expiresIn, err := authService.Login(ctx, email, password, userAgent, ipAddress)

		require.NoError(t, err)
		assert.NotEmpty(t, accessToken)
		assert.NotEmpty(t, refreshToken)
		assert.Greater(t, expiresIn, int64(0))
		mockUserService.AssertExpectations(t)
		mockSessionRepo.AssertExpectations(t)
	})

	t.Run("user not found", func(t *testing.T) {
		mockUserService.On("AuthenticateUser", ctx, email, password).Return(sqlc.User{}, service.ErrUserNotFound).Once()

		_, _, _, err := authService.Login(ctx, email, password, userAgent, ipAddress)

		assert.ErrorIs(t, err, service.ErrUserNotFound)
		mockUserService.AssertExpectations(t)
	})

	t.Run("invalid password", func(t *testing.T) {
		mockUserService.On("AuthenticateUser", ctx, email, password).Return(sqlc.User{}, service.ErrInvalidPassword).Once()

		_, _, _, err := authService.Login(ctx, email, password, userAgent, ipAddress)

		assert.ErrorIs(t, err, service.ErrInvalidPassword)
		mockUserService.AssertExpectations(t)
	})
}

func TestAuthService_RefreshToken(t *testing.T) {
	mockUserService := new(MockUserService)
	mockSessionRepo := new(MockSessionRepository)
	jwtService := service.NewJWTService("test-secret-key-32-chars-long", 15*time.Minute, 24*time.Hour)
	authService := service.NewAuthService(mockUserService, mockSessionRepo, jwtService, 15*time.Minute, 24*time.Hour)

	ctx := context.Background()
	userAgent := "TestAgent"
	ipAddress := "127.0.0.1"

	userID := pgtype.UUID{}
	_ = userID.Scan("550e8400-e29b-41d4-a716-446655440000")

	sessionID := pgtype.UUID{}
	_ = sessionID.Scan("660e8400-e29b-41d4-a716-446655440000")

	// Generate a valid refresh token
	refreshToken, _ := jwtService.GenerateRefreshToken()

	t.Run("successful token refresh", func(t *testing.T) {
		session := sqlc.Session{
			ID:     sessionID,
			UserID: userID,
		}

		mockSessionRepo.On("GetSessionByRefreshTokenHash", ctx, mock.Anything).Return(session, nil).Once()
		mockSessionRepo.On("UpdateSessionLastUsed", ctx, sessionID).Return(nil).Once()

		accessToken, expiresIn, err := authService.RefreshToken(ctx, refreshToken, userAgent, ipAddress)

		require.NoError(t, err)
		assert.NotEmpty(t, accessToken)
		assert.Greater(t, expiresIn, int64(0))
		mockSessionRepo.AssertExpectations(t)
	})

	t.Run("session not found", func(t *testing.T) {
		mockSessionRepo.On("GetSessionByRefreshTokenHash", ctx, mock.Anything).Return(sqlc.Session{}, service.ErrSessionNotFound).Once()

		_, _, err := authService.RefreshToken(ctx, refreshToken, userAgent, ipAddress)

		assert.ErrorIs(t, err, service.ErrSessionNotFound)
		mockSessionRepo.AssertExpectations(t)
	})
}

func TestAuthService_Logout(t *testing.T) {
	mockUserService := new(MockUserService)
	mockSessionRepo := new(MockSessionRepository)
	jwtService := service.NewJWTService("test-secret-key-32-chars-long", 15*time.Minute, 24*time.Hour)
	authService := service.NewAuthService(mockUserService, mockSessionRepo, jwtService, 15*time.Minute, 24*time.Hour)

	ctx := context.Background()
	refreshToken := "refresh-token"
	accessTokenJTI := "jti-123"

	userID := pgtype.UUID{}
	_ = userID.Scan("550e8400-e29b-41d4-a716-446655440000")

	t.Run("successful logout with token blacklisting", func(t *testing.T) {
		mockSessionRepo.On("DeleteSession", ctx, refreshToken).Return(nil).Once()
		mockSessionRepo.On("BlacklistToken", ctx, mock.AnythingOfType("sqlc.BlacklistTokenParams")).Return(nil).Once()

		err := authService.Logout(ctx, refreshToken, accessTokenJTI, userID)

		require.NoError(t, err)
		mockSessionRepo.AssertExpectations(t)
	})

	t.Run("logout without JTI", func(t *testing.T) {
		mockSessionRepo.On("DeleteSession", ctx, refreshToken).Return(nil).Once()

		err := authService.Logout(ctx, refreshToken, "", userID)

		require.NoError(t, err)
		mockSessionRepo.AssertExpectations(t)
	})

	t.Run("session deletion fails", func(t *testing.T) {
		mockSessionRepo.On("DeleteSession", ctx, refreshToken).Return(errors.New("db error")).Once()

		err := authService.Logout(ctx, refreshToken, accessTokenJTI, userID)

		assert.Error(t, err)
		mockSessionRepo.AssertExpectations(t)
	})
}

func TestAuthService_LogoutAllSessions(t *testing.T) {
	mockUserService := new(MockUserService)
	mockSessionRepo := new(MockSessionRepository)
	jwtService := service.NewJWTService("test-secret-key-32-chars-long", 15*time.Minute, 24*time.Hour)
	authService := service.NewAuthService(mockUserService, mockSessionRepo, jwtService, 15*time.Minute, 24*time.Hour)

	ctx := context.Background()
	userID := pgtype.UUID{}
	_ = userID.Scan("550e8400-e29b-41d4-a716-446655440000")

	t.Run("successful logout all sessions", func(t *testing.T) {
		mockSessionRepo.On("DeleteAllUserSessions", ctx, userID).Return(nil).Once()

		err := authService.LogoutAllSessions(ctx, userID)

		require.NoError(t, err)
		mockSessionRepo.AssertExpectations(t)
	})

	t.Run("deletion fails", func(t *testing.T) {
		mockSessionRepo.On("DeleteAllUserSessions", ctx, userID).Return(errors.New("db error")).Once()

		err := authService.LogoutAllSessions(ctx, userID)

		assert.Error(t, err)
		mockSessionRepo.AssertExpectations(t)
	})
}

func TestAuthService_CleanupExpiredSessions(t *testing.T) {
	mockUserService := new(MockUserService)
	mockSessionRepo := new(MockSessionRepository)
	jwtService := service.NewJWTService("test-secret-key-32-chars-long", 15*time.Minute, 24*time.Hour)
	authService := service.NewAuthService(mockUserService, mockSessionRepo, jwtService, 15*time.Minute, 24*time.Hour)

	ctx := context.Background()

	t.Run("successful cleanup", func(t *testing.T) {
		mockSessionRepo.On("DeleteExpiredSessions", ctx).Return(nil).Once()
		mockSessionRepo.On("DeleteExpiredBlacklistedTokens", ctx).Return(nil).Once()

		err := authService.CleanupExpiredSessions(ctx)

		require.NoError(t, err)
		mockSessionRepo.AssertExpectations(t)
	})

	t.Run("session cleanup fails", func(t *testing.T) {
		mockSessionRepo.On("DeleteExpiredSessions", ctx).Return(errors.New("db error")).Once()

		err := authService.CleanupExpiredSessions(ctx)

		assert.Error(t, err)
		mockSessionRepo.AssertExpectations(t)
	})

	t.Run("token cleanup fails", func(t *testing.T) {
		mockSessionRepo.On("DeleteExpiredSessions", ctx).Return(nil).Once()
		mockSessionRepo.On("DeleteExpiredBlacklistedTokens", ctx).Return(errors.New("db error")).Once()

		err := authService.CleanupExpiredSessions(ctx)

		assert.Error(t, err)
		mockSessionRepo.AssertExpectations(t)
	})
}
