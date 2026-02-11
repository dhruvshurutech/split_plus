package router

import (
	"github.com/dhruvsaxena1998/splitplus/internal/http/handlers"
	"github.com/dhruvsaxena1998/splitplus/internal/http/middleware"
	"github.com/dhruvsaxena1998/splitplus/internal/repository"
	"github.com/dhruvsaxena1998/splitplus/internal/service"
	"github.com/go-chi/chi/v5"
)

func WithThemeRoutes(themeService service.ThemeService, jwtService service.JWTService, sessionRepo repository.SessionRepository) Option {
	return optionFunc(func(r chi.Router) {
		r.With(middleware.RequireAuth(jwtService, sessionRepo)).Get("/themes/presets", handlers.ListThemePresetsHandler(themeService))
	})
}
