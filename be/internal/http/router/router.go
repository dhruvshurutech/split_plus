package router

import (
	"net/http"
	"os"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

type Option interface {
	Apply(r chi.Router)
}
type optionFunc func(r chi.Router)

func (f optionFunc) Apply(r chi.Router) {
	f(r)
}

func New(opts ...Option) http.Handler {
	r := chi.NewRouter()

	// global middleware
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// CORS configuration
	// Get allowed origins from environment variable, default to localhost:3000 for development
	allowedOrigins := os.Getenv("CORS_ALLOWED_ORIGINS")
	if allowedOrigins == "" {
		allowedOrigins = "http://localhost:3000"
	}

	// Convert comma-separated origins to slice
	origins := []string{}
	for _, origin := range strings.Split(allowedOrigins, ",") {
		origins = append(origins, strings.TrimSpace(origin))
	}

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   origins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           86400, // 24 hours
	}))

	for _, opt := range opts {
		opt.Apply(r)
	}

	return r
}
