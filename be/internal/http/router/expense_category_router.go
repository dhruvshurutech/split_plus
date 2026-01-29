package router

import (
	"github.com/dhruvsaxena1998/splitplus/internal/http/handlers"
	"github.com/dhruvsaxena1998/splitplus/internal/http/middleware"
	"github.com/dhruvsaxena1998/splitplus/internal/service"
	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
)

func WithExpenseCategoryRoutes(categoryService service.ExpenseCategoryService) Option {
	return optionFunc(func(r chi.Router) {
		v := validator.New()

		// Public route - get category presets
		r.Get("/categories/presets", handlers.GetCategoryPresetsHandler(categoryService))

		// Group category routes - require authentication
		r.Route("/groups/{group_id}/categories", func(r chi.Router) {
			r.Use(middleware.RequireAuth)

			r.Get("/", handlers.ListGroupCategoriesHandler(categoryService))
			r.Post("/", middleware.ValidateBody[handlers.CreateCategoryRequest](v)(handlers.CreateGroupCategoryHandler(categoryService)).ServeHTTP)
			r.Post("/from-presets", middleware.ValidateBody[handlers.CreateFromPresetsRequest](v)(handlers.CreateCategoriesFromPresetsHandler(categoryService)).ServeHTTP)
			r.Put("/{id}", middleware.ValidateBody[handlers.UpdateCategoryRequest](v)(handlers.UpdateGroupCategoryHandler(categoryService)).ServeHTTP)
			r.Delete("/{id}", handlers.DeleteGroupCategoryHandler(categoryService))
		})
	})
}
