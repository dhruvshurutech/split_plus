package handlers

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/dhruvsaxena1998/splitplus/internal/http/middleware"
	"github.com/dhruvsaxena1998/splitplus/internal/http/response"
	"github.com/dhruvsaxena1998/splitplus/internal/service"
)

type SendFriendRequestRequest struct {
	FriendID string `json:"friend_id" validate:"required,uuid"`
}

type FriendSummaryResponse struct {
	ID          pgtype.UUID `json:"id"`
	FriendID    pgtype.UUID `json:"friend_id"`
	FriendEmail string      `json:"friend_email"`
	FriendName  string      `json:"friend_name,omitempty"`
	AvatarURL   string      `json:"avatar_url,omitempty"`
}

type FriendRequestResponse struct {
	ID           pgtype.UUID `json:"id"`
	UserID       pgtype.UUID `json:"user_id"`
	FriendUserID pgtype.UUID `json:"friend_user_id"`
	Status       string      `json:"status"`
}

func SendFriendRequestHandler(friendService service.FriendService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req, ok := middleware.GetBody[SendFriendRequestRequest](r)
		if !ok {
			response.SendError(w, http.StatusInternalServerError, "invalid request context")
			return
		}

		userID, ok := middleware.GetUserID(r)
		if !ok {
			response.SendError(w, http.StatusUnauthorized, "unauthorized")
			return
		}

		friendID, err := parseUUID(req.FriendID)
		if err != nil {
			response.SendError(w, http.StatusBadRequest, "invalid friend_id")
			return
		}

		fr, err := friendService.SendFriendRequest(r.Context(), userID, friendID)
		if err != nil {
			status := http.StatusBadRequest
			switch err {
			case service.ErrInvalidFriendAction:
				status = http.StatusUnprocessableEntity
			case service.ErrFriendRequestExists:
				status = http.StatusConflict
			}
			response.SendError(w, status, err.Error())
			return
		}

		resp := FriendRequestResponse{
			ID:           fr.ID,
			UserID:       fr.UserID,
			FriendUserID: fr.FriendUserID,
			Status:       fr.Status,
		}

		response.SendSuccess(w, http.StatusCreated, resp)
	}
}

func ListFriendsHandler(friendService service.FriendService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := middleware.GetUserID(r)
		if !ok {
			response.SendError(w, http.StatusUnauthorized, "unauthorized")
			return
		}

		friends, err := friendService.ListFriends(r.Context(), userID)
		if err != nil {
			response.SendError(w, http.StatusBadRequest, err.Error())
			return
		}

		resp := make([]FriendSummaryResponse, len(friends))
		for i, f := range friends {
			resp[i] = FriendSummaryResponse{
				ID:          f.ID,
				FriendID:    f.FriendID,
				FriendEmail: f.FriendEmail,
				FriendName:  f.FriendName,
				AvatarURL:   f.AvatarURL,
			}
		}

		response.SendSuccess(w, http.StatusOK, resp)
	}
}

func ListIncomingFriendRequestsHandler(friendService service.FriendService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := middleware.GetUserID(r)
		if !ok {
			response.SendError(w, http.StatusUnauthorized, "unauthorized")
			return
		}

		requests, err := friendService.ListIncomingRequests(r.Context(), userID)
		if err != nil {
			response.SendError(w, http.StatusBadRequest, err.Error())
			return
		}

		resp := make([]FriendRequestResponse, len(requests))
		for i, fr := range requests {
			resp[i] = FriendRequestResponse{
				ID:           fr.ID,
				UserID:       fr.UserID,
				FriendUserID: fr.FriendUserID,
				Status:       fr.Status,
			}
		}

		response.SendSuccess(w, http.StatusOK, resp)
	}
}

func ListOutgoingFriendRequestsHandler(friendService service.FriendService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := middleware.GetUserID(r)
		if !ok {
			response.SendError(w, http.StatusUnauthorized, "unauthorized")
			return
		}

		requests, err := friendService.ListOutgoingRequests(r.Context(), userID)
		if err != nil {
			response.SendError(w, http.StatusBadRequest, err.Error())
			return
		}

		resp := make([]FriendRequestResponse, len(requests))
		for i, fr := range requests {
			resp[i] = FriendRequestResponse{
				ID:           fr.ID,
				UserID:       fr.UserID,
				FriendUserID: fr.FriendUserID,
				Status:       fr.Status,
			}
		}

		response.SendSuccess(w, http.StatusOK, resp)
	}
}

func AcceptFriendRequestHandler(friendService service.FriendService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := middleware.GetUserID(r)
		if !ok {
			response.SendError(w, http.StatusUnauthorized, "unauthorized")
			return
		}

		idStr := chi.URLParam(r, "id")
		requestID, err := parseUUID(idStr)
		if err != nil {
			response.SendError(w, http.StatusBadRequest, "invalid id")
			return
		}

		fr, err := friendService.AcceptFriendRequest(r.Context(), userID, requestID)
		if err != nil {
			status := http.StatusBadRequest
			switch err {
			case service.ErrFriendNotFound:
				status = http.StatusNotFound
			case service.ErrInvalidFriendAction:
				status = http.StatusUnprocessableEntity
			}
			response.SendError(w, status, err.Error())
			return
		}

		resp := FriendRequestResponse{
			ID:           fr.ID,
			UserID:       fr.UserID,
			FriendUserID: fr.FriendUserID,
			Status:       fr.Status,
		}

		response.SendSuccess(w, http.StatusOK, resp)
	}
}

func DeclineFriendRequestHandler(friendService service.FriendService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := middleware.GetUserID(r)
		if !ok {
			response.SendError(w, http.StatusUnauthorized, "unauthorized")
			return
		}

		idStr := chi.URLParam(r, "id")
		requestID, err := parseUUID(idStr)
		if err != nil {
			response.SendError(w, http.StatusBadRequest, "invalid id")
			return
		}

		err = friendService.DeclineFriendRequest(r.Context(), userID, requestID)
		if err != nil {
			status := http.StatusBadRequest
			switch err {
			case service.ErrFriendNotFound:
				status = http.StatusNotFound
			case service.ErrInvalidFriendAction:
				status = http.StatusUnprocessableEntity
			}
			response.SendError(w, status, err.Error())
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

func RemoveFriendHandler(friendService service.FriendService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := middleware.GetUserID(r)
		if !ok {
			response.SendError(w, http.StatusUnauthorized, "unauthorized")
			return
		}

		friendIDStr := chi.URLParam(r, "friend_id")
		friendID, err := parseUUID(friendIDStr)
		if err != nil {
			response.SendError(w, http.StatusBadRequest, "invalid friend_id")
			return
		}

		err = friendService.RemoveFriend(r.Context(), userID, friendID)
		if err != nil {
			status := http.StatusBadRequest
			switch err {
			case service.ErrFriendNotFound:
				status = http.StatusNotFound
			case service.ErrInvalidFriendAction:
				status = http.StatusUnprocessableEntity
			}
			response.SendError(w, status, err.Error())
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}
