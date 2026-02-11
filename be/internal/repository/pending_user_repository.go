package repository

import (
	"context"

	"github.com/dhruvsaxena1998/splitplus/internal/db/sqlc"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PendingUserRepository interface {
	BeginTx(ctx context.Context) (pgx.Tx, error)
	WithTx(tx pgx.Tx) PendingUserRepository

	CreatePendingUser(ctx context.Context, params sqlc.CreatePendingUserParams) (sqlc.PendingUser, error)
	GetPendingUserByEmail(ctx context.Context, email string) (sqlc.PendingUser, error)
	GetPendingUserByID(ctx context.Context, id pgtype.UUID) (sqlc.PendingUser, error)
	UpdatePendingPaymentUserID(ctx context.Context, params sqlc.UpdatePendingPaymentUserIDParams) error
	UpdatePendingSplitUserID(ctx context.Context, params sqlc.UpdatePendingSplitUserIDParams) error
	UpdatePendingSettlementPayerUserID(ctx context.Context, params sqlc.UpdatePendingSettlementPayerUserIDParams) error
	UpdatePendingSettlementPayeeUserID(ctx context.Context, params sqlc.UpdatePendingSettlementPayeeUserIDParams) error
	DeletePendingUserByID(ctx context.Context, id pgtype.UUID) error
}

type pendingUserRepository struct {
	pool    *pgxpool.Pool
	queries *sqlc.Queries
}

func NewPendingUserRepository(pool *pgxpool.Pool, queries *sqlc.Queries) PendingUserRepository {
	return &pendingUserRepository{pool: pool, queries: queries}
}

func (r *pendingUserRepository) BeginTx(ctx context.Context) (pgx.Tx, error) {
	return r.pool.Begin(ctx)
}

func (r *pendingUserRepository) WithTx(tx pgx.Tx) PendingUserRepository {
	return &pendingUserRepository{
		pool:    r.pool,
		queries: r.queries.WithTx(tx),
	}
}

func (r *pendingUserRepository) CreatePendingUser(ctx context.Context, params sqlc.CreatePendingUserParams) (sqlc.PendingUser, error) {
	return r.queries.CreatePendingUser(ctx, params)
}

func (r *pendingUserRepository) GetPendingUserByEmail(ctx context.Context, email string) (sqlc.PendingUser, error) {
	return r.queries.GetPendingUserByEmail(ctx, email)
}

func (r *pendingUserRepository) GetPendingUserByID(ctx context.Context, id pgtype.UUID) (sqlc.PendingUser, error) {
	return r.queries.GetPendingUserByID(ctx, id)
}

func (r *pendingUserRepository) UpdatePendingPaymentUserID(ctx context.Context, params sqlc.UpdatePendingPaymentUserIDParams) error {
	return r.queries.UpdatePendingPaymentUserID(ctx, params)
}

func (r *pendingUserRepository) UpdatePendingSplitUserID(ctx context.Context, params sqlc.UpdatePendingSplitUserIDParams) error {
	return r.queries.UpdatePendingSplitUserID(ctx, params)
}

func (r *pendingUserRepository) UpdatePendingSettlementPayerUserID(ctx context.Context, params sqlc.UpdatePendingSettlementPayerUserIDParams) error {
	return r.queries.UpdatePendingSettlementPayerUserID(ctx, params)
}

func (r *pendingUserRepository) UpdatePendingSettlementPayeeUserID(ctx context.Context, params sqlc.UpdatePendingSettlementPayeeUserIDParams) error {
	return r.queries.UpdatePendingSettlementPayeeUserID(ctx, params)
}

func (r *pendingUserRepository) DeletePendingUserByID(ctx context.Context, id pgtype.UUID) error {
	return r.queries.DeletePendingUserByID(ctx, id)
}
