package handlers

import (
	"net/http"

	"github.com/dhruvsaxena1998/splitplus/internal/http/middleware"
	"github.com/dhruvsaxena1998/splitplus/internal/http/response"
	"github.com/dhruvsaxena1998/splitplus/internal/service"
	"github.com/jackc/pgx/v5/pgtype"
)

type CreateUserRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
}

type UserResponse struct {
	ID        pgtype.UUID `json:"id"`
	Email     string      `json:"email"`
	CreatedAt string      `json:"created_at"`
}

func CreateUserHandler(userService service.UserService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req, ok := middleware.GetBody[CreateUserRequest](r)
		if !ok {
			response.SendError(w, http.StatusInternalServerError, "invalid request context")
			return
		}

		user, err := userService.CreateUser(r.Context(), req.Email, req.Password)
		if err != nil {
			var statusCode int
			switch err {
			case service.ErrUserAlreadyExists:
				statusCode = http.StatusConflict
			default:
				statusCode = http.StatusBadRequest
			}
			response.SendError(w, statusCode, err.Error())
			return
		}

		resp := UserResponse{
			ID:        user.ID,
			Email:     user.Email,
			CreatedAt: user.CreatedAt.Time.String(),
		}

		response.SendSuccess(w, http.StatusCreated, resp)
	}
}
