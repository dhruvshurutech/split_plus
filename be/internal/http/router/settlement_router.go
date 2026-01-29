package router

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"

	"github.com/dhruvsaxena1998/splitplus/internal/http/handlers"
	"github.com/dhruvsaxena1998/splitplus/internal/http/middleware"
	"github.com/dhruvsaxena1998/splitplus/internal/service"
)

func WithSettlementRoutes(settlementService service.SettlementService) Option {
	return optionFunc(func(r chi.Router) {
		v := validator.New()

		// All settlement routes require authentication
		r.Route("/groups/{group_id}/settlements", func(r chi.Router) {
			r.Use(middleware.RequireAuth)

			// GET /groups/{group_id}/settlements - List all settlements in group
			r.Get("/", handlers.ListSettlementsByGroupHandler(settlementService))

			// POST /groups/{group_id}/settlements - Create a new settlement
			r.Post("/",
				middleware.ValidateBody[handlers.CreateSettlementRequest](v)(
					handlers.CreateSettlementHandler(settlementService),
				).ServeHTTP,
			)

			// Settlement-specific routes with {settlement_id}
			r.Route("/{settlement_id}", func(r chi.Router) {
				// GET /groups/{group_id}/settlements/{settlement_id} - Get settlement by ID
				r.Get("/", handlers.GetSettlementHandler(settlementService))

				// PUT /groups/{group_id}/settlements/{settlement_id} - Update settlement
				r.Put("/",
					middleware.ValidateBody[handlers.UpdateSettlementRequest](v)(
						handlers.UpdateSettlementHandler(settlementService),
					).ServeHTTP,
				)

				// PATCH /groups/{group_id}/settlements/{settlement_id}/status - Update settlement status
				r.Patch("/status",
					middleware.ValidateBody[handlers.UpdateSettlementStatusRequest](v)(
						handlers.UpdateSettlementStatusHandler(settlementService),
					).ServeHTTP,
				)

				// DELETE /groups/{group_id}/settlements/{settlement_id} - Delete settlement
				r.Delete("/", handlers.DeleteSettlementHandler(settlementService))
			})
		})
	})
}
