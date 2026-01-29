package router

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"

	"github.com/dhruvsaxena1998/splitplus/internal/http/handlers"
	"github.com/dhruvsaxena1998/splitplus/internal/http/middleware"
	"github.com/dhruvsaxena1998/splitplus/internal/service"
)

func WithGroupRoutes(groupService service.GroupService, invitationService service.GroupInvitationService) Option {
	return optionFunc(func(r chi.Router) {
		v := validator.New()

		r.Route("/groups", func(r chi.Router) {
			// All group routes require authentication
			r.Use(middleware.RequireAuth)

			// GET /groups - List all groups for the authenticated user
			r.Get("/", handlers.ListUserGroupsHandler(groupService))

			// POST /groups - Create a new group
			r.Post("/",
				middleware.ValidateBody[handlers.CreateGroupRequest](v)(
					handlers.CreateGroupHandler(groupService),
				).ServeHTTP,
			)

			// Group-specific routes with {group_id}
			r.Route("/{group_id}", func(r chi.Router) {
				// POST /groups/{group_id}/invitations - Invite user (email-based)
				r.Post("/invitations",
					middleware.ValidateBody[handlers.CreateInvitationRequest](v)(
						handlers.CreateInvitationHandler(invitationService),
					).ServeHTTP,
				)

				// GET /groups/{group_id}/members - List group members
				r.Get("/members", handlers.ListGroupMembersHandler(groupService))
			})
		})

		// Invitation routes (Public/Protected mixed)
		r.Route("/invitations", func(r chi.Router) {
			// GET /invitations/{token} - Get invitation details (Public)
			r.Get("/{token}", handlers.GetInvitationHandler(invitationService))

			// POST /invitations/{token}/accept - Accept invitation (Authenticated)
			r.With(middleware.RequireAuth).Post("/{token}/accept",
				handlers.AcceptInvitationHandler(invitationService),
			)

			// POST /invitations/{token}/join - Smart Join (Public, handles auth/registration internally)
			r.With(middleware.ParseAuth).Post("/{token}/join", handlers.JoinGroupHandler(invitationService))
		})
	})
}
