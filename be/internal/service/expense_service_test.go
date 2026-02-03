package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/dhruvsaxena1998/splitplus/internal/db/sqlc"
	"github.com/dhruvsaxena1998/splitplus/internal/repository"
	"github.com/dhruvsaxena1998/splitplus/internal/testutil"
)

// MockExpenseRepository for testing
type MockExpenseRepository struct {
	BeginTxFunc               func(ctx context.Context) (pgx.Tx, error)
	WithTxFunc                func(tx pgx.Tx) repository.ExpenseRepository
	CreateExpenseFunc         func(ctx context.Context, params sqlc.CreateExpenseParams) (sqlc.Expense, error)
	GetExpenseByIDFunc        func(ctx context.Context, id pgtype.UUID) (sqlc.Expense, error)
	ListExpensesByGroupFunc   func(ctx context.Context, groupID pgtype.UUID) ([]sqlc.Expense, error)
	ListFriendExpensesFunc    func(ctx context.Context, params sqlc.ListFriendExpensesParams) ([]sqlc.Expense, error)
	UpdateExpenseFunc         func(ctx context.Context, params sqlc.UpdateExpenseParams) (sqlc.Expense, error)
	DeleteExpenseFunc         func(ctx context.Context, id pgtype.UUID) error
	CreateExpensePaymentFunc  func(ctx context.Context, params sqlc.CreateExpensePaymentParams) (sqlc.ExpensePayment, error)
	ListExpensePaymentsFunc   func(ctx context.Context, expenseID pgtype.UUID) ([]sqlc.ListExpensePaymentsRow, error)
	DeleteExpensePaymentsFunc func(ctx context.Context, expenseID pgtype.UUID) error
	CreateExpenseSplitFunc    func(ctx context.Context, params sqlc.CreateExpenseSplitParams) (sqlc.ExpenseSplit, error)
	ListExpenseSplitsFunc     func(ctx context.Context, expenseID pgtype.UUID) ([]sqlc.ListExpenseSplitsRow, error)
	DeleteExpenseSplitsFunc   func(ctx context.Context, expenseID pgtype.UUID) error
	GetGroupByIDFunc          func(ctx context.Context, id pgtype.UUID) (sqlc.Group, error)
	GetGroupMemberFunc        func(ctx context.Context, params sqlc.GetGroupMemberParams) (sqlc.GroupMember, error)
	SearchExpensesFunc        func(ctx context.Context, params sqlc.SearchExpensesParams) ([]sqlc.Expense, error)
}

func (m *MockExpenseRepository) BeginTx(ctx context.Context) (pgx.Tx, error) {
	if m.BeginTxFunc != nil {
		return m.BeginTxFunc(ctx)
	}
	return &testutil.MockTx{}, nil
}

func (m *MockExpenseRepository) WithTx(tx pgx.Tx) repository.ExpenseRepository {
	if m.WithTxFunc != nil {
		return m.WithTxFunc(tx)
	}
	return m
}

func (m *MockExpenseRepository) CreateExpense(ctx context.Context, params sqlc.CreateExpenseParams) (sqlc.Expense, error) {
	if m.CreateExpenseFunc != nil {
		return m.CreateExpenseFunc(ctx, params)
	}
	return sqlc.Expense{}, nil
}

func (m *MockExpenseRepository) GetExpenseByID(ctx context.Context, id pgtype.UUID) (sqlc.Expense, error) {
	if m.GetExpenseByIDFunc != nil {
		return m.GetExpenseByIDFunc(ctx, id)
	}
	return sqlc.Expense{}, errors.New("not found")
}

func (m *MockExpenseRepository) ListExpensesByGroup(ctx context.Context, groupID pgtype.UUID) ([]sqlc.Expense, error) {
	if m.ListExpensesByGroupFunc != nil {
		return m.ListExpensesByGroupFunc(ctx, groupID)
	}
	return []sqlc.Expense{}, nil
}

func (m *MockExpenseRepository) UpdateExpense(ctx context.Context, params sqlc.UpdateExpenseParams) (sqlc.Expense, error) {
	if m.UpdateExpenseFunc != nil {
		return m.UpdateExpenseFunc(ctx, params)
	}
	return sqlc.Expense{}, nil
}

func (m *MockExpenseRepository) DeleteExpense(ctx context.Context, id pgtype.UUID) error {
	if m.DeleteExpenseFunc != nil {
		return m.DeleteExpenseFunc(ctx, id)
	}
	return nil
}

func (m *MockExpenseRepository) CreateExpensePayment(ctx context.Context, params sqlc.CreateExpensePaymentParams) (sqlc.ExpensePayment, error) {
	if m.CreateExpensePaymentFunc != nil {
		return m.CreateExpensePaymentFunc(ctx, params)
	}
	return sqlc.ExpensePayment{}, nil
}

func (m *MockExpenseRepository) ListExpensePayments(ctx context.Context, expenseID pgtype.UUID) ([]sqlc.ListExpensePaymentsRow, error) {
	if m.ListExpensePaymentsFunc != nil {
		return m.ListExpensePaymentsFunc(ctx, expenseID)
	}
	return []sqlc.ListExpensePaymentsRow{}, nil
}

func (m *MockExpenseRepository) DeleteExpensePayments(ctx context.Context, expenseID pgtype.UUID) error {
	if m.DeleteExpensePaymentsFunc != nil {
		return m.DeleteExpensePaymentsFunc(ctx, expenseID)
	}
	return nil
}

func (m *MockExpenseRepository) CreateExpenseSplit(ctx context.Context, params sqlc.CreateExpenseSplitParams) (sqlc.ExpenseSplit, error) {
	if m.CreateExpenseSplitFunc != nil {
		return m.CreateExpenseSplitFunc(ctx, params)
	}
	return sqlc.ExpenseSplit{}, nil
}

func (m *MockExpenseRepository) ListExpenseSplits(ctx context.Context, expenseID pgtype.UUID) ([]sqlc.ListExpenseSplitsRow, error) {
	if m.ListExpenseSplitsFunc != nil {
		return m.ListExpenseSplitsFunc(ctx, expenseID)
	}
	return []sqlc.ListExpenseSplitsRow{}, nil
}

func (m *MockExpenseRepository) ListFriendExpenses(ctx context.Context, params sqlc.ListFriendExpensesParams) ([]sqlc.Expense, error) {
	if m.ListFriendExpensesFunc != nil {
		return m.ListFriendExpensesFunc(ctx, params)
	}
	return []sqlc.Expense{}, nil
}

func (m *MockExpenseRepository) DeleteExpenseSplits(ctx context.Context, expenseID pgtype.UUID) error {
	if m.DeleteExpenseSplitsFunc != nil {
		return m.DeleteExpenseSplitsFunc(ctx, expenseID)
	}
	return nil
}

func (m *MockExpenseRepository) GetGroupByID(ctx context.Context, id pgtype.UUID) (sqlc.Group, error) {
	if m.GetGroupByIDFunc != nil {
		return m.GetGroupByIDFunc(ctx, id)
	}
	return sqlc.Group{}, errors.New("not found")
}

func (m *MockExpenseRepository) GetGroupMember(ctx context.Context, params sqlc.GetGroupMemberParams) (sqlc.GroupMember, error) {
	if m.GetGroupMemberFunc != nil {
		return m.GetGroupMemberFunc(ctx, params)
	}
	return sqlc.GroupMember{}, errors.New("not found")
}

func (m *MockExpenseRepository) SearchExpenses(ctx context.Context, params sqlc.SearchExpensesParams) ([]sqlc.Expense, error) {
	if m.SearchExpensesFunc != nil {
		return m.SearchExpensesFunc(ctx, params)
	}
	return []sqlc.Expense{}, nil
}

var _ repository.ExpenseRepository = (*MockExpenseRepository)(nil)

func strPtr(s string) *string {
	return &s
}

func TestExpenseService_CreateExpense(t *testing.T) {
	userID := testutil.CreateTestUUID(1)
	otherUserID := testutil.CreateTestUUID(2)
	groupID := testutil.CreateTestUUID(10)

	tests := []struct {
		name          string
		input         CreateExpenseInput
		mockSetup     func(*MockExpenseRepository)
		mockUserSetup func(*testutil.MockUserRepository, *testutil.MockPendingUserRepository)
		expectedError error
	}{
		{
			name: "empty title",
			input: CreateExpenseInput{
				Title:     "",
				Amount:    "100.00",
				Date:      time.Now(),
				CreatedBy: userID,
				Payments:  []PaymentInput{{UserID: userID, Amount: "100.00"}},
				Splits:    []SplitInput{{UserID: userID, Type: "equal"}},
			},
			mockSetup: func(mock *MockExpenseRepository) {
				mock.GetGroupByIDFunc = func(ctx context.Context, id pgtype.UUID) (sqlc.Group, error) {
					return testutil.CreateTestGroup(groupID, "Test Group", userID), nil
				}
				mock.GetGroupMemberFunc = func(ctx context.Context, params sqlc.GetGroupMemberParams) (sqlc.GroupMember, error) {
					return testutil.CreateTestGroupMember(testutil.CreateTestUUID(200), params.GroupID, params.UserID, "member", "active"), nil
				}
			},
			expectedError: errors.New("title is required"),
		},
		{
			name: "invalid amount - zero",
			input: CreateExpenseInput{
				Title:     "Test Expense",
				Amount:    "0",
				Date:      time.Now(),
				CreatedBy: userID,
				Payments:  []PaymentInput{{UserID: userID, Amount: "0"}},
				Splits:    []SplitInput{{UserID: userID, Type: "equal"}},
			},
			mockSetup: func(mock *MockExpenseRepository) {
				mock.GetGroupByIDFunc = func(ctx context.Context, id pgtype.UUID) (sqlc.Group, error) {
					return testutil.CreateTestGroup(groupID, "Test Group", userID), nil
				}
				mock.GetGroupMemberFunc = func(ctx context.Context, params sqlc.GetGroupMemberParams) (sqlc.GroupMember, error) {
					return testutil.CreateTestGroupMember(testutil.CreateTestUUID(200), params.GroupID, params.UserID, "member", "active"), nil
				}
			},
			expectedError: ErrInvalidAmount,
		},
		{
			name: "invalid amount - negative",
			input: CreateExpenseInput{
				Title:     "Test Expense",
				Amount:    "-10.00",
				Date:      time.Now(),
				CreatedBy: userID,
				Payments:  []PaymentInput{{UserID: userID, Amount: "-10.00"}},
				Splits:    []SplitInput{{UserID: userID, Type: "equal"}},
			},
			mockSetup: func(mock *MockExpenseRepository) {
				mock.GetGroupByIDFunc = func(ctx context.Context, id pgtype.UUID) (sqlc.Group, error) {
					return testutil.CreateTestGroup(groupID, "Test Group", userID), nil
				}
				mock.GetGroupMemberFunc = func(ctx context.Context, params sqlc.GetGroupMemberParams) (sqlc.GroupMember, error) {
					return testutil.CreateTestGroupMember(testutil.CreateTestUUID(200), params.GroupID, params.UserID, "member", "active"), nil
				}
			},
			expectedError: ErrInvalidAmount,
		},
		{
			name: "invalid amount - invalid string",
			input: CreateExpenseInput{
				Title:     "Test Expense",
				Amount:    "not-a-number",
				Date:      time.Now(),
				CreatedBy: userID,
				Payments:  []PaymentInput{{UserID: userID, Amount: "not-a-number"}},
				Splits:    []SplitInput{{UserID: userID, Type: "equal"}},
			},
			mockSetup: func(mock *MockExpenseRepository) {
				mock.GetGroupByIDFunc = func(ctx context.Context, id pgtype.UUID) (sqlc.Group, error) {
					return testutil.CreateTestGroup(groupID, "Test Group", userID), nil
				}
				mock.GetGroupMemberFunc = func(ctx context.Context, params sqlc.GetGroupMemberParams) (sqlc.GroupMember, error) {
					return testutil.CreateTestGroupMember(testutil.CreateTestUUID(200), params.GroupID, params.UserID, "member", "active"), nil
				}
			},
			expectedError: ErrInvalidAmount,
		},
		{
			name: "no payments",
			input: CreateExpenseInput{
				Title:     "Test Expense",
				Amount:    "100.00",
				Date:      time.Now(),
				CreatedBy: userID,
				Payments:  []PaymentInput{},
				Splits:    []SplitInput{{UserID: userID, Type: "equal"}},
			},
			mockSetup: func(mock *MockExpenseRepository) {
				mock.GetGroupByIDFunc = func(ctx context.Context, id pgtype.UUID) (sqlc.Group, error) {
					return testutil.CreateTestGroup(groupID, "Test Group", userID), nil
				}
				mock.GetGroupMemberFunc = func(ctx context.Context, params sqlc.GetGroupMemberParams) (sqlc.GroupMember, error) {
					return testutil.CreateTestGroupMember(testutil.CreateTestUUID(200), params.GroupID, params.UserID, "member", "active"), nil
				}
			},
			expectedError: errors.New("at least one payment is required"),
		},
		{
			name: "no splits",
			input: CreateExpenseInput{
				Title:     "Test Expense",
				Amount:    "100.00",
				Date:      time.Now(),
				CreatedBy: userID,
				Payments:  []PaymentInput{{UserID: userID, Amount: "100.00"}},
				Splits:    []SplitInput{},
			},
			mockSetup: func(mock *MockExpenseRepository) {
				mock.GetGroupByIDFunc = func(ctx context.Context, id pgtype.UUID) (sqlc.Group, error) {
					return testutil.CreateTestGroup(groupID, "Test Group", userID), nil
				}
				mock.GetGroupMemberFunc = func(ctx context.Context, params sqlc.GetGroupMemberParams) (sqlc.GroupMember, error) {
					return testutil.CreateTestGroupMember(testutil.CreateTestUUID(200), params.GroupID, params.UserID, "member", "active"), nil
				}
			},
			expectedError: errors.New("at least one split is required"),
		},
		{
			name: "payment total mismatch - less than expense",
			input: CreateExpenseInput{
				Title:     "Test Expense",
				Amount:    "100.00",
				Date:      time.Now(),
				CreatedBy: userID,
				Payments:  []PaymentInput{{UserID: userID, Amount: "50.00"}},
				Splits:    []SplitInput{{UserID: userID, Type: "equal"}},
			},
			mockSetup: func(mock *MockExpenseRepository) {
				mock.GetGroupByIDFunc = func(ctx context.Context, id pgtype.UUID) (sqlc.Group, error) {
					return testutil.CreateTestGroup(groupID, "Test Group", userID), nil
				}
				mock.GetGroupMemberFunc = func(ctx context.Context, params sqlc.GetGroupMemberParams) (sqlc.GroupMember, error) {
					return testutil.CreateTestGroupMember(testutil.CreateTestUUID(200), params.GroupID, params.UserID, "member", "active"), nil
				}
			},
			expectedError: ErrPaymentTotalMismatch,
		},
		{
			name: "payment total mismatch - exceeds expense",
			input: CreateExpenseInput{
				Title:     "Test Expense",
				Amount:    "100.00",
				Date:      time.Now(),
				CreatedBy: userID,
				Payments:  []PaymentInput{{UserID: userID, Amount: "150.00"}},
				Splits:    []SplitInput{{UserID: userID, Type: "equal"}},
			},
			mockSetup: func(mock *MockExpenseRepository) {
				mock.GetGroupByIDFunc = func(ctx context.Context, id pgtype.UUID) (sqlc.Group, error) {
					return testutil.CreateTestGroup(groupID, "Test Group", userID), nil
				}
				mock.GetGroupMemberFunc = func(ctx context.Context, params sqlc.GetGroupMemberParams) (sqlc.GroupMember, error) {
					return testutil.CreateTestGroupMember(testutil.CreateTestUUID(200), params.GroupID, params.UserID, "member", "active"), nil
				}
			},
			expectedError: ErrPaymentTotalMismatch,
		},
		{
			name: "invalid payment amount - zero",
			input: CreateExpenseInput{
				Title:     "Test Expense",
				Amount:    "100.00",
				Date:      time.Now(),
				CreatedBy: userID,
				Payments:  []PaymentInput{{UserID: userID, Amount: "0"}},
				Splits:    []SplitInput{{UserID: userID, Type: "equal"}},
			},
			mockSetup: func(mock *MockExpenseRepository) {
				mock.GetGroupByIDFunc = func(ctx context.Context, id pgtype.UUID) (sqlc.Group, error) {
					return testutil.CreateTestGroup(groupID, "Test Group", userID), nil
				}
				mock.GetGroupMemberFunc = func(ctx context.Context, params sqlc.GetGroupMemberParams) (sqlc.GroupMember, error) {
					return testutil.CreateTestGroupMember(testutil.CreateTestUUID(200), params.GroupID, params.UserID, "member", "active"), nil
				}
			},
			expectedError: ErrInvalidAmount,
		},
		{
			name: "invalid payment amount - negative",
			input: CreateExpenseInput{
				Title:     "Test Expense",
				Amount:    "100.00",
				Date:      time.Now(),
				CreatedBy: userID,
				Payments:  []PaymentInput{{UserID: userID, Amount: "-10.00"}},
				Splits:    []SplitInput{{UserID: userID, Type: "equal"}},
			},
			mockSetup: func(mock *MockExpenseRepository) {
				mock.GetGroupByIDFunc = func(ctx context.Context, id pgtype.UUID) (sqlc.Group, error) {
					return testutil.CreateTestGroup(groupID, "Test Group", userID), nil
				}
				mock.GetGroupMemberFunc = func(ctx context.Context, params sqlc.GetGroupMemberParams) (sqlc.GroupMember, error) {
					return testutil.CreateTestGroupMember(testutil.CreateTestUUID(200), params.GroupID, params.UserID, "member", "active"), nil
				}
			},
			expectedError: ErrInvalidAmount,
		},
		{
			name: "split total mismatch - less than expense",
			input: CreateExpenseInput{
				Title:     "Test Expense",
				Amount:    "100.00",
				Date:      time.Now(),
				CreatedBy: userID,
				Payments:  []PaymentInput{{UserID: userID, Amount: "100.00"}},
				Splits:    []SplitInput{{UserID: userID, Type: "fixed", Amount: strPtr("50.00")}},
			},
			mockSetup: func(mock *MockExpenseRepository) {
				mock.GetGroupByIDFunc = func(ctx context.Context, id pgtype.UUID) (sqlc.Group, error) {
					return testutil.CreateTestGroup(groupID, "Test Group", userID), nil
				}
				mock.GetGroupMemberFunc = func(ctx context.Context, params sqlc.GetGroupMemberParams) (sqlc.GroupMember, error) {
					return testutil.CreateTestGroupMember(testutil.CreateTestUUID(200), params.GroupID, params.UserID, "member", "active"), nil
				}
			},
			expectedError: ErrSplitTotalMismatch,
		},
		{
			name: "split total mismatch - exceeds expense",
			input: CreateExpenseInput{
				Title:     "Test Expense",
				Amount:    "100.00",
				Date:      time.Now(),
				CreatedBy: userID,
				Payments:  []PaymentInput{{UserID: userID, Amount: "100.00"}},
				Splits:    []SplitInput{{UserID: userID, Type: "fixed", Amount: strPtr("150.00")}},
			},
			mockSetup: func(mock *MockExpenseRepository) {
				mock.GetGroupByIDFunc = func(ctx context.Context, id pgtype.UUID) (sqlc.Group, error) {
					return testutil.CreateTestGroup(groupID, "Test Group", userID), nil
				}
				mock.GetGroupMemberFunc = func(ctx context.Context, params sqlc.GetGroupMemberParams) (sqlc.GroupMember, error) {
					return testutil.CreateTestGroupMember(testutil.CreateTestUUID(200), params.GroupID, params.UserID, "member", "active"), nil
				}
			},
			expectedError: ErrSplitTotalMismatch,
		},
		{
			name: "invalid split amount - negative",
			input: CreateExpenseInput{
				Title:     "Test Expense",
				Amount:    "100.00",
				Date:      time.Now(),
				CreatedBy: userID,
				Payments:  []PaymentInput{{UserID: userID, Amount: "100.00"}},
				Splits:    []SplitInput{{UserID: userID, Type: "fixed", Amount: strPtr("-10.00")}},
			},
			mockSetup: func(mock *MockExpenseRepository) {
				mock.GetGroupByIDFunc = func(ctx context.Context, id pgtype.UUID) (sqlc.Group, error) {
					return testutil.CreateTestGroup(groupID, "Test Group", userID), nil
				}
				mock.GetGroupMemberFunc = func(ctx context.Context, params sqlc.GetGroupMemberParams) (sqlc.GroupMember, error) {
					return testutil.CreateTestGroupMember(testutil.CreateTestUUID(200), params.GroupID, params.UserID, "member", "active"), nil
				}
			},
			expectedError: ErrInvalidAmount,
		},
		{
			name: "not group member",
			input: CreateExpenseInput{
				Title:     "Test Expense",
				Amount:    "100.00",
				Date:      time.Now(),
				CreatedBy: userID,
				Payments:  []PaymentInput{{UserID: userID, Amount: "100.00"}},
				Splits:    []SplitInput{{UserID: userID, Type: "equal"}},
			},
			mockSetup: func(mock *MockExpenseRepository) {
				mock.GetGroupByIDFunc = func(ctx context.Context, id pgtype.UUID) (sqlc.Group, error) {
					return testutil.CreateTestGroup(groupID, "Test Group", userID), nil
				}
				mock.GetGroupMemberFunc = func(ctx context.Context, params sqlc.GetGroupMemberParams) (sqlc.GroupMember, error) {
					return sqlc.GroupMember{}, errors.New("not found")
				}
			},
			expectedError: ErrNotGroupMember,
		},
		{
			name: "valid equal split single payer",
			input: CreateExpenseInput{
				Title:     "Dinner",
				Amount:    "100.00",
				Date:      time.Now(),
				CreatedBy: userID,
				Payments: []PaymentInput{
					{UserID: userID, Amount: "100.00"},
				},
				Splits: []SplitInput{
					{UserID: userID, Type: "equal"},
					{UserID: otherUserID, Type: "equal"},
				},
			},
			mockSetup: func(mock *MockExpenseRepository) {
				expenseID := testutil.CreateTestUUID(100)
				mock.GetGroupByIDFunc = func(ctx context.Context, id pgtype.UUID) (sqlc.Group, error) {
					return testutil.CreateTestGroup(groupID, "Test Group", userID), nil
				}
				mock.GetGroupMemberFunc = func(ctx context.Context, params sqlc.GetGroupMemberParams) (sqlc.GroupMember, error) {
					return testutil.CreateTestGroupMember(testutil.CreateTestUUID(200), params.GroupID, params.UserID, "member", "active"), nil
				}
				mock.CreateExpenseFunc = func(ctx context.Context, params sqlc.CreateExpenseParams) (sqlc.Expense, error) {
					return sqlc.Expense{ID: expenseID, GroupID: params.GroupID, Title: params.Title, Amount: params.Amount, CurrencyCode: params.CurrencyCode}, nil
				}
				mock.CreateExpensePaymentFunc = func(ctx context.Context, params sqlc.CreateExpensePaymentParams) (sqlc.ExpensePayment, error) {
					return sqlc.ExpensePayment{ID: testutil.CreateTestUUID(300), ExpenseID: params.ExpenseID, UserID: params.UserID, Amount: params.Amount}, nil
				}
				mock.CreateExpenseSplitFunc = func(ctx context.Context, params sqlc.CreateExpenseSplitParams) (sqlc.ExpenseSplit, error) {
					return sqlc.ExpenseSplit{ID: testutil.CreateTestUUID(400), ExpenseID: params.ExpenseID, UserID: params.UserID, AmountOwned: params.AmountOwned, SplitType: params.SplitType}, nil
				}
			},
			expectedError: nil,
		},
		{
			name: "valid split by percentage multiple payers",
			input: CreateExpenseInput{
				Title:     "Trip",
				Amount:    "100.00",
				Date:      time.Now(),
				CreatedBy: userID,
				Payments: []PaymentInput{
					// multiple payers: 60 + 40 = 100
					{UserID: userID, Amount: "60.00"},
					{UserID: otherUserID, Amount: "40.00"},
				},
				Splits: []SplitInput{
					// 60% / 40% split
					{UserID: userID, Type: "percentage", Percentage: strPtr("60.00")},
					{UserID: otherUserID, Type: "percentage", Percentage: strPtr("40.00")},
				},
			},
			mockSetup: func(mock *MockExpenseRepository) {
				expenseID := testutil.CreateTestUUID(101)
				mock.GetGroupByIDFunc = func(ctx context.Context, id pgtype.UUID) (sqlc.Group, error) {
					return testutil.CreateTestGroup(groupID, "Test Group", userID), nil
				}
				mock.GetGroupMemberFunc = func(ctx context.Context, params sqlc.GetGroupMemberParams) (sqlc.GroupMember, error) {
					return testutil.CreateTestGroupMember(testutil.CreateTestUUID(201), params.GroupID, params.UserID, "member", "active"), nil
				}
				mock.CreateExpenseFunc = func(ctx context.Context, params sqlc.CreateExpenseParams) (sqlc.Expense, error) {
					return sqlc.Expense{ID: expenseID, GroupID: params.GroupID, Title: params.Title, Amount: params.Amount, CurrencyCode: params.CurrencyCode}, nil
				}
				mock.CreateExpensePaymentFunc = func(ctx context.Context, params sqlc.CreateExpensePaymentParams) (sqlc.ExpensePayment, error) {
					return sqlc.ExpensePayment{ID: testutil.CreateTestUUID(301), ExpenseID: params.ExpenseID, UserID: params.UserID, Amount: params.Amount}, nil
				}
				mock.CreateExpenseSplitFunc = func(ctx context.Context, params sqlc.CreateExpenseSplitParams) (sqlc.ExpenseSplit, error) {
					return sqlc.ExpenseSplit{ID: testutil.CreateTestUUID(401), ExpenseID: params.ExpenseID, UserID: params.UserID, AmountOwned: params.AmountOwned, SplitType: params.SplitType}, nil
				}
			},
			expectedError: nil,
		},
		{
			name: "valid fixed unequal split single payer",
			input: CreateExpenseInput{
				Title:     "Groceries",
				Amount:    "150.00",
				Date:      time.Now(),
				CreatedBy: userID,
				Payments: []PaymentInput{
					{UserID: userID, Amount: "150.00"},
				},
				Splits: []SplitInput{
					// Fixed unequal amounts: 100 + 50 = 150
					{UserID: userID, Type: "fixed", Amount: strPtr("100.00")},
					{UserID: otherUserID, Type: "fixed", Amount: strPtr("50.00")},
				},
			},
			mockSetup: func(mock *MockExpenseRepository) {
				expenseID := testutil.CreateTestUUID(102)
				mock.GetGroupByIDFunc = func(ctx context.Context, id pgtype.UUID) (sqlc.Group, error) {
					return testutil.CreateTestGroup(groupID, "Test Group", userID), nil
				}
				mock.GetGroupMemberFunc = func(ctx context.Context, params sqlc.GetGroupMemberParams) (sqlc.GroupMember, error) {
					return testutil.CreateTestGroupMember(testutil.CreateTestUUID(202), params.GroupID, params.UserID, "member", "active"), nil
				}
				mock.CreateExpenseFunc = func(ctx context.Context, params sqlc.CreateExpenseParams) (sqlc.Expense, error) {
					return sqlc.Expense{ID: expenseID, GroupID: params.GroupID, Title: params.Title, Amount: params.Amount, CurrencyCode: params.CurrencyCode}, nil
				}
				mock.CreateExpensePaymentFunc = func(ctx context.Context, params sqlc.CreateExpensePaymentParams) (sqlc.ExpensePayment, error) {
					return sqlc.ExpensePayment{ID: testutil.CreateTestUUID(302), ExpenseID: params.ExpenseID, UserID: params.UserID, Amount: params.Amount}, nil
				}
				mock.CreateExpenseSplitFunc = func(ctx context.Context, params sqlc.CreateExpenseSplitParams) (sqlc.ExpenseSplit, error) {
					return sqlc.ExpenseSplit{ID: testutil.CreateTestUUID(402), ExpenseID: params.ExpenseID, UserID: params.UserID, AmountOwned: params.AmountOwned, SplitType: params.SplitType}, nil
				}
			},
			expectedError: nil,
		},
		{
			name: "valid multiple payers with equal split",
			input: CreateExpenseInput{
				Title:     "Restaurant",
				Amount:    "200.00",
				Date:      time.Now(),
				CreatedBy: userID,
				Payments: []PaymentInput{
					// Three payers: 100 + 50 + 50 = 200
					{UserID: userID, Amount: "100.00"},
					{UserID: otherUserID, Amount: "50.00"},
					{UserID: testutil.CreateTestUUID(3), Amount: "50.00"},
				},
				Splits: []SplitInput{
					// Equal split: 66.67 + 66.67 + 66.66 = 200 (rounded)
					{UserID: userID, Type: "equal"},
					{UserID: otherUserID, Type: "equal"},
					{UserID: testutil.CreateTestUUID(3), Type: "equal"},
				},
			},
			mockSetup: func(mock *MockExpenseRepository) {
				expenseID := testutil.CreateTestUUID(103)
				mock.GetGroupByIDFunc = func(ctx context.Context, id pgtype.UUID) (sqlc.Group, error) {
					return testutil.CreateTestGroup(groupID, "Test Group", userID), nil
				}
				mock.GetGroupMemberFunc = func(ctx context.Context, params sqlc.GetGroupMemberParams) (sqlc.GroupMember, error) {
					return testutil.CreateTestGroupMember(testutil.CreateTestUUID(203), params.GroupID, params.UserID, "member", "active"), nil
				}
				mock.CreateExpenseFunc = func(ctx context.Context, params sqlc.CreateExpenseParams) (sqlc.Expense, error) {
					return sqlc.Expense{ID: expenseID, GroupID: params.GroupID, Title: params.Title, Amount: params.Amount, CurrencyCode: params.CurrencyCode}, nil
				}
				mock.CreateExpensePaymentFunc = func(ctx context.Context, params sqlc.CreateExpensePaymentParams) (sqlc.ExpensePayment, error) {
					return sqlc.ExpensePayment{ID: testutil.CreateTestUUID(303), ExpenseID: params.ExpenseID, UserID: params.UserID, Amount: params.Amount}, nil
				}
				mock.CreateExpenseSplitFunc = func(ctx context.Context, params sqlc.CreateExpenseSplitParams) (sqlc.ExpenseSplit, error) {
					return sqlc.ExpenseSplit{ID: testutil.CreateTestUUID(403), ExpenseID: params.ExpenseID, UserID: params.UserID, AmountOwned: params.AmountOwned, SplitType: params.SplitType}, nil
				}
			},
			expectedError: nil,
		},
		{
			name: "valid multiple payers with percentage split",
			input: CreateExpenseInput{
				Title:     "Hotel",
				Amount:    "300.00",
				Date:      time.Now(),
				CreatedBy: userID,
				Payments: []PaymentInput{
					// Three payers: 150 + 100 + 50 = 300
					{UserID: userID, Amount: "150.00"},
					{UserID: otherUserID, Amount: "100.00"},
					{UserID: testutil.CreateTestUUID(4), Amount: "50.00"},
				},
				Splits: []SplitInput{
					// Percentage split: 50% + 33.33% + 16.67% = 100%
					{UserID: userID, Type: "percentage", Percentage: strPtr("50.00")},
					{UserID: otherUserID, Type: "percentage", Percentage: strPtr("33.33")},
					{UserID: testutil.CreateTestUUID(4), Type: "percentage", Percentage: strPtr("16.67")},
				},
			},
			mockSetup: func(mock *MockExpenseRepository) {
				expenseID := testutil.CreateTestUUID(104)
				mock.GetGroupByIDFunc = func(ctx context.Context, id pgtype.UUID) (sqlc.Group, error) {
					return testutil.CreateTestGroup(groupID, "Test Group", userID), nil
				}
				mock.GetGroupMemberFunc = func(ctx context.Context, params sqlc.GetGroupMemberParams) (sqlc.GroupMember, error) {
					return testutil.CreateTestGroupMember(testutil.CreateTestUUID(205), params.GroupID, params.UserID, "member", "active"), nil
				}
				mock.CreateExpenseFunc = func(ctx context.Context, params sqlc.CreateExpenseParams) (sqlc.Expense, error) {
					return sqlc.Expense{ID: expenseID, GroupID: params.GroupID, Title: params.Title, Amount: params.Amount, CurrencyCode: params.CurrencyCode}, nil
				}
				mock.CreateExpensePaymentFunc = func(ctx context.Context, params sqlc.CreateExpensePaymentParams) (sqlc.ExpensePayment, error) {
					return sqlc.ExpensePayment{ID: testutil.CreateTestUUID(304), ExpenseID: params.ExpenseID, UserID: params.UserID, Amount: params.Amount}, nil
				}
				mock.CreateExpenseSplitFunc = func(ctx context.Context, params sqlc.CreateExpenseSplitParams) (sqlc.ExpenseSplit, error) {
					return sqlc.ExpenseSplit{ID: testutil.CreateTestUUID(404), ExpenseID: params.ExpenseID, UserID: params.UserID, AmountOwned: params.AmountOwned, SplitType: params.SplitType}, nil
				}
			},
			expectedError: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &MockExpenseRepository{}
			tt.mockSetup(mock)

			mockCategoryRepo := &MockExpenseCategoryRepository{}
			mockActivitySvc := &MockGroupActivityService{}
			mockUserRepo := &testutil.MockUserRepository{}
			mockPendingUserRepo := &testutil.MockPendingUserRepository{}
			if tt.mockUserSetup != nil {
				tt.mockUserSetup(mockUserRepo, mockPendingUserRepo)
			}
			service := NewExpenseService(mock, mockCategoryRepo, mockActivitySvc, mockUserRepo, mockPendingUserRepo)
			tt.input.GroupID = groupID

			_, err := service.CreateExpense(context.Background(), tt.input)
			if tt.expectedError != nil {
				if err == nil {
					t.Errorf("expected error %v, got nil", tt.expectedError)
					return
				}
				if err.Error() != tt.expectedError.Error() && !errors.Is(err, tt.expectedError) {
					t.Errorf("expected error %v, got %v", tt.expectedError, err)
				}
			} else if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestExpenseService_GetExpenseByID(t *testing.T) {
	userID := testutil.CreateTestUUID(1)
	expenseID := testutil.CreateTestUUID(10)
	groupID := testutil.CreateTestUUID(20)

	baseExpense := sqlc.Expense{
		ID:      expenseID,
		GroupID: groupID,
		Title:   "Test Expense",
	}

	tests := []struct {
		name          string
		expenseID     pgtype.UUID
		requesterID   pgtype.UUID
		mockSetup     func(*MockExpenseRepository)
		expectedError error
	}{
		{
			name:        "success",
			expenseID:   expenseID,
			requesterID: userID,
			mockSetup: func(mock *MockExpenseRepository) {
				mock.GetExpenseByIDFunc = func(ctx context.Context, id pgtype.UUID) (sqlc.Expense, error) {
					return baseExpense, nil
				}
				mock.GetGroupMemberFunc = func(ctx context.Context, params sqlc.GetGroupMemberParams) (sqlc.GroupMember, error) {
					return testutil.CreateTestGroupMember(testutil.CreateTestUUID(100), params.GroupID, params.UserID, "member", "active"), nil
				}
			},
			expectedError: nil,
		},
		{
			name:        "expense not found",
			expenseID:   expenseID,
			requesterID: userID,
			mockSetup: func(mock *MockExpenseRepository) {
				mock.GetExpenseByIDFunc = func(ctx context.Context, id pgtype.UUID) (sqlc.Expense, error) {
					return sqlc.Expense{}, errors.New("not found")
				}
			},
			expectedError: ErrExpenseNotFound,
		},
		{
			name:        "requester not group member",
			expenseID:   expenseID,
			requesterID: userID,
			mockSetup: func(mock *MockExpenseRepository) {
				mock.GetExpenseByIDFunc = func(ctx context.Context, id pgtype.UUID) (sqlc.Expense, error) {
					return baseExpense, nil
				}
				mock.GetGroupMemberFunc = func(ctx context.Context, params sqlc.GetGroupMemberParams) (sqlc.GroupMember, error) {
					return sqlc.GroupMember{}, errors.New("not found")
				}
			},
			expectedError: ErrNotGroupMember,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &MockExpenseRepository{}
			tt.mockSetup(mock)

			mockCategoryRepo := &MockExpenseCategoryRepository{}
			mockActivitySvc := &MockGroupActivityService{}
			mockUserRepo := &testutil.MockUserRepository{}
			mockPendingUserRepo := &testutil.MockPendingUserRepository{}
			svc := NewExpenseService(mock, mockCategoryRepo, mockActivitySvc, mockUserRepo, mockPendingUserRepo)
			_, err := svc.GetExpenseByID(context.Background(), tt.expenseID, tt.requesterID)

			if tt.expectedError != nil {
				if err == nil {
					t.Fatalf("expected error %v, got nil", tt.expectedError)
				}
				if !errors.Is(err, tt.expectedError) {
					t.Fatalf("expected error %v, got %v", tt.expectedError, err)
				}
			} else if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestExpenseService_ListExpensesByGroup(t *testing.T) {
	userID := testutil.CreateTestUUID(1)
	groupID := testutil.CreateTestUUID(20)

	tests := []struct {
		name          string
		groupID       pgtype.UUID
		requesterID   pgtype.UUID
		mockSetup     func(*MockExpenseRepository)
		expectedError error
	}{
		{
			name:        "success",
			groupID:     groupID,
			requesterID: userID,
			mockSetup: func(mock *MockExpenseRepository) {
				mock.GetGroupByIDFunc = func(ctx context.Context, id pgtype.UUID) (sqlc.Group, error) {
					return testutil.CreateTestGroup(groupID, "Test Group", userID), nil
				}
				mock.GetGroupMemberFunc = func(ctx context.Context, params sqlc.GetGroupMemberParams) (sqlc.GroupMember, error) {
					return testutil.CreateTestGroupMember(testutil.CreateTestUUID(100), params.GroupID, params.UserID, "member", "active"), nil
				}
				mock.ListExpensesByGroupFunc = func(ctx context.Context, gid pgtype.UUID) ([]sqlc.Expense, error) {
					return []sqlc.Expense{{ID: testutil.CreateTestUUID(30), GroupID: gid, Title: "Expense 1"}}, nil
				}
			},
			expectedError: nil,
		},
		{
			name:        "group not found",
			groupID:     groupID,
			requesterID: userID,
			mockSetup: func(mock *MockExpenseRepository) {
				mock.GetGroupByIDFunc = func(ctx context.Context, id pgtype.UUID) (sqlc.Group, error) {
					return sqlc.Group{}, errors.New("not found")
				}
			},
			expectedError: ErrExpenseNotFound,
		},
		{
			name:        "requester not group member",
			groupID:     groupID,
			requesterID: userID,
			mockSetup: func(mock *MockExpenseRepository) {
				mock.GetGroupByIDFunc = func(ctx context.Context, id pgtype.UUID) (sqlc.Group, error) {
					return testutil.CreateTestGroup(groupID, "Test Group", userID), nil
				}
				mock.GetGroupMemberFunc = func(ctx context.Context, params sqlc.GetGroupMemberParams) (sqlc.GroupMember, error) {
					return sqlc.GroupMember{}, errors.New("not found")
				}
			},
			expectedError: ErrNotGroupMember,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &MockExpenseRepository{}
			tt.mockSetup(mock)

			mockCategoryRepo := &MockExpenseCategoryRepository{}
			mockActivitySvc := &MockGroupActivityService{}
			mockUserRepo := &testutil.MockUserRepository{}
			mockPendingUserRepo := &testutil.MockPendingUserRepository{}
			svc := NewExpenseService(mock, mockCategoryRepo, mockActivitySvc, mockUserRepo, mockPendingUserRepo)
			_, err := svc.ListExpensesByGroup(context.Background(), tt.groupID, tt.requesterID)

			if tt.expectedError != nil {
				if err == nil {
					t.Fatalf("expected error %v, got nil", tt.expectedError)
				}
				if !errors.Is(err, tt.expectedError) {
					t.Fatalf("expected error %v, got %v", tt.expectedError, err)
				}
			} else if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestExpenseService_UpdateExpense(t *testing.T) {
	userID := testutil.CreateTestUUID(1)
	otherUserID := testutil.CreateTestUUID(2)
	groupID := testutil.CreateTestUUID(10)
	expenseID := testutil.CreateTestUUID(50)

	baseExpense := sqlc.Expense{
		ID:      expenseID,
		GroupID: groupID,
		Title:   "Original Expense",
		Amount:  pgtype.Numeric{},
	}

	tests := []struct {
		name          string
		input         UpdateExpenseInput
		mockSetup     func(*MockExpenseRepository)
		expectedError error
	}{
		{
			name: "expense not found",
			input: UpdateExpenseInput{
				ExpenseID: expenseID,
				Title:     "Updated Expense",
				Amount:    "100.00",
				Date:      time.Now(),
				UpdatedBy: userID,
				Payments:  []PaymentInput{{UserID: userID, Amount: "100.00"}},
				Splits:    []SplitInput{{UserID: userID, Type: "equal"}},
			},
			mockSetup: func(mock *MockExpenseRepository) {
				mock.GetExpenseByIDFunc = func(ctx context.Context, id pgtype.UUID) (sqlc.Expense, error) {
					return sqlc.Expense{}, errors.New("not found")
				}
			},
			expectedError: ErrExpenseNotFound,
		},
		{
			name: "not group member",
			input: UpdateExpenseInput{
				ExpenseID: expenseID,
				Title:     "Updated Expense",
				Amount:    "100.00",
				Date:      time.Now(),
				UpdatedBy: userID,
				Payments:  []PaymentInput{{UserID: userID, Amount: "100.00"}},
				Splits:    []SplitInput{{UserID: userID, Type: "equal"}},
			},
			mockSetup: func(mock *MockExpenseRepository) {
				mock.GetExpenseByIDFunc = func(ctx context.Context, id pgtype.UUID) (sqlc.Expense, error) {
					return baseExpense, nil
				}
				mock.GetGroupMemberFunc = func(ctx context.Context, params sqlc.GetGroupMemberParams) (sqlc.GroupMember, error) {
					return sqlc.GroupMember{}, errors.New("not found")
				}
			},
			expectedError: ErrNotGroupMember,
		},
		{
			name: "empty title",
			input: UpdateExpenseInput{
				ExpenseID: expenseID,
				Title:     "",
				Amount:    "100.00",
				Date:      time.Now(),
				UpdatedBy: userID,
				Payments:  []PaymentInput{{UserID: userID, Amount: "100.00"}},
				Splits:    []SplitInput{{UserID: userID, Type: "equal"}},
			},
			mockSetup: func(mock *MockExpenseRepository) {
				mock.GetExpenseByIDFunc = func(ctx context.Context, id pgtype.UUID) (sqlc.Expense, error) {
					return baseExpense, nil
				}
				mock.GetGroupMemberFunc = func(ctx context.Context, params sqlc.GetGroupMemberParams) (sqlc.GroupMember, error) {
					return testutil.CreateTestGroupMember(testutil.CreateTestUUID(200), params.GroupID, params.UserID, "member", "active"), nil
				}
			},
			expectedError: errors.New("title is required"),
		},
		{
			name: "invalid amount - zero",
			input: UpdateExpenseInput{
				ExpenseID: expenseID,
				Title:     "Updated Expense",
				Amount:    "0",
				Date:      time.Now(),
				UpdatedBy: userID,
				Payments:  []PaymentInput{{UserID: userID, Amount: "0"}},
				Splits:    []SplitInput{{UserID: userID, Type: "equal"}},
			},
			mockSetup: func(mock *MockExpenseRepository) {
				mock.GetExpenseByIDFunc = func(ctx context.Context, id pgtype.UUID) (sqlc.Expense, error) {
					return baseExpense, nil
				}
				mock.GetGroupMemberFunc = func(ctx context.Context, params sqlc.GetGroupMemberParams) (sqlc.GroupMember, error) {
					return testutil.CreateTestGroupMember(testutil.CreateTestUUID(200), params.GroupID, params.UserID, "member", "active"), nil
				}
			},
			expectedError: ErrInvalidAmount,
		},
		{
			name: "invalid amount - negative",
			input: UpdateExpenseInput{
				ExpenseID: expenseID,
				Title:     "Updated Expense",
				Amount:    "-10.00",
				Date:      time.Now(),
				UpdatedBy: userID,
				Payments:  []PaymentInput{{UserID: userID, Amount: "-10.00"}},
				Splits:    []SplitInput{{UserID: userID, Type: "equal"}},
			},
			mockSetup: func(mock *MockExpenseRepository) {
				mock.GetExpenseByIDFunc = func(ctx context.Context, id pgtype.UUID) (sqlc.Expense, error) {
					return baseExpense, nil
				}
				mock.GetGroupMemberFunc = func(ctx context.Context, params sqlc.GetGroupMemberParams) (sqlc.GroupMember, error) {
					return testutil.CreateTestGroupMember(testutil.CreateTestUUID(200), params.GroupID, params.UserID, "member", "active"), nil
				}
			},
			expectedError: ErrInvalidAmount,
		},
		{
			name: "no payments",
			input: UpdateExpenseInput{
				ExpenseID: expenseID,
				Title:     "Updated Expense",
				Amount:    "100.00",
				Date:      time.Now(),
				UpdatedBy: userID,
				Payments:  []PaymentInput{},
				Splits:    []SplitInput{{UserID: userID, Type: "equal"}},
			},
			mockSetup: func(mock *MockExpenseRepository) {
				mock.GetExpenseByIDFunc = func(ctx context.Context, id pgtype.UUID) (sqlc.Expense, error) {
					return baseExpense, nil
				}
				mock.GetGroupMemberFunc = func(ctx context.Context, params sqlc.GetGroupMemberParams) (sqlc.GroupMember, error) {
					return testutil.CreateTestGroupMember(testutil.CreateTestUUID(200), params.GroupID, params.UserID, "member", "active"), nil
				}
			},
			expectedError: errors.New("at least one payment is required"),
		},
		{
			name: "no splits",
			input: UpdateExpenseInput{
				ExpenseID: expenseID,
				Title:     "Updated Expense",
				Amount:    "100.00",
				Date:      time.Now(),
				UpdatedBy: userID,
				Payments:  []PaymentInput{{UserID: userID, Amount: "100.00"}},
				Splits:    []SplitInput{},
			},
			mockSetup: func(mock *MockExpenseRepository) {
				mock.GetExpenseByIDFunc = func(ctx context.Context, id pgtype.UUID) (sqlc.Expense, error) {
					return baseExpense, nil
				}
				mock.GetGroupMemberFunc = func(ctx context.Context, params sqlc.GetGroupMemberParams) (sqlc.GroupMember, error) {
					return testutil.CreateTestGroupMember(testutil.CreateTestUUID(200), params.GroupID, params.UserID, "member", "active"), nil
				}
			},
			expectedError: errors.New("at least one split is required"),
		},
		{
			name: "payment total mismatch - less than expense",
			input: UpdateExpenseInput{
				ExpenseID: expenseID,
				Title:     "Updated Expense",
				Amount:    "100.00",
				Date:      time.Now(),
				UpdatedBy: userID,
				Payments:  []PaymentInput{{UserID: userID, Amount: "50.00"}},
				Splits:    []SplitInput{{UserID: userID, Type: "equal"}},
			},
			mockSetup: func(mock *MockExpenseRepository) {
				mock.GetExpenseByIDFunc = func(ctx context.Context, id pgtype.UUID) (sqlc.Expense, error) {
					return baseExpense, nil
				}
				mock.GetGroupMemberFunc = func(ctx context.Context, params sqlc.GetGroupMemberParams) (sqlc.GroupMember, error) {
					return testutil.CreateTestGroupMember(testutil.CreateTestUUID(200), params.GroupID, params.UserID, "member", "active"), nil
				}
			},
			expectedError: ErrPaymentTotalMismatch,
		},
		{
			name: "payment total mismatch - exceeds expense",
			input: UpdateExpenseInput{
				ExpenseID: expenseID,
				Title:     "Updated Expense",
				Amount:    "100.00",
				Date:      time.Now(),
				UpdatedBy: userID,
				Payments:  []PaymentInput{{UserID: userID, Amount: "150.00"}},
				Splits:    []SplitInput{{UserID: userID, Type: "equal"}},
			},
			mockSetup: func(mock *MockExpenseRepository) {
				mock.GetExpenseByIDFunc = func(ctx context.Context, id pgtype.UUID) (sqlc.Expense, error) {
					return baseExpense, nil
				}
				mock.GetGroupMemberFunc = func(ctx context.Context, params sqlc.GetGroupMemberParams) (sqlc.GroupMember, error) {
					return testutil.CreateTestGroupMember(testutil.CreateTestUUID(200), params.GroupID, params.UserID, "member", "active"), nil
				}
			},
			expectedError: ErrPaymentTotalMismatch,
		},
		{
			name: "split total mismatch - less than expense",
			input: UpdateExpenseInput{
				ExpenseID: expenseID,
				Title:     "Updated Expense",
				Amount:    "100.00",
				Date:      time.Now(),
				UpdatedBy: userID,
				Payments:  []PaymentInput{{UserID: userID, Amount: "100.00"}},
				Splits:    []SplitInput{{UserID: userID, Type: "fixed", Amount: strPtr("50.00")}},
			},
			mockSetup: func(mock *MockExpenseRepository) {
				mock.GetExpenseByIDFunc = func(ctx context.Context, id pgtype.UUID) (sqlc.Expense, error) {
					return baseExpense, nil
				}
				mock.GetGroupMemberFunc = func(ctx context.Context, params sqlc.GetGroupMemberParams) (sqlc.GroupMember, error) {
					return testutil.CreateTestGroupMember(testutil.CreateTestUUID(200), params.GroupID, params.UserID, "member", "active"), nil
				}
			},
			expectedError: ErrSplitTotalMismatch,
		},
		{
			name: "split total mismatch - exceeds expense",
			input: UpdateExpenseInput{
				ExpenseID: expenseID,
				Title:     "Updated Expense",
				Amount:    "100.00",
				Date:      time.Now(),
				UpdatedBy: userID,
				Payments:  []PaymentInput{{UserID: userID, Amount: "100.00"}},
				Splits:    []SplitInput{{UserID: userID, Type: "fixed", Amount: strPtr("150.00")}},
			},
			mockSetup: func(mock *MockExpenseRepository) {
				mock.GetExpenseByIDFunc = func(ctx context.Context, id pgtype.UUID) (sqlc.Expense, error) {
					return baseExpense, nil
				}
				mock.GetGroupMemberFunc = func(ctx context.Context, params sqlc.GetGroupMemberParams) (sqlc.GroupMember, error) {
					return testutil.CreateTestGroupMember(testutil.CreateTestUUID(200), params.GroupID, params.UserID, "member", "active"), nil
				}
			},
			expectedError: ErrSplitTotalMismatch,
		},
		{
			name: "group not found for currency",
			input: UpdateExpenseInput{
				ExpenseID:    expenseID,
				Title:        "Updated Expense",
				Amount:       "100.00",
				Date:         time.Now(),
				UpdatedBy:    userID,
				CurrencyCode: "",
				Payments:     []PaymentInput{{UserID: userID, Amount: "100.00"}},
				Splits:       []SplitInput{{UserID: userID, Type: "equal"}},
			},
			mockSetup: func(mock *MockExpenseRepository) {
				mock.GetExpenseByIDFunc = func(ctx context.Context, id pgtype.UUID) (sqlc.Expense, error) {
					return baseExpense, nil
				}
				mock.GetGroupMemberFunc = func(ctx context.Context, params sqlc.GetGroupMemberParams) (sqlc.GroupMember, error) {
					return testutil.CreateTestGroupMember(testutil.CreateTestUUID(200), params.GroupID, params.UserID, "member", "active"), nil
				}
				mock.GetGroupByIDFunc = func(ctx context.Context, id pgtype.UUID) (sqlc.Group, error) {
					return sqlc.Group{}, errors.New("not found")
				}
			},
			expectedError: ErrExpenseNotFound,
		},
		{
			name: "valid update - equal split",
			input: UpdateExpenseInput{
				ExpenseID: expenseID,
				Title:     "Updated Dinner",
				Amount:    "120.00",
				Date:      time.Now(),
				UpdatedBy: userID,
				Payments:  []PaymentInput{{UserID: userID, Amount: "120.00"}},
				Splits: []SplitInput{
					{UserID: userID, Type: "equal"},
					{UserID: otherUserID, Type: "equal"},
				},
			},
			mockSetup: func(mock *MockExpenseRepository) {
				mock.GetExpenseByIDFunc = func(ctx context.Context, id pgtype.UUID) (sqlc.Expense, error) {
					return baseExpense, nil
				}
				mock.GetGroupMemberFunc = func(ctx context.Context, params sqlc.GetGroupMemberParams) (sqlc.GroupMember, error) {
					return testutil.CreateTestGroupMember(testutil.CreateTestUUID(200), params.GroupID, params.UserID, "member", "active"), nil
				}
				mock.GetGroupByIDFunc = func(ctx context.Context, id pgtype.UUID) (sqlc.Group, error) {
					return testutil.CreateTestGroup(groupID, "Test Group", userID), nil
				}
				mock.UpdateExpenseFunc = func(ctx context.Context, params sqlc.UpdateExpenseParams) (sqlc.Expense, error) {
					return sqlc.Expense{ID: expenseID, GroupID: groupID, Title: params.Title, Amount: params.Amount, CurrencyCode: params.CurrencyCode}, nil
				}
				mock.DeleteExpensePaymentsFunc = func(ctx context.Context, id pgtype.UUID) error {
					return nil
				}
				mock.DeleteExpenseSplitsFunc = func(ctx context.Context, id pgtype.UUID) error {
					return nil
				}
				mock.CreateExpensePaymentFunc = func(ctx context.Context, params sqlc.CreateExpensePaymentParams) (sqlc.ExpensePayment, error) {
					return sqlc.ExpensePayment{ID: testutil.CreateTestUUID(300), ExpenseID: params.ExpenseID, UserID: params.UserID, Amount: params.Amount}, nil
				}
				mock.CreateExpenseSplitFunc = func(ctx context.Context, params sqlc.CreateExpenseSplitParams) (sqlc.ExpenseSplit, error) {
					return sqlc.ExpenseSplit{ID: testutil.CreateTestUUID(400), ExpenseID: params.ExpenseID, UserID: params.UserID, AmountOwned: params.AmountOwned, SplitType: params.SplitType}, nil
				}
			},
			expectedError: nil,
		},
		{
			name: "valid update - percentage split multiple payers",
			input: UpdateExpenseInput{
				ExpenseID: expenseID,
				Title:     "Updated Trip",
				Amount:    "200.00",
				Date:      time.Now(),
				UpdatedBy: userID,
				Payments: []PaymentInput{
					{UserID: userID, Amount: "120.00"},
					{UserID: otherUserID, Amount: "80.00"},
				},
				Splits: []SplitInput{
					{UserID: userID, Type: "percentage", Percentage: strPtr("60.00")},
					{UserID: otherUserID, Type: "percentage", Percentage: strPtr("40.00")},
				},
			},
			mockSetup: func(mock *MockExpenseRepository) {
				mock.GetExpenseByIDFunc = func(ctx context.Context, id pgtype.UUID) (sqlc.Expense, error) {
					return baseExpense, nil
				}
				mock.GetGroupMemberFunc = func(ctx context.Context, params sqlc.GetGroupMemberParams) (sqlc.GroupMember, error) {
					return testutil.CreateTestGroupMember(testutil.CreateTestUUID(200), params.GroupID, params.UserID, "member", "active"), nil
				}
				mock.GetGroupByIDFunc = func(ctx context.Context, id pgtype.UUID) (sqlc.Group, error) {
					return testutil.CreateTestGroup(groupID, "Test Group", userID), nil
				}
				mock.UpdateExpenseFunc = func(ctx context.Context, params sqlc.UpdateExpenseParams) (sqlc.Expense, error) {
					return sqlc.Expense{ID: expenseID, GroupID: groupID, Title: params.Title, Amount: params.Amount, CurrencyCode: params.CurrencyCode}, nil
				}
				mock.DeleteExpensePaymentsFunc = func(ctx context.Context, id pgtype.UUID) error {
					return nil
				}
				mock.DeleteExpenseSplitsFunc = func(ctx context.Context, id pgtype.UUID) error {
					return nil
				}
				mock.CreateExpensePaymentFunc = func(ctx context.Context, params sqlc.CreateExpensePaymentParams) (sqlc.ExpensePayment, error) {
					return sqlc.ExpensePayment{ID: testutil.CreateTestUUID(301), ExpenseID: params.ExpenseID, UserID: params.UserID, Amount: params.Amount}, nil
				}
				mock.CreateExpenseSplitFunc = func(ctx context.Context, params sqlc.CreateExpenseSplitParams) (sqlc.ExpenseSplit, error) {
					return sqlc.ExpenseSplit{ID: testutil.CreateTestUUID(401), ExpenseID: params.ExpenseID, UserID: params.UserID, AmountOwned: params.AmountOwned, SplitType: params.SplitType}, nil
				}
			},
			expectedError: nil,
		},
		{
			name: "valid update - fixed unequal split",
			input: UpdateExpenseInput{
				ExpenseID: expenseID,
				Title:     "Updated Groceries",
				Amount:    "180.00",
				Date:      time.Now(),
				UpdatedBy: userID,
				Payments:  []PaymentInput{{UserID: userID, Amount: "180.00"}},
				Splits: []SplitInput{
					{UserID: userID, Type: "fixed", Amount: strPtr("120.00")},
					{UserID: otherUserID, Type: "fixed", Amount: strPtr("60.00")},
				},
			},
			mockSetup: func(mock *MockExpenseRepository) {
				mock.GetExpenseByIDFunc = func(ctx context.Context, id pgtype.UUID) (sqlc.Expense, error) {
					return baseExpense, nil
				}
				mock.GetGroupMemberFunc = func(ctx context.Context, params sqlc.GetGroupMemberParams) (sqlc.GroupMember, error) {
					return testutil.CreateTestGroupMember(testutil.CreateTestUUID(200), params.GroupID, params.UserID, "member", "active"), nil
				}
				mock.GetGroupByIDFunc = func(ctx context.Context, id pgtype.UUID) (sqlc.Group, error) {
					return testutil.CreateTestGroup(groupID, "Test Group", userID), nil
				}
				mock.UpdateExpenseFunc = func(ctx context.Context, params sqlc.UpdateExpenseParams) (sqlc.Expense, error) {
					return sqlc.Expense{ID: expenseID, GroupID: groupID, Title: params.Title, Amount: params.Amount, CurrencyCode: params.CurrencyCode}, nil
				}
				mock.DeleteExpensePaymentsFunc = func(ctx context.Context, id pgtype.UUID) error {
					return nil
				}
				mock.DeleteExpenseSplitsFunc = func(ctx context.Context, id pgtype.UUID) error {
					return nil
				}
				mock.CreateExpensePaymentFunc = func(ctx context.Context, params sqlc.CreateExpensePaymentParams) (sqlc.ExpensePayment, error) {
					return sqlc.ExpensePayment{ID: testutil.CreateTestUUID(302), ExpenseID: params.ExpenseID, UserID: params.UserID, Amount: params.Amount}, nil
				}
				mock.CreateExpenseSplitFunc = func(ctx context.Context, params sqlc.CreateExpenseSplitParams) (sqlc.ExpenseSplit, error) {
					return sqlc.ExpenseSplit{ID: testutil.CreateTestUUID(402), ExpenseID: params.ExpenseID, UserID: params.UserID, AmountOwned: params.AmountOwned, SplitType: params.SplitType}, nil
				}
			},
			expectedError: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &MockExpenseRepository{}
			tt.mockSetup(mock)

			mockCategoryRepo := &MockExpenseCategoryRepository{}
			mockActivitySvc := &MockGroupActivityService{}
			mockUserRepo := &testutil.MockUserRepository{}
			mockPendingUserRepo := &testutil.MockPendingUserRepository{}
			service := NewExpenseService(mock, mockCategoryRepo, mockActivitySvc, mockUserRepo, mockPendingUserRepo)
			_, err := service.UpdateExpense(context.Background(), tt.input)

			if tt.expectedError != nil {
				if err == nil {
					t.Errorf("expected error %v, got nil", tt.expectedError)
					return
				}
				if err.Error() != tt.expectedError.Error() && !errors.Is(err, tt.expectedError) {
					t.Errorf("expected error %v, got %v", tt.expectedError, err)
				}
			} else if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestExpenseService_DeleteExpense(t *testing.T) {
	userID := testutil.CreateTestUUID(1)
	groupID := testutil.CreateTestUUID(20)
	expenseID := testutil.CreateTestUUID(30)

	baseExpense := sqlc.Expense{
		ID:      expenseID,
		GroupID: groupID,
		Title:   "Test Expense",
	}

	tests := []struct {
		name          string
		expenseID     pgtype.UUID
		requesterID   pgtype.UUID
		mockSetup     func(*MockExpenseRepository)
		expectedError error
	}{
		{
			name:        "success",
			expenseID:   expenseID,
			requesterID: userID,
			mockSetup: func(mock *MockExpenseRepository) {
				mock.GetExpenseByIDFunc = func(ctx context.Context, id pgtype.UUID) (sqlc.Expense, error) {
					return baseExpense, nil
				}
				mock.GetGroupMemberFunc = func(ctx context.Context, params sqlc.GetGroupMemberParams) (sqlc.GroupMember, error) {
					return testutil.CreateTestGroupMember(testutil.CreateTestUUID(100), params.GroupID, params.UserID, "member", "active"), nil
				}
				mock.DeleteExpenseFunc = func(ctx context.Context, id pgtype.UUID) error {
					return nil
				}
			},
			expectedError: nil,
		},
		{
			name:        "expense not found",
			expenseID:   expenseID,
			requesterID: userID,
			mockSetup: func(mock *MockExpenseRepository) {
				mock.GetExpenseByIDFunc = func(ctx context.Context, id pgtype.UUID) (sqlc.Expense, error) {
					return sqlc.Expense{}, errors.New("not found")
				}
			},
			expectedError: ErrExpenseNotFound,
		},
		{
			name:        "requester not group member",
			expenseID:   expenseID,
			requesterID: userID,
			mockSetup: func(mock *MockExpenseRepository) {
				mock.GetExpenseByIDFunc = func(ctx context.Context, id pgtype.UUID) (sqlc.Expense, error) {
					return baseExpense, nil
				}
				mock.GetGroupMemberFunc = func(ctx context.Context, params sqlc.GetGroupMemberParams) (sqlc.GroupMember, error) {
					return sqlc.GroupMember{}, errors.New("not found")
				}
			},
			expectedError: ErrNotGroupMember,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &MockExpenseRepository{}
			tt.mockSetup(mock)

			mockCategoryRepo := &MockExpenseCategoryRepository{}
			mockActivitySvc := &MockGroupActivityService{}
			mockUserRepo := &testutil.MockUserRepository{}
			mockPendingUserRepo := &testutil.MockPendingUserRepository{}
			svc := NewExpenseService(mock, mockCategoryRepo, mockActivitySvc, mockUserRepo, mockPendingUserRepo)
			err := svc.DeleteExpense(context.Background(), tt.expenseID, tt.requesterID)

			if tt.expectedError != nil {
				if err == nil {
					t.Fatalf("expected error %v, got nil", tt.expectedError)
				}
				if !errors.Is(err, tt.expectedError) {
					t.Fatalf("expected error %v, got %v", tt.expectedError, err)
				}
			} else if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestExpenseService_GetExpensePaymentsAndSplits(t *testing.T) {
	userID := testutil.CreateTestUUID(1)
	groupID := testutil.CreateTestUUID(20)
	expenseID := testutil.CreateTestUUID(30)

	baseExpense := sqlc.Expense{
		ID:      expenseID,
		GroupID: groupID,
		Title:   "Test Expense",
	}

	tests := []struct {
		name             string
		expenseID        pgtype.UUID
		requesterID      pgtype.UUID
		mockSetup        func(*MockExpenseRepository)
		expectPaymentsOK bool
		expectSplitsOK   bool
		expectedError    error
	}{
		{
			name:        "payments success",
			expenseID:   expenseID,
			requesterID: userID,
			mockSetup: func(mock *MockExpenseRepository) {
				mock.GetExpenseByIDFunc = func(ctx context.Context, id pgtype.UUID) (sqlc.Expense, error) {
					return baseExpense, nil
				}
				mock.GetGroupMemberFunc = func(ctx context.Context, params sqlc.GetGroupMemberParams) (sqlc.GroupMember, error) {
					return testutil.CreateTestGroupMember(testutil.CreateTestUUID(100), params.GroupID, params.UserID, "member", "active"), nil
				}
				mock.ListExpensePaymentsFunc = func(ctx context.Context, id pgtype.UUID) ([]sqlc.ListExpensePaymentsRow, error) {
					return []sqlc.ListExpensePaymentsRow{
						{ID: testutil.CreateTestUUID(1), ExpenseID: id},
					}, nil
				}
			},
			expectPaymentsOK: true,
			expectedError:    nil,
		},
		{
			name:        "payments expense not found",
			expenseID:   expenseID,
			requesterID: userID,
			mockSetup: func(mock *MockExpenseRepository) {
				mock.GetExpenseByIDFunc = func(ctx context.Context, id pgtype.UUID) (sqlc.Expense, error) {
					return sqlc.Expense{}, errors.New("not found")
				}
			},
			expectPaymentsOK: false,
			expectedError:    ErrExpenseNotFound,
		},
		{
			name:        "payments requester not group member",
			expenseID:   expenseID,
			requesterID: userID,
			mockSetup: func(mock *MockExpenseRepository) {
				mock.GetExpenseByIDFunc = func(ctx context.Context, id pgtype.UUID) (sqlc.Expense, error) {
					return baseExpense, nil
				}
				mock.GetGroupMemberFunc = func(ctx context.Context, params sqlc.GetGroupMemberParams) (sqlc.GroupMember, error) {
					return sqlc.GroupMember{}, errors.New("not found")
				}
			},
			expectPaymentsOK: false,
			expectedError:    ErrNotGroupMember,
		},
		{
			name:        "splits success",
			expenseID:   expenseID,
			requesterID: userID,
			mockSetup: func(mock *MockExpenseRepository) {
				mock.GetExpenseByIDFunc = func(ctx context.Context, id pgtype.UUID) (sqlc.Expense, error) {
					return baseExpense, nil
				}
				mock.GetGroupMemberFunc = func(ctx context.Context, params sqlc.GetGroupMemberParams) (sqlc.GroupMember, error) {
					return testutil.CreateTestGroupMember(testutil.CreateTestUUID(100), params.GroupID, params.UserID, "member", "active"), nil
				}
				mock.ListExpenseSplitsFunc = func(ctx context.Context, id pgtype.UUID) ([]sqlc.ListExpenseSplitsRow, error) {
					return []sqlc.ListExpenseSplitsRow{
						{ID: testutil.CreateTestUUID(1), ExpenseID: id},
					}, nil
				}
			},
			expectSplitsOK: true,
			expectedError:  nil,
		},
		{
			name:        "splits expense not found",
			expenseID:   expenseID,
			requesterID: userID,
			mockSetup: func(mock *MockExpenseRepository) {
				mock.GetExpenseByIDFunc = func(ctx context.Context, id pgtype.UUID) (sqlc.Expense, error) {
					return sqlc.Expense{}, errors.New("not found")
				}
			},
			expectSplitsOK: false,
			expectedError:  ErrExpenseNotFound,
		},
		{
			name:        "splits requester not group member",
			expenseID:   expenseID,
			requesterID: userID,
			mockSetup: func(mock *MockExpenseRepository) {
				mock.GetExpenseByIDFunc = func(ctx context.Context, id pgtype.UUID) (sqlc.Expense, error) {
					return baseExpense, nil
				}
				mock.GetGroupMemberFunc = func(ctx context.Context, params sqlc.GetGroupMemberParams) (sqlc.GroupMember, error) {
					return sqlc.GroupMember{}, errors.New("not found")
				}
			},
			expectSplitsOK: false,
			expectedError:  ErrNotGroupMember,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &MockExpenseRepository{}
			tt.mockSetup(mock)

			mockCategoryRepo := &MockExpenseCategoryRepository{}
			mockActivitySvc := &MockGroupActivityService{}
			mockUserRepo := &testutil.MockUserRepository{}
			mockPendingUserRepo := &testutil.MockPendingUserRepository{}
			svc := NewExpenseService(mock, mockCategoryRepo, mockActivitySvc, mockUserRepo, mockPendingUserRepo)

			if tt.expectPaymentsOK || tt.expectedError != nil {
				_, err := svc.GetExpensePayments(context.Background(), tt.expenseID, tt.requesterID)
				if tt.expectedError != nil {
					if err == nil {
						t.Fatalf("[payments] expected error %v, got nil", tt.expectedError)
					}
					if !errors.Is(err, tt.expectedError) {
						t.Fatalf("[payments] expected error %v, got %v", tt.expectedError, err)
					}
				} else if err != nil {
					t.Fatalf("[payments] unexpected error: %v", err)
				}
			}

			if tt.expectSplitsOK || tt.expectedError != nil {
				_, err := svc.GetExpenseSplits(context.Background(), tt.expenseID, tt.requesterID)
				if tt.expectedError != nil {
					if err == nil {
						t.Fatalf("[splits] expected error %v, got nil", tt.expectedError)
					}
					if !errors.Is(err, tt.expectedError) {
						t.Fatalf("[splits] expected error %v, got %v", tt.expectedError, err)
					}
				} else if err != nil {
					t.Fatalf("[splits] unexpected error: %v", err)
				}
			}
		})
	}
}
