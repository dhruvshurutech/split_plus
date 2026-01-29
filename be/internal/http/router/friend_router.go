package router

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"

	"github.com/dhruvsaxena1998/splitplus/internal/http/handlers"
	"github.com/dhruvsaxena1998/splitplus/internal/http/middleware"
	"github.com/dhruvsaxena1998/splitplus/internal/service"
)

func WithFriendRoutes(
	friendService service.FriendService,
	friendExpenseService service.FriendExpenseService,
	friendSettlementService service.FriendSettlementService,
) Option {
	return optionFunc(func(r chi.Router) {
		v := validator.New()

		r.Group(func(r chi.Router) {
			r.Use(middleware.RequireAuth)

			r.Route("/friends", func(r chi.Router) {
				// Friend requests / contacts
				r.Post("/requests", middleware.ValidateBody[handlers.SendFriendRequestRequest](v)(handlers.SendFriendRequestHandler(friendService)).ServeHTTP)
				r.Get("/", handlers.ListFriendsHandler(friendService))
				r.Get("/requests/incoming", handlers.ListIncomingFriendRequestsHandler(friendService))
				r.Get("/requests/outgoing", handlers.ListOutgoingFriendRequestsHandler(friendService))
				r.Post("/requests/{id}/accept", handlers.AcceptFriendRequestHandler(friendService))
				r.Post("/requests/{id}/decline", handlers.DeclineFriendRequestHandler(friendService))
				r.Delete("/{friend_id}", handlers.RemoveFriendHandler(friendService))

				// Friend expenses
				r.Route("/{friend_id}", func(r chi.Router) {
					r.Post("/expenses",
						middleware.ValidateBody[handlers.CreateExpenseRequest](v)(
							handlers.CreateFriendExpenseHandler(friendExpenseService),
						).ServeHTTP,
					)
					r.Get("/expenses", handlers.ListFriendExpensesHandler(friendExpenseService))

					// Friend settlements
					r.Post("/settlements",
						middleware.ValidateBody[handlers.CreateFriendSettlementRequest](v)(
							handlers.CreateFriendSettlementHandler(friendSettlementService),
						).ServeHTTP,
					)
					r.Get("/settlements", handlers.ListFriendSettlementsHandler(friendSettlementService))
				})
			})
		})
	})
}
