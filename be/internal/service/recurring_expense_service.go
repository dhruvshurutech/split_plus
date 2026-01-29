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
	ErrRecurringExpenseNotFound = errors.New("recurring expense not found")
	ErrInvalidRecurringInterval = errors.New("invalid recurring interval")
	ErrInvalidDateConfiguration = errors.New("invalid date configuration for interval")
)

type RecurringPaymentInput struct {
	UserID        pgtype.UUID
	Amount        string
	PaymentMethod string
}

type RecurringSplitInput struct {
	UserID      pgtype.UUID
	AmountOwned string
	SplitType   string
}

type CreateRecurringExpenseInput struct {
	GroupID        pgtype.UUID
	Title          string
	Notes          string
	Amount         string
	CurrencyCode   string
	RepeatInterval string
	DayOfMonth     *int
	DayOfWeek      *int
	StartDate      time.Time
	EndDate        *time.Time
	CreatedBy      pgtype.UUID
	Payments       []RecurringPaymentInput
	Splits         []RecurringSplitInput
}

type UpdateRecurringExpenseInput struct {
	RecurringExpenseID pgtype.UUID
	Title              string
	Notes              string
	Amount             string
	CurrencyCode       string
	RepeatInterval     string
	DayOfMonth         *int
	DayOfWeek          *int
	StartDate          time.Time
	EndDate            *time.Time
	IsActive           bool
	UpdatedBy          pgtype.UUID
	Payments           []RecurringPaymentInput
	Splits             []RecurringSplitInput
}

type RecurringExpenseService interface {
	CreateRecurringExpense(ctx context.Context, input CreateRecurringExpenseInput) (sqlc.RecurringExpense, error)
	GetRecurringExpenseByID(ctx context.Context, id, requesterID pgtype.UUID) (sqlc.RecurringExpense, error)
	ListRecurringExpensesByGroup(ctx context.Context, groupID, requesterID pgtype.UUID) ([]sqlc.RecurringExpense, error)
	UpdateRecurringExpense(ctx context.Context, input UpdateRecurringExpenseInput) (sqlc.RecurringExpense, error)
	DeleteRecurringExpense(ctx context.Context, id, requesterID pgtype.UUID) error
	GenerateExpenseFromRecurring(ctx context.Context, recurringID, requesterID pgtype.UUID) (CreateExpenseResult, error)
	ProcessDueRecurringExpenses(ctx context.Context) error
}

type recurringExpenseService struct {
	repo           repository.RecurringExpenseRepository
	expenseService ExpenseService
}

func NewRecurringExpenseService(repo repository.RecurringExpenseRepository, expenseService ExpenseService) RecurringExpenseService {
	return &recurringExpenseService{
		repo:           repo,
		expenseService: expenseService,
	}
}

func (s *recurringExpenseService) validateGroupMembership(ctx context.Context, groupID, userID pgtype.UUID) error {
	_, err := s.repo.GetGroupMember(ctx, sqlc.GetGroupMemberParams{
		GroupID: groupID,
		UserID:  userID,
	})
	if err != nil {
		return ErrNotGroupMember
	}
	return nil
}

func (s *recurringExpenseService) validatePaymentsTotal(expenseAmount string, payments []PaymentInput) error {
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

func (s *recurringExpenseService) validateSplitsTotal(expenseAmount string, splits []SplitInput) error {
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

func (s *recurringExpenseService) validateIntervalFields(interval string, dayOfMonth, dayOfWeek *int) error {
	switch interval {
	case "daily":
		if dayOfMonth != nil || dayOfWeek != nil {
			return errors.New("day_of_month and day_of_week must be NULL for daily interval")
		}
	case "weekly":
		if dayOfWeek == nil {
			return errors.New("day_of_week is required for weekly interval")
		}
		if *dayOfWeek < 0 || *dayOfWeek > 6 {
			return errors.New("day_of_week must be between 0 (Sunday) and 6 (Saturday)")
		}
		if dayOfMonth != nil {
			return errors.New("day_of_month must be NULL for weekly interval")
		}
	case "monthly", "yearly":
		if dayOfMonth == nil {
			return errors.New("day_of_month is required for monthly/yearly interval")
		}
		if *dayOfMonth < 1 || *dayOfMonth > 31 {
			return errors.New("day_of_month must be between 1 and 31")
		}
		if dayOfWeek != nil {
			return errors.New("day_of_week must be NULL for monthly/yearly interval")
		}
	default:
		return ErrInvalidRecurringInterval
	}
	return nil
}

func (s *recurringExpenseService) calculateNextOccurrenceDate(currentDate time.Time, interval string, dayOfMonth, dayOfWeek *int) (time.Time, error) {
	switch interval {
	case "daily":
		return currentDate.AddDate(0, 0, 1), nil
	case "weekly":
		return currentDate.AddDate(0, 0, 7), nil
	case "monthly":
		nextDate := currentDate.AddDate(0, 1, 0)
		// Handle month-end edge cases
		if dayOfMonth != nil {
			// If the day doesn't exist in the next month, use the last day
			lastDayOfMonth := time.Date(nextDate.Year(), nextDate.Month()+1, 0, 0, 0, 0, 0, nextDate.Location()).Day()
			day := *dayOfMonth
			if day > lastDayOfMonth {
				day = lastDayOfMonth
			}
			nextDate = time.Date(nextDate.Year(), nextDate.Month(), day, 0, 0, 0, 0, nextDate.Location())
		}
		return nextDate, nil
	case "yearly":
		nextDate := currentDate.AddDate(1, 0, 0)
		// Handle month-end edge cases (especially for Feb 29/30)
		if dayOfMonth != nil {
			lastDayOfMonth := time.Date(nextDate.Year(), nextDate.Month()+1, 0, 0, 0, 0, 0, nextDate.Location()).Day()
			day := *dayOfMonth
			if day > lastDayOfMonth {
				day = lastDayOfMonth
			}
			nextDate = time.Date(nextDate.Year(), nextDate.Month(), day, 0, 0, 0, 0, nextDate.Location())
		}
		return nextDate, nil
	default:
		return time.Time{}, ErrInvalidRecurringInterval
	}
}

func (s *recurringExpenseService) CreateRecurringExpense(ctx context.Context, input CreateRecurringExpenseInput) (sqlc.RecurringExpense, error) {
	// Validate group exists
	group, err := s.repo.GetGroupByID(ctx, input.GroupID)
	if err != nil {
		return sqlc.RecurringExpense{}, ErrGroupNotFound
	}

	// Validate user is group member
	if err := s.validateGroupMembership(ctx, input.GroupID, input.CreatedBy); err != nil {
		return sqlc.RecurringExpense{}, err
	}

	// Validate interval
	interval := strings.TrimSpace(strings.ToLower(input.RepeatInterval))
	validIntervals := map[string]bool{"daily": true, "weekly": true, "monthly": true, "yearly": true}
	if !validIntervals[interval] {
		return sqlc.RecurringExpense{}, ErrInvalidRecurringInterval
	}

	// Validate interval-specific fields
	if err := s.validateIntervalFields(interval, input.DayOfMonth, input.DayOfWeek); err != nil {
		return sqlc.RecurringExpense{}, err
	}

	// Validate amount
	amount, err := decimal.NewFromString(input.Amount)
	if err != nil || amount.LessThanOrEqual(decimal.Zero) {
		return sqlc.RecurringExpense{}, ErrInvalidAmount
	}

	// Use group currency if not provided
	currencyCode := strings.TrimSpace(input.CurrencyCode)
	if currencyCode == "" {
		currencyCode = group.CurrencyCode
	}

	// Validate title
	title := strings.TrimSpace(input.Title)
	if title == "" {
		return sqlc.RecurringExpense{}, errors.New("title is required")
	}

	// Validate dates
	if input.EndDate != nil && input.StartDate.After(*input.EndDate) {
		return sqlc.RecurringExpense{}, errors.New("start_date must be before or equal to end_date")
	}

	// Calculate next occurrence date
	nextOccurrenceDate := input.StartDate

	// Convert amount to numeric
	amountNumeric, err := stringToNumeric(input.Amount)
	if err != nil {
		return sqlc.RecurringExpense{}, ErrInvalidAmount
	}

	// Convert dates
	startDate := pgtype.Date{Time: input.StartDate, Valid: true}
	var endDate pgtype.Date
	if input.EndDate != nil {
		endDate = pgtype.Date{Time: *input.EndDate, Valid: true}
	}
	nextOccurrenceDatePg := pgtype.Date{Time: nextOccurrenceDate, Valid: true}

	// Convert day_of_month and day_of_week
	var dayOfMonthPg pgtype.Int4
	if input.DayOfMonth != nil {
		dayOfMonthPg = pgtype.Int4{Int32: int32(*input.DayOfMonth), Valid: true}
	}
	var dayOfWeekPg pgtype.Int4
	if input.DayOfWeek != nil {
		dayOfWeekPg = pgtype.Int4{Int32: int32(*input.DayOfWeek), Valid: true}
	}

	// Start transaction
	tx, err := s.repo.BeginTx(ctx)
	if err != nil {
		return sqlc.RecurringExpense{}, err
	}
	defer tx.Rollback(ctx)

	txRepo := s.repo.WithTx(tx)

	// Create recurring expense
	recurringExpense, err := txRepo.CreateRecurringExpense(ctx, sqlc.CreateRecurringExpenseParams{
		GroupID:            input.GroupID,
		Title:              title,
		Notes:              pgtype.Text{String: input.Notes, Valid: input.Notes != ""},
		Amount:             amountNumeric,
		CurrencyCode:       currencyCode,
		RepeatInterval:     interval,
		DayOfMonth:         dayOfMonthPg,
		DayOfWeek:          dayOfWeekPg,
		StartDate:          startDate,
		EndDate:            endDate,
		NextOccurrenceDate: nextOccurrenceDatePg,
		IsActive:           pgtype.Bool{Bool: true, Valid: true},
		CreatedBy:          input.CreatedBy,
	})
	if err != nil {
		return sqlc.RecurringExpense{}, err
	}

	// Validate and create payments
	if len(input.Payments) == 0 {
		return sqlc.RecurringExpense{}, errors.New("at least one payment is required")
	}
	if err := s.validatePaymentsTotal(input.Amount, convertRecurringPayments(input.Payments)); err != nil {
		return sqlc.RecurringExpense{}, err
	}

	for _, paymentInput := range input.Payments {
		paymentAmount, err := stringToNumeric(paymentInput.Amount)
		if err != nil {
			return sqlc.RecurringExpense{}, ErrInvalidAmount
		}

		_, err = txRepo.CreateRecurringExpensePayment(ctx, sqlc.CreateRecurringExpensePaymentParams{
			RecurringExpenseID: recurringExpense.ID,
			UserID:             paymentInput.UserID,
			Amount:             paymentAmount,
			PaymentMethod:      pgtype.Text{String: paymentInput.PaymentMethod, Valid: paymentInput.PaymentMethod != ""},
		})
		if err != nil {
			return sqlc.RecurringExpense{}, err
		}
	}

	// Validate and create splits
	if len(input.Splits) == 0 {
		return sqlc.RecurringExpense{}, errors.New("at least one split is required")
	}
	if err := s.validateSplitsTotal(input.Amount, convertRecurringSplits(input.Splits)); err != nil {
		return sqlc.RecurringExpense{}, err
	}

	for _, splitInput := range input.Splits {
		splitAmount, err := stringToNumeric(splitInput.AmountOwned)
		if err != nil {
			return sqlc.RecurringExpense{}, ErrInvalidAmount
		}

		_, err = txRepo.CreateRecurringExpenseSplit(ctx, sqlc.CreateRecurringExpenseSplitParams{
			RecurringExpenseID: recurringExpense.ID,
			UserID:             splitInput.UserID,
			AmountOwned:        splitAmount,
			SplitType:          splitInput.SplitType,
		})
		if err != nil {
			return sqlc.RecurringExpense{}, err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return sqlc.RecurringExpense{}, err
	}

	return recurringExpense, nil
}

func convertRecurringPayments(payments []RecurringPaymentInput) []PaymentInput {
	result := make([]PaymentInput, len(payments))
	for i, p := range payments {
		result[i] = PaymentInput{
			UserID:        p.UserID,
			Amount:        p.Amount,
			PaymentMethod: p.PaymentMethod,
		}
	}
	return result
}

func convertRecurringSplits(splits []RecurringSplitInput) []SplitInput {
	result := make([]SplitInput, len(splits))
	for i, s := range splits {
		result[i] = SplitInput{
			UserID:      s.UserID,
			AmountOwned: s.AmountOwned,
			SplitType:   s.SplitType,
		}
	}
	return result
}

func (s *recurringExpenseService) GetRecurringExpenseByID(ctx context.Context, id, requesterID pgtype.UUID) (sqlc.RecurringExpense, error) {
	recurringExpense, err := s.repo.GetRecurringExpenseByID(ctx, id)
	if err != nil {
		return sqlc.RecurringExpense{}, ErrRecurringExpenseNotFound
	}

	// Validate requester is group member
	if err := s.validateGroupMembership(ctx, recurringExpense.GroupID, requesterID); err != nil {
		return sqlc.RecurringExpense{}, err
	}

	return recurringExpense, nil
}

func (s *recurringExpenseService) ListRecurringExpensesByGroup(ctx context.Context, groupID, requesterID pgtype.UUID) ([]sqlc.RecurringExpense, error) {
	// Validate requester is group member
	if err := s.validateGroupMembership(ctx, groupID, requesterID); err != nil {
		return nil, err
	}

	return s.repo.ListRecurringExpensesByGroup(ctx, groupID)
}

func (s *recurringExpenseService) UpdateRecurringExpense(ctx context.Context, input UpdateRecurringExpenseInput) (sqlc.RecurringExpense, error) {
	// Get existing recurring expense
	existing, err := s.repo.GetRecurringExpenseByID(ctx, input.RecurringExpenseID)
	if err != nil {
		return sqlc.RecurringExpense{}, ErrRecurringExpenseNotFound
	}

	// Validate requester is group member
	if err := s.validateGroupMembership(ctx, existing.GroupID, input.UpdatedBy); err != nil {
		return sqlc.RecurringExpense{}, err
	}

	// Validate interval
	interval := strings.TrimSpace(strings.ToLower(input.RepeatInterval))
	validIntervals := map[string]bool{"daily": true, "weekly": true, "monthly": true, "yearly": true}
	if !validIntervals[interval] {
		return sqlc.RecurringExpense{}, ErrInvalidRecurringInterval
	}

	// Validate interval-specific fields
	if err := s.validateIntervalFields(interval, input.DayOfMonth, input.DayOfWeek); err != nil {
		return sqlc.RecurringExpense{}, err
	}

	// Validate amount
	amount, err := decimal.NewFromString(input.Amount)
	if err != nil || amount.LessThanOrEqual(decimal.Zero) {
		return sqlc.RecurringExpense{}, ErrInvalidAmount
	}

	// Validate title
	title := strings.TrimSpace(input.Title)
	if title == "" {
		return sqlc.RecurringExpense{}, errors.New("title is required")
	}

	// Validate dates
	if input.EndDate != nil && input.StartDate.After(*input.EndDate) {
		return sqlc.RecurringExpense{}, errors.New("start_date must be before or equal to end_date")
	}

	// Convert amount to numeric
	amountNumeric, err := stringToNumeric(input.Amount)
	if err != nil {
		return sqlc.RecurringExpense{}, ErrInvalidAmount
	}

	// Convert dates
	startDate := pgtype.Date{Time: input.StartDate, Valid: true}
	var endDate pgtype.Date
	if input.EndDate != nil {
		endDate = pgtype.Date{Time: *input.EndDate, Valid: true}
	}

	// Keep existing next_occurrence_date or recalculate if needed
	nextOccurrenceDate := existing.NextOccurrenceDate

	// Convert day_of_month and day_of_week
	var dayOfMonthPg pgtype.Int4
	if input.DayOfMonth != nil {
		dayOfMonthPg = pgtype.Int4{Int32: int32(*input.DayOfMonth), Valid: true}
	}
	var dayOfWeekPg pgtype.Int4
	if input.DayOfWeek != nil {
		dayOfWeekPg = pgtype.Int4{Int32: int32(*input.DayOfWeek), Valid: true}
	}

	// Update recurring expense
	updated, err := s.repo.UpdateRecurringExpense(ctx, sqlc.UpdateRecurringExpenseParams{
		ID:                 input.RecurringExpenseID,
		Title:              title,
		Notes:              pgtype.Text{String: input.Notes, Valid: input.Notes != ""},
		Amount:             amountNumeric,
		CurrencyCode:       strings.TrimSpace(input.CurrencyCode),
		RepeatInterval:     interval,
		DayOfMonth:         dayOfMonthPg,
		DayOfWeek:          dayOfWeekPg,
		StartDate:          startDate,
		EndDate:            endDate,
		NextOccurrenceDate: nextOccurrenceDate,
		IsActive:           pgtype.Bool{Bool: input.IsActive, Valid: true},
		UpdatedBy:          input.UpdatedBy,
	})
	if err != nil {
		return sqlc.RecurringExpense{}, err
	}

	return updated, nil
}

func (s *recurringExpenseService) DeleteRecurringExpense(ctx context.Context, id, requesterID pgtype.UUID) error {
	// Get existing recurring expense
	existing, err := s.repo.GetRecurringExpenseByID(ctx, id)
	if err != nil {
		return ErrRecurringExpenseNotFound
	}

	// Validate requester is group member
	if err := s.validateGroupMembership(ctx, existing.GroupID, requesterID); err != nil {
		return err
	}

	return s.repo.DeleteRecurringExpense(ctx, id)
}

func (s *recurringExpenseService) GenerateExpenseFromRecurring(ctx context.Context, recurringID, requesterID pgtype.UUID) (CreateExpenseResult, error) {
	// Get recurring expense
	recurringExpense, err := s.repo.GetRecurringExpenseByID(ctx, recurringID)
	if err != nil {
		return CreateExpenseResult{}, ErrRecurringExpenseNotFound
	}

	// Validate requester is group member
	if err := s.validateGroupMembership(ctx, recurringExpense.GroupID, requesterID); err != nil {
		return CreateExpenseResult{}, err
	}

	// Check if recurring expense is active
	if !recurringExpense.IsActive.Bool {
		return CreateExpenseResult{}, errors.New("recurring expense is not active")
	}

	// Check if it's due
	if recurringExpense.NextOccurrenceDate.Time.After(time.Now()) {
		return CreateExpenseResult{}, errors.New("recurring expense is not due yet")
	}

	// Get payments and splits
	paymentRows, err := s.repo.ListRecurringExpensePayments(ctx, recurringID)
	if err != nil {
		return CreateExpenseResult{}, err
	}

	splitRows, err := s.repo.ListRecurringExpenseSplits(ctx, recurringID)
	if err != nil {
		return CreateExpenseResult{}, err
	}

	// Convert to expense input format
	payments := make([]PaymentInput, len(paymentRows))
	for i, p := range paymentRows {
		amountStr, _ := numericToString(p.Amount)
		payments[i] = PaymentInput{
			UserID:        p.UserID,
			Amount:        amountStr,
			PaymentMethod: getStringValue(p.PaymentMethod),
		}
	}

	splits := make([]SplitInput, len(splitRows))
	for i, s := range splitRows {
		amountStr, _ := numericToString(s.AmountOwned)
		splits[i] = SplitInput{
			UserID:      s.UserID,
			AmountOwned: amountStr,
			SplitType:   s.SplitType,
		}
	}

	// Get amount as string
	amountStr, _ := numericToString(recurringExpense.Amount)

	// Create expense from template
	expenseResult, err := s.expenseService.CreateExpense(ctx, CreateExpenseInput{
		GroupID:      recurringExpense.GroupID,
		Title:        recurringExpense.Title,
		Notes:        getStringValue(recurringExpense.Notes),
		Amount:       amountStr,
		CurrencyCode: recurringExpense.CurrencyCode,
		Date:         recurringExpense.NextOccurrenceDate.Time,
		CreatedBy:    requesterID,
		Payments:     payments,
		Splits:       splits,
	})
	if err != nil {
		return CreateExpenseResult{}, err
	}

	// Calculate next occurrence date
	var dayOfMonth *int
	if recurringExpense.DayOfMonth.Valid {
		day := int(recurringExpense.DayOfMonth.Int32)
		dayOfMonth = &day
	}
	var dayOfWeek *int
	if recurringExpense.DayOfWeek.Valid {
		day := int(recurringExpense.DayOfWeek.Int32)
		dayOfWeek = &day
	}

	nextDate, err := s.calculateNextOccurrenceDate(recurringExpense.NextOccurrenceDate.Time, recurringExpense.RepeatInterval, dayOfMonth, dayOfWeek)
	if err != nil {
		return CreateExpenseResult{}, err
	}

	// Check if end_date reached
	isActive := true
	if recurringExpense.EndDate.Valid && !nextDate.After(recurringExpense.EndDate.Time) {
		// If next date would be after end_date, check if we should deactivate
		if nextDate.After(recurringExpense.EndDate.Time) || nextDate.Equal(recurringExpense.EndDate.Time) {
			isActive = false
		}
	}

	// Update next occurrence date
	nextDatePg := pgtype.Date{Time: nextDate, Valid: true}
	_, err = s.repo.UpdateNextOccurrenceDate(ctx, sqlc.UpdateNextOccurrenceDateParams{
		ID:                 recurringID,
		NextOccurrenceDate: nextDatePg,
	})
	if err != nil {
		return CreateExpenseResult{}, err
	}

	// Update active status if needed
	if !isActive {
		_, err = s.repo.UpdateRecurringExpenseActiveStatus(ctx, sqlc.UpdateRecurringExpenseActiveStatusParams{
			ID:       recurringID,
			IsActive: pgtype.Bool{Bool: false, Valid: true},
		})
		if err != nil {
			return CreateExpenseResult{}, err
		}
	}

	return expenseResult, nil
}

func (s *recurringExpenseService) ProcessDueRecurringExpenses(ctx context.Context) error {
	// Get all due recurring expenses
	dueExpenses, err := s.repo.GetRecurringExpensesDue(ctx)
	if err != nil {
		return err
	}

	// Process each due expense
	for _, recurringExpense := range dueExpenses {
		// Use the original creator as requesterID
		_, err := s.GenerateExpenseFromRecurring(ctx, recurringExpense.ID, recurringExpense.CreatedBy)
		if err != nil {
			// Log error but continue processing other expenses
			// In production, you'd want proper logging here
			continue
		}
	}

	return nil
}

func getStringValue(text pgtype.Text) string {
	if text.Valid {
		return text.String
	}
	return ""
}
