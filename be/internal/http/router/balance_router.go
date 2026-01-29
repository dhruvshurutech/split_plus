package router

import (
	"github.com/go-chi/chi/v5"

	"github.com/dhruvsaxena1998/splitplus/internal/http/handlers"
	"github.com/dhruvsaxena1998/splitplus/internal/http/middleware"
	"github.com/dhruvsaxena1998/splitplus/internal/service"
)

func WithBalanceRoutes(balanceService service.BalanceService) Option {
	return optionFunc(func(r chi.Router) {
		// All balance routes require authentication
		r.Route("/groups/{group_id}/balances", func(r chi.Router) {
			r.Use(middleware.RequireAuth)

			// GET /groups/{group_id}/balances - List all balances in group
			r.Get("/", handlers.ListGroupBalancesHandler(balanceService))

			// GET /groups/{group_id}/balances/{user_id} - Get specific user's balance
			r.Get("/{user_id}", handlers.GetUserBalanceInGroupHandler(balanceService))
		})

		// GET /groups/{group_id}/debts - Get simplified "who owes whom" view
		r.Route("/groups/{group_id}/debts", func(r chi.Router) {
			r.Use(middleware.RequireAuth)
			r.Get("/", handlers.GetSimplifiedDebtsHandler(balanceService))
		})

		// GET /users/me/balances - Get user's balances across all groups
		r.Route("/users/me/balances", func(r chi.Router) {
			r.Use(middleware.RequireAuth)
			r.Get("/", handlers.GetOverallUserBalanceHandler(balanceService))
		})
	})
}
