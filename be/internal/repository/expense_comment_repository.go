package repository

import (
	"context"

	"github.com/dhruvsaxena1998/splitplus/internal/db/sqlc"
	"github.com/jackc/pgx/v5/pgtype"
)

type ExpenseCommentRepository interface {
	CreateComment(ctx context.Context, params sqlc.CreateExpenseCommentParams) (sqlc.ExpenseComment, error)
	GetCommentByID(ctx context.Context, id pgtype.UUID) (sqlc.GetExpenseCommentByIDRow, error)
	ListComments(ctx context.Context, expenseID pgtype.UUID) ([]sqlc.ListExpenseCommentsRow, error)
	UpdateComment(ctx context.Context, params sqlc.UpdateExpenseCommentParams) (sqlc.ExpenseComment, error)
	DeleteComment(ctx context.Context, id pgtype.UUID) error
	CountComments(ctx context.Context, expenseID pgtype.UUID) (int64, error)
}

type expenseCommentRepository struct {
	q *sqlc.Queries
}

func NewExpenseCommentRepository(q *sqlc.Queries) ExpenseCommentRepository {
	return &expenseCommentRepository{q: q}
}

func (r *expenseCommentRepository) CreateComment(ctx context.Context, params sqlc.CreateExpenseCommentParams) (sqlc.ExpenseComment, error) {
	return r.q.CreateExpenseComment(ctx, params)
}

func (r *expenseCommentRepository) GetCommentByID(ctx context.Context, id pgtype.UUID) (sqlc.GetExpenseCommentByIDRow, error) {
	return r.q.GetExpenseCommentByID(ctx, id)
}

func (r *expenseCommentRepository) ListComments(ctx context.Context, expenseID pgtype.UUID) ([]sqlc.ListExpenseCommentsRow, error) {
	return r.q.ListExpenseComments(ctx, expenseID)
}

func (r *expenseCommentRepository) UpdateComment(ctx context.Context, params sqlc.UpdateExpenseCommentParams) (sqlc.ExpenseComment, error) {
	return r.q.UpdateExpenseComment(ctx, params)
}

func (r *expenseCommentRepository) DeleteComment(ctx context.Context, id pgtype.UUID) error {
	return r.q.DeleteExpenseComment(ctx, id)
}

func (r *expenseCommentRepository) CountComments(ctx context.Context, expenseID pgtype.UUID) (int64, error) {
	return r.q.CountExpenseComments(ctx, expenseID)
}
