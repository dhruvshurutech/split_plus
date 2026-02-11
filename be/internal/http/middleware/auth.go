package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/dhruvsaxena1998/splitplus/internal/http/response"
	"github.com/dhruvsaxena1998/splitplus/internal/repository"
	"github.com/dhruvsaxena1998/splitplus/internal/service"
	"github.com/jackc/pgx/v5/pgtype"
)

type userIDKey struct{}
type jtiKey struct{}

// SetUserID sets the authenticated user ID in the request context
func SetUserID(ctx context.Context, userID pgtype.UUID) context.Context {
	return context.WithValue(ctx, userIDKey{}, userID)
}

// GetUserID retrieves the authenticated user ID from the request context
func GetUserID(r *http.Request) (pgtype.UUID, bool) {
	userID, ok := r.Context().Value(userIDKey{}).(pgtype.UUID)
	return userID, ok
}

// SetJTI sets the JWT ID in the request context
func SetJTI(ctx context.Context, jti string) context.Context {
	return context.WithValue(ctx, jtiKey{}, jti)
}

// GetJTI retrieves the JWT ID from the request context
func GetJTI(r *http.Request) (string, bool) {
	jti, ok := r.Context().Value(jtiKey{}).(string)
	return jti, ok
}

// RequireAuth is a middleware that ensures the user is authenticated via JWT
func RequireAuth(jwtService service.JWTService, sessionRepo repository.SessionRepository) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Extract token from Authorization header
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				response.SendErrorWithCode(w, http.StatusUnauthorized, "auth.authorization.missing_header", "Authorization header is required.")
				return
			}

			// Check if it's a Bearer token
			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || parts[0] != "Bearer" {
				response.SendErrorWithCode(w, http.StatusUnauthorized, "auth.authorization.invalid_format", "Authorization header must use Bearer token format.")
				return
			}

			tokenString := parts[1]

			// Validate JWT token
			claims, err := jwtService.ValidateAccessToken(tokenString)
			if err != nil {
				var code string
				var message string
				switch err {
				case service.ErrExpiredToken:
					code = "auth.token.expired"
					message = "token has expired"
				case service.ErrInvalidToken:
					code = "auth.token.invalid"
					message = "invalid token"
				default:
					code = "auth.authorization.unauthorized"
					message = "unauthorized"
				}
				response.SendErrorWithCode(w, http.StatusUnauthorized, code, message)
				return
			}

			// Check if token is blacklisted
			isBlacklisted, err := sessionRepo.IsTokenBlacklisted(r.Context(), claims.ID)
			if err != nil {
				response.SendErrorWithCode(w, http.StatusInternalServerError, "auth.internal.check_failed", "Authentication check failed.")
				return
			}

			if isBlacklisted {
				response.SendErrorWithCode(w, http.StatusUnauthorized, "auth.token.revoked", "Token has been revoked.")
				return
			}

			// Parse user ID from claims
			var userID pgtype.UUID
			if err := userID.Scan(claims.UserID); err != nil {
				response.SendErrorWithCode(w, http.StatusUnauthorized, "auth.token.user_invalid", "Invalid user in token.")
				return
			}

			// Set user ID and JTI in context
			ctx := SetUserID(r.Context(), userID)
			ctx = SetJTI(ctx, claims.ID)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// ParseAuth is a middleware that parses the authenticated user ID but doesn't require it
func ParseAuth(jwtService service.JWTService, sessionRepo repository.SessionRepository) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				next.ServeHTTP(w, r)
				return
			}

			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || parts[0] != "Bearer" {
				next.ServeHTTP(w, r)
				return
			}

			tokenString := parts[1]

			// Validate JWT token
			claims, err := jwtService.ValidateAccessToken(tokenString)
			if err != nil {
				// Invalid token, but we don't error on optional auth
				next.ServeHTTP(w, r)
				return
			}

			// Check if token is blacklisted
			isBlacklisted, err := sessionRepo.IsTokenBlacklisted(r.Context(), claims.ID)
			if err != nil || isBlacklisted {
				// Blacklisted or error, but we don't error on optional auth
				next.ServeHTTP(w, r)
				return
			}

			// Parse user ID from claims
			var userID pgtype.UUID
			if err := userID.Scan(claims.UserID); err != nil {
				next.ServeHTTP(w, r)
				return
			}

			// Set user ID and JTI in context
			ctx := SetUserID(r.Context(), userID)
			ctx = SetJTI(ctx, claims.ID)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
