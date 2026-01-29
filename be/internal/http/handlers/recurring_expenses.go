package handlers

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/dhruvsaxena1998/splitplus/internal/db/sqlc"
	"github.com/dhruvsaxena1998/splitplus/internal/http/middleware"
	"github.com/dhruvsaxena1998/splitplus/internal/http/response"
	"github.com/dhruvsaxena1998/splitplus/internal/service"
)

// Request structs

type CreateRecurringExpenseRequest struct {
	Title          string           `json:"title" validate:"required,min=1,max=200"`
	Notes          string           `json:"notes,omitempty" validate:"max=1000"`
	Amount         string           `json:"amount" validate:"required"`
	CurrencyCode   string           `json:"currency_code,omitempty" validate:"omitempty,len=3"`
	RepeatInterval string           `json:"repeat_interval" validate:"required,oneof=daily weekly monthly yearly"`
	DayOfMonth     *int             `json:"day_of_month,omitempty" validate:"omitempty,min=1,max=31"`
	DayOfWeek      *int             `json:"day_of_week,omitempty" validate:"omitempty,min=0,max=6"`
	StartDate      string           `json:"start_date" validate:"required"`
	EndDate        *string          `json:"end_date,omitempty"`
	Payments       []PaymentRequest `json:"payments" validate:"required,min=1,dive"`
	Splits         []SplitRequest   `json:"splits" validate:"required,min=1,dive"`
}

type UpdateRecurringExpenseRequest struct {
	Title          string           `json:"title" validate:"required,min=1,max=200"`
	Notes          string           `json:"notes,omitempty" validate:"max=1000"`
	Amount         string           `json:"amount" validate:"required"`
	CurrencyCode   string           `json:"currency_code,omitempty" validate:"omitempty,len=3"`
	RepeatInterval string           `json:"repeat_interval" validate:"required,oneof=daily weekly monthly yearly"`
	DayOfMonth     *int             `json:"day_of_month,omitempty" validate:"omitempty,min=1,max=31"`
	DayOfWeek      *int             `json:"day_of_week,omitempty" validate:"omitempty,min=0,max=6"`
	StartDate      string           `json:"start_date" validate:"required"`
	EndDate        *string          `json:"end_date,omitempty"`
	IsActive       bool             `json:"is_active"`
	Payments       []PaymentRequest `json:"payments" validate:"required,min=1,dive"`
	Splits         []SplitRequest   `json:"splits" validate:"required,min=1,dive"`
}

// Response structs

type RecurringExpenseResponse struct {
	ID                 pgtype.UUID       `json:"id"`
	GroupID            pgtype.UUID       `json:"group_id"`
	Title              string            `json:"title"`
	Notes              string            `json:"notes,omitempty"`
	Amount             string            `json:"amount"`
	CurrencyCode       string            `json:"currency_code"`
	RepeatInterval     string            `json:"repeat_interval"`
	DayOfMonth         *int              `json:"day_of_month,omitempty"`
	DayOfWeek          *int              `json:"day_of_week,omitempty"`
	StartDate          string            `json:"start_date"`
	EndDate            *string           `json:"end_date,omitempty"`
	NextOccurrenceDate string            `json:"next_occurrence_date"`
	IsActive           bool              `json:"is_active"`
	CreatedAt          string            `json:"created_at"`
	CreatedBy          pgtype.UUID       `json:"created_by"`
	UpdatedAt          string            `json:"updated_at"`
	UpdatedBy          pgtype.UUID       `json:"updated_by"`
	Payments           []PaymentResponse `json:"payments,omitempty"`
	Splits             []SplitResponse   `json:"splits,omitempty"`
}

// Handlers

func CreateRecurringExpenseHandler(recurringExpenseService service.RecurringExpenseService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req, ok := middleware.GetBody[CreateRecurringExpenseRequest](r)
		if !ok {
			response.SendError(w, http.StatusInternalServerError, "invalid request context")
			return
		}

		groupID, err := parseUUID(chi.URLParam(r, "group_id"))
		if err != nil {
			response.SendError(w, http.StatusBadRequest, "invalid group_id")
			return
		}

		userID, ok := middleware.GetUserID(r)
		if !ok {
			response.SendError(w, http.StatusUnauthorized, "unauthorized")
			return
		}

		// Parse dates
		startDate, err := time.Parse("2006-01-02", req.StartDate)
		if err != nil {
			response.SendError(w, http.StatusBadRequest, "invalid start_date format, expected YYYY-MM-DD")
			return
		}

		var endDate *time.Time
		if req.EndDate != nil {
			parsed, err := time.Parse("2006-01-02", *req.EndDate)
			if err != nil {
				response.SendError(w, http.StatusBadRequest, "invalid end_date format, expected YYYY-MM-DD")
				return
			}
			endDate = &parsed
		}

		// Convert payments
		payments := make([]service.RecurringPaymentInput, len(req.Payments))
		for i, p := range req.Payments {
			paymentUserID, err := parseUUID(p.UserID)
			if err != nil {
				response.SendError(w, http.StatusBadRequest, "invalid payment user_id")
				return
			}
			payments[i] = service.RecurringPaymentInput{
				UserID:        paymentUserID,
				Amount:        p.Amount,
				PaymentMethod: p.PaymentMethod,
			}
		}

		// Convert splits
		splits := make([]service.RecurringSplitInput, len(req.Splits))
		for i, s := range req.Splits {
			splitUserID, err := parseUUID(s.UserID)
			if err != nil {
				response.SendError(w, http.StatusBadRequest, "invalid split user_id")
				return
			}
			splits[i] = service.RecurringSplitInput{
				UserID:      splitUserID,
				AmountOwned: s.AmountOwned,
				SplitType:   s.SplitType,
			}
		}

		result, err := recurringExpenseService.CreateRecurringExpense(r.Context(), service.CreateRecurringExpenseInput{
			GroupID:        groupID,
			Title:          req.Title,
			Notes:          req.Notes,
			Amount:         req.Amount,
			CurrencyCode:   req.CurrencyCode,
			RepeatInterval: req.RepeatInterval,
			DayOfMonth:     req.DayOfMonth,
			DayOfWeek:      req.DayOfWeek,
			StartDate:      startDate,
			EndDate:        endDate,
			CreatedBy:      userID,
			Payments:       payments,
			Splits:         splits,
		})
		if err != nil {
			statusCode := http.StatusBadRequest
			switch err {
			case service.ErrRecurringExpenseNotFound:
				statusCode = http.StatusNotFound
			case service.ErrGroupNotFound:
				statusCode = http.StatusNotFound
			case service.ErrNotGroupMember:
				statusCode = http.StatusForbidden
			case service.ErrInvalidRecurringInterval, service.ErrInvalidDateConfiguration:
				statusCode = http.StatusUnprocessableEntity
			case service.ErrInvalidAmount, service.ErrPaymentTotalMismatch, service.ErrSplitTotalMismatch:
				statusCode = http.StatusUnprocessableEntity
			}
			response.SendError(w, statusCode, err.Error())
			return
		}

		resp := recurringExpenseToResponse(result)
		response.SendSuccess(w, http.StatusCreated, resp)
	}
}

func GetRecurringExpenseHandler(recurringExpenseService service.RecurringExpenseService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		recurringExpenseID, err := parseUUID(chi.URLParam(r, "recurring_expense_id"))
		if err != nil {
			response.SendError(w, http.StatusBadRequest, "invalid recurring_expense_id")
			return
		}

		userID, ok := middleware.GetUserID(r)
		if !ok {
			response.SendError(w, http.StatusUnauthorized, "unauthorized")
			return
		}

		recurringExpense, err := recurringExpenseService.GetRecurringExpenseByID(r.Context(), recurringExpenseID, userID)
		if err != nil {
			statusCode := http.StatusBadRequest
			switch err {
			case service.ErrRecurringExpenseNotFound:
				statusCode = http.StatusNotFound
			case service.ErrNotGroupMember:
				statusCode = http.StatusForbidden
			}
			response.SendError(w, statusCode, err.Error())
			return
		}

		resp := recurringExpenseToResponse(recurringExpense)
		response.SendSuccess(w, http.StatusOK, resp)
	}
}

func ListRecurringExpensesHandler(recurringExpenseService service.RecurringExpenseService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		groupID, err := parseUUID(chi.URLParam(r, "group_id"))
		if err != nil {
			response.SendError(w, http.StatusBadRequest, "invalid group_id")
			return
		}

		userID, ok := middleware.GetUserID(r)
		if !ok {
			response.SendError(w, http.StatusUnauthorized, "unauthorized")
			return
		}

		recurringExpenses, err := recurringExpenseService.ListRecurringExpensesByGroup(r.Context(), groupID, userID)
		if err != nil {
			statusCode := http.StatusBadRequest
			switch err {
			case service.ErrNotGroupMember:
				statusCode = http.StatusForbidden
			}
			response.SendError(w, statusCode, err.Error())
			return
		}

		resp := make([]RecurringExpenseResponse, len(recurringExpenses))
		for i, re := range recurringExpenses {
			resp[i] = recurringExpenseToResponse(re)
		}

		response.SendSuccess(w, http.StatusOK, resp)
	}
}

func UpdateRecurringExpenseHandler(recurringExpenseService service.RecurringExpenseService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req, ok := middleware.GetBody[UpdateRecurringExpenseRequest](r)
		if !ok {
			response.SendError(w, http.StatusInternalServerError, "invalid request context")
			return
		}

		recurringExpenseID, err := parseUUID(chi.URLParam(r, "recurring_expense_id"))
		if err != nil {
			response.SendError(w, http.StatusBadRequest, "invalid recurring_expense_id")
			return
		}

		userID, ok := middleware.GetUserID(r)
		if !ok {
			response.SendError(w, http.StatusUnauthorized, "unauthorized")
			return
		}

		// Parse dates
		startDate, err := time.Parse("2006-01-02", req.StartDate)
		if err != nil {
			response.SendError(w, http.StatusBadRequest, "invalid start_date format, expected YYYY-MM-DD")
			return
		}

		var endDate *time.Time
		if req.EndDate != nil {
			parsed, err := time.Parse("2006-01-02", *req.EndDate)
			if err != nil {
				response.SendError(w, http.StatusBadRequest, "invalid end_date format, expected YYYY-MM-DD")
				return
			}
			endDate = &parsed
		}

		// Convert payments
		payments := make([]service.RecurringPaymentInput, len(req.Payments))
		for i, p := range req.Payments {
			paymentUserID, err := parseUUID(p.UserID)
			if err != nil {
				response.SendError(w, http.StatusBadRequest, "invalid payment user_id")
				return
			}
			payments[i] = service.RecurringPaymentInput{
				UserID:        paymentUserID,
				Amount:        p.Amount,
				PaymentMethod: p.PaymentMethod,
			}
		}

		// Convert splits
		splits := make([]service.RecurringSplitInput, len(req.Splits))
		for i, s := range req.Splits {
			splitUserID, err := parseUUID(s.UserID)
			if err != nil {
				response.SendError(w, http.StatusBadRequest, "invalid split user_id")
				return
			}
			splits[i] = service.RecurringSplitInput{
				UserID:      splitUserID,
				AmountOwned: s.AmountOwned,
				SplitType:   s.SplitType,
			}
		}

		result, err := recurringExpenseService.UpdateRecurringExpense(r.Context(), service.UpdateRecurringExpenseInput{
			RecurringExpenseID: recurringExpenseID,
			Title:              req.Title,
			Notes:              req.Notes,
			Amount:             req.Amount,
			CurrencyCode:       req.CurrencyCode,
			RepeatInterval:     req.RepeatInterval,
			DayOfMonth:         req.DayOfMonth,
			DayOfWeek:          req.DayOfWeek,
			StartDate:          startDate,
			EndDate:            endDate,
			IsActive:           req.IsActive,
			UpdatedBy:          userID,
			Payments:           payments,
			Splits:             splits,
		})
		if err != nil {
			statusCode := http.StatusBadRequest
			switch err {
			case service.ErrRecurringExpenseNotFound:
				statusCode = http.StatusNotFound
			case service.ErrNotGroupMember:
				statusCode = http.StatusForbidden
			case service.ErrInvalidRecurringInterval, service.ErrInvalidDateConfiguration:
				statusCode = http.StatusUnprocessableEntity
			case service.ErrInvalidAmount:
				statusCode = http.StatusUnprocessableEntity
			}
			response.SendError(w, statusCode, err.Error())
			return
		}

		resp := recurringExpenseToResponse(result)
		response.SendSuccess(w, http.StatusOK, resp)
	}
}

func DeleteRecurringExpenseHandler(recurringExpenseService service.RecurringExpenseService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		recurringExpenseID, err := parseUUID(chi.URLParam(r, "recurring_expense_id"))
		if err != nil {
			response.SendError(w, http.StatusBadRequest, "invalid recurring_expense_id")
			return
		}

		userID, ok := middleware.GetUserID(r)
		if !ok {
			response.SendError(w, http.StatusUnauthorized, "unauthorized")
			return
		}

		err = recurringExpenseService.DeleteRecurringExpense(r.Context(), recurringExpenseID, userID)
		if err != nil {
			statusCode := http.StatusBadRequest
			switch err {
			case service.ErrRecurringExpenseNotFound:
				statusCode = http.StatusNotFound
			case service.ErrNotGroupMember:
				statusCode = http.StatusForbidden
			}
			response.SendError(w, statusCode, err.Error())
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

func GenerateExpenseFromRecurringHandler(recurringExpenseService service.RecurringExpenseService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		recurringExpenseID, err := parseUUID(chi.URLParam(r, "recurring_expense_id"))
		if err != nil {
			response.SendError(w, http.StatusBadRequest, "invalid recurring_expense_id")
			return
		}

		userID, ok := middleware.GetUserID(r)
		if !ok {
			response.SendError(w, http.StatusUnauthorized, "unauthorized")
			return
		}

		result, err := recurringExpenseService.GenerateExpenseFromRecurring(r.Context(), recurringExpenseID, userID)
		if err != nil {
			statusCode := http.StatusBadRequest
			switch err {
			case service.ErrRecurringExpenseNotFound:
				statusCode = http.StatusNotFound
			case service.ErrNotGroupMember:
				statusCode = http.StatusForbidden
			}
			response.SendError(w, statusCode, err.Error())
			return
		}

		// Convert to expense response format
		resp := CreateExpenseResponse{
			Expense:  expenseToResponse(result.Expense),
			Payments: paymentsToResponse(result.Payments),
			Splits:   splitsToResponse(result.Splits),
		}

		response.SendSuccess(w, http.StatusCreated, resp)
	}
}

// Helper functions

func recurringExpenseToResponse(re sqlc.RecurringExpense) RecurringExpenseResponse {
	amountStr, _ := numericToString(re.Amount)

	var dayOfMonth *int
	if re.DayOfMonth.Valid {
		day := int(re.DayOfMonth.Int32)
		dayOfMonth = &day
	}

	var dayOfWeek *int
	if re.DayOfWeek.Valid {
		day := int(re.DayOfWeek.Int32)
		dayOfWeek = &day
	}

	var endDate *string
	if re.EndDate.Valid {
		dateStr := re.EndDate.Time.Format("2006-01-02")
		endDate = &dateStr
	}

	return RecurringExpenseResponse{
		ID:                 re.ID,
		GroupID:            re.GroupID,
		Title:              re.Title,
		Notes:              re.Notes.String,
		Amount:             amountStr,
		CurrencyCode:       re.CurrencyCode,
		RepeatInterval:     re.RepeatInterval,
		DayOfMonth:         dayOfMonth,
		DayOfWeek:          dayOfWeek,
		StartDate:          re.StartDate.Time.Format("2006-01-02"),
		EndDate:            endDate,
		NextOccurrenceDate: re.NextOccurrenceDate.Time.Format("2006-01-02"),
		IsActive:           re.IsActive.Bool,
		CreatedAt:          formatTimestamp(re.CreatedAt),
		CreatedBy:          re.CreatedBy,
		UpdatedAt:          formatTimestamp(re.UpdatedAt),
		UpdatedBy:          re.UpdatedBy,
	}
}
