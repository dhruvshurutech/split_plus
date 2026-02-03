package service

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/shopspring/decimal"

	"github.com/dhruvsaxena1998/splitplus/internal/db/sqlc"
	"github.com/dhruvsaxena1998/splitplus/internal/repository"
)

var (
	ErrExpenseNotFound         = errors.New("expense not found")
	ErrInvalidAmount           = errors.New("invalid amount")
	ErrPaymentTotalMismatch    = errors.New("payment total does not match expense amount")
	ErrSplitTotalMismatch      = errors.New("split total does not match expense amount")
	ErrCategoryNotInGroup      = errors.New("category does not belong to this group")
	ErrPercentageTotalMismatch = errors.New("percentages must sum to 100")
	ErrAmountRequired          = errors.New("amount required for fixed/custom splits")
	ErrPercentageRequired      = errors.New("percentage required for percentage splits")
	ErrSharesRequired          = errors.New("shares required and must be > 0 for shares splits")
	ErrMixedSplitTypes         = errors.New("all splits must have the same type")
	ErrInvalidSplitType        = errors.New("invalid split type")
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
	Type          string
	Percentage    *string
	Shares        *int
	Amount        *string
}

// calculatedSplit is the result after backend calculation
type calculatedSplit struct {
	UserID        pgtype.UUID
	PendingUserID *pgtype.UUID
	Type          string
	Amount        string
	ShareValue    *string
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
	GroupID    pgtype.UUID
	Query      *string
	StartDate  *time.Time
	EndDate    *time.Time
	CategoryID *pgtype.UUID
	CreatedBy  *pgtype.UUID
	MinAmount  *string
	MaxAmount  *string
	PayerID    *pgtype.UUID
	OwerID     *pgtype.UUID
	Limit      int32
	Offset     int32
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
	userRepo        repository.UserRepository
	pendingUserRepo repository.PendingUserRepository
}

func NewExpenseService(
	repo repository.ExpenseRepository,
	categoryRepo repository.ExpenseCategoryRepository,
	activityService GroupActivityService,
	userRepo repository.UserRepository,
	pendingUserRepo repository.PendingUserRepository,
) ExpenseService {
	return &expenseService{
		repo:            repo,
		categoryRepo:    categoryRepo,
		activityService: activityService,
		userRepo:        userRepo,
		pendingUserRepo: pendingUserRepo,
	}
}

// resolveUserOrPendingUser checks if the given ID belongs to a user or a pending user.
// Returns (userID, pendingUserID, error). One of userID or pendingUserID will be valid, passed ID is treated as "target".
func (s *expenseService) resolveUserOrPendingUser(ctx context.Context, targetID pgtype.UUID) (pgtype.UUID, pgtype.UUID, error) {
	if !targetID.Valid {
		return pgtype.UUID{}, pgtype.UUID{}, nil
	}

	// 1. Check if it's a real user
	_, err := s.userRepo.GetUserByID(ctx, targetID)
	if err == nil {
		return targetID, pgtype.UUID{}, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return pgtype.UUID{}, pgtype.UUID{}, err
	}

	// 2. Check if it's a pending user
	_, err = s.pendingUserRepo.GetPendingUserByID(ctx, targetID)
	if err == nil {
		return pgtype.UUID{}, targetID, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return pgtype.UUID{}, pgtype.UUID{}, err
	}

	// Not found in either
	return pgtype.UUID{}, pgtype.UUID{}, ErrUserNotFound
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

func calculateSplitAmounts(expenseAmount string, splits []SplitInput) ([]calculatedSplit, error) {
	if len(splits) == 0 {
		return nil, errors.New("at least one split is required")
	}

	// 1. Validate all splits have same type
	splitType := splits[0].Type
	for _, split := range splits {
		if split.Type != splitType {
			return nil, ErrMixedSplitTypes
		}
	}

	// 2. Calculate based on type
	switch splitType {
	case "equal":
		return calculateEqualSplits(expenseAmount, splits)
	case "percentage":
		return calculatePercentageSplits(expenseAmount, splits)
	case "shares":
		return calculateSharesSplits(expenseAmount, splits)
	case "fixed", "custom":
		return validateFixedSplits(expenseAmount, splits)
	default:
		return nil, ErrInvalidSplitType
	}
}

func calculateEqualSplits(total string, splits []SplitInput) ([]calculatedSplit, error) {
	totalDec, err := decimal.NewFromString(total)
	if err != nil {
		return nil, ErrInvalidAmount
	}

	count := len(splits)

	// Calculate base amount (rounded down to 2 decimal places)
	baseAmount := totalDec.Div(decimal.NewFromInt(int64(count))).Round(2)

	// Calculate remainder to add to last split
	allocatedTotal := baseAmount.Mul(decimal.NewFromInt(int64(count - 1)))
	lastAmount := totalDec.Sub(allocatedTotal)

	result := make([]calculatedSplit, count)
	for i, split := range splits {
		if i == count-1 {
			result[i].Amount = lastAmount.String()
		} else {
			result[i].Amount = baseAmount.String()
		}
		result[i].UserID = split.UserID
		result[i].PendingUserID = split.PendingUserID
		result[i].Type = "equal"
		result[i].ShareValue = nil
	}
	return result, nil
}

func calculatePercentageSplits(total string, splits []SplitInput) ([]calculatedSplit, error) {
	totalDec, err := decimal.NewFromString(total)
	if err != nil {
		return nil, ErrInvalidAmount
	}

	// Validate percentages sum to 100
	var percentageSum decimal.Decimal
	for _, split := range splits {
		if split.Percentage == nil {
			return nil, ErrPercentageRequired
		}
		pct, err := decimal.NewFromString(*split.Percentage)
		if err != nil || pct.LessThanOrEqual(decimal.Zero) {
			return nil, ErrInvalidAmount
		}
		percentageSum = percentageSum.Add(pct)
	}
	if !percentageSum.Equal(decimal.NewFromInt(100)) {
		return nil, ErrPercentageTotalMismatch
	}

	// Calculate amounts, adjust last for rounding
	result := make([]calculatedSplit, len(splits))
	var allocatedTotal decimal.Decimal

	for i, split := range splits {
		pct, _ := decimal.NewFromString(*split.Percentage)

		if i == len(splits)-1 {
			// Last split gets remainder
			result[i].Amount = totalDec.Sub(allocatedTotal).String()
		} else {
			amount := totalDec.Mul(pct).Div(decimal.NewFromInt(100)).Round(2)
			result[i].Amount = amount.String()
			allocatedTotal = allocatedTotal.Add(amount)
		}
		result[i].UserID = split.UserID
		result[i].PendingUserID = split.PendingUserID
		result[i].Type = "percentage"
		result[i].ShareValue = split.Percentage
	}
	return result, nil
}

func calculateSharesSplits(total string, splits []SplitInput) ([]calculatedSplit, error) {
	totalDec, err := decimal.NewFromString(total)
	if err != nil {
		return nil, ErrInvalidAmount
	}

	// Calculate total shares
	var totalShares int64
	for _, split := range splits {
		if split.Shares == nil || *split.Shares <= 0 {
			return nil, ErrSharesRequired
		}
		totalShares += int64(*split.Shares)
	}

	// Calculate amounts based on share proportion
	result := make([]calculatedSplit, len(splits))
	var allocatedTotal decimal.Decimal

	for i, split := range splits {
		if i == len(splits)-1 {
			// Last split gets remainder to ensure exact total
			result[i].Amount = totalDec.Sub(allocatedTotal).String()
		} else {
			// amount = total * (shares / totalShares)
			shareRatio := decimal.NewFromInt(int64(*split.Shares)).Div(decimal.NewFromInt(totalShares))
			amount := totalDec.Mul(shareRatio).Round(2)
			result[i].Amount = amount.String()
			allocatedTotal = allocatedTotal.Add(amount)
		}
		result[i].UserID = split.UserID
		result[i].PendingUserID = split.PendingUserID
		result[i].Type = "shares"
		shareValueStr := decimal.NewFromInt(int64(*split.Shares)).String()
		result[i].ShareValue = &shareValueStr
	}
	return result, nil
}

func validateFixedSplits(total string, splits []SplitInput) ([]calculatedSplit, error) {
	totalDec, err := decimal.NewFromString(total)
	if err != nil {
		return nil, ErrInvalidAmount
	}

	var splitSum decimal.Decimal
	result := make([]calculatedSplit, len(splits))

	for i, split := range splits {
		if split.Amount == nil {
			return nil, ErrAmountRequired
		}
		amount, err := decimal.NewFromString(*split.Amount)
		if err != nil || amount.LessThan(decimal.Zero) {
			return nil, ErrInvalidAmount
		}
		splitSum = splitSum.Add(amount)

		result[i].UserID = split.UserID
		result[i].PendingUserID = split.PendingUserID
		result[i].Amount = *split.Amount
		result[i].Type = split.Type
		result[i].ShareValue = nil
	}

	if !splitSum.Equal(totalDec) {
		return nil, ErrSplitTotalMismatch
	}

	return result, nil
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

	// Always use group currency for group expenses
	currencyCode := group.CurrencyCode

	// Validate payments
	if len(input.Payments) == 0 {
		return CreateExpenseResult{}, errors.New("at least one payment is required")
	}
	if err := s.validatePaymentsTotal(input.Amount, input.Payments); err != nil {
		return CreateExpenseResult{}, err
	}

	// Calculate split amounts
	calculatedSplits, err := calculateSplitAmounts(input.Amount, input.Splits)
	if err != nil {
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

		var userID, pendingUserID pgtype.UUID
		if paymentInput.UserID.Valid {
			// Resolve UserID which might be a PendingUserID
			userID, pendingUserID, err = s.resolveUserOrPendingUser(ctx, paymentInput.UserID)
			if err != nil {
				return CreateExpenseResult{}, err
			}
		} else if paymentInput.PendingUserID != nil && paymentInput.PendingUserID.Valid {
			pendingUserID = *paymentInput.PendingUserID
		} else {
			return CreateExpenseResult{}, errors.New("payment must have user_id or pending_user_id")
		}

		payment, err := txRepo.CreateExpensePayment(ctx, sqlc.CreateExpensePaymentParams{
			ExpenseID:     expense.ID,
			UserID:        userID,
			PendingUserID: pendingUserID,
			Amount:        paymentAmount,
			PaymentMethod: pgtype.Text{String: paymentInput.PaymentMethod, Valid: paymentInput.PaymentMethod != ""},
		})
		if err != nil {
			return CreateExpenseResult{}, err
		}
		payments = append(payments, payment)
	}

	// Create splits using calculated amounts
	splits := make([]sqlc.ExpenseSplit, 0, len(calculatedSplits))
	for _, calcSplit := range calculatedSplits {
		splitAmount, err := stringToNumeric(calcSplit.Amount)
		if err != nil {
			return CreateExpenseResult{}, ErrInvalidAmount
		}

		var userID, pendingUserID pgtype.UUID
		if calcSplit.UserID.Valid {
			// Resolve UserID which might be a PendingUserID
			userID, pendingUserID, err = s.resolveUserOrPendingUser(ctx, calcSplit.UserID)
			if err != nil {
				return CreateExpenseResult{}, err
			}
		} else if calcSplit.PendingUserID != nil && calcSplit.PendingUserID.Valid {
			pendingUserID = *calcSplit.PendingUserID
		} else {
			return CreateExpenseResult{}, errors.New("split must have user_id or pending_user_id")
		}

		var shareValue pgtype.Numeric
		if calcSplit.ShareValue != nil {
			shareValue, err = stringToNumeric(*calcSplit.ShareValue)
			if err != nil {
				return CreateExpenseResult{}, ErrInvalidAmount
			}
		}

		split, err := txRepo.CreateExpenseSplit(ctx, sqlc.CreateExpenseSplitParams{
			ExpenseID:     expense.ID,
			UserID:        userID,
			PendingUserID: pendingUserID,
			AmountOwned:   splitAmount,
			SplitType:     calcSplit.Type,
			ShareValue:    shareValue,
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
		Metadata: buildExpenseActivityMetadata(expense, input.Amount, currencyCode, payments, splits, nil),
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

	// Calculate split amounts
	calculatedSplits, err := calculateSplitAmounts(input.Amount, input.Splits)
	if err != nil {
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

	// Always use group currency for group expenses
	group, err := s.repo.GetGroupByID(ctx, expense.GroupID)
	if err != nil {
		return CreateExpenseResult{}, ErrExpenseNotFound
	}
	currencyCode := group.CurrencyCode

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

		var userID, pendingUserID pgtype.UUID
		if paymentInput.UserID.Valid {
			userID, pendingUserID, err = s.resolveUserOrPendingUser(ctx, paymentInput.UserID)
			if err != nil {
				return CreateExpenseResult{}, err
			}
		} else if paymentInput.PendingUserID != nil && paymentInput.PendingUserID.Valid {
			pendingUserID = *paymentInput.PendingUserID
		} else {
			return CreateExpenseResult{}, errors.New("payment must have user_id or pending_user_id")
		}

		payment, err := txRepo.CreateExpensePayment(ctx, sqlc.CreateExpensePaymentParams{
			ExpenseID:     updatedExpense.ID,
			UserID:        userID,
			PendingUserID: pendingUserID,
			Amount:        paymentAmount,
			PaymentMethod: pgtype.Text{String: paymentInput.PaymentMethod, Valid: paymentInput.PaymentMethod != ""},
		})
		if err != nil {
			return CreateExpenseResult{}, err
		}
		payments = append(payments, payment)
	}

	// Create new splits using calculated amounts
	splits := make([]sqlc.ExpenseSplit, 0, len(calculatedSplits))
	for _, calcSplit := range calculatedSplits {
		splitAmount, err := stringToNumeric(calcSplit.Amount)
		if err != nil {
			return CreateExpenseResult{}, ErrInvalidAmount
		}

		var userID, pendingUserID pgtype.UUID
		if calcSplit.UserID.Valid {
			userID, pendingUserID, err = s.resolveUserOrPendingUser(ctx, calcSplit.UserID)
			if err != nil {
				return CreateExpenseResult{}, err
			}
		} else if calcSplit.PendingUserID != nil && calcSplit.PendingUserID.Valid {
			pendingUserID = *calcSplit.PendingUserID
		} else {
			return CreateExpenseResult{}, errors.New("split must have user_id or pending_user_id")
		}

		var shareValue pgtype.Numeric
		if calcSplit.ShareValue != nil {
			shareValue, err = stringToNumeric(*calcSplit.ShareValue)
			if err != nil {
				return CreateExpenseResult{}, ErrInvalidAmount
			}
		}

		split, err := txRepo.CreateExpenseSplit(ctx, sqlc.CreateExpenseSplitParams{
			ExpenseID:     updatedExpense.ID,
			UserID:        userID,
			PendingUserID: pendingUserID,
			AmountOwned:   splitAmount,
			SplitType:     calcSplit.Type,
			ShareValue:    shareValue,
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
		Metadata: buildExpenseActivityMetadata(updatedExpense, input.Amount, currencyCode, payments, splits, &expense),
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
		Metadata: buildExpenseActivityMetadata(expense, numericToStringSafe(expense.Amount), expense.CurrencyCode, nil, nil, nil),
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

func buildExpenseActivityMetadata(
	expense sqlc.Expense,
	amount string,
	currencyCode string,
	payments []sqlc.ExpensePayment,
	splits []sqlc.ExpenseSplit,
	before *sqlc.Expense,
) map[string]interface{} {
	metadata := map[string]interface{}{
		"version": 1,
		"summary": map[string]interface{}{
			"title":         expense.Title,
			"amount":        amount,
			"currency_code": currencyCode,
		},
		"split_type": splitTypeFromSplits(splits),
		"splits":     buildActivitySplits(splits),
		"payments":   buildActivityPayments(payments),
	}

	if before != nil {
		metadata["before"] = map[string]interface{}{
			"title":         before.Title,
			"amount":        numericToStringSafe(before.Amount),
			"currency_code": before.CurrencyCode,
		}
		metadata["after"] = map[string]interface{}{
			"title":         expense.Title,
			"amount":        amount,
			"currency_code": currencyCode,
		}
	}

	return metadata
}

func splitTypeFromSplits(splits []sqlc.ExpenseSplit) string {
	if len(splits) == 0 {
		return ""
	}
	return splits[0].SplitType
}

func buildActivitySplits(splits []sqlc.ExpenseSplit) []map[string]interface{} {
	out := make([]map[string]interface{}, 0, len(splits))
	for _, split := range splits {
		entry := map[string]interface{}{
			"id":               uuidToString(split.ID),
			"user_id":          uuidToString(split.UserID),
			"pending_user_id":  uuidToString(split.PendingUserID),
			"amount":           numericToStringSafe(split.AmountOwned),
			"type":             split.SplitType,
		}
		if split.ShareValue.Valid {
			entry["share_value"] = numericToStringSafe(split.ShareValue)
		}
		out = append(out, entry)
	}
	return out
}

func buildActivityPayments(payments []sqlc.ExpensePayment) []map[string]interface{} {
	out := make([]map[string]interface{}, 0, len(payments))
	for _, payment := range payments {
		entry := map[string]interface{}{
			"id":               uuidToString(payment.ID),
			"user_id":          uuidToString(payment.UserID),
			"pending_user_id":  uuidToString(payment.PendingUserID),
			"amount":           numericToStringSafe(payment.Amount),
		}
		if payment.PaymentMethod.Valid {
			entry["payment_method"] = payment.PaymentMethod.String
		}
		out = append(out, entry)
	}
	return out
}

func uuidToString(id pgtype.UUID) string {
	if !id.Valid {
		return ""
	}
	return id.String()
}

func numericToStringSafe(n pgtype.Numeric) string {
	if s, err := numericToString(n); err == nil {
		return s
	}
	return ""
}
