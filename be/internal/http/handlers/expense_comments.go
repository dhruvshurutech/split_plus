package handlers

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/dhruvsaxena1998/splitplus/internal/http/middleware"
	"github.com/dhruvsaxena1998/splitplus/internal/http/response"
	"github.com/dhruvsaxena1998/splitplus/internal/service"
)

type CommentRequest struct {
	Comment string `json:"comment" validate:"required,min=1,max=1000"`
}

type CommentResponse struct {
	ID        pgtype.UUID `json:"id"`
	ExpenseID pgtype.UUID `json:"expense_id"`
	UserID    pgtype.UUID `json:"user_id"`
	Comment   string      `json:"comment"`
	CreatedAt string      `json:"created_at"`
	UpdatedAt string      `json:"updated_at"`
	User      UserInfo    `json:"user,omitempty"`
}

func CreateCommentHandler(commentService service.ExpenseCommentService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req, ok := middleware.GetBody[CommentRequest](r)
		if !ok {
			response.SendError(w, http.StatusInternalServerError, "invalid request context")
			return
		}

		expenseID, err := parseUUID(chi.URLParam(r, "expense_id"))
		if err != nil {
			response.SendError(w, http.StatusBadRequest, "invalid expense_id")
			return
		}

		userID, ok := middleware.GetUserID(r)
		if !ok {
			response.SendError(w, http.StatusUnauthorized, "unauthorized")
			return
		}

		result, err := commentService.CreateComment(r.Context(), expenseID, userID, req.Comment)
		if err != nil {
			statusCode := http.StatusBadRequest
			switch err {
			case service.ErrExpenseNotFound:
				statusCode = http.StatusNotFound
			case service.ErrNotGroupMember:
				statusCode = http.StatusForbidden
			}
			response.SendError(w, statusCode, err.Error())
			return
		}

		// Since CreateComment returns sqlc.ExpenseComment without user info, we might want to return simple response
		// or fetch the full comment with user info. For now, let's construct what we have.
		// Note: The UI might need the user info immediately.
		// The service returns the created comment. To get user info we'd need to fetch user details or just return what we have?
		// The `CreateComment` returns `sqlc.ExpenseComment`.
		// Let's rely on client refetching or just return basic info.
		// Wait, the client usually wants to append the comment immediately.
		// Ideally `CreateComment` should return the comment with user info or we construct it if we have user info from context (we have ID, maybe name/avatar is in JWT claim? No).
		// For simplicity, let's just return the comment data. The user created it so they know who they are.

		resp := CommentResponse{
			ID:        result.ID,
			ExpenseID: result.ExpenseID,
			UserID:    result.UserID,
			Comment:   result.Comment,
			CreatedAt: formatTimestamp(result.CreatedAt),
			UpdatedAt: formatTimestamp(result.UpdatedAt),
			// User info omitted for simple create response unless we fetch it.
			// Ideally we fetch the comment we just created to get the join.
		}

		response.SendSuccess(w, http.StatusCreated, resp)
	}
}

func ListCommentsHandler(commentService service.ExpenseCommentService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		expenseID, err := parseUUID(chi.URLParam(r, "expense_id"))
		if err != nil {
			response.SendError(w, http.StatusBadRequest, "invalid expense_id")
			return
		}

		comments, err := commentService.ListComments(r.Context(), expenseID)
		if err != nil {
			response.SendError(w, http.StatusInternalServerError, err.Error())
			return
		}

		resp := make([]CommentResponse, len(comments))
		for i, c := range comments {
			resp[i] = CommentResponse{
				ID:        c.ID,
				ExpenseID: c.ExpenseID,
				UserID:    c.UserID,
				Comment:   c.Comment,
				CreatedAt: formatTimestamp(c.CreatedAt),
				UpdatedAt: formatTimestamp(c.UpdatedAt),
				User: UserInfo{
					Email:     c.UserEmail,
					Name:      c.UserName.String,
					AvatarURL: c.UserAvatarUrl.String,
				},
			}
		}

		response.SendSuccess(w, http.StatusOK, resp)
	}
}

func UpdateCommentHandler(commentService service.ExpenseCommentService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req, ok := middleware.GetBody[CommentRequest](r)
		if !ok {
			response.SendError(w, http.StatusInternalServerError, "invalid request context")
			return
		}

		commentID, err := parseUUID(chi.URLParam(r, "comment_id"))
		if err != nil {
			response.SendError(w, http.StatusBadRequest, "invalid comment_id")
			return
		}

		userID, ok := middleware.GetUserID(r)
		if !ok {
			response.SendError(w, http.StatusUnauthorized, "unauthorized")
			return
		}

		result, err := commentService.UpdateComment(r.Context(), commentID, userID, req.Comment)
		if err != nil {
			statusCode := http.StatusBadRequest
			switch err {
			case service.ErrCommentNotFound:
				statusCode = http.StatusNotFound
			case service.ErrCommentPermissioDenied:
				statusCode = http.StatusForbidden
			}
			response.SendError(w, statusCode, err.Error())
			return
		}

		resp := CommentResponse{
			ID:        result.ID,
			ExpenseID: result.ExpenseID,
			UserID:    result.UserID,
			Comment:   result.Comment,
			CreatedAt: formatTimestamp(result.CreatedAt),
			UpdatedAt: formatTimestamp(result.UpdatedAt),
		}

		response.SendSuccess(w, http.StatusOK, resp)
	}
}

func DeleteCommentHandler(commentService service.ExpenseCommentService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		commentID, err := parseUUID(chi.URLParam(r, "comment_id"))
		if err != nil {
			response.SendError(w, http.StatusBadRequest, "invalid comment_id")
			return
		}

		userID, ok := middleware.GetUserID(r)
		if !ok {
			response.SendError(w, http.StatusUnauthorized, "unauthorized")
			return
		}

		err = commentService.DeleteComment(r.Context(), commentID, userID)
		if err != nil {
			statusCode := http.StatusBadRequest
			switch err {
			case service.ErrCommentNotFound:
				statusCode = http.StatusNotFound
			case service.ErrCommentPermissioDenied:
				statusCode = http.StatusForbidden
			}
			response.SendError(w, statusCode, err.Error())
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}
