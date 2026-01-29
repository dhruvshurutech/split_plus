package service

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/shopspring/decimal"

	"github.com/dhruvsaxena1998/splitplus/internal/db/sqlc"
	"github.com/dhruvsaxena1998/splitplus/internal/repository"
)

var (
	ErrExpenseNotFound      = errors.New("expense not found")
	ErrInvalidAmount        = errors.New("invalid amount")
	ErrPaymentTotalMismatch = errors.New("payment total does not match expense amount")
	ErrSplitTotalMismatch   = errors.New("split total does not match expense amount")
	ErrCategoryNotInGroup   = errors.New("category does not belong to this group")
)

type PaymentInput struct {
	UserID        pgtype.UUID
	PendingUserID *pgtype.UUID
	Amount        string
	PaymentMethod string
}

type SplitInput struct {
	UserID        pgtype.UUID
	PendingUserID *pgtype.UUID
	AmountOwned   string
	SplitType     string
}

type CreateExpenseInput struct {
	GroupID      pgtype.UUID
	Title        string
	Notes        string
	Amount       string
	CurrencyCode string
	Date         time.Time
	CategoryID   *pgtype.UUID
	Tags         []string
	CreatedBy    pgtype.UUID
	Payments     []PaymentInput
	Splits       []SplitInput
}

type UpdateExpenseInput struct {
	ExpenseID    pgtype.UUID
	Title        string
	Notes        string
	Amount       string
	CurrencyCode string
	Date         time.Time
	CategoryID   *pgtype.UUID
	Tags         []string
	UpdatedBy    pgtype.UUID
	Payments     []PaymentInput
	Splits       []SplitInput
}

type SearchExpensesInput struct {
	GroupID     pgtype.UUID
	Query       *string
	StartDate   *time.Time
	EndDate     *time.Time
	CategoryID  *pgtype.UUID
	CreatedBy   *pgtype.UUID
	MinAmount   *string
	MaxAmount   *string
	PayerID     *pgtype.UUID
	OwerID      *pgtype.UUID
	Limit       int32
	Offset      int32
}

type CreateExpenseResult struct {
	Expense  sqlc.Expense
	Payments []sqlc.ExpensePayment
	Splits   []sqlc.ExpenseSplit
}

type ExpenseService interface {
	CreateExpense(ctx context.Context, input CreateExpenseInput) (CreateExpenseResult, error)
	GetExpenseByID(ctx context.Context, expenseID, requesterID pgtype.UUID) (sqlc.Expense, error)
	ListExpensesByGroup(ctx context.Context, groupID, requesterID pgtype.UUID) ([]sqlc.Expense, error)
	UpdateExpense(ctx context.Context, input UpdateExpenseInput) (CreateExpenseResult, error)
	DeleteExpense(ctx context.Context, expenseID, requesterID pgtype.UUID) error
	GetExpensePayments(ctx context.Context, expenseID, requesterID pgtype.UUID) ([]sqlc.ListExpensePaymentsRow, error)
	GetExpenseSplits(ctx context.Context, expenseID, requesterID pgtype.UUID) ([]sqlc.ListExpenseSplitsRow, error)
	SearchExpenses(ctx context.Context, input SearchExpensesInput, requesterID pgtype.UUID) ([]sqlc.Expense, error)
}

type expenseService struct {
	repo            repository.ExpenseRepository
	categoryRepo    repository.ExpenseCategoryRepository
	activityService GroupActivityService
}

func NewExpenseService(
	repo repository.ExpenseRepository,
	categoryRepo repository.ExpenseCategoryRepository,
	activityService GroupActivityService,
) ExpenseService {
	return &expenseService{
		repo:            repo,
		categoryRepo:    categoryRepo,
		activityService: activityService,
	}
}

// Helper functions for numeric conversion
func stringToNumeric(s string) (pgtype.Numeric, error) {
	d, err := decimal.NewFromString(s)
	if err != nil {
		return pgtype.Numeric{}, err
	}
	var n pgtype.Numeric
	err = n.Scan(d.String())
	return n, err
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
	return "0", errors.New("invalid numeric value")
}

func numericToDecimal(n pgtype.Numeric) (decimal.Decimal, error) {
	if !n.Valid {
		return decimal.Zero, nil
	}
	val, err := n.Value()
	if err != nil {
		return decimal.Zero, err
	}
	if str, ok := val.(string); ok {
		return decimal.NewFromString(str)
	}
	return decimal.Zero, errors.New("invalid numeric value")
}

func (s *expenseService) validateGroupMembership(ctx context.Context, groupID, userID pgtype.UUID) error {
	_, err := s.repo.GetGroupMember(ctx, sqlc.GetGroupMemberParams{
		GroupID: groupID,
		UserID:  userID,
	})
	if err != nil {
		return ErrNotGroupMember
	}
	return nil
}

func (s *expenseService) validatePaymentsTotal(expenseAmount string, payments []PaymentInput) error {
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

func (s *expenseService) validateSplitsTotal(expenseAmount string, splits []SplitInput) error {
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

func (s *expenseService) validateCategory(ctx context.Context, categoryID *pgtype.UUID, groupID pgtype.UUID) error {
	if categoryID == nil || !categoryID.Valid {
		return nil // Category is optional
	}

	category, err := s.categoryRepo.GetCategoryByID(ctx, *categoryID)
	if err != nil {
		return ErrCategoryNotFound
	}

	// Validate category belongs to the group
	if category.GroupID.Bytes != groupID.Bytes {
		return ErrCategoryNotInGroup
	}

	return nil
}

func (s *expenseService) CreateExpense(ctx context.Context, input CreateExpenseInput) (CreateExpenseResult, error) {
	// Validate group exists
	group, err := s.repo.GetGroupByID(ctx, input.GroupID)
	if err != nil {
		return CreateExpenseResult{}, ErrExpenseNotFound
	}

	// Validate user is group member
	if err := s.validateGroupMembership(ctx, input.GroupID, input.CreatedBy); err != nil {
		return CreateExpenseResult{}, err
	}

	// Validate amount
	expenseAmount, err := decimal.NewFromString(input.Amount)
	if err != nil || expenseAmount.LessThanOrEqual(decimal.Zero) {
		return CreateExpenseResult{}, ErrInvalidAmount
	}

	// Use group currency if not provided
	currencyCode := strings.TrimSpace(input.CurrencyCode)
	if currencyCode == "" {
		currencyCode = group.CurrencyCode
	}

	// Validate payments
	if len(input.Payments) == 0 {
		return CreateExpenseResult{}, errors.New("at least one payment is required")
	}
	if err := s.validatePaymentsTotal(input.Amount, input.Payments); err != nil {
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

	// Validate category if provided
	if err := s.validateCategory(ctx, input.CategoryID, input.GroupID); err != nil {
		return CreateExpenseResult{}, err
	}

	// Convert amount to numeric
	amountNumeric, err := stringToNumeric(input.Amount)
	if err != nil {
		return CreateExpenseResult{}, ErrInvalidAmount
	}

	// Convert date
	date := pgtype.Date{Time: input.Date, Valid: true}

	// Convert category_id
	var categoryID pgtype.UUID
	if input.CategoryID != nil && input.CategoryID.Valid {
		categoryID = *input.CategoryID
	}

	// Convert tags to pgtype array
	tags := []string{}
	if input.Tags != nil {
		tags = input.Tags
	}

	// Start transaction
	tx, err := s.repo.BeginTx(ctx)
	if err != nil {
		return CreateExpenseResult{}, err
	}
	defer tx.Rollback(ctx)

	txRepo := s.repo.WithTx(tx)

	// Create expense
	expense, err := txRepo.CreateExpense(ctx, sqlc.CreateExpenseParams{
		GroupID:      input.GroupID,
		Type:         "group",
		Title:        title,
		Notes:        pgtype.Text{String: input.Notes, Valid: input.Notes != ""},
		Amount:       amountNumeric,
		CurrencyCode: currencyCode,
		Date:         date,
		CategoryID:   categoryID,
		Tags:         tags,
		CreatedBy:    input.CreatedBy,
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

		var pendingUserID pgtype.UUID
		if paymentInput.PendingUserID != nil {
			pendingUserID = *paymentInput.PendingUserID
		}

		if !paymentInput.UserID.Valid && !pendingUserID.Valid {
			return CreateExpenseResult{}, errors.New("payment must have user_id or pending_user_id")
		}

		payment, err := txRepo.CreateExpensePayment(ctx, sqlc.CreateExpensePaymentParams{
			ExpenseID:     expense.ID,
			UserID:        paymentInput.UserID,
			PendingUserID: pendingUserID,
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

		splitType := strings.TrimSpace(splitInput.SplitType)
		if splitType == "" {
			splitType = "equal"
		}

		var pendingUserID pgtype.UUID
		if splitInput.PendingUserID != nil {
			pendingUserID = *splitInput.PendingUserID
		}

		if !splitInput.UserID.Valid && !pendingUserID.Valid {
			return CreateExpenseResult{}, errors.New("split must have user_id or pending_user_id")
		}

		split, err := txRepo.CreateExpenseSplit(ctx, sqlc.CreateExpenseSplitParams{
			ExpenseID:     expense.ID,
			UserID:        splitInput.UserID,
			PendingUserID: pendingUserID,
			AmountOwned:   splitAmount,
			SplitType:     splitType,
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

	// Log activity
	_ = s.activityService.LogActivity(ctx, LogActivityInput{
		GroupID:    input.GroupID,
		UserID:     input.CreatedBy,
		Action:     "expense_created",
		EntityType: "expense",
		EntityID:   expense.ID,
		Metadata: map[string]interface{}{
			"title":  expense.Title,
			"amount": input.Amount,
		},
	})

	return CreateExpenseResult{
		Expense:  expense,
		Payments: payments,
		Splits:   splits,
	}, nil
}

func (s *expenseService) GetExpenseByID(ctx context.Context, expenseID, requesterID pgtype.UUID) (sqlc.Expense, error) {
	expense, err := s.repo.GetExpenseByID(ctx, expenseID)
	if err != nil {
		return sqlc.Expense{}, ErrExpenseNotFound
	}

	// Validate user is group member
	if err := s.validateGroupMembership(ctx, expense.GroupID, requesterID); err != nil {
		return sqlc.Expense{}, err
	}

	return expense, nil
}

func (s *expenseService) ListExpensesByGroup(ctx context.Context, groupID, requesterID pgtype.UUID) ([]sqlc.Expense, error) {
	// Validate group exists
	_, err := s.repo.GetGroupByID(ctx, groupID)
	if err != nil {
		return nil, ErrExpenseNotFound
	}

	// Validate user is group member
	if err := s.validateGroupMembership(ctx, groupID, requesterID); err != nil {
		return nil, err
	}

	return s.repo.ListExpensesByGroup(ctx, groupID)
}

func (s *expenseService) UpdateExpense(ctx context.Context, input UpdateExpenseInput) (CreateExpenseResult, error) {
	// Get existing expense
	expense, err := s.repo.GetExpenseByID(ctx, input.ExpenseID)
	if err != nil {
		return CreateExpenseResult{}, ErrExpenseNotFound
	}

	// Validate user is group member
	if err := s.validateGroupMembership(ctx, expense.GroupID, input.UpdatedBy); err != nil {
		return CreateExpenseResult{}, err
	}

	// Validate amount
	expenseAmount, err := decimal.NewFromString(input.Amount)
	if err != nil || expenseAmount.LessThanOrEqual(decimal.Zero) {
		return CreateExpenseResult{}, ErrInvalidAmount
	}

	// Validate payments
	if len(input.Payments) == 0 {
		return CreateExpenseResult{}, errors.New("at least one payment is required")
	}
	if err := s.validatePaymentsTotal(input.Amount, input.Payments); err != nil {
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

	// Validate category if provided
	if err := s.validateCategory(ctx, input.CategoryID, expense.GroupID); err != nil {
		return CreateExpenseResult{}, err
	}

	// Use group currency if not provided
	currencyCode := strings.TrimSpace(input.CurrencyCode)
	if currencyCode == "" {
		group, err := s.repo.GetGroupByID(ctx, expense.GroupID)
		if err != nil {
			return CreateExpenseResult{}, ErrExpenseNotFound
		}
		currencyCode = group.CurrencyCode
	}

	// Convert amount to numeric
	amountNumeric, err := stringToNumeric(input.Amount)
	if err != nil {
		return CreateExpenseResult{}, ErrInvalidAmount
	}

	// Convert date
	date := pgtype.Date{Time: input.Date, Valid: true}

	// Convert category_id
	var categoryID pgtype.UUID
	if input.CategoryID != nil && input.CategoryID.Valid {
		categoryID = *input.CategoryID
	}

	// Convert tags to pgtype array
	tags := []string{}
	if input.Tags != nil {
		tags = input.Tags
	}

	// Start transaction
	tx, err := s.repo.BeginTx(ctx)
	if err != nil {
		return CreateExpenseResult{}, err
	}
	defer tx.Rollback(ctx)

	txRepo := s.repo.WithTx(tx)

	// Update expense
	updatedExpense, err := txRepo.UpdateExpense(ctx, sqlc.UpdateExpenseParams{
		ID:           input.ExpenseID,
		Title:        title,
		Notes:        pgtype.Text{String: input.Notes, Valid: input.Notes != ""},
		Amount:       amountNumeric,
		CurrencyCode: currencyCode,
		Date:         date,
		CategoryID:   categoryID,
		Tags:         tags,
		UpdatedBy:    input.UpdatedBy,
	})
	if err != nil {
		return CreateExpenseResult{}, err
	}

	// Delete old payments and splits
	if err := txRepo.DeleteExpensePayments(ctx, input.ExpenseID); err != nil {
		return CreateExpenseResult{}, err
	}
	if err := txRepo.DeleteExpenseSplits(ctx, input.ExpenseID); err != nil {
		return CreateExpenseResult{}, err
	}

	// Create new payments
	payments := make([]sqlc.ExpensePayment, 0, len(input.Payments))
	for _, paymentInput := range input.Payments {
		paymentAmount, err := stringToNumeric(paymentInput.Amount)
		if err != nil {
			return CreateExpenseResult{}, ErrInvalidAmount
		}

		var pendingUserID pgtype.UUID
		if paymentInput.PendingUserID != nil {
			pendingUserID = *paymentInput.PendingUserID
		}

		if !paymentInput.UserID.Valid && !pendingUserID.Valid {
			return CreateExpenseResult{}, errors.New("payment must have user_id or pending_user_id")
		}

		payment, err := txRepo.CreateExpensePayment(ctx, sqlc.CreateExpensePaymentParams{
			ExpenseID:     updatedExpense.ID,
			UserID:        paymentInput.UserID,
			PendingUserID: pendingUserID,
			Amount:        paymentAmount,
			PaymentMethod: pgtype.Text{String: paymentInput.PaymentMethod, Valid: paymentInput.PaymentMethod != ""},
		})
		if err != nil {
			return CreateExpenseResult{}, err
		}
		payments = append(payments, payment)
	}

	// Create new splits
	splits := make([]sqlc.ExpenseSplit, 0, len(input.Splits))
	for _, splitInput := range input.Splits {
		splitAmount, err := stringToNumeric(splitInput.AmountOwned)
		if err != nil {
			return CreateExpenseResult{}, ErrInvalidAmount
		}

		splitType := strings.TrimSpace(splitInput.SplitType)
		if splitType == "" {
			splitType = "equal"
		}

		var pendingUserID pgtype.UUID
		if splitInput.PendingUserID != nil {
			pendingUserID = *splitInput.PendingUserID
		}

		if !splitInput.UserID.Valid && !pendingUserID.Valid {
			return CreateExpenseResult{}, errors.New("split must have user_id or pending_user_id")
		}

		split, err := txRepo.CreateExpenseSplit(ctx, sqlc.CreateExpenseSplitParams{
			ExpenseID:   updatedExpense.ID,
			UserID:      splitInput.UserID,
			PendingUserID: pendingUserID,
			AmountOwned: splitAmount,
			SplitType:   splitType,
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

	// Log activity
	_ = s.activityService.LogActivity(ctx, LogActivityInput{
		GroupID:    expense.GroupID,
		UserID:     input.UpdatedBy,
		Action:     "expense_updated",
		EntityType: "expense",
		EntityID:   updatedExpense.ID,
		Metadata: map[string]interface{}{
			"title":  updatedExpense.Title,
			"amount": input.Amount,
		},
	})

	return CreateExpenseResult{
		Expense:  updatedExpense,
		Payments: payments,
		Splits:   splits,
	}, nil
}

func (s *expenseService) DeleteExpense(ctx context.Context, expenseID, requesterID pgtype.UUID) error {
	expense, err := s.repo.GetExpenseByID(ctx, expenseID)
	if err != nil {
		return ErrExpenseNotFound
	}

	// Validate user is group member
	if err := s.validateGroupMembership(ctx, expense.GroupID, requesterID); err != nil {
		return err
	}

	if err := s.repo.DeleteExpense(ctx, expenseID); err != nil {
		return err
	}

	// Log activity
	_ = s.activityService.LogActivity(ctx, LogActivityInput{
		GroupID:    expense.GroupID,
		UserID:     requesterID,
		Action:     "expense_deleted",
		EntityType: "expense",
		EntityID:   expenseID,
		Metadata: map[string]interface{}{
			"title": expense.Title,
		},
	})

	return nil
}

func (s *expenseService) GetExpensePayments(ctx context.Context, expenseID, requesterID pgtype.UUID) ([]sqlc.ListExpensePaymentsRow, error) {
	expense, err := s.repo.GetExpenseByID(ctx, expenseID)
	if err != nil {
		return nil, ErrExpenseNotFound
	}

	// Validate user is group member
	if err := s.validateGroupMembership(ctx, expense.GroupID, requesterID); err != nil {
		return nil, err
	}

	return s.repo.ListExpensePayments(ctx, expenseID)
}

func (s *expenseService) GetExpenseSplits(ctx context.Context, expenseID, requesterID pgtype.UUID) ([]sqlc.ListExpenseSplitsRow, error) {
	expense, err := s.repo.GetExpenseByID(ctx, expenseID)
	if err != nil {
		return nil, ErrExpenseNotFound
	}

	// Validate user is group member
	if err := s.validateGroupMembership(ctx, expense.GroupID, requesterID); err != nil {
		return nil, err
	}

	return s.repo.ListExpenseSplits(ctx, expenseID)
}

func (s *expenseService) SearchExpenses(ctx context.Context, input SearchExpensesInput, requesterID pgtype.UUID) ([]sqlc.Expense, error) {
	// Validate group exists
	_, err := s.repo.GetGroupByID(ctx, input.GroupID)
	if err != nil {
		return nil, ErrExpenseNotFound
	}

	// Validate user is group member
	if err := s.validateGroupMembership(ctx, input.GroupID, requesterID); err != nil {
		return nil, err
	}

	params := sqlc.SearchExpensesParams{
		GroupID: input.GroupID,
		Limit:   input.Limit,
		Offset:  input.Offset,
	}

	if input.Query != nil {
		params.Query = pgtype.Text{String: *input.Query, Valid: true}
	}

	if input.StartDate != nil {
		params.StartDate = pgtype.Date{Time: *input.StartDate, Valid: true}
	}

	if input.EndDate != nil {
		params.EndDate = pgtype.Date{Time: *input.EndDate, Valid: true}
	}

	if input.CategoryID != nil {
		params.CategoryID = *input.CategoryID
	}

	if input.CreatedBy != nil {
		params.CreatedBy = *input.CreatedBy
	}

	if input.MinAmount != nil {
		if n, err := stringToNumeric(*input.MinAmount); err == nil {
			params.MinAmount = n
		}
	}

	if input.MaxAmount != nil {
		if n, err := stringToNumeric(*input.MaxAmount); err == nil {
			params.MaxAmount = n
		}
	}
	
	if input.PayerID != nil {
		params.PayerID = *input.PayerID
	}
	
	if input.OwerID != nil {
		params.OwerID = *input.OwerID
	}

	if params.Limit <= 0 {
		params.Limit = 20 // Default limit
	}
	if params.Limit > 100 {
		params.Limit = 100 // Max limit
	}

	return s.repo.SearchExpenses(ctx, params)
}
