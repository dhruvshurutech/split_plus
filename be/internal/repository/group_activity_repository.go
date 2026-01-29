package repository

import (
	"context"

	"github.com/dhruvsaxena1998/splitplus/internal/db/sqlc"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type GroupActivityRepository interface {
	CreateActivity(ctx context.Context, params sqlc.CreateGroupActivityParams) (sqlc.GroupActivity, error)
	ListGroupActivities(ctx context.Context, params sqlc.ListGroupActivitiesParams) ([]sqlc.ListGroupActivitiesRow, error)
	GetExpenseHistory(ctx context.Context, expenseID pgtype.UUID) ([]sqlc.GetExpenseHistoryRow, error)
}

type groupActivityRepository struct {
	queries *sqlc.Queries
	pool    *pgxpool.Pool
}

func NewGroupActivityRepository(pool *pgxpool.Pool) GroupActivityRepository {
	return &groupActivityRepository{
		queries: sqlc.New(pool),
		pool:    pool,
	}
}

func (r *groupActivityRepository) CreateActivity(ctx context.Context, params sqlc.CreateGroupActivityParams) (sqlc.GroupActivity, error) {
	return r.queries.CreateGroupActivity(ctx, params)
}

func (r *groupActivityRepository) ListGroupActivities(ctx context.Context, params sqlc.ListGroupActivitiesParams) ([]sqlc.ListGroupActivitiesRow, error) {
	return r.queries.ListGroupActivities(ctx, params)
}

func (r *groupActivityRepository) GetExpenseHistory(ctx context.Context, expenseID pgtype.UUID) ([]sqlc.GetExpenseHistoryRow, error) {
	return r.queries.GetExpenseHistory(ctx, expenseID)
}
