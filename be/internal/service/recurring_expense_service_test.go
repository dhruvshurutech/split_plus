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

// MockRecurringExpenseRepository for testing
type MockRecurringExpenseRepository struct {
	BeginTxFunc                            func(ctx context.Context) (pgx.Tx, error)
	WithTxFunc                             func(tx pgx.Tx) repository.RecurringExpenseRepository
	CreateRecurringExpenseFunc             func(ctx context.Context, params sqlc.CreateRecurringExpenseParams) (sqlc.RecurringExpense, error)
	GetRecurringExpenseByIDFunc            func(ctx context.Context, id pgtype.UUID) (sqlc.RecurringExpense, error)
	ListRecurringExpensesByGroupFunc       func(ctx context.Context, groupID pgtype.UUID) ([]sqlc.RecurringExpense, error)
	UpdateRecurringExpenseFunc             func(ctx context.Context, params sqlc.UpdateRecurringExpenseParams) (sqlc.RecurringExpense, error)
	DeleteRecurringExpenseFunc             func(ctx context.Context, id pgtype.UUID) error
	GetRecurringExpensesDueFunc            func(ctx context.Context) ([]sqlc.RecurringExpense, error)
	UpdateNextOccurrenceDateFunc           func(ctx context.Context, params sqlc.UpdateNextOccurrenceDateParams) (sqlc.RecurringExpense, error)
	UpdateRecurringExpenseActiveStatusFunc func(ctx context.Context, params sqlc.UpdateRecurringExpenseActiveStatusParams) (sqlc.RecurringExpense, error)
	CreateRecurringExpensePaymentFunc      func(ctx context.Context, params sqlc.CreateRecurringExpensePaymentParams) (sqlc.RecurringExpensePayment, error)
	ListRecurringExpensePaymentsFunc       func(ctx context.Context, recurringExpenseID pgtype.UUID) ([]sqlc.ListRecurringExpensePaymentsRow, error)
	CreateRecurringExpenseSplitFunc        func(ctx context.Context, params sqlc.CreateRecurringExpenseSplitParams) (sqlc.RecurringExpenseSplit, error)
	ListRecurringExpenseSplitsFunc         func(ctx context.Context, recurringExpenseID pgtype.UUID) ([]sqlc.ListRecurringExpenseSplitsRow, error)
	GetGroupByIDFunc                       func(ctx context.Context, id pgtype.UUID) (sqlc.Group, error)
	GetGroupMemberFunc                     func(ctx context.Context, params sqlc.GetGroupMemberParams) (sqlc.GroupMember, error)
}

func (m *MockRecurringExpenseRepository) BeginTx(ctx context.Context) (pgx.Tx, error) {
	if m.BeginTxFunc != nil {
		return m.BeginTxFunc(ctx)
	}
	return &testutil.MockTx{}, nil
}

func (m *MockRecurringExpenseRepository) WithTx(tx pgx.Tx) repository.RecurringExpenseRepository {
	if m.WithTxFunc != nil {
		return m.WithTxFunc(tx)
	}
	return m
}

func (m *MockRecurringExpenseRepository) CreateRecurringExpense(ctx context.Context, params sqlc.CreateRecurringExpenseParams) (sqlc.RecurringExpense, error) {
	if m.CreateRecurringExpenseFunc != nil {
		return m.CreateRecurringExpenseFunc(ctx, params)
	}
	return sqlc.RecurringExpense{}, nil
}

func (m *MockRecurringExpenseRepository) GetRecurringExpenseByID(ctx context.Context, id pgtype.UUID) (sqlc.RecurringExpense, error) {
	if m.GetRecurringExpenseByIDFunc != nil {
		return m.GetRecurringExpenseByIDFunc(ctx, id)
	}
	return sqlc.RecurringExpense{}, errors.New("not found")
}

func (m *MockRecurringExpenseRepository) ListRecurringExpensesByGroup(ctx context.Context, groupID pgtype.UUID) ([]sqlc.RecurringExpense, error) {
	if m.ListRecurringExpensesByGroupFunc != nil {
		return m.ListRecurringExpensesByGroupFunc(ctx, groupID)
	}
	return []sqlc.RecurringExpense{}, nil
}

func (m *MockRecurringExpenseRepository) UpdateRecurringExpense(ctx context.Context, params sqlc.UpdateRecurringExpenseParams) (sqlc.RecurringExpense, error) {
	if m.UpdateRecurringExpenseFunc != nil {
		return m.UpdateRecurringExpenseFunc(ctx, params)
	}
	return sqlc.RecurringExpense{}, nil
}

func (m *MockRecurringExpenseRepository) DeleteRecurringExpense(ctx context.Context, id pgtype.UUID) error {
	if m.DeleteRecurringExpenseFunc != nil {
		return m.DeleteRecurringExpenseFunc(ctx, id)
	}
	return nil
}

func (m *MockRecurringExpenseRepository) GetRecurringExpensesDue(ctx context.Context) ([]sqlc.RecurringExpense, error) {
	if m.GetRecurringExpensesDueFunc != nil {
		return m.GetRecurringExpensesDueFunc(ctx)
	}
	return []sqlc.RecurringExpense{}, nil
}

func (m *MockRecurringExpenseRepository) UpdateNextOccurrenceDate(ctx context.Context, params sqlc.UpdateNextOccurrenceDateParams) (sqlc.RecurringExpense, error) {
	if m.UpdateNextOccurrenceDateFunc != nil {
		return m.UpdateNextOccurrenceDateFunc(ctx, params)
	}
	return sqlc.RecurringExpense{}, nil
}

func (m *MockRecurringExpenseRepository) UpdateRecurringExpenseActiveStatus(ctx context.Context, params sqlc.UpdateRecurringExpenseActiveStatusParams) (sqlc.RecurringExpense, error) {
	if m.UpdateRecurringExpenseActiveStatusFunc != nil {
		return m.UpdateRecurringExpenseActiveStatusFunc(ctx, params)
	}
	return sqlc.RecurringExpense{}, nil
}

func (m *MockRecurringExpenseRepository) CreateRecurringExpensePayment(ctx context.Context, params sqlc.CreateRecurringExpensePaymentParams) (sqlc.RecurringExpensePayment, error) {
	if m.CreateRecurringExpensePaymentFunc != nil {
		return m.CreateRecurringExpensePaymentFunc(ctx, params)
	}
	return sqlc.RecurringExpensePayment{}, nil
}

func (m *MockRecurringExpenseRepository) ListRecurringExpensePayments(ctx context.Context, recurringExpenseID pgtype.UUID) ([]sqlc.ListRecurringExpensePaymentsRow, error) {
	if m.ListRecurringExpensePaymentsFunc != nil {
		return m.ListRecurringExpensePaymentsFunc(ctx, recurringExpenseID)
	}
	return []sqlc.ListRecurringExpensePaymentsRow{}, nil
}

func (m *MockRecurringExpenseRepository) CreateRecurringExpenseSplit(ctx context.Context, params sqlc.CreateRecurringExpenseSplitParams) (sqlc.RecurringExpenseSplit, error) {
	if m.CreateRecurringExpenseSplitFunc != nil {
		return m.CreateRecurringExpenseSplitFunc(ctx, params)
	}
	return sqlc.RecurringExpenseSplit{}, nil
}

func (m *MockRecurringExpenseRepository) ListRecurringExpenseSplits(ctx context.Context, recurringExpenseID pgtype.UUID) ([]sqlc.ListRecurringExpenseSplitsRow, error) {
	if m.ListRecurringExpenseSplitsFunc != nil {
		return m.ListRecurringExpenseSplitsFunc(ctx, recurringExpenseID)
	}
	return []sqlc.ListRecurringExpenseSplitsRow{}, nil
}

func (m *MockRecurringExpenseRepository) GetGroupByID(ctx context.Context, id pgtype.UUID) (sqlc.Group, error) {
	if m.GetGroupByIDFunc != nil {
		return m.GetGroupByIDFunc(ctx, id)
	}
	return sqlc.Group{}, errors.New("not found")
}

func (m *MockRecurringExpenseRepository) GetGroupMember(ctx context.Context, params sqlc.GetGroupMemberParams) (sqlc.GroupMember, error) {
	if m.GetGroupMemberFunc != nil {
		return m.GetGroupMemberFunc(ctx, params)
	}
	return sqlc.GroupMember{}, errors.New("not found")
}

var _ repository.RecurringExpenseRepository = (*MockRecurringExpenseRepository)(nil)

func TestRecurringExpenseService_CreateRecurringExpense(t *testing.T) {
	groupID := testutil.CreateTestUUID(1)
	userID := testutil.CreateTestUUID(2)
	recurringExpenseID := testutil.CreateTestUUID(10)
	startDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name          string
		input         CreateRecurringExpenseInput
		mockSetup     func(*MockRecurringExpenseRepository, *MockExpenseRepository)
		expectedError error
		validate      func(*testing.T, sqlc.RecurringExpense)
	}{
		{
			name: "successful daily recurring expense creation",
			input: CreateRecurringExpenseInput{
				GroupID:        groupID,
				Title:          "Daily Coffee",
				Notes:          "Morning coffee",
				Amount:         "5.00",
				CurrencyCode:   "USD",
				RepeatInterval: "daily",
				DayOfMonth:     nil,
				DayOfWeek:      nil,
				StartDate:      startDate,
				EndDate:        nil,
				CreatedBy:      userID,
				Payments: []RecurringPaymentInput{
					{UserID: userID, Amount: "5.00", PaymentMethod: "card"},
				},
				Splits: []RecurringSplitInput{
					{UserID: userID, Type: "equal"},
				},
			},
			mockSetup: func(repo *MockRecurringExpenseRepository, expenseRepo *MockExpenseRepository) {
				repo.GetGroupByIDFunc = func(ctx context.Context, id pgtype.UUID) (sqlc.Group, error) {
					return testutil.CreateTestGroup(groupID, "Test Group", userID), nil
				}
				repo.GetGroupMemberFunc = func(ctx context.Context, params sqlc.GetGroupMemberParams) (sqlc.GroupMember, error) {
					return testutil.CreateTestGroupMember(testutil.CreateTestUUID(100), groupID, userID, "member", "active"), nil
				}
				repo.BeginTxFunc = func(ctx context.Context) (pgx.Tx, error) {
					return &testutil.MockTx{}, nil
				}
				repo.CreateRecurringExpenseFunc = func(ctx context.Context, params sqlc.CreateRecurringExpenseParams) (sqlc.RecurringExpense, error) {
					return sqlc.RecurringExpense{
						ID:                 recurringExpenseID,
						GroupID:            params.GroupID,
						Title:              params.Title,
						RepeatInterval:     params.RepeatInterval,
						NextOccurrenceDate: params.NextOccurrenceDate,
						IsActive:           params.IsActive,
					}, nil
				}
				repo.CreateRecurringExpensePaymentFunc = func(ctx context.Context, params sqlc.CreateRecurringExpensePaymentParams) (sqlc.RecurringExpensePayment, error) {
					return sqlc.RecurringExpensePayment{}, nil
				}
				repo.CreateRecurringExpenseSplitFunc = func(ctx context.Context, params sqlc.CreateRecurringExpenseSplitParams) (sqlc.RecurringExpenseSplit, error) {
					return sqlc.RecurringExpenseSplit{}, nil
				}
			},
			expectedError: nil,
			validate: func(t *testing.T, re sqlc.RecurringExpense) {
				if re.Title != "Daily Coffee" {
					t.Errorf("expected title 'Daily Coffee', got '%s'", re.Title)
				}
				if re.RepeatInterval != "daily" {
					t.Errorf("expected interval 'daily', got '%s'", re.RepeatInterval)
				}
			},
		},
		{
			name: "successful monthly recurring expense with day_of_month",
			input: CreateRecurringExpenseInput{
				GroupID:        groupID,
				Title:          "Monthly Subscription",
				Amount:         "9.99",
				CurrencyCode:   "USD",
				RepeatInterval: "monthly",
				DayOfMonth:     intPtr(3),
				DayOfWeek:      nil,
				StartDate:      startDate,
				CreatedBy:      userID,
				Payments: []RecurringPaymentInput{
					{UserID: userID, Amount: "9.99"},
				},
				Splits: []RecurringSplitInput{
					{UserID: userID, Type: "equal"},
				},
			},
			mockSetup: func(repo *MockRecurringExpenseRepository, expenseRepo *MockExpenseRepository) {
				repo.GetGroupByIDFunc = func(ctx context.Context, id pgtype.UUID) (sqlc.Group, error) {
					return testutil.CreateTestGroup(groupID, "Test Group", userID), nil
				}
				repo.GetGroupMemberFunc = func(ctx context.Context, params sqlc.GetGroupMemberParams) (sqlc.GroupMember, error) {
					return testutil.CreateTestGroupMember(testutil.CreateTestUUID(100), groupID, userID, "member", "active"), nil
				}
				repo.BeginTxFunc = func(ctx context.Context) (pgx.Tx, error) {
					return &testutil.MockTx{}, nil
				}
				repo.CreateRecurringExpenseFunc = func(ctx context.Context, params sqlc.CreateRecurringExpenseParams) (sqlc.RecurringExpense, error) {
					return sqlc.RecurringExpense{ID: recurringExpenseID, Title: params.Title, RepeatInterval: params.RepeatInterval}, nil
				}
				repo.CreateRecurringExpensePaymentFunc = func(ctx context.Context, params sqlc.CreateRecurringExpensePaymentParams) (sqlc.RecurringExpensePayment, error) {
					return sqlc.RecurringExpensePayment{}, nil
				}
				repo.CreateRecurringExpenseSplitFunc = func(ctx context.Context, params sqlc.CreateRecurringExpenseSplitParams) (sqlc.RecurringExpenseSplit, error) {
					return sqlc.RecurringExpenseSplit{}, nil
				}
			},
			expectedError: nil,
		},
		{
			name: "invalid interval - daily with day_of_month",
			input: CreateRecurringExpenseInput{
				GroupID:        groupID,
				Title:          "Test",
				Amount:         "10.00",
				RepeatInterval: "daily",
				DayOfMonth:     intPtr(5), // Invalid for daily
				StartDate:      startDate,
				CreatedBy:      userID,
				Payments:       []RecurringPaymentInput{{UserID: userID, Amount: "10.00"}},
				Splits:         []RecurringSplitInput{{UserID: userID, Type: "equal"}},
			},
			mockSetup: func(repo *MockRecurringExpenseRepository, expenseRepo *MockExpenseRepository) {
				repo.GetGroupByIDFunc = func(ctx context.Context, id pgtype.UUID) (sqlc.Group, error) {
					return testutil.CreateTestGroup(groupID, "Test Group", userID), nil
				}
				repo.GetGroupMemberFunc = func(ctx context.Context, params sqlc.GetGroupMemberParams) (sqlc.GroupMember, error) {
					return testutil.CreateTestGroupMember(testutil.CreateTestUUID(100), groupID, userID, "member", "active"), nil
				}
			},
			expectedError: errors.New("day_of_month and day_of_week must be NULL for daily interval"),
		},
		{
			name: "invalid interval - weekly without day_of_week",
			input: CreateRecurringExpenseInput{
				GroupID:        groupID,
				Title:          "Test",
				Amount:         "10.00",
				RepeatInterval: "weekly",
				DayOfWeek:      nil, // Required for weekly
				StartDate:      startDate,
				CreatedBy:      userID,
				Payments:       []RecurringPaymentInput{{UserID: userID, Amount: "10.00"}},
				Splits:         []RecurringSplitInput{{UserID: userID, Type: "equal"}},
			},
			mockSetup: func(repo *MockRecurringExpenseRepository, expenseRepo *MockExpenseRepository) {
				repo.GetGroupByIDFunc = func(ctx context.Context, id pgtype.UUID) (sqlc.Group, error) {
					return testutil.CreateTestGroup(groupID, "Test Group", userID), nil
				}
				repo.GetGroupMemberFunc = func(ctx context.Context, params sqlc.GetGroupMemberParams) (sqlc.GroupMember, error) {
					return testutil.CreateTestGroupMember(testutil.CreateTestUUID(100), groupID, userID, "member", "active"), nil
				}
			},
			expectedError: errors.New("day_of_week is required for weekly interval"),
		},
		{
			name: "empty title",
			input: CreateRecurringExpenseInput{
				GroupID:        groupID,
				Title:          "",
				Amount:         "10.00",
				RepeatInterval: "daily",
				StartDate:      startDate,
				CreatedBy:      userID,
				Payments:       []RecurringPaymentInput{{UserID: userID, Amount: "10.00"}},
				Splits:         []RecurringSplitInput{{UserID: userID, Type: "equal"}},
			},
			mockSetup: func(repo *MockRecurringExpenseRepository, expenseRepo *MockExpenseRepository) {
				repo.GetGroupByIDFunc = func(ctx context.Context, id pgtype.UUID) (sqlc.Group, error) {
					return testutil.CreateTestGroup(groupID, "Test Group", userID), nil
				}
				repo.GetGroupMemberFunc = func(ctx context.Context, params sqlc.GetGroupMemberParams) (sqlc.GroupMember, error) {
					return testutil.CreateTestGroupMember(testutil.CreateTestUUID(100), groupID, userID, "member", "active"), nil
				}
			},
			expectedError: errors.New("title is required"),
		},
		{
			name: "invalid amount",
			input: CreateRecurringExpenseInput{
				GroupID:        groupID,
				Title:          "Test",
				Amount:         "invalid",
				RepeatInterval: "daily",
				StartDate:      startDate,
				CreatedBy:      userID,
				Payments:       []RecurringPaymentInput{{UserID: userID, Amount: "10.00"}},
				Splits:         []RecurringSplitInput{{UserID: userID, Type: "equal"}},
			},
			mockSetup: func(repo *MockRecurringExpenseRepository, expenseRepo *MockExpenseRepository) {
				repo.GetGroupByIDFunc = func(ctx context.Context, id pgtype.UUID) (sqlc.Group, error) {
					return testutil.CreateTestGroup(groupID, "Test Group", userID), nil
				}
				repo.GetGroupMemberFunc = func(ctx context.Context, params sqlc.GetGroupMemberParams) (sqlc.GroupMember, error) {
					return testutil.CreateTestGroupMember(testutil.CreateTestUUID(100), groupID, userID, "member", "active"), nil
				}
			},
			expectedError: ErrInvalidAmount,
		},
		{
			name: "not a group member",
			input: CreateRecurringExpenseInput{
				GroupID:        groupID,
				Title:          "Test",
				Amount:         "10.00",
				RepeatInterval: "daily",
				StartDate:      startDate,
				CreatedBy:      userID,
				Payments:       []RecurringPaymentInput{{UserID: userID, Amount: "10.00"}},
				Splits:         []RecurringSplitInput{{UserID: userID, Type: "equal"}},
			},
			mockSetup: func(repo *MockRecurringExpenseRepository, expenseRepo *MockExpenseRepository) {
				repo.GetGroupByIDFunc = func(ctx context.Context, id pgtype.UUID) (sqlc.Group, error) {
					return testutil.CreateTestGroup(groupID, "Test Group", userID), nil
				}
				repo.GetGroupMemberFunc = func(ctx context.Context, params sqlc.GetGroupMemberParams) (sqlc.GroupMember, error) {
					return sqlc.GroupMember{}, errors.New("not found")
				}
			},
			expectedError: ErrNotGroupMember,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &MockRecurringExpenseRepository{}
			mockExpenseRepo := &MockExpenseRepository{}
			if tt.mockSetup != nil {
				tt.mockSetup(mockRepo, mockExpenseRepo)
			}

			mockCategoryRepo := &MockExpenseCategoryRepository{}
			mockActivitySvc := &MockGroupActivityService{}
			mockUserRepo := &testutil.MockUserRepository{}
			mockPendingUserRepo := &testutil.MockPendingUserRepository{}
			expenseService := NewExpenseService(mockExpenseRepo, mockCategoryRepo, mockActivitySvc, mockUserRepo, mockPendingUserRepo)
			service := NewRecurringExpenseService(mockRepo, expenseService)
			ctx := context.Background()

			result, err := service.CreateRecurringExpense(ctx, tt.input)

			if tt.expectedError != nil {
				if err == nil {
					t.Errorf("expected error %v, got nil", tt.expectedError)
					return
				}
				if !errors.Is(err, tt.expectedError) && err.Error() != tt.expectedError.Error() {
					t.Errorf("expected error %v, got %v", tt.expectedError, err)
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if tt.validate != nil {
				tt.validate(t, result)
			}
		})
	}
}

func intPtr(i int) *int {
	return &i
}
