package router

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"

	"github.com/dhruvsaxena1998/splitplus/internal/http/handlers"
	"github.com/dhruvsaxena1998/splitplus/internal/http/middleware"
	"github.com/dhruvsaxena1998/splitplus/internal/service"
)

func WithExpenseRoutes(expenseService service.ExpenseService) Option {
	return optionFunc(func(r chi.Router) {
		v := validator.New()

		// All expense routes require authentication
		r.Route("/groups/{group_id}/expenses", func(r chi.Router) {
			r.Use(middleware.RequireAuth)

			// GET /groups/{group_id}/expenses - List all expenses for a group
			r.Get("/", handlers.ListExpensesHandler(expenseService))

			// POST /groups/{group_id}/expenses - Create a new expense
			r.Post("/",
				middleware.ValidateBody[handlers.CreateExpenseRequest](v)(
					handlers.CreateExpenseHandler(expenseService),
				).ServeHTTP,
			)

			// GET /groups/{group_id}/expenses/search - Search expenses
			r.Get("/search", handlers.SearchExpensesHandler(expenseService))

			// Expense-specific routes with {expense_id}
			r.Route("/{expense_id}", func(r chi.Router) {
				// GET /groups/{group_id}/expenses/{expense_id} - Get expense by ID
				r.Get("/", handlers.GetExpenseHandler(expenseService))

				// PUT /groups/{group_id}/expenses/{expense_id} - Update expense
				r.Put("/",
					middleware.ValidateBody[handlers.UpdateExpenseRequest](v)(
						handlers.UpdateExpenseHandler(expenseService),
					).ServeHTTP,
				)

				// DELETE /groups/{group_id}/expenses/{expense_id} - Delete expense
				r.Delete("/", handlers.DeleteExpenseHandler(expenseService))
			})
		})
	})
}
