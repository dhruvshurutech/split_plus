package repository

import (
	"context"

	"github.com/dhruvsaxena1998/splitplus/internal/db/sqlc"
	"github.com/jackc/pgx/v5/pgtype"
)

type SettlementRepository interface {
	CreateSettlement(ctx context.Context, params sqlc.CreateSettlementParams) (sqlc.Settlement, error)
	GetSettlementByID(ctx context.Context, id pgtype.UUID) (sqlc.Settlement, error)
	ListSettlementsByGroup(ctx context.Context, groupID pgtype.UUID) ([]sqlc.ListSettlementsByGroupRow, error)
	ListSettlementsByUser(ctx context.Context, userID pgtype.UUID) ([]sqlc.ListSettlementsByUserRow, error)
	UpdateSettlementStatus(ctx context.Context, params sqlc.UpdateSettlementStatusParams) (sqlc.Settlement, error)
	UpdateSettlement(ctx context.Context, params sqlc.UpdateSettlementParams) (sqlc.Settlement, error)
	DeleteSettlement(ctx context.Context, id pgtype.UUID) error

	// Friend settlement operations
	ListFriendSettlements(ctx context.Context, params sqlc.ListFriendSettlementsParams) ([]sqlc.ListFriendSettlementsRow, error)

	// Group lookup (for validation)
	GetGroupByID(ctx context.Context, id pgtype.UUID) (sqlc.Group, error)
	GetGroupMember(ctx context.Context, params sqlc.GetGroupMemberParams) (sqlc.GroupMember, error)
}

type settlementRepository struct {
	queries *sqlc.Queries
}

func NewSettlementRepository(queries *sqlc.Queries) SettlementRepository {
	return &settlementRepository{queries: queries}
}

func (r *settlementRepository) CreateSettlement(ctx context.Context, params sqlc.CreateSettlementParams) (sqlc.Settlement, error) {
	return r.queries.CreateSettlement(ctx, params)
}

func (r *settlementRepository) GetSettlementByID(ctx context.Context, id pgtype.UUID) (sqlc.Settlement, error) {
	return r.queries.GetSettlementByID(ctx, id)
}

func (r *settlementRepository) ListSettlementsByGroup(ctx context.Context, groupID pgtype.UUID) ([]sqlc.ListSettlementsByGroupRow, error) {
	return r.queries.ListSettlementsByGroup(ctx, groupID)
}

func (r *settlementRepository) ListSettlementsByUser(ctx context.Context, userID pgtype.UUID) ([]sqlc.ListSettlementsByUserRow, error) {
	return r.queries.ListSettlementsByUser(ctx, userID)
}

func (r *settlementRepository) UpdateSettlementStatus(ctx context.Context, params sqlc.UpdateSettlementStatusParams) (sqlc.Settlement, error) {
	return r.queries.UpdateSettlementStatus(ctx, params)
}

func (r *settlementRepository) UpdateSettlement(ctx context.Context, params sqlc.UpdateSettlementParams) (sqlc.Settlement, error) {
	return r.queries.UpdateSettlement(ctx, params)
}

func (r *settlementRepository) DeleteSettlement(ctx context.Context, id pgtype.UUID) error {
	return r.queries.DeleteSettlement(ctx, id)
}

func (r *settlementRepository) ListFriendSettlements(ctx context.Context, params sqlc.ListFriendSettlementsParams) ([]sqlc.ListFriendSettlementsRow, error) {
	return r.queries.ListFriendSettlements(ctx, params)
}

func (r *settlementRepository) GetGroupByID(ctx context.Context, id pgtype.UUID) (sqlc.Group, error) {
	return r.queries.GetGroupByID(ctx, id)
}

func (r *settlementRepository) GetGroupMember(ctx context.Context, params sqlc.GetGroupMemberParams) (sqlc.GroupMember, error) {
	return r.queries.GetGroupMember(ctx, params)
}
