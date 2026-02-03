package handlers

import (
	"net/http"

	"github.com/dhruvsaxena1998/splitplus/internal/http/middleware"
	"github.com/dhruvsaxena1998/splitplus/internal/http/response"
	"github.com/dhruvsaxena1998/splitplus/internal/service"
	"github.com/jackc/pgx/v5/pgtype"
)

type CreateUserRequest struct {
	Name    string `json:"name" validate:"required"`
	Email   string `json:"email" validate:"required,email"`
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
			response.SendError(w, http.StatusInternalServerError, "invalid request context")
			return
		}

		user, err := userService.CreateUser(r.Context(), req.Name, req.Email, req.Password)
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
			response.SendError(w, http.StatusUnauthorized, "unauthorized")
			return
		}

		user, err := userService.GetUser(r.Context(), userID)
		if err != nil {
			response.SendError(w, http.StatusInternalServerError, err.Error())
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
