package router

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"

	"github.com/dhruvsaxena1998/splitplus/internal/http/handlers"
	"github.com/dhruvsaxena1998/splitplus/internal/http/middleware"
	"github.com/dhruvsaxena1998/splitplus/internal/service"
)

func WithUserRoutes(userService service.UserService) Option {
	return optionFunc(func(r chi.Router) {
		v := validator.New()

		r.Route("/users", func(r chi.Router) {
			r.Post("/", middleware.ValidateBody[handlers.CreateUserRequest](v)(handlers.CreateUserHandler(userService)).ServeHTTP)
			r.Get("/me", handlers.GetMeHandler(userService))
		})
	})
}
