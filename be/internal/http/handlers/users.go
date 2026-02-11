package handlers

import (
	"net/http"

	"github.com/dhruvsaxena1998/splitplus/internal/http/middleware"
	"github.com/dhruvsaxena1998/splitplus/internal/http/response"
	"github.com/dhruvsaxena1998/splitplus/internal/service"
	"github.com/jackc/pgx/v5/pgtype"
)

type CreateUserRequest struct {
	Name     string `json:"name" validate:"required"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
}

type UserResponse struct {
	ID        pgtype.UUID `json:"id"`
	Name      string      `json:"name"`
	Email     string      `json:"email"`
	CreatedAt string      `json:"created_at"`
}

func CreateUserHandler(userService service.UserService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req, ok := middleware.GetBody[CreateUserRequest](r)
		if !ok {
			response.SendErrorWithCode(w, http.StatusInternalServerError, "system.request.context_invalid", "Invalid request context.")
			return
		}

		user, err := userService.CreateUser(r.Context(), req.Name, req.Email, req.Password)
		if err != nil {
			var statusCode int
			var code string
			var message string
			switch err {
			case service.ErrUserAlreadyExists:
				statusCode = http.StatusConflict
				code = "conflict.user.email_already_exists"
				message = "An account with this email already exists."
			case service.ErrUserEmailRequired:
				statusCode = http.StatusUnprocessableEntity
				code = "validation.user.email.required"
				message = "Email is required."
			case service.ErrUserNotFound:
				statusCode = http.StatusNotFound
				code = "resource.user.not_found"
				message = "User not found."
			default:
				statusCode = http.StatusBadRequest
				code = "system.user.create_failed"
				message = "Unable to create account right now."
			}
			response.SendErrorWithCode(w, statusCode, code, message)
			return
		}

		resp := UserResponse{
			ID:        user.ID,
			Name:      user.Name.String,
			Email:     user.Email,
			CreatedAt: user.CreatedAt.Time.String(),
		}

		response.SendSuccess(w, http.StatusCreated, resp)
	}
}

func GetMeHandler(userService service.UserService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := middleware.GetUserID(r)
		if !ok {
			response.SendErrorWithCode(w, http.StatusUnauthorized, "auth.authorization.unauthorized", "Unauthorized.")
			return
		}

		user, err := userService.GetUser(r.Context(), userID)
		if err != nil {
			if err == service.ErrUserNotFound {
				response.SendErrorWithCode(w, http.StatusNotFound, "resource.user.not_found", "User not found.")
				return
			}
			response.SendErrorWithCode(w, http.StatusInternalServerError, "system.user.fetch_failed", "Unable to load user profile.")
			return
		}

		resp := UserResponse{
			ID:        user.ID,
			Name:      user.Name.String,
			Email:     user.Email,
			CreatedAt: user.CreatedAt.Time.String(),
		}

		response.SendSuccess(w, http.StatusOK, resp)
	}
}
