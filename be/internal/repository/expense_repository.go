package repository

import (
	"context"

	"github.com/dhruvsaxena1998/splitplus/internal/db/sqlc"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ExpenseRepository interface {
	// Transaction support
	BeginTx(ctx context.Context) (pgx.Tx, error)
	WithTx(tx pgx.Tx) ExpenseRepository

	// Expense operations
	CreateExpense(ctx context.Context, params sqlc.CreateExpenseParams) (sqlc.Expense, error)
	GetExpenseByID(ctx context.Context, id pgtype.UUID) (sqlc.Expense, error)
	ListExpensesByGroup(ctx context.Context, groupID pgtype.UUID) ([]sqlc.Expense, error)
	UpdateExpense(ctx context.Context, params sqlc.UpdateExpenseParams) (sqlc.Expense, error)
	DeleteExpense(ctx context.Context, id pgtype.UUID) error

	// Payment operations
	CreateExpensePayment(ctx context.Context, params sqlc.CreateExpensePaymentParams) (sqlc.ExpensePayment, error)
	ListExpensePayments(ctx context.Context, expenseID pgtype.UUID) ([]sqlc.ListExpensePaymentsRow, error)
	DeleteExpensePayments(ctx context.Context, expenseID pgtype.UUID) error

	// Split operations
	CreateExpenseSplit(ctx context.Context, params sqlc.CreateExpenseSplitParams) (sqlc.ExpenseSplit, error)
	ListExpenseSplits(ctx context.Context, expenseID pgtype.UUID) ([]sqlc.ListExpenseSplitsRow, error)
	DeleteExpenseSplits(ctx context.Context, expenseID pgtype.UUID) error

	// Friend expense operations
	ListFriendExpenses(ctx context.Context, params sqlc.ListFriendExpensesParams) ([]sqlc.Expense, error)

	// Search operations
	SearchExpenses(ctx context.Context, params sqlc.SearchExpensesParams) ([]sqlc.Expense, error)

	// Group lookup (for validation)
	GetGroupByID(ctx context.Context, id pgtype.UUID) (sqlc.Group, error)
	GetGroupMember(ctx context.Context, params sqlc.GetGroupMemberParams) (sqlc.GroupMember, error)
}

type expenseRepository struct {
	pool    *pgxpool.Pool
	queries *sqlc.Queries
}

func NewExpenseRepository(pool *pgxpool.Pool, queries *sqlc.Queries) ExpenseRepository {
	return &expenseRepository{pool: pool, queries: queries}
}

func (r *expenseRepository) BeginTx(ctx context.Context) (pgx.Tx, error) {
	return r.pool.Begin(ctx)
}

func (r *expenseRepository) WithTx(tx pgx.Tx) ExpenseRepository {
	return &expenseRepository{
		pool:    r.pool,
		queries: r.queries.WithTx(tx),
	}
}

func (r *expenseRepository) CreateExpense(ctx context.Context, params sqlc.CreateExpenseParams) (sqlc.Expense, error) {
	return r.queries.CreateExpense(ctx, params)
}

func (r *expenseRepository) GetExpenseByID(ctx context.Context, id pgtype.UUID) (sqlc.Expense, error) {
	return r.queries.GetExpenseByID(ctx, id)
}

func (r *expenseRepository) ListExpensesByGroup(ctx context.Context, groupID pgtype.UUID) ([]sqlc.Expense, error) {
	return r.queries.ListExpensesByGroup(ctx, groupID)
}

func (r *expenseRepository) UpdateExpense(ctx context.Context, params sqlc.UpdateExpenseParams) (sqlc.Expense, error) {
	return r.queries.UpdateExpense(ctx, params)
}

func (r *expenseRepository) DeleteExpense(ctx context.Context, id pgtype.UUID) error {
	return r.queries.DeleteExpense(ctx, id)
}

func (r *expenseRepository) CreateExpensePayment(ctx context.Context, params sqlc.CreateExpensePaymentParams) (sqlc.ExpensePayment, error) {
	return r.queries.CreateExpensePayment(ctx, params)
}

func (r *expenseRepository) ListExpensePayments(ctx context.Context, expenseID pgtype.UUID) ([]sqlc.ListExpensePaymentsRow, error) {
	return r.queries.ListExpensePayments(ctx, expenseID)
}

func (r *expenseRepository) DeleteExpensePayments(ctx context.Context, expenseID pgtype.UUID) error {
	return r.queries.DeleteExpensePayments(ctx, expenseID)
}

func (r *expenseRepository) CreateExpenseSplit(ctx context.Context, params sqlc.CreateExpenseSplitParams) (sqlc.ExpenseSplit, error) {
	return r.queries.CreateExpenseSplit(ctx, params)
}

func (r *expenseRepository) ListExpenseSplits(ctx context.Context, expenseID pgtype.UUID) ([]sqlc.ListExpenseSplitsRow, error) {
	return r.queries.ListExpenseSplits(ctx, expenseID)
}

func (r *expenseRepository) DeleteExpenseSplits(ctx context.Context, expenseID pgtype.UUID) error {
	return r.queries.DeleteExpenseSplits(ctx, expenseID)
}

func (r *expenseRepository) ListFriendExpenses(ctx context.Context, params sqlc.ListFriendExpensesParams) ([]sqlc.Expense, error) {
	return r.queries.ListFriendExpenses(ctx, params)
}

func (r *expenseRepository) SearchExpenses(ctx context.Context, params sqlc.SearchExpensesParams) ([]sqlc.Expense, error) {
	return r.queries.SearchExpenses(ctx, params)
}

func (r *expenseRepository) GetGroupByID(ctx context.Context, id pgtype.UUID) (sqlc.Group, error) {
	return r.queries.GetGroupByID(ctx, id)
}

func (r *expenseRepository) GetGroupMember(ctx context.Context, params sqlc.GetGroupMemberParams) (sqlc.GroupMember, error) {
	return r.queries.GetGroupMember(ctx, params)
}
