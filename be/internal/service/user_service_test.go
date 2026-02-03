package service

import (
	"context"
	"errors"
	"testing"

	"github.com/jackc/pgconn"

	"github.com/dhruvsaxena1998/splitplus/internal/db/sqlc"
	"github.com/dhruvsaxena1998/splitplus/internal/testutil"
)

func TestUserService_CreateUser(t *testing.T) {
	tests := []struct {
		name          string
		email         string
		password      string
		mockSetup     func(*testutil.MockUserRepository)
		expectedUser  sqlc.User
		expectedError error
	}{
		{
			name:     "Test User",
			email:    "test@example.com",
			password: "password123",
			mockSetup: func(mock *testutil.MockUserRepository) {
				mock.CreateUserFunc = func(ctx context.Context, params sqlc.CreateUserParams) (sqlc.User, error) {
					return testutil.CreateTestUser(testutil.CreateTestUUID(1), params.Email), nil
				}
			},
			expectedUser:  testutil.CreateTestUser(testutil.CreateTestUUID(1), "test@example.com"),
			expectedError: nil,
		},
		{
			name:     "",
			email:    "",
			password: "password123",
			mockSetup: func(mock *testutil.MockUserRepository) {
				// Repository should not be called
			},
			expectedError: errors.New("email is required"),
		},
		{
			name:     "",
			email:    "   ",
			password: "password123",
			mockSetup: func(mock *testutil.MockUserRepository) {
				// Repository should not be called
			},
			expectedError: errors.New("email is required"),
		},
		{
			name:     "Test User",
			email:    "  test@example.com  ",
			password: "password123",
			mockSetup: func(mock *testutil.MockUserRepository) {
				mock.CreateUserFunc = func(ctx context.Context, params sqlc.CreateUserParams) (sqlc.User, error) {
					// Email should be trimmed
					if params.Email != "test@example.com" {
						t.Errorf("expected trimmed email, got %q", params.Email)
					}
					return testutil.CreateTestUser(testutil.CreateTestUUID(1), params.Email), nil
				}
			},
			expectedUser:  testutil.CreateTestUser(testutil.CreateTestUUID(1), "test@example.com"),
			expectedError: nil,
		},
		{
			name:     "Test User",
			email:    "existing@example.com",
			password: "password123",
			mockSetup: func(mock *testutil.MockUserRepository) {
				mock.CreateUserFunc = func(ctx context.Context, params sqlc.CreateUserParams) (sqlc.User, error) {
					pgErr := &pgconn.PgError{
						Code: "23505", // unique_violation
					}
					return sqlc.User{}, pgErr
				}
			},
			expectedError: ErrUserAlreadyExists,
		},
		{
			name:     "Test User",
			email:    "test@example.com",
			password: "password123",
			mockSetup: func(mock *testutil.MockUserRepository) {
				mock.CreateUserFunc = func(ctx context.Context, params sqlc.CreateUserParams) (sqlc.User, error) {
					return sqlc.User{}, errors.New("database connection error")
				}
			},
			expectedError: errors.New("database connection error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &testutil.MockUserRepository{}
			if tt.mockSetup != nil {
				tt.mockSetup(mockRepo)
			}

			service := NewUserService(mockRepo)
			ctx := context.Background()

			user, err := service.CreateUser(ctx, tt.name, tt.email, tt.password)

			if tt.expectedError != nil {
				if err == nil {
					t.Errorf("expected error %v, got nil", tt.expectedError)
					return
				}
				if !errors.Is(err, tt.expectedError) && err.Error() != tt.expectedError.Error() {
					t.Errorf("expected error %v, got %v", tt.expectedError, err)
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if user.Email != tt.expectedUser.Email {
				t.Errorf("expected email %q, got %q", tt.expectedUser.Email, user.Email)
			}
			if user.ID != tt.expectedUser.ID {
				t.Errorf("expected ID %v, got %v", tt.expectedUser.ID, user.ID)
			}
		})
	}
}
