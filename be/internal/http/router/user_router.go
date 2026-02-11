package router

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"

	"github.com/dhruvsaxena1998/splitplus/internal/http/handlers"
	"github.com/dhruvsaxena1998/splitplus/internal/http/middleware"
	"github.com/dhruvsaxena1998/splitplus/internal/repository"
	"github.com/dhruvsaxena1998/splitplus/internal/service"
)

func WithUserRoutes(userService service.UserService, themeService service.ThemeService, jwtService service.JWTService, sessionRepo repository.SessionRepository) Option {
	return optionFunc(func(r chi.Router) {
		v := validator.New()

		r.Route("/users", func(r chi.Router) {
			r.Post("/", middleware.ValidateBodyWithScope[handlers.CreateUserRequest](v, "user")(handlers.CreateUserHandler(userService)).ServeHTTP)
			r.With(middleware.RequireAuth(jwtService, sessionRepo)).Get("/me", handlers.GetMeHandler(userService))
			r.With(middleware.RequireAuth(jwtService, sessionRepo)).Get("/me/theme/preferences", handlers.GetThemePreferencesHandler(themeService))
			r.With(middleware.RequireAuth(jwtService, sessionRepo)).Put("/me/theme/preferences", middleware.ValidateBodyWithScope[handlers.UpdateThemePreferencesRequest](v, "theme")(handlers.UpdateThemePreferencesHandler(themeService)).ServeHTTP)
			r.With(middleware.RequireAuth(jwtService, sessionRepo)).Get("/me/themes", handlers.ListUserThemesHandler(themeService))
			r.With(middleware.RequireAuth(jwtService, sessionRepo)).Post("/me/themes", middleware.ValidateBodyWithScope[handlers.CreateThemeRequest](v, "theme")(handlers.CreateUserThemeHandler(themeService)).ServeHTTP)
			r.With(middleware.RequireAuth(jwtService, sessionRepo)).Get("/me/themes/{theme_id}", handlers.GetUserThemeHandler(themeService))
			r.With(middleware.RequireAuth(jwtService, sessionRepo)).Patch("/me/themes/{theme_id}", middleware.ValidateBodyWithScope[handlers.UpdateThemeRequest](v, "theme")(handlers.UpdateUserThemeHandler(themeService)).ServeHTTP)
			r.With(middleware.RequireAuth(jwtService, sessionRepo)).Delete("/me/themes/{theme_id}", handlers.DeleteUserThemeHandler(themeService))
		})
	})
}
