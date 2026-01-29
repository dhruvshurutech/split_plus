package repository

import (
	"context"

	"github.com/dhruvsaxena1998/splitplus/internal/db/sqlc"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type RecurringExpenseRepository interface {
	// Transaction support
	BeginTx(ctx context.Context) (pgx.Tx, error)
	WithTx(tx pgx.Tx) RecurringExpenseRepository

	// Recurring expense operations
	CreateRecurringExpense(ctx context.Context, params sqlc.CreateRecurringExpenseParams) (sqlc.RecurringExpense, error)
	GetRecurringExpenseByID(ctx context.Context, id pgtype.UUID) (sqlc.RecurringExpense, error)
	ListRecurringExpensesByGroup(ctx context.Context, groupID pgtype.UUID) ([]sqlc.RecurringExpense, error)
	UpdateRecurringExpense(ctx context.Context, params sqlc.UpdateRecurringExpenseParams) (sqlc.RecurringExpense, error)
	DeleteRecurringExpense(ctx context.Context, id pgtype.UUID) error
	GetRecurringExpensesDue(ctx context.Context) ([]sqlc.RecurringExpense, error)
	UpdateNextOccurrenceDate(ctx context.Context, params sqlc.UpdateNextOccurrenceDateParams) (sqlc.RecurringExpense, error)
	UpdateRecurringExpenseActiveStatus(ctx context.Context, params sqlc.UpdateRecurringExpenseActiveStatusParams) (sqlc.RecurringExpense, error)

	// Payment operations
	CreateRecurringExpensePayment(ctx context.Context, params sqlc.CreateRecurringExpensePaymentParams) (sqlc.RecurringExpensePayment, error)
	ListRecurringExpensePayments(ctx context.Context, recurringExpenseID pgtype.UUID) ([]sqlc.ListRecurringExpensePaymentsRow, error)

	// Split operations
	CreateRecurringExpenseSplit(ctx context.Context, params sqlc.CreateRecurringExpenseSplitParams) (sqlc.RecurringExpenseSplit, error)
	ListRecurringExpenseSplits(ctx context.Context, recurringExpenseID pgtype.UUID) ([]sqlc.ListRecurringExpenseSplitsRow, error)

	// Group lookup (for validation)
	GetGroupByID(ctx context.Context, id pgtype.UUID) (sqlc.Group, error)
	GetGroupMember(ctx context.Context, params sqlc.GetGroupMemberParams) (sqlc.GroupMember, error)
}

type recurringExpenseRepository struct {
	pool    *pgxpool.Pool
	queries *sqlc.Queries
}

func NewRecurringExpenseRepository(pool *pgxpool.Pool, queries *sqlc.Queries) RecurringExpenseRepository {
	return &recurringExpenseRepository{pool: pool, queries: queries}
}

func (r *recurringExpenseRepository) BeginTx(ctx context.Context) (pgx.Tx, error) {
	return r.pool.Begin(ctx)
}

func (r *recurringExpenseRepository) WithTx(tx pgx.Tx) RecurringExpenseRepository {
	return &recurringExpenseRepository{
		pool:    r.pool,
		queries: r.queries.WithTx(tx),
	}
}

func (r *recurringExpenseRepository) CreateRecurringExpense(ctx context.Context, params sqlc.CreateRecurringExpenseParams) (sqlc.RecurringExpense, error) {
	return r.queries.CreateRecurringExpense(ctx, params)
}

func (r *recurringExpenseRepository) GetRecurringExpenseByID(ctx context.Context, id pgtype.UUID) (sqlc.RecurringExpense, error) {
	return r.queries.GetRecurringExpenseByID(ctx, id)
}

func (r *recurringExpenseRepository) ListRecurringExpensesByGroup(ctx context.Context, groupID pgtype.UUID) ([]sqlc.RecurringExpense, error) {
	return r.queries.ListRecurringExpensesByGroup(ctx, groupID)
}

func (r *recurringExpenseRepository) UpdateRecurringExpense(ctx context.Context, params sqlc.UpdateRecurringExpenseParams) (sqlc.RecurringExpense, error) {
	return r.queries.UpdateRecurringExpense(ctx, params)
}

func (r *recurringExpenseRepository) DeleteRecurringExpense(ctx context.Context, id pgtype.UUID) error {
	return r.queries.DeleteRecurringExpense(ctx, id)
}

func (r *recurringExpenseRepository) GetRecurringExpensesDue(ctx context.Context) ([]sqlc.RecurringExpense, error) {
	return r.queries.GetRecurringExpensesDue(ctx)
}

func (r *recurringExpenseRepository) UpdateNextOccurrenceDate(ctx context.Context, params sqlc.UpdateNextOccurrenceDateParams) (sqlc.RecurringExpense, error) {
	return r.queries.UpdateNextOccurrenceDate(ctx, params)
}

func (r *recurringExpenseRepository) UpdateRecurringExpenseActiveStatus(ctx context.Context, params sqlc.UpdateRecurringExpenseActiveStatusParams) (sqlc.RecurringExpense, error) {
	return r.queries.UpdateRecurringExpenseActiveStatus(ctx, params)
}

func (r *recurringExpenseRepository) CreateRecurringExpensePayment(ctx context.Context, params sqlc.CreateRecurringExpensePaymentParams) (sqlc.RecurringExpensePayment, error) {
	return r.queries.CreateRecurringExpensePayment(ctx, params)
}

func (r *recurringExpenseRepository) ListRecurringExpensePayments(ctx context.Context, recurringExpenseID pgtype.UUID) ([]sqlc.ListRecurringExpensePaymentsRow, error) {
	return r.queries.ListRecurringExpensePayments(ctx, recurringExpenseID)
}

func (r *recurringExpenseRepository) CreateRecurringExpenseSplit(ctx context.Context, params sqlc.CreateRecurringExpenseSplitParams) (sqlc.RecurringExpenseSplit, error) {
	return r.queries.CreateRecurringExpenseSplit(ctx, params)
}

func (r *recurringExpenseRepository) ListRecurringExpenseSplits(ctx context.Context, recurringExpenseID pgtype.UUID) ([]sqlc.ListRecurringExpenseSplitsRow, error) {
	return r.queries.ListRecurringExpenseSplits(ctx, recurringExpenseID)
}

func (r *recurringExpenseRepository) GetGroupByID(ctx context.Context, id pgtype.UUID) (sqlc.Group, error) {
	return r.queries.GetGroupByID(ctx, id)
}

func (r *recurringExpenseRepository) GetGroupMember(ctx context.Context, params sqlc.GetGroupMemberParams) (sqlc.GroupMember, error) {
	return r.queries.GetGroupMember(ctx, params)
}
