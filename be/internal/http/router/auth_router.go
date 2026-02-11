package router

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"

	"github.com/dhruvsaxena1998/splitplus/internal/http/handlers"
	"github.com/dhruvsaxena1998/splitplus/internal/http/middleware"
	"github.com/dhruvsaxena1998/splitplus/internal/repository"
	"github.com/dhruvsaxena1998/splitplus/internal/service"
)

func WithAuthRoutes(
	authService service.AuthService,
	jwtService service.JWTService,
	sessionRepo repository.SessionRepository,
) Option {
	return optionFunc(func(r chi.Router) {
		v := validator.New()

		r.Route("/auth", func(r chi.Router) {
			// Public routes
			r.Post("/login", middleware.ValidateBodyWithScope[handlers.LoginRequest](v, "auth")(handlers.LoginHandler(authService)).ServeHTTP)
			r.Post("/refresh", middleware.ValidateBodyWithScope[handlers.RefreshTokenRequest](v, "auth")(handlers.RefreshTokenHandler(authService)).ServeHTTP)

			// Protected routes
			r.Group(func(r chi.Router) {
				r.Use(middleware.RequireAuth(jwtService, sessionRepo))
				r.Post("/logout", middleware.ValidateBodyWithScope[handlers.LogoutRequest](v, "auth")(handlers.LogoutHandler(authService, jwtService)).ServeHTTP)
				r.Post("/logout-all", handlers.LogoutAllHandler(authService))
			})
		})
	})
}
