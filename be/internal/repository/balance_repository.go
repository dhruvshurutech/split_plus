package repository

import (
	"context"

	"github.com/dhruvsaxena1998/splitplus/internal/db/sqlc"
	"github.com/jackc/pgx/v5/pgtype"
)

type BalanceRepository interface {
	GetGroupBalances(ctx context.Context, groupID pgtype.UUID) ([]sqlc.GetGroupBalancesRow, error)
	GetGroupBalancesWithPending(ctx context.Context, groupID pgtype.UUID) ([]sqlc.GetGroupBalancesWithPendingRow, error)
	GetUserBalanceInGroup(ctx context.Context, params sqlc.GetUserBalanceInGroupParams) (sqlc.GetUserBalanceInGroupRow, error)
	GetOverallUserBalance(ctx context.Context, userID pgtype.UUID) ([]sqlc.GetOverallUserBalanceRow, error)

	// Group lookup (for validation)
	GetGroupByID(ctx context.Context, id pgtype.UUID) (sqlc.Group, error)
	GetGroupMember(ctx context.Context, params sqlc.GetGroupMemberParams) (sqlc.GroupMember, error)
}

type balanceRepository struct {
	queries *sqlc.Queries
}

func NewBalanceRepository(queries *sqlc.Queries) BalanceRepository {
	return &balanceRepository{queries: queries}
}

func (r *balanceRepository) GetGroupBalances(ctx context.Context, groupID pgtype.UUID) ([]sqlc.GetGroupBalancesRow, error) {
	return r.queries.GetGroupBalances(ctx, groupID)
}

func (r *balanceRepository) GetGroupBalancesWithPending(ctx context.Context, groupID pgtype.UUID) ([]sqlc.GetGroupBalancesWithPendingRow, error) {
	return r.queries.GetGroupBalancesWithPending(ctx, groupID)
}

func (r *balanceRepository) GetUserBalanceInGroup(ctx context.Context, params sqlc.GetUserBalanceInGroupParams) (sqlc.GetUserBalanceInGroupRow, error) {
	return r.queries.GetUserBalanceInGroup(ctx, params)
}

func (r *balanceRepository) GetOverallUserBalance(ctx context.Context, userID pgtype.UUID) ([]sqlc.GetOverallUserBalanceRow, error) {
	return r.queries.GetOverallUserBalance(ctx, userID)
}

func (r *balanceRepository) GetGroupByID(ctx context.Context, id pgtype.UUID) (sqlc.Group, error) {
	return r.queries.GetGroupByID(ctx, id)
}

func (r *balanceRepository) GetGroupMember(ctx context.Context, params sqlc.GetGroupMemberParams) (sqlc.GroupMember, error) {
	return r.queries.GetGroupMember(ctx, params)
}
