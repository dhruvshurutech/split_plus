package handlers

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/dhruvsaxena1998/splitplus/internal/http/middleware"
	"github.com/dhruvsaxena1998/splitplus/internal/http/response"
	"github.com/dhruvsaxena1998/splitplus/internal/service"
)

// Handlers

func ListGroupBalancesHandler(balanceService service.BalanceService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		requesterID, ok := middleware.GetUserID(r)
		if !ok {
			response.SendError(w, http.StatusUnauthorized, "unauthorized")
			return
		}

		groupIDStr := chi.URLParam(r, "group_id")
		var groupID pgtype.UUID
		if err := groupID.Scan(groupIDStr); err != nil {
			response.SendError(w, http.StatusBadRequest, "invalid group id")
			return
		}

		balances, err := balanceService.GetGroupBalances(r.Context(), groupID, requesterID)
		if err != nil {
			var statusCode int
			switch err {
			case service.ErrGroupNotFound:
				statusCode = http.StatusNotFound
			case service.ErrNotGroupMember:
				statusCode = http.StatusForbidden
			default:
				statusCode = http.StatusInternalServerError
			}
			response.SendError(w, statusCode, err.Error())
			return
		}

		response.SendSuccess(w, http.StatusOK, balances)
	}
}

func GetUserBalanceInGroupHandler(balanceService service.BalanceService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		requesterID, ok := middleware.GetUserID(r)
		if !ok {
			response.SendError(w, http.StatusUnauthorized, "unauthorized")
			return
		}

		groupIDStr := chi.URLParam(r, "group_id")
		var groupID pgtype.UUID
		if err := groupID.Scan(groupIDStr); err != nil {
			response.SendError(w, http.StatusBadRequest, "invalid group id")
			return
		}

		userIDStr := chi.URLParam(r, "user_id")
		var userID pgtype.UUID
		if err := userID.Scan(userIDStr); err != nil {
			response.SendError(w, http.StatusBadRequest, "invalid user id")
			return
		}

		balance, err := balanceService.GetUserBalanceInGroup(r.Context(), groupID, userID, requesterID)
		if err != nil {
			var statusCode int
			switch err {
			case service.ErrGroupNotFound, service.ErrBalanceNotFound:
				statusCode = http.StatusNotFound
			case service.ErrNotGroupMember:
				statusCode = http.StatusForbidden
			default:
				statusCode = http.StatusInternalServerError
			}
			response.SendError(w, statusCode, err.Error())
			return
		}

		response.SendSuccess(w, http.StatusOK, balance)
	}
}

func GetOverallUserBalanceHandler(balanceService service.BalanceService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := middleware.GetUserID(r)
		if !ok {
			response.SendError(w, http.StatusUnauthorized, "unauthorized")
			return
		}

		balances, err := balanceService.GetOverallUserBalance(r.Context(), userID)
		if err != nil {
			response.SendError(w, http.StatusInternalServerError, err.Error())
			return
		}

		response.SendSuccess(w, http.StatusOK, balances)
	}
}

func GetSimplifiedDebtsHandler(balanceService service.BalanceService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		requesterID, ok := middleware.GetUserID(r)
		if !ok {
			response.SendError(w, http.StatusUnauthorized, "unauthorized")
			return
		}

		groupIDStr := chi.URLParam(r, "group_id")
		var groupID pgtype.UUID
		if err := groupID.Scan(groupIDStr); err != nil {
			response.SendError(w, http.StatusBadRequest, "invalid group id")
			return
		}

		debts, err := balanceService.GetSimplifiedDebts(r.Context(), groupID, requesterID)
		if err != nil {
			var statusCode int
			switch err {
			case service.ErrGroupNotFound:
				statusCode = http.StatusNotFound
			case service.ErrNotGroupMember:
				statusCode = http.StatusForbidden
			default:
				statusCode = http.StatusInternalServerError
			}
			response.SendError(w, statusCode, err.Error())
			return
		}

		response.SendSuccess(w, http.StatusOK, debts)
	}
}
