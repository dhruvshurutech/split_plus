package handlers

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/dhruvsaxena1998/splitplus/internal/http/middleware"
	"github.com/dhruvsaxena1998/splitplus/internal/http/response"
	"github.com/dhruvsaxena1998/splitplus/internal/service"
)

type CreateInvitationRequest struct {
	Email string `json:"email" validate:"required,email"`
	Name  string `json:"name,omitempty"` // Optional name for pending user
	Role  string `json:"role" validate:"omitempty,oneof=member admin"`
}

type AcceptInvitationRequest struct {
	Token string `json:"token" validate:"required"`
}

type JoinGroupRequest struct {
	Password string `json:"password,omitempty"`
	Name     string `json:"name,omitempty"`
}

type InvitationResponse struct {
	ID           pgtype.UUID `json:"id"`
	GroupID      pgtype.UUID `json:"group_id"`
	GroupName    string      `json:"group_name"`
	Email        string      `json:"email"`
	Role         string      `json:"role"`
	Status       string      `json:"status"`
	ExpiresAt    string      `json:"expires_at"`
	InvitedBy    pgtype.UUID `json:"invited_by"`
	InviterName  string      `json:"inviter_name,omitempty"`
	InviterEmail string      `json:"inviter_email,omitempty"`
}

func CreateInvitationHandler(invitationService service.GroupInvitationService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req, ok := middleware.GetBody[CreateInvitationRequest](r)
		if !ok {
			response.SendErrorWithCode(w, http.StatusInternalServerError, "system.request.context_invalid", "Invalid request context.")
			return
		}

		groupID, err := parseUUID(chi.URLParam(r, "group_id"))
		if err != nil {
			response.SendErrorWithCode(w, http.StatusBadRequest, "validation.invitation.group_id.invalid", "Invalid group id.")
			return
		}

		userID, ok := middleware.GetUserID(r)
		if !ok {
			response.SendErrorWithCode(w, http.StatusUnauthorized, "auth.authorization.unauthorized", "Unauthorized.")
			return
		}

		token, err := invitationService.CreateInvitation(r.Context(), service.CreateInvitationInput{
			GroupID:   groupID,
			InvitedBy: userID,
			Email:     req.Email,
			Role:      req.Role,
			Name:      req.Name,
		})
		if err != nil {
			statusCode := http.StatusBadRequest
			code := "system.invitation.create_failed"
			message := "Unable to create invitation."
			switch err {
			case service.ErrNotGroupMember, service.ErrInsufficientPermissions:
				statusCode = http.StatusForbidden
				code = "permission.group.invitation_denied"
				message = "You do not have permission to invite users to this group."
			}
			response.SendErrorWithCode(w, statusCode, code, message)
			return
		}

		response.SendSuccess(w, http.StatusCreated, map[string]string{
			"message": "invitation sent (mock)",
			"token":   token, // Returning token for testing purposes
		})
	}
}

func GetInvitationHandler(invitationService service.GroupInvitationService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token := chi.URLParam(r, "token")
		if token == "" {
			response.SendErrorWithCode(w, http.StatusBadRequest, "validation.invitation.token.required", "Invitation token is required.")
			return
		}

		inv, err := invitationService.GetInvitation(r.Context(), token)
		if err != nil {
			statusCode := http.StatusBadRequest
			code := "system.invitation.fetch_failed"
			message := "Unable to fetch invitation."
			if err == service.ErrInvitationNotFound {
				statusCode = http.StatusNotFound
				code = "resource.invitation.not_found"
				message = "Invitation not found."
			}
			response.SendErrorWithCode(w, statusCode, code, message)
			return
		}

		resp := InvitationResponse{
			ID:           inv.Invitation.ID,
			GroupID:      inv.Invitation.GroupID,
			GroupName:    inv.GroupName,
			Email:        inv.Invitation.Email,
			Role:         inv.Invitation.Role,
			Status:       inv.Invitation.Status,
			ExpiresAt:    formatTimestamp(inv.Invitation.ExpiresAt),
			InvitedBy:    inv.Invitation.InvitedBy,
			InviterName:  inv.InviterName,
			InviterEmail: inv.InviterEmail,
		}

		response.SendSuccess(w, http.StatusOK, resp)
	}
}

func AcceptInvitationHandler(invitationService service.GroupInvitationService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token := chi.URLParam(r, "token")
		if token == "" {
			response.SendErrorWithCode(w, http.StatusBadRequest, "validation.invitation.token.required", "Invitation token is required.")
			return
		}

		userID, ok := middleware.GetUserID(r)
		if !ok {
			response.SendErrorWithCode(w, http.StatusUnauthorized, "auth.authorization.unauthorized", "Unauthorized.")
			return
		}

		member, err := invitationService.AcceptInvitation(r.Context(), service.AcceptInvitationInput{
			Token:  token,
			UserID: userID,
		})
		if err != nil {
			statusCode := http.StatusBadRequest
			code := "system.invitation.accept_failed"
			message := "Unable to accept invitation."
			switch err {
			case service.ErrInvitationNotFound:
				statusCode = http.StatusNotFound
				code = "resource.invitation.not_found"
				message = "Invitation not found."
			case service.ErrInvitationExpired:
				statusCode = http.StatusGone
				code = "resource.invitation.expired"
				message = "Invitation has expired."
			case service.ErrAlreadyMember:
				statusCode = http.StatusConflict
				code = "conflict.group.member_already_exists"
				message = "You are already a member of this group."
			}
			response.SendErrorWithCode(w, statusCode, code, message)
			return
		}

		response.SendSuccess(w, http.StatusOK, map[string]interface{}{
			"message": "invitation accepted",
			"member":  member,
		})
	}
}

func JoinGroupHandler(invitationService service.GroupInvitationService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token := chi.URLParam(r, "token")
		if token == "" {
			response.SendErrorWithCode(w, http.StatusBadRequest, "validation.invitation.token.required", "Invitation token is required.")
			return
		}

		// Body is optional (only needed if not authenticated)
		var req JoinGroupRequest
		_ = middleware.DecodeBody(r, &req) // Ignore error if body is empty

		userID, isAuth := middleware.GetUserID(r)

		user, member, err := invitationService.JoinGroup(r.Context(), service.JoinGroupInput{
			Token:               token,
			Password:            req.Password,
			Name:                req.Name,
			AuthenticatedUserID: pgtype.UUID{Bytes: userID.Bytes, Valid: isAuth},
		})

		if err != nil {
			statusCode := http.StatusBadRequest
			code := "system.invitation.join_failed"
			message := "Unable to join group."
			switch err {
			case service.ErrInvitationNotFound:
				statusCode = http.StatusNotFound
				code = "resource.invitation.not_found"
				message = "Invitation not found."
			case service.ErrUserNotFound:
				statusCode = http.StatusNotFound
				code = "resource.user.not_found"
				message = "User not found."
			case service.ErrInvalidPassword:
				statusCode = http.StatusUnauthorized
				code = "auth.credentials.invalid"
				message = "Invalid email or password."
			case service.ErrInvitationEmailMismatch:
				statusCode = http.StatusForbidden
				code = "permission.invitation.email_mismatch"
				message = "This invitation is for a different email address."
			case service.ErrPasswordRequiredForExistingAccount:
				statusCode = http.StatusUnprocessableEntity
				code = "validation.invitation.password.required_existing_account"
				message = "Password is required for existing account."
			case service.ErrPasswordRequiredToCreateAccount:
				statusCode = http.StatusUnprocessableEntity
				code = "validation.invitation.password.required_new_account"
				message = "Password is required to create account."
			}
			response.SendErrorWithCode(w, statusCode, code, message)
			return
		}

		response.SendSuccess(w, http.StatusOK, map[string]interface{}{
			"message": "joined group successfully",
			"user":    user,
			"member":  member,
		})
	}
}

func ListPendingInvitationsHandler(invitationService service.GroupInvitationService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get email from query or from user info?
		// Since pending users are by email, logic says we should query by user's email.
		// But in this app, user's have emails. Using authenticated user's email is safest.

		// Wait, user struct might not be in context fully, just ID.
		// We might need to fetch user to get their email?
		// Or we trust the client? No.
		// For now, assuming user is logged in, their email in `users` table matches invitation.

		// Actually, `ListPendingInvitations` in service takes `email`.
		// But we don't have user's email in `GetUserID` (just returns UUID).
		// We'd need to fetch user.
		// Maybe skip this endpoint for now as it wasn't explicitly requested in main tasks, but useful.
		// The prompt mentioned "ListPendingInvitations" in service.

		// I'll skip implementing this handler for now as I need `userRepo` to fetch email, which is not available in basic handler.
		// Or pass email in request? That is insecure if I can see other's invites.
		// I will stick to Create, Get, Accept handlers.
		response.SendErrorWithCode(w, http.StatusNotImplemented, "system.invitation.pending_list.not_implemented", "Not implemented.")
	}
}

// Helper to reuse formatTimestamp from expenses.go?
// No, expenses.go helper is private. I should redefine or export it.
// I'll redefine it for now to avoid cross-file dependency on private symbol.
