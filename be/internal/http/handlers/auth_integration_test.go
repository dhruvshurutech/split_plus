package handlers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/dhruvsaxena1998/splitplus/internal/app"
	"github.com/dhruvsaxena1998/splitplus/internal/db"
	"github.com/dhruvsaxena1998/splitplus/internal/db/sqlc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestAuthenticationFlow tests the complete authentication flow
func TestAuthenticationFlow(t *testing.T) {
	// Skip if no database connection
	ctx := context.Background()
	pool, err := db.NewPool(ctx, "postgres://app:app@localhost:5432/app_db?sslmode=disable")
	if err != nil {
		t.Skip("Skipping integration test: database not available")
	}
	defer pool.Close()

	queries := sqlc.New(pool)
	
	// Initialize app with test JWT secret
	application := app.New(pool, queries, "test-secret-key-for-integration-tests-32-chars", 
		15*time.Minute, // 15 minutes for access token
		24*time.Hour, // 24 hours for refresh token
	)

	// Create test user first
	testEmail := "integration-test@example.com"
	testPassword := "testpassword123"

	// Clean up any existing test user
	_, _ = pool.Exec(ctx, "DELETE FROM users WHERE email = $1", testEmail)

	// Register user
	registerBody := map[string]string{
		"name":     "Test User",
		"email":    testEmail,
		"password": testPassword,
	}
	registerJSON, _ := json.Marshal(registerBody)
	
	registerReq := httptest.NewRequest(http.MethodPost, "/users", bytes.NewBuffer(registerJSON))
	registerReq.Header.Set("Content-Type", "application/json")
	registerRec := httptest.NewRecorder()
	
	application.Router.ServeHTTP(registerRec, registerReq)
	require.Equal(t, http.StatusCreated, registerRec.Code, "User registration should succeed")

	t.Run("Login Flow", func(t *testing.T) {
		// Test login
		loginBody := map[string]string{
			"email":    testEmail,
			"password": testPassword,
		}
		loginJSON, _ := json.Marshal(loginBody)
		
		req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewBuffer(loginJSON))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		
		application.Router.ServeHTTP(rec, req)
		
		require.Equal(t, http.StatusOK, rec.Code, "Login should succeed")
		
		var response struct {
			Status bool `json:"status"`
			Data   *struct {
				AccessToken  string `json:"access_token"`
				RefreshToken string `json:"refresh_token"`
				ExpiresIn    int    `json:"expires_in"`
			} `json:"data"`
		}
		err := json.Unmarshal(rec.Body.Bytes(), &response)
		require.NoError(t, err)
		
		assert.True(t, response.Status)
		require.NotNil(t, response.Data)
		assert.NotEmpty(t, response.Data.AccessToken)
		assert.NotEmpty(t, response.Data.RefreshToken)
		assert.NotZero(t, response.Data.ExpiresIn)
	})

	t.Run("Login with Invalid Credentials", func(t *testing.T) {
		loginBody := map[string]string{
			"email":    testEmail,
			"password": "wrongpassword",
		}
		loginJSON, _ := json.Marshal(loginBody)
		
		req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewBuffer(loginJSON))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		
		application.Router.ServeHTTP(rec, req)
		
		assert.Equal(t, http.StatusUnauthorized, rec.Code)
	})

	t.Run("Token Refresh Flow", func(t *testing.T) {
		// Login to get tokens
		loginBody := map[string]string{
			"email":    testEmail,
			"password": testPassword,
		}
		loginJSON, _ := json.Marshal(loginBody)
		
		loginReq := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewBuffer(loginJSON))
		loginReq.Header.Set("Content-Type", "application/json")
		loginRec := httptest.NewRecorder()
		
		application.Router.ServeHTTP(loginRec, loginReq)
		
		var loginResponse struct {
			Data *struct {
				RefreshToken string `json:"refresh_token"`
			} `json:"data"`
		}
		json.Unmarshal(loginRec.Body.Bytes(), &loginResponse)
		
		refreshToken := loginResponse.Data.RefreshToken
		
		// Refresh token
		refreshBody := map[string]string{
			"refresh_token": refreshToken,
		}
		refreshJSON, _ := json.Marshal(refreshBody)
		
		req := httptest.NewRequest(http.MethodPost, "/auth/refresh", bytes.NewBuffer(refreshJSON))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		
		application.Router.ServeHTTP(rec, req)
		
		require.Equal(t, http.StatusOK, rec.Code, "Token refresh should succeed")
		
		var refreshResponse struct {
			Status bool `json:"status"`
			Data   *struct {
				AccessToken string `json:"access_token"`
				ExpiresIn   int    `json:"expires_in"`
			} `json:"data"`
		}
		err := json.Unmarshal(rec.Body.Bytes(), &refreshResponse)
		require.NoError(t, err)
		
		assert.True(t, refreshResponse.Status)
		require.NotNil(t, refreshResponse.Data)
		assert.NotEmpty(t, refreshResponse.Data.AccessToken)
		assert.NotZero(t, refreshResponse.Data.ExpiresIn)
	})

	// Cleanup
	_, _ = pool.Exec(ctx, "DELETE FROM users WHERE email = $1", testEmail)
}
