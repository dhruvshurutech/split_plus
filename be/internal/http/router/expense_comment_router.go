package router

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"

	"github.com/dhruvsaxena1998/splitplus/internal/http/handlers"
	"github.com/dhruvsaxena1998/splitplus/internal/http/middleware"
	"github.com/dhruvsaxena1998/splitplus/internal/service"
)

func WithExpenseCommentRoutes(commentService service.ExpenseCommentService) Option {
	return optionFunc(func(r chi.Router) {
		v := validator.New()

		// Routes for accessing comments via group->expense hierarchy
		r.Route("/groups/{group_id}/expenses/{expense_id}/comments", func(r chi.Router) {
			r.Use(middleware.RequireAuth)

			// GET / - List comments
			r.Get("/", handlers.ListCommentsHandler(commentService))

			// POST / - Create comment
			r.Post("/",
				middleware.ValidateBody[handlers.CommentRequest](v)(
					handlers.CreateCommentHandler(commentService),
				).ServeHTTP,
			)

			// Routes for specific comment
			r.Route("/{comment_id}", func(r chi.Router) {
				// PUT / - Update comment
				r.Put("/",
					middleware.ValidateBody[handlers.CommentRequest](v)(
						handlers.UpdateCommentHandler(commentService),
					).ServeHTTP,
				)

				// DELETE / - Delete comment
				r.Delete("/", handlers.DeleteCommentHandler(commentService))
			})
		})
	})
}
