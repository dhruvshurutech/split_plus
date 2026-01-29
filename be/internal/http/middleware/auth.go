package middleware

import (
	"context"
	"net/http"

	"github.com/dhruvsaxena1998/splitplus/internal/http/response"
	"github.com/jackc/pgx/v5/pgtype"
)

type userIDKey struct{}

// SetUserID sets the authenticated user ID in the request context
func SetUserID(ctx context.Context, userID pgtype.UUID) context.Context {
	return context.WithValue(ctx, userIDKey{}, userID)
}

// GetUserID retrieves the authenticated user ID from the request context
func GetUserID(r *http.Request) (pgtype.UUID, bool) {
	userID, ok := r.Context().Value(userIDKey{}).(pgtype.UUID)
	return userID, ok
}

// RequireAuth is a middleware that ensures the user is authenticated
// For now, this is a placeholder using X-User-ID header for testing
// TODO: Replace with actual JWT validation
func RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userIDStr := r.Header.Get("X-User-ID")
		if userIDStr == "" {
			response.SendError(w, http.StatusUnauthorized, "unauthorized")
			return
		}

		var userID pgtype.UUID
		if err := userID.Scan(userIDStr); err != nil {
			response.SendError(w, http.StatusUnauthorized, "invalid user id")
			return
		}

		ctx := SetUserID(r.Context(), userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// ParseAuth is a middleware that parses the authenticated user ID but doesn't require it
func ParseAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userIDStr := r.Header.Get("X-User-ID")
		if userIDStr == "" {
			next.ServeHTTP(w, r)
			return
		}

		var userID pgtype.UUID
		if err := userID.Scan(userIDStr); err != nil {
			// If provided but invalid, we ignore it for now or could error.
			// Given it's "optional", ignoring it is safer for public routes.
			next.ServeHTTP(w, r)
			return
		}

		ctx := SetUserID(r.Context(), userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
