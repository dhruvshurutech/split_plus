package router

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"

	"github.com/dhruvsaxena1998/splitplus/internal/http/handlers"
	"github.com/dhruvsaxena1998/splitplus/internal/http/middleware"
	"github.com/dhruvsaxena1998/splitplus/internal/service"
)

func WithRecurringExpenseRoutes(recurringExpenseService service.RecurringExpenseService) Option {
	return optionFunc(func(r chi.Router) {
		v := validator.New()

		// All recurring expense routes require authentication
		r.Route("/groups/{group_id}/recurring-expenses", func(r chi.Router) {
			r.Use(middleware.RequireAuth)

			// GET /groups/{group_id}/recurring-expenses - List all recurring expenses for a group
			r.Get("/", handlers.ListRecurringExpensesHandler(recurringExpenseService))

			// POST /groups/{group_id}/recurring-expenses - Create a new recurring expense
			r.Post("/",
				middleware.ValidateBody[handlers.CreateRecurringExpenseRequest](v)(
					handlers.CreateRecurringExpenseHandler(recurringExpenseService),
				).ServeHTTP,
			)

			// Recurring expense-specific routes with {recurring_expense_id}
			r.Route("/{recurring_expense_id}", func(r chi.Router) {
				// GET /groups/{group_id}/recurring-expenses/{recurring_expense_id} - Get recurring expense by ID
				r.Get("/", handlers.GetRecurringExpenseHandler(recurringExpenseService))

				// PUT /groups/{group_id}/recurring-expenses/{recurring_expense_id} - Update recurring expense
				r.Put("/",
					middleware.ValidateBody[handlers.UpdateRecurringExpenseRequest](v)(
						handlers.UpdateRecurringExpenseHandler(recurringExpenseService),
					).ServeHTTP,
				)

				// DELETE /groups/{group_id}/recurring-expenses/{recurring_expense_id} - Delete recurring expense
				r.Delete("/", handlers.DeleteRecurringExpenseHandler(recurringExpenseService))

				// POST /groups/{group_id}/recurring-expenses/{recurring_expense_id}/generate - Manually generate expense
				r.Post("/generate", handlers.GenerateExpenseFromRecurringHandler(recurringExpenseService))
			})
		})
	})
}
