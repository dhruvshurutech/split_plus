package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/dhruvsaxena1998/splitplus/internal/db/sqlc"
	"github.com/dhruvsaxena1998/splitplus/internal/http/middleware"
	"github.com/dhruvsaxena1998/splitplus/internal/service"
	"github.com/dhruvsaxena1998/splitplus/internal/testutil"
)

// MockExpenseService for testing
type MockExpenseService struct {
	CreateExpenseFunc       func(ctx context.Context, input service.CreateExpenseInput) (service.CreateExpenseResult, error)
	GetExpenseByIDFunc      func(ctx context.Context, expenseID, requesterID pgtype.UUID) (sqlc.Expense, error)
	ListExpensesByGroupFunc func(ctx context.Context, groupID, requesterID pgtype.UUID) ([]sqlc.Expense, error)
	UpdateExpenseFunc       func(ctx context.Context, input service.UpdateExpenseInput) (service.CreateExpenseResult, error)
	DeleteExpenseFunc       func(ctx context.Context, expenseID, requesterID pgtype.UUID) error
	GetExpensePaymentsFunc  func(ctx context.Context, expenseID, requesterID pgtype.UUID) ([]sqlc.ListExpensePaymentsRow, error)
	GetExpenseSplitsFunc    func(ctx context.Context, expenseID, requesterID pgtype.UUID) ([]sqlc.ListExpenseSplitsRow, error)
	SearchExpensesFunc      func(ctx context.Context, input service.SearchExpensesInput, requesterID pgtype.UUID) ([]sqlc.Expense, error)
}

func (m *MockExpenseService) CreateExpense(ctx context.Context, input service.CreateExpenseInput) (service.CreateExpenseResult, error) {
	if m.CreateExpenseFunc != nil {
		return m.CreateExpenseFunc(ctx, input)
	}
	return service.CreateExpenseResult{}, nil
}

func (m *MockExpenseService) GetExpenseByID(ctx context.Context, expenseID, requesterID pgtype.UUID) (sqlc.Expense, error) {
	if m.GetExpenseByIDFunc != nil {
		return m.GetExpenseByIDFunc(ctx, expenseID, requesterID)
	}
	return sqlc.Expense{}, errors.New("not found")
}

func (m *MockExpenseService) ListExpensesByGroup(ctx context.Context, groupID, requesterID pgtype.UUID) ([]sqlc.Expense, error) {
	if m.ListExpensesByGroupFunc != nil {
		return m.ListExpensesByGroupFunc(ctx, groupID, requesterID)
	}
	return []sqlc.Expense{}, nil
}

func (m *MockExpenseService) UpdateExpense(ctx context.Context, input service.UpdateExpenseInput) (service.CreateExpenseResult, error) {
	if m.UpdateExpenseFunc != nil {
		return m.UpdateExpenseFunc(ctx, input)
	}
	return service.CreateExpenseResult{}, nil
}

func (m *MockExpenseService) DeleteExpense(ctx context.Context, expenseID, requesterID pgtype.UUID) error {
	if m.DeleteExpenseFunc != nil {
		return m.DeleteExpenseFunc(ctx, expenseID, requesterID)
	}
	return nil
}

func (m *MockExpenseService) GetExpensePayments(ctx context.Context, expenseID, requesterID pgtype.UUID) ([]sqlc.ListExpensePaymentsRow, error) {
	if m.GetExpensePaymentsFunc != nil {
		return m.GetExpensePaymentsFunc(ctx, expenseID, requesterID)
	}
	return []sqlc.ListExpensePaymentsRow{}, nil
}

func (m *MockExpenseService) GetExpenseSplits(ctx context.Context, expenseID, requesterID pgtype.UUID) ([]sqlc.ListExpenseSplitsRow, error) {
	if m.GetExpenseSplitsFunc != nil {
		return m.GetExpenseSplitsFunc(ctx, expenseID, requesterID)
	}
	return []sqlc.ListExpenseSplitsRow{}, nil
}

func (m *MockExpenseService) SearchExpenses(ctx context.Context, input service.SearchExpensesInput, requesterID pgtype.UUID) ([]sqlc.Expense, error) {
	if m.SearchExpensesFunc != nil {
		return m.SearchExpensesFunc(ctx, input, requesterID)
	}
	return []sqlc.Expense{}, nil
}

var _ service.ExpenseService = (*MockExpenseService)(nil)

func createExpenseRequest(method, path string, body interface{}, userID pgtype.UUID) *http.Request {
	var req *http.Request
	if body != nil {
		bodyBytes, _ := json.Marshal(body)
		req = httptest.NewRequest(method, path, bytes.NewReader(bodyBytes))
	} else {
		req = httptest.NewRequest(method, path, nil)
	}
	req.Header.Set("Content-Type", "application/json")
	ctx := middleware.SetUserID(req.Context(), userID)
	return req.WithContext(ctx)
}

func TestCreateExpenseHandler(t *testing.T) {
	userID := testutil.CreateTestUUID(1)
	groupID := testutil.CreateTestUUID(10)

	tests := []struct {
		name           string
		requestBody    CreateExpenseRequest
		userID         pgtype.UUID
		groupID        pgtype.UUID
		mockSetup      func(*MockExpenseService)
		expectedStatus int
	}{
		{
			name: "invalid date format",
			requestBody: CreateExpenseRequest{
				Title:    "Test Expense",
				Amount:   "100.00",
				Date:     "invalid-date",
				Payments: []PaymentRequest{{UserID: userID.String(), Amount: "100.00"}},
				Splits:   []SplitRequest{{UserID: userID.String(), AmountOwned: "100.00"}},
			},
			userID:         userID,
			groupID:        groupID,
			mockSetup:      func(mock *MockExpenseService) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "invalid payment user_id",
			requestBody: CreateExpenseRequest{
				Title:    "Test Expense",
				Amount:   "100.00",
				Date:     time.Now().Format("2006-01-02"),
				Payments: []PaymentRequest{{UserID: "invalid-uuid", Amount: "100.00"}},
				Splits:   []SplitRequest{{UserID: userID.String(), AmountOwned: "100.00"}},
			},
			userID:    userID,
			groupID:   groupID,
			mockSetup: func(mock *MockExpenseService) {},
			// Fails in validation middleware due to invalid UUID format
			expectedStatus: http.StatusUnprocessableEntity,
		},
		{
			name: "service error - not group member",
			requestBody: CreateExpenseRequest{
				Title:    "Test Expense",
				Amount:   "100.00",
				Date:     time.Now().Format("2006-01-02"),
				Payments: []PaymentRequest{{UserID: userID.String(), Amount: "100.00"}},
				Splits:   []SplitRequest{{UserID: userID.String(), AmountOwned: "100.00"}},
			},
			userID:  userID,
			groupID: groupID,
			mockSetup: func(mock *MockExpenseService) {
				mock.CreateExpenseFunc = func(ctx context.Context, input service.CreateExpenseInput) (service.CreateExpenseResult, error) {
					return service.CreateExpenseResult{}, service.ErrNotGroupMember
				}
			},
			expectedStatus: http.StatusForbidden,
		},
		{
			name: "successful creation with user info",
			requestBody: CreateExpenseRequest{
				Title:    "Test Expense",
				Amount:   "100.00",
				Date:     time.Now().Format("2006-01-02"),
				Payments: []PaymentRequest{{UserID: userID.String(), Amount: "100.00"}},
				Splits:   []SplitRequest{{UserID: userID.String(), AmountOwned: "100.00"}},
			},
			userID:  userID,
			groupID: groupID,
			mockSetup: func(mock *MockExpenseService) {
				expenseID := testutil.CreateTestUUID(100)
				mock.CreateExpenseFunc = func(ctx context.Context, input service.CreateExpenseInput) (service.CreateExpenseResult, error) {
					return service.CreateExpenseResult{
						Expense: sqlc.Expense{
							ID:           expenseID,
							GroupID:      input.GroupID,
							Title:        input.Title,
							Amount:       pgtype.Numeric{},
							CurrencyCode: "USD",
						},
						Payments: []sqlc.ExpensePayment{},
						Splits:   []sqlc.ExpenseSplit{},
					}, nil
				}
				mock.GetExpensePaymentsFunc = func(ctx context.Context, expenseID, requesterID pgtype.UUID) ([]sqlc.ListExpensePaymentsRow, error) {
					return []sqlc.ListExpensePaymentsRow{
						{
							ID:        testutil.CreateTestUUID(200),
							ExpenseID: expenseID,
							UserID:    userID,
							Amount:    pgtype.Numeric{},
							UserEmail: pgtype.Text{String: "test@example.com", Valid: true},
							UserName:  pgtype.Text{String: "Test User", Valid: true},
						},
					}, nil
				}
				mock.GetExpenseSplitsFunc = func(ctx context.Context, expenseID, requesterID pgtype.UUID) ([]sqlc.ListExpenseSplitsRow, error) {
					return []sqlc.ListExpenseSplitsRow{
						{
							ID:          testutil.CreateTestUUID(300),
							ExpenseID:   expenseID,
							UserID:      userID,
							AmountOwned: pgtype.Numeric{},
							SplitType:   "equal",
							UserEmail:   pgtype.Text{String: "test@example.com", Valid: true},
							UserName:    pgtype.Text{String: "Test User", Valid: true},
						},
					}, nil
				}
			},
			expectedStatus: http.StatusCreated,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &MockExpenseService{}
			tt.mockSetup(mock)

			// In production, the handler is wrapped with ValidateBody middleware.
			// Mirror that setup here so middleware.GetBody works as expected.
			v := validator.New()
			baseHandler := CreateExpenseHandler(mock)
			handler := middleware.ValidateBody[CreateExpenseRequest](v)(baseHandler)

			req := createExpenseRequest("POST", "/groups/"+tt.groupID.String()+"/expenses", tt.requestBody, tt.userID)

			// Set up chi context with group_id
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("group_id", tt.groupID.String())
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, rr.Code)
			}
		})
	}
}
