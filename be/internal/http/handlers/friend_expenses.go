package handlers

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/dhruvsaxena1998/splitplus/internal/http/middleware"
	"github.com/dhruvsaxena1998/splitplus/internal/http/response"
	"github.com/dhruvsaxena1998/splitplus/internal/service"
)

// Friend expense handlers reuse CreateExpenseRequest / responses from expenses.go

func CreateFriendExpenseHandler(friendExpenseService service.FriendExpenseService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req, ok := middleware.GetBody[CreateExpenseRequest](r)
		if !ok {
			response.SendError(w, http.StatusInternalServerError, "invalid request context")
			return
		}

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

		// Parse date
		date, err := time.Parse("2006-01-02", req.Date)
		if err != nil {
			response.SendError(w, http.StatusBadRequest, "invalid date format, expected YYYY-MM-DD")
			return
		}

		// Convert payments
		payments := make([]service.PaymentInput, len(req.Payments))
		for i, p := range req.Payments {
			paymentUserID, err := parseUUID(p.UserID)
			if err != nil {
				response.SendError(w, http.StatusBadRequest, "invalid payment user_id")
				return
			}
			payments[i] = service.PaymentInput{
				UserID:        paymentUserID,
				Amount:        p.Amount,
				PaymentMethod: p.PaymentMethod,
			}
		}

		// Convert splits
		splits := make([]service.SplitInput, len(req.Splits))
		for i, s := range req.Splits {
			splitUserID, err := parseUUID(s.UserID)
			if err != nil {
				response.SendError(w, http.StatusBadRequest, "invalid split user_id")
				return
			}
			splits[i] = service.SplitInput{
				UserID:      splitUserID,
				AmountOwned: s.AmountOwned,
				SplitType:   s.SplitType,
			}
		}

		result, err := friendExpenseService.CreateFriendExpense(r.Context(), userID, friendID, service.CreateExpenseInput{
			Title:        req.Title,
			Notes:        req.Notes,
			Amount:       req.Amount,
			CurrencyCode: req.CurrencyCode,
			Date:         date,
			CreatedBy:    userID,
			Payments:     payments,
			Splits:       splits,
		})
		if err != nil {
			statusCode := http.StatusBadRequest
			switch err {
			case service.ErrInvalidAmount, service.ErrPaymentTotalMismatch, service.ErrSplitTotalMismatch:
				statusCode = http.StatusUnprocessableEntity
			case service.ErrFriendNotFound, service.ErrInvalidFriendAction:
				statusCode = http.StatusForbidden
			}
			response.SendError(w, statusCode, err.Error())
			return
		}

		// Load payments and splits with user info
		paymentRows, err := friendExpenseService.GetFriendExpensePayments(r.Context(), result.Expense.ID, userID, friendID)
		if err != nil {
			response.SendError(w, http.StatusBadRequest, err.Error())
			return
		}
		splitRows, err := friendExpenseService.GetFriendExpenseSplits(r.Context(), result.Expense.ID, userID, friendID)
		if err != nil {
			response.SendError(w, http.StatusBadRequest, err.Error())
			return
		}

		resp := CreateExpenseResponse{
			Expense:  expenseToResponse(result.Expense),
			Payments: paymentsWithUserToResponse(paymentRows),
			Splits:   splitsWithUserToResponse(splitRows),
		}

		response.SendSuccess(w, http.StatusCreated, resp)
	}
}

func ListFriendExpensesHandler(friendExpenseService service.FriendExpenseService) http.HandlerFunc {
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

		expenses, err := friendExpenseService.ListFriendExpenses(r.Context(), userID, friendID)
		if err != nil {
			statusCode := http.StatusBadRequest
			switch err {
			case service.ErrFriendNotFound, service.ErrInvalidFriendAction:
				statusCode = http.StatusForbidden
			}
			response.SendError(w, statusCode, err.Error())
			return
		}

		// For each expense, load payments with user info (similar to group list)
		resp := make([]GroupExpenseResponse, len(expenses))
		for i, e := range expenses {
			paymentRows, err := friendExpenseService.GetFriendExpensePayments(r.Context(), e.ID, userID, friendID)
			if err != nil {
				response.SendError(w, http.StatusBadRequest, err.Error())
				return
			}

			resp[i] = GroupExpenseResponse{
				Expense:  expenseToResponse(e),
				Payments: paymentsWithUserToResponse(paymentRows),
			}
		}

		response.SendSuccess(w, http.StatusOK, resp)
	}
}
