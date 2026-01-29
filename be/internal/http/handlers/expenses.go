package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/dhruvsaxena1998/splitplus/internal/db/sqlc"
	"github.com/dhruvsaxena1998/splitplus/internal/http/middleware"
	"github.com/dhruvsaxena1998/splitplus/internal/http/response"
	"github.com/dhruvsaxena1998/splitplus/internal/service"
)

// Request structs

type PaymentRequest struct {
	UserID        string `json:"user_id,omitempty" validate:"omitempty,uuid"`
	PendingUserID string `json:"pending_user_id,omitempty" validate:"omitempty,uuid"`
	Amount        string `json:"amount" validate:"required"`
	PaymentMethod string `json:"payment_method,omitempty"`
}

type SplitRequest struct {
	UserID        string `json:"user_id,omitempty" validate:"omitempty,uuid"`
	PendingUserID string `json:"pending_user_id,omitempty" validate:"omitempty,uuid"`
	AmountOwned   string `json:"amount_owned" validate:"required"`
	SplitType     string `json:"split_type,omitempty"`
}

type CreateExpenseRequest struct {
	Title        string           `json:"title" validate:"required,min=1,max=200"`
	Notes        string           `json:"notes,omitempty" validate:"max=1000"`
	Amount       string           `json:"amount" validate:"required"`
	CurrencyCode string           `json:"currency_code,omitempty" validate:"omitempty,len=3"`
	Date         string           `json:"date" validate:"required"`
	CategoryID   string           `json:"category_id,omitempty" validate:"omitempty,uuid"`
	Tags         []string         `json:"tags,omitempty"`
	Payments     []PaymentRequest `json:"payments" validate:"required,min=1,dive"`
	Splits       []SplitRequest   `json:"splits" validate:"required,min=1,dive"`
}

type UpdateExpenseRequest struct {
	Title        string           `json:"title" validate:"required,min=1,max=200"`
	Notes        string           `json:"notes,omitempty" validate:"max=1000"`
	Amount       string           `json:"amount" validate:"required"`
	CurrencyCode string           `json:"currency_code,omitempty" validate:"omitempty,len=3"`
	Date         string           `json:"date" validate:"required"`
	CategoryID   string           `json:"category_id,omitempty" validate:"omitempty,uuid"`
	Tags         []string         `json:"tags,omitempty"`
	Payments     []PaymentRequest `json:"payments" validate:"required,min=1,dive"`
	Splits       []SplitRequest   `json:"splits" validate:"required,min=1,dive"`
}

// Response structs

type ExpenseResponse struct {
	ID           pgtype.UUID  `json:"id"`
	GroupID      pgtype.UUID  `json:"group_id"`
	Title        string       `json:"title"`
	Notes        string       `json:"notes,omitempty"`
	Amount       string       `json:"amount"`
	CurrencyCode string       `json:"currency_code"`
	Date         string       `json:"date"`
	CategoryID   *pgtype.UUID `json:"category_id,omitempty"`
	Tags         []string     `json:"tags,omitempty"`
	CreatedAt    string       `json:"created_at"`
	CreatedBy    pgtype.UUID  `json:"created_by"`
	UpdatedAt    string       `json:"updated_at"`
	UpdatedBy    pgtype.UUID  `json:"updated_by"`
}

type PaymentResponse struct {
	ID            pgtype.UUID  `json:"id"`
	ExpenseID     pgtype.UUID  `json:"expense_id"`
	UserID        pgtype.UUID  `json:"user_id"`
	PendingUserID *pgtype.UUID `json:"pending_user_id,omitempty"`
	Amount        string       `json:"amount"`
	PaymentMethod string       `json:"payment_method,omitempty"`
	CreatedAt     string       `json:"created_at"`
	User          UserInfo     `json:"user,omitempty"`
	PendingUser   *UserInfo    `json:"pending_user,omitempty"`
}

type SplitResponse struct {
	ID            pgtype.UUID  `json:"id"`
	ExpenseID     pgtype.UUID  `json:"expense_id"`
	UserID        pgtype.UUID  `json:"user_id"`
	PendingUserID *pgtype.UUID `json:"pending_user_id,omitempty"`
	AmountOwned   string       `json:"amount_owned"`
	SplitType     string       `json:"split_type"`
	CreatedAt     string       `json:"created_at"`
	User          UserInfo     `json:"user,omitempty"`
	PendingUser   *UserInfo    `json:"pending_user,omitempty"`
}

type CreateExpenseResponse struct {
	Expense  ExpenseResponse   `json:"expense"`
	Payments []PaymentResponse `json:"payments"`
	Splits   []SplitResponse   `json:"splits"`
}

// List responses

type GroupExpenseResponse struct {
	Expense  ExpenseResponse   `json:"expense"`
	Payments []PaymentResponse `json:"payments"`
}

// Handlers

func CreateExpenseHandler(expenseService service.ExpenseService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req, ok := middleware.GetBody[CreateExpenseRequest](r)
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

		// Parse date
		date, err := time.Parse("2006-01-02", req.Date)
		if err != nil {
			response.SendError(w, http.StatusBadRequest, "invalid date format, expected YYYY-MM-DD")
			return
		}

		// Convert payments
		payments := make([]service.PaymentInput, len(req.Payments))
		for i, p := range req.Payments {
			var userID pgtype.UUID
			if p.UserID != "" {
				var err error
				userID, err = parseUUID(p.UserID)
				if err != nil {
					response.SendError(w, http.StatusBadRequest, "invalid payment user_id")
					return
				}
			}

			var pendingUserID *pgtype.UUID
			if p.PendingUserID != "" {
				id, err := parseUUID(p.PendingUserID)
				if err != nil {
					response.SendError(w, http.StatusBadRequest, "invalid payment pending_user_id")
					return
				}
				pendingUserID = &id
			}

			payments[i] = service.PaymentInput{
				UserID:        userID,
				PendingUserID: pendingUserID,
				Amount:        p.Amount,
				PaymentMethod: p.PaymentMethod,
			}
		}

		// Convert splits
		splits := make([]service.SplitInput, len(req.Splits))
		for i, s := range req.Splits {
			var userID pgtype.UUID
			if s.UserID != "" {
				var err error
				userID, err = parseUUID(s.UserID)
				if err != nil {
					response.SendError(w, http.StatusBadRequest, "invalid split user_id")
					return
				}
			}

			var pendingUserID *pgtype.UUID
			if s.PendingUserID != "" {
				id, err := parseUUID(s.PendingUserID)
				if err != nil {
					response.SendError(w, http.StatusBadRequest, "invalid split pending_user_id")
					return
				}
				pendingUserID = &id
			}

			splits[i] = service.SplitInput{
				UserID:        userID,
				PendingUserID: pendingUserID,
				AmountOwned:   s.AmountOwned,
				SplitType:     s.SplitType,
			}
		}

		// Parse category_id if provided
		var categoryID *pgtype.UUID
		if req.CategoryID != "" {
			parsedCategoryID, err := parseUUID(req.CategoryID)
			if err != nil {
				response.SendError(w, http.StatusBadRequest, "invalid category_id")
				return
			}
			categoryID = &parsedCategoryID
		}

		result, err := expenseService.CreateExpense(r.Context(), service.CreateExpenseInput{
			GroupID:      groupID,
			Title:        req.Title,
			Notes:        req.Notes,
			Amount:       req.Amount,
			CurrencyCode: req.CurrencyCode,
			Date:         date,
			CategoryID:   categoryID,
			Tags:         req.Tags,
			CreatedBy:    userID,
			Payments:     payments,
			Splits:       splits,
		})
		if err != nil {
			statusCode := http.StatusBadRequest
			switch err {
			case service.ErrExpenseNotFound:
				statusCode = http.StatusNotFound
			case service.ErrNotGroupMember:
				statusCode = http.StatusForbidden
			case service.ErrCategoryNotFound, service.ErrCategoryNotInGroup:
				statusCode = http.StatusBadRequest
			case service.ErrInvalidAmount, service.ErrPaymentTotalMismatch, service.ErrSplitTotalMismatch:
				statusCode = http.StatusUnprocessableEntity
			}
			response.SendError(w, statusCode, err.Error())
			return
		}

		// Load payments and splits with user info so response contains full user details
		paymentRows, err := expenseService.GetExpensePayments(r.Context(), result.Expense.ID, userID)
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

		splitRows, err := expenseService.GetExpenseSplits(r.Context(), result.Expense.ID, userID)
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

		resp := CreateExpenseResponse{
			Expense:  expenseToResponse(result.Expense),
			Payments: paymentsWithUserToResponse(paymentRows),
			Splits:   splitsWithUserToResponse(splitRows),
		}

		response.SendSuccess(w, http.StatusCreated, resp)
	}
}

func GetExpenseHandler(expenseService service.ExpenseService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
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

		expense, err := expenseService.GetExpenseByID(r.Context(), expenseID, userID)
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

		// Load payments and splits with user info
		payments, err := expenseService.GetExpensePayments(r.Context(), expenseID, userID)
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

		splits, err := expenseService.GetExpenseSplits(r.Context(), expenseID, userID)
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

		resp := CreateExpenseResponse{
			Expense:  expenseToResponse(expense),
			Payments: paymentsWithUserToResponse(payments),
			Splits:   splitsWithUserToResponse(splits),
		}

		response.SendSuccess(w, http.StatusOK, resp)
	}
}

func ListExpensesHandler(expenseService service.ExpenseService) http.HandlerFunc {
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

		expenses, err := expenseService.ListExpensesByGroup(r.Context(), groupID, userID)
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

		resp := make([]GroupExpenseResponse, len(expenses))
		for i, e := range expenses {
			// For each expense, load payments with user info so we can show who paid
			paymentRows, err := expenseService.GetExpensePayments(r.Context(), e.ID, userID)
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

			resp[i] = GroupExpenseResponse{
				Expense:  expenseToResponse(e),
				Payments: paymentsWithUserToResponse(paymentRows),
			}
		}

		response.SendSuccess(w, http.StatusOK, resp)
	}
}

func UpdateExpenseHandler(expenseService service.ExpenseService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req, ok := middleware.GetBody[UpdateExpenseRequest](r)
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

		// Parse date
		date, err := time.Parse("2006-01-02", req.Date)
		if err != nil {
			response.SendError(w, http.StatusBadRequest, "invalid date format, expected YYYY-MM-DD")
			return
		}

		// Convert payments
		payments := make([]service.PaymentInput, len(req.Payments))
		for i, p := range req.Payments {
			var paymentUserID pgtype.UUID
			if p.UserID != "" {
				var err error
				paymentUserID, err = parseUUID(p.UserID)
				if err != nil {
					response.SendError(w, http.StatusBadRequest, "invalid payment user_id")
					return
				}
			}

			var pendingUserID *pgtype.UUID
			if p.PendingUserID != "" {
				id, err := parseUUID(p.PendingUserID)
				if err != nil {
					response.SendError(w, http.StatusBadRequest, "invalid payment pending_user_id")
					return
				}
				pendingUserID = &id
			}

			payments[i] = service.PaymentInput{
				UserID:        paymentUserID,
				PendingUserID: pendingUserID,
				Amount:        p.Amount,
				PaymentMethod: p.PaymentMethod,
			}
		}

		// Convert splits
		splits := make([]service.SplitInput, len(req.Splits))
		for i, s := range req.Splits {
			var splitUserID pgtype.UUID
			if s.UserID != "" {
				var err error
				splitUserID, err = parseUUID(s.UserID)
				if err != nil {
					response.SendError(w, http.StatusBadRequest, "invalid split user_id")
					return
				}
			}

			var pendingUserID *pgtype.UUID
			if s.PendingUserID != "" {
				id, err := parseUUID(s.PendingUserID)
				if err != nil {
					response.SendError(w, http.StatusBadRequest, "invalid split pending_user_id")
					return
				}
				pendingUserID = &id
			}

			splits[i] = service.SplitInput{
				UserID:        splitUserID,
				PendingUserID: pendingUserID,
				AmountOwned:   s.AmountOwned,
				SplitType:     s.SplitType,
			}
		}

		// Parse category_id if provided
		var categoryID *pgtype.UUID
		if req.CategoryID != "" {
			parsedCategoryID, err := parseUUID(req.CategoryID)
			if err != nil {
				response.SendError(w, http.StatusBadRequest, "invalid category_id")
				return
			}
			categoryID = &parsedCategoryID
		}

		result, err := expenseService.UpdateExpense(r.Context(), service.UpdateExpenseInput{
			ExpenseID:    expenseID,
			Title:        req.Title,
			Notes:        req.Notes,
			Amount:       req.Amount,
			CurrencyCode: req.CurrencyCode,
			Date:         date,
			CategoryID:   categoryID,
			Tags:         req.Tags,
			UpdatedBy:    userID,
			Payments:     payments,
			Splits:       splits,
		})
		if err != nil {
			statusCode := http.StatusBadRequest
			switch err {
			case service.ErrExpenseNotFound:
				statusCode = http.StatusNotFound
			case service.ErrNotGroupMember:
				statusCode = http.StatusForbidden
			case service.ErrCategoryNotFound, service.ErrCategoryNotInGroup:
				statusCode = http.StatusBadRequest
			case service.ErrInvalidAmount, service.ErrPaymentTotalMismatch, service.ErrSplitTotalMismatch:
				statusCode = http.StatusUnprocessableEntity
			}
			response.SendError(w, statusCode, err.Error())
			return
		}

		// Load payments and splits with user info so response contains full user details
		paymentRows, err := expenseService.GetExpensePayments(r.Context(), result.Expense.ID, userID)
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

		splitRows, err := expenseService.GetExpenseSplits(r.Context(), result.Expense.ID, userID)
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

		resp := CreateExpenseResponse{
			Expense:  expenseToResponse(result.Expense),
			Payments: paymentsWithUserToResponse(paymentRows),
			Splits:   splitsWithUserToResponse(splitRows),
		}

		response.SendSuccess(w, http.StatusOK, resp)
	}
}

func SearchExpensesHandler(expenseService service.ExpenseService) http.HandlerFunc {
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

		// Parse query params
		var queryPtr *string
		if q := r.URL.Query().Get("q"); q != "" {
			queryPtr = &q
		}

		var startDate *time.Time
		if s := r.URL.Query().Get("start_date"); s != "" {
			t, err := time.Parse("2006-01-02", s)
			if err != nil {
				response.SendError(w, http.StatusBadRequest, "invalid start_date format")
				return
			}
			startDate = &t
		}

		var endDate *time.Time
		if s := r.URL.Query().Get("end_date"); s != "" {
			t, err := time.Parse("2006-01-02", s)
			if err != nil {
				response.SendError(w, http.StatusBadRequest, "invalid end_date format")
				return
			}
			endDate = &t
		}

		var categoryID *pgtype.UUID
		if s := r.URL.Query().Get("category_id"); s != "" {
			id, err := parseUUID(s)
			if err != nil {
				response.SendError(w, http.StatusBadRequest, "invalid category_id")
				return
			}
			categoryID = &id
		}

		var createdBy *pgtype.UUID
		if s := r.URL.Query().Get("created_by"); s != "" {
			id, err := parseUUID(s)
			if err != nil {
				response.SendError(w, http.StatusBadRequest, "invalid created_by")
				return
			}
			createdBy = &id
		}

		var minAmount *string
		if s := r.URL.Query().Get("min_amount"); s != "" {
			minAmount = &s
		}

		var maxAmount *string
		if s := r.URL.Query().Get("max_amount"); s != "" {
			maxAmount = &s
		}

		var payerID *pgtype.UUID
		if s := r.URL.Query().Get("payer_id"); s != "" {
			id, err := parseUUID(s)
			if err != nil {
				response.SendError(w, http.StatusBadRequest, "invalid payer_id")
				return
			}
			payerID = &id
		}

		var owerID *pgtype.UUID
		if s := r.URL.Query().Get("ower_id"); s != "" {
			id, err := parseUUID(s)
			if err != nil {
				response.SendError(w, http.StatusBadRequest, "invalid ower_id")
				return
			}
			owerID = &id
		}
		
		limit := 20
		if s := r.URL.Query().Get("limit"); s != "" {
			if l, err := strconv.Atoi(s); err == nil && l > 0 {
				limit = l
			}
		}

		offset := 0
		if s := r.URL.Query().Get("offset"); s != "" {
			if o, err := strconv.Atoi(s); err == nil && o >= 0 {
				offset = o
			}
		}
		
		input := service.SearchExpensesInput{
			GroupID:    groupID,
			Query:      queryPtr,
			StartDate:  startDate,
			EndDate:    endDate,
			CategoryID: categoryID,
			CreatedBy:  createdBy,
			MinAmount:  minAmount,
			MaxAmount:  maxAmount,
			PayerID:    payerID,
			OwerID:     owerID,
			Limit:      int32(limit),
			Offset:     int32(offset),
		}

		expenses, err := expenseService.SearchExpenses(r.Context(), input, userID)
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

		resp := make([]GroupExpenseResponse, len(expenses))
		for i, e := range expenses {
			paymentRows, err := expenseService.GetExpensePayments(r.Context(), e.ID, userID)
			if err != nil {
				continue
			}
			
			resp[i] = GroupExpenseResponse{
				Expense:  expenseToResponse(e),
				Payments: paymentsWithUserToResponse(paymentRows),
			}
		}

		response.SendSuccess(w, http.StatusOK, resp)
	}
}

func DeleteExpenseHandler(expenseService service.ExpenseService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
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

		err = expenseService.DeleteExpense(r.Context(), expenseID, userID)
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

		w.WriteHeader(http.StatusNoContent)
	}
}

// Helper functions

func expenseToResponse(expense sqlc.Expense) ExpenseResponse {
	amount, _ := numericToString(expense.Amount)
	dateStr := ""
	if expense.Date.Valid {
		dateStr = expense.Date.Time.Format("2006-01-02")
	}

	var categoryID *pgtype.UUID
	if expense.CategoryID.Valid {
		categoryID = &expense.CategoryID
	}

	tags := []string{}
	if expense.Tags != nil {
		tags = expense.Tags
	}

	return ExpenseResponse{
		ID:           expense.ID,
		GroupID:      expense.GroupID,
		Title:        expense.Title,
		Notes:        expense.Notes.String,
		Amount:       amount,
		CurrencyCode: expense.CurrencyCode,
		Date:         dateStr,
		CategoryID:   categoryID,
		Tags:         tags,
		CreatedAt:    formatTimestamp(expense.CreatedAt),
		CreatedBy:    expense.CreatedBy,
		UpdatedAt:    formatTimestamp(expense.UpdatedAt),
		UpdatedBy:    expense.UpdatedBy,
	}
}

func paymentsToResponse(payments []sqlc.ExpensePayment) []PaymentResponse {
	resp := make([]PaymentResponse, len(payments))
	for i, p := range payments {
		amount, _ := numericToString(p.Amount)
		resp[i] = PaymentResponse{
			ID:            p.ID,
			ExpenseID:     p.ExpenseID,
			UserID:        p.UserID,
			Amount:        amount,
			PaymentMethod: p.PaymentMethod.String,
			CreatedAt:     formatTimestamp(p.CreatedAt),
		}
	}
	return resp
}

func paymentsWithUserToResponse(payments []sqlc.ListExpensePaymentsRow) []PaymentResponse {
	resp := make([]PaymentResponse, len(payments))
	for i, p := range payments {
		amount, _ := numericToString(p.Amount)

		var pendingUserID *pgtype.UUID
		if p.PendingUserID.Valid {
			pendingUserID = &p.PendingUserID
		}

		var pendingUser *UserInfo
		if p.PendingUserID.Valid {
			pendingUser = &UserInfo{
				Email: p.PendingUserEmail.String,
				Name:  p.PendingUserName.String,
			}
		}

		resp[i] = PaymentResponse{
			ID:            p.ID,
			ExpenseID:     p.ExpenseID,
			UserID:        p.UserID,
			PendingUserID: pendingUserID,
			Amount:        amount,
			PaymentMethod: p.PaymentMethod.String,
			CreatedAt:     formatTimestamp(p.CreatedAt),
			User: UserInfo{
				Email:     p.UserEmail.String,
				Name:      p.UserName.String,
				AvatarURL: p.UserAvatarUrl.String,
			},
			PendingUser: pendingUser,
		}
	}
	return resp
}

func splitsToResponse(splits []sqlc.ExpenseSplit) []SplitResponse {
	resp := make([]SplitResponse, len(splits))
	for i, s := range splits {
		amount, _ := numericToString(s.AmountOwned)
		resp[i] = SplitResponse{
			ID:          s.ID,
			ExpenseID:   s.ExpenseID,
			UserID:      s.UserID,
			AmountOwned: amount,
			SplitType:   s.SplitType,
			CreatedAt:   formatTimestamp(s.CreatedAt),
		}
	}
	return resp
}

func splitsWithUserToResponse(splits []sqlc.ListExpenseSplitsRow) []SplitResponse {
	resp := make([]SplitResponse, len(splits))
	for i, s := range splits {
		amount, _ := numericToString(s.AmountOwned)

		var pendingUserID *pgtype.UUID
		if s.PendingUserID.Valid {
			pendingUserID = &s.PendingUserID
		}

		var pendingUser *UserInfo
		if s.PendingUserID.Valid {
			pendingUser = &UserInfo{
				Email: s.PendingUserEmail.String,
				Name:  s.PendingUserName.String,
			}
		}

		resp[i] = SplitResponse{
			ID:            s.ID,
			ExpenseID:     s.ExpenseID,
			UserID:        s.UserID,
			PendingUserID: pendingUserID,
			AmountOwned:   amount,
			SplitType:     s.SplitType,
			CreatedAt:     formatTimestamp(s.CreatedAt),
			User: UserInfo{
				Email:     s.UserEmail.String,
				Name:      s.UserName.String,
				AvatarURL: s.UserAvatarUrl.String,
			},
			PendingUser: pendingUser,
		}
	}
	return resp
}

func numericToString(n pgtype.Numeric) (string, error) {
	if !n.Valid {
		return "0", nil
	}
	val, err := n.Value()
	if err != nil {
		return "0", err
	}
	if str, ok := val.(string); ok {
		return str, nil
	}
	return "0", nil
}
