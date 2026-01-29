package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dhruvsaxena1998/splitplus/internal/db/sqlc"
	"github.com/dhruvsaxena1998/splitplus/internal/http/middleware"
	"github.com/dhruvsaxena1998/splitplus/internal/http/response"
	"github.com/dhruvsaxena1998/splitplus/internal/service"
	"github.com/dhruvsaxena1998/splitplus/internal/testutil"
	"github.com/go-playground/validator/v10"
)

func TestCreateUserHandler(t *testing.T) {
	tests := []struct {
		name             string
		requestBody      CreateUserRequest
		mockSetup        func(*testutil.MockUserService)
		expectedStatus   int
		validateResponse func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name: "successful user creation",
			requestBody: CreateUserRequest{
				Email:    "test@example.com",
				Password: "password123",
			},
			mockSetup: func(mock *testutil.MockUserService) {
				mock.CreateUserFunc = func(ctx context.Context, email string, password string) (sqlc.User, error) {
					return testutil.CreateTestUser(testutil.CreateTestUUID(1), email), nil
				}
			},
			expectedStatus: http.StatusCreated,
			validateResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var resp response.StandardResponse[UserResponse, response.ErrorDetail]
				if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
					t.Fatalf("failed to unmarshal response: %v", err)
				}

				if !resp.Status {
					t.Errorf("expected status true, got false")
				}
				if resp.Data == nil {
					t.Fatal("expected data, got nil")
				}
				if resp.Data.Email != "test@example.com" {
					t.Errorf("expected email test@example.com, got %s", resp.Data.Email)
				}
				if !resp.Data.ID.Valid {
					t.Errorf("expected valid ID, got invalid")
				}
			},
		},
		{
			name: "duplicate email",
			requestBody: CreateUserRequest{
				Email:    "existing@example.com",
				Password: "password123",
			},
			mockSetup: func(mock *testutil.MockUserService) {
				mock.CreateUserFunc = func(ctx context.Context, email string, password string) (sqlc.User, error) {
					return sqlc.User{}, service.ErrUserAlreadyExists
				}
			},
			expectedStatus: http.StatusConflict,
			validateResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var resp response.StandardResponse[interface{}, response.ErrorDetail]
				if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
					t.Fatalf("failed to unmarshal response: %v", err)
				}

				if resp.Status {
					t.Errorf("expected status false, got true")
				}
				if resp.Error == nil {
					t.Fatal("expected error, got nil")
				}
			},
		},
		{
			name: "service error",
			requestBody: CreateUserRequest{
				Email:    "test@example.com",
				Password: "password123",
			},
			mockSetup: func(mock *testutil.MockUserService) {
				mock.CreateUserFunc = func(ctx context.Context, email string, password string) (sqlc.User, error) {
					return sqlc.User{}, errors.New("database error")
				}
			},
			expectedStatus: http.StatusBadRequest,
			validateResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var resp response.StandardResponse[interface{}, response.ErrorDetail]
				if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
					t.Fatalf("failed to unmarshal response: %v", err)
				}

				if resp.Status {
					t.Errorf("expected status false, got true")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &testutil.MockUserService{}
			if tt.mockSetup != nil {
				tt.mockSetup(mockService)
			}

			handler := CreateUserHandler(mockService)
			validate := validator.New()

			// Create request body
			body, err := json.Marshal(tt.requestBody)
			if err != nil {
				t.Fatalf("failed to marshal request body: %v", err)
			}

			req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")

			// Apply validation middleware
			wrappedHandler := middleware.ValidateBody[CreateUserRequest](validate)(handler)
			w := httptest.NewRecorder()

			wrappedHandler.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if tt.validateResponse != nil {
				tt.validateResponse(t, w)
			}
		})
	}
}
