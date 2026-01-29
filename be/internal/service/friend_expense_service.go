package service

import (
	"context"
	"errors"
	"strings"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/shopspring/decimal"

	"github.com/dhruvsaxena1998/splitplus/internal/db/sqlc"
	"github.com/dhruvsaxena1998/splitplus/internal/repository"
)

type FriendExpenseService interface {
	CreateFriendExpense(ctx context.Context, creatorID, friendID pgtype.UUID, input CreateExpenseInput) (CreateExpenseResult, error)
	ListFriendExpenses(ctx context.Context, userID, friendID pgtype.UUID) ([]sqlc.Expense, error)
	GetFriendExpenseByID(ctx context.Context, expenseID, requesterID, friendID pgtype.UUID) (sqlc.Expense, error)
	GetFriendExpensePayments(ctx context.Context, expenseID, requesterID, friendID pgtype.UUID) ([]sqlc.ListExpensePaymentsRow, error)
	GetFriendExpenseSplits(ctx context.Context, expenseID, requesterID, friendID pgtype.UUID) ([]sqlc.ListExpenseSplitsRow, error)
}

type friendExpenseService struct {
	expenseRepo repository.ExpenseRepository
	friendRepo  repository.FriendRepository
}

func NewFriendExpenseService(expenseRepo repository.ExpenseRepository, friendRepo repository.FriendRepository) FriendExpenseService {
	return &friendExpenseService{
		expenseRepo: expenseRepo,
		friendRepo:  friendRepo,
	}
}

func canonicalPairForFriends(a, b pgtype.UUID) (pgtype.UUID, pgtype.UUID) {
	if lessOrEqualUUIDForFriends(a, b) {
		return a, b
	}
	return b, a
}

func lessOrEqualUUIDForFriends(a, b pgtype.UUID) bool {
	for i := 0; i < len(a.Bytes); i++ {
		if a.Bytes[i] < b.Bytes[i] {
			return true
		}
		if a.Bytes[i] > b.Bytes[i] {
			return false
		}
	}
	return true
}

func (s *friendExpenseService) validateFriendship(ctx context.Context, userID, friendID pgtype.UUID) error {
	a, b := canonicalPairForFriends(userID, friendID)
	// Check if friendship is accepted
	fr, err := s.friendRepo.GetFriendship(ctx, a, b)
	if err != nil {
		return ErrFriendNotFound
	}
	if fr.Status != "accepted" {
		return ErrInvalidFriendAction
	}
	return nil
}

func (s *friendExpenseService) validateTwoUsersOnly(payments []PaymentInput, splits []SplitInput, userID, friendID pgtype.UUID) error {
	// Check payments only involve these two users
	for _, p := range payments {
		if p.UserID != userID && p.UserID != friendID {
			return errors.New("payments must only involve the two friends")
		}
	}
	// Check splits only involve these two users
	for _, sp := range splits {
		if sp.UserID != userID && sp.UserID != friendID {
			return errors.New("splits must only involve the two friends")
		}
	}
	return nil
}

func (s *friendExpenseService) CreateFriendExpense(ctx context.Context, creatorID, friendID pgtype.UUID, input CreateExpenseInput) (CreateExpenseResult, error) {
	// Validate friendship
	if err := s.validateFriendship(ctx, creatorID, friendID); err != nil {
		return CreateExpenseResult{}, err
	}

	// Validate amount
	expenseAmount, err := decimal.NewFromString(input.Amount)
	if err != nil || expenseAmount.LessThanOrEqual(decimal.Zero) {
		return CreateExpenseResult{}, ErrInvalidAmount
	}

	// Use default currency if not provided
	currencyCode := strings.TrimSpace(input.CurrencyCode)
	if currencyCode == "" {
		currencyCode = "USD" // Default for friend expenses
	}

	// Validate payments
	if len(input.Payments) == 0 {
		return CreateExpenseResult{}, errors.New("at least one payment is required")
	}
	if err := s.validatePaymentsTotal(input.Amount, input.Payments); err != nil {
		return CreateExpenseResult{}, err
	}
	if err := s.validateTwoUsersOnly(input.Payments, input.Splits, creatorID, friendID); err != nil {
		return CreateExpenseResult{}, err
	}

	// Validate splits
	if len(input.Splits) == 0 {
		return CreateExpenseResult{}, errors.New("at least one split is required")
	}
	if err := s.validateSplitsTotal(input.Amount, input.Splits); err != nil {
		return CreateExpenseResult{}, err
	}

	// Validate title
	title := strings.TrimSpace(input.Title)
	if title == "" {
		return CreateExpenseResult{}, errors.New("title is required")
	}

	// Convert amount to numeric
	amountNumeric, err := stringToNumeric(input.Amount)
	if err != nil {
		return CreateExpenseResult{}, ErrInvalidAmount
	}

	// Convert date
	date := pgtype.Date{Time: input.Date, Valid: true}

	// Start transaction
	tx, err := s.expenseRepo.BeginTx(ctx)
	if err != nil {
		return CreateExpenseResult{}, err
	}
	defer tx.Rollback(ctx)

	txRepo := s.expenseRepo.WithTx(tx)

	// Create expense with type='friend' and group_id=NULL
	expense, err := txRepo.CreateExpense(ctx, sqlc.CreateExpenseParams{
		GroupID:      pgtype.UUID{Valid: false}, // NULL for friend expenses
		Type:         "friend",
		Title:        title,
		Notes:        pgtype.Text{String: input.Notes, Valid: input.Notes != ""},
		Amount:       amountNumeric,
		CurrencyCode: currencyCode,
		Date:         date,
		CreatedBy:    creatorID,
	})
	if err != nil {
		return CreateExpenseResult{}, err
	}

	// Create payments
	payments := make([]sqlc.ExpensePayment, 0, len(input.Payments))
	for _, paymentInput := range input.Payments {
		paymentAmount, err := stringToNumeric(paymentInput.Amount)
		if err != nil {
			return CreateExpenseResult{}, ErrInvalidAmount
		}

		payment, err := txRepo.CreateExpensePayment(ctx, sqlc.CreateExpensePaymentParams{
			ExpenseID:     expense.ID,
			UserID:        paymentInput.UserID,
			PendingUserID: pgtype.UUID{Valid: false},
			Amount:        paymentAmount,
			PaymentMethod: pgtype.Text{String: paymentInput.PaymentMethod, Valid: paymentInput.PaymentMethod != ""},
		})
		if err != nil {
			return CreateExpenseResult{}, err
		}
		payments = append(payments, payment)
	}

	// Create splits
	splits := make([]sqlc.ExpenseSplit, 0, len(input.Splits))
	for _, splitInput := range input.Splits {
		splitAmount, err := stringToNumeric(splitInput.AmountOwned)
		if err != nil {
			return CreateExpenseResult{}, ErrInvalidAmount
		}

		split, err := txRepo.CreateExpenseSplit(ctx, sqlc.CreateExpenseSplitParams{
			ExpenseID:     expense.ID,
			UserID:        splitInput.UserID,
			PendingUserID: pgtype.UUID{Valid: false},
			AmountOwned:   splitAmount,
			SplitType:     splitInput.SplitType,
		})
		if err != nil {
			return CreateExpenseResult{}, err
		}
		splits = append(splits, split)
	}

	// Commit transaction
	if err := tx.Commit(ctx); err != nil {
		return CreateExpenseResult{}, err
	}

	return CreateExpenseResult{
		Expense:  expense,
		Payments: payments,
		Splits:   splits,
	}, nil
}

func (s *friendExpenseService) ListFriendExpenses(ctx context.Context, userID, friendID pgtype.UUID) ([]sqlc.Expense, error) {
	// Validate friendship
	if err := s.validateFriendship(ctx, userID, friendID); err != nil {
		return nil, err
	}

	return s.expenseRepo.ListFriendExpenses(ctx, sqlc.ListFriendExpensesParams{
		UserID:   userID,
		UserID_2: friendID,
	})
}

func (s *friendExpenseService) GetFriendExpenseByID(ctx context.Context, expenseID, requesterID, friendID pgtype.UUID) (sqlc.Expense, error) {
	// Validate friendship
	if err := s.validateFriendship(ctx, requesterID, friendID); err != nil {
		return sqlc.Expense{}, err
	}

	expense, err := s.expenseRepo.GetExpenseByID(ctx, expenseID)
	if err != nil {
		return sqlc.Expense{}, ErrExpenseNotFound
	}

	// Verify it's a friend expense and involves these two users
	if expense.Type != "friend" {
		return sqlc.Expense{}, ErrExpenseNotFound
	}

	// Verify requester is one of the two friends
	if requesterID != friendID && requesterID != expense.CreatedBy {
		// Check if requester is in payments or splits
		payments, _ := s.expenseRepo.ListExpensePayments(ctx, expenseID)
		splits, _ := s.expenseRepo.ListExpenseSplits(ctx, expenseID)
		found := false
		for _, p := range payments {
			if p.UserID == requesterID {
				found = true
				break
			}
		}
		if !found {
			for _, sp := range splits {
				if sp.UserID == requesterID {
					found = true
					break
				}
			}
		}
		if !found {
			return sqlc.Expense{}, ErrExpenseNotFound
		}
	}

	return expense, nil
}

func (s *friendExpenseService) GetFriendExpensePayments(ctx context.Context, expenseID, requesterID, friendID pgtype.UUID) ([]sqlc.ListExpensePaymentsRow, error) {
	// Validate expense access
	if _, err := s.GetFriendExpenseByID(ctx, expenseID, requesterID, friendID); err != nil {
		return nil, err
	}

	return s.expenseRepo.ListExpensePayments(ctx, expenseID)
}

func (s *friendExpenseService) GetFriendExpenseSplits(ctx context.Context, expenseID, requesterID, friendID pgtype.UUID) ([]sqlc.ListExpenseSplitsRow, error) {
	// Validate expense access
	if _, err := s.GetFriendExpenseByID(ctx, expenseID, requesterID, friendID); err != nil {
		return nil, err
	}

	return s.expenseRepo.ListExpenseSplits(ctx, expenseID)
}

// Helper functions reused from expense_service
func (s *friendExpenseService) validatePaymentsTotal(expenseAmount string, payments []PaymentInput) error {
	expenseDecimal, err := decimal.NewFromString(expenseAmount)
	if err != nil {
		return ErrInvalidAmount
	}

	var total decimal.Decimal
	for _, payment := range payments {
		amount, err := decimal.NewFromString(payment.Amount)
		if err != nil {
			return ErrInvalidAmount
		}
		if amount.LessThanOrEqual(decimal.Zero) {
			return ErrInvalidAmount
		}
		total = total.Add(amount)
	}

	if !total.Equal(expenseDecimal) {
		return ErrPaymentTotalMismatch
	}

	return nil
}

func (s *friendExpenseService) validateSplitsTotal(expenseAmount string, splits []SplitInput) error {
	expenseDecimal, err := decimal.NewFromString(expenseAmount)
	if err != nil {
		return ErrInvalidAmount
	}

	var total decimal.Decimal
	for _, split := range splits {
		amount, err := decimal.NewFromString(split.AmountOwned)
		if err != nil {
			return ErrInvalidAmount
		}
		if amount.LessThan(decimal.Zero) {
			return ErrInvalidAmount
		}
		total = total.Add(amount)
	}

	if !total.Equal(expenseDecimal) {
		return ErrSplitTotalMismatch
	}

	return nil
}
