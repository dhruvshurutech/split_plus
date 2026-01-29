package repository

import (
	"context"

	"github.com/dhruvsaxena1998/splitplus/internal/db/sqlc"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type GroupInvitationRepository interface {
	BeginTx(ctx context.Context) (pgx.Tx, error)
	WithTx(tx pgx.Tx) GroupInvitationRepository

	CreateInvitation(ctx context.Context, params sqlc.CreateInvitationParams) (sqlc.GroupInvitation, error)
	GetInvitationByToken(ctx context.Context, token string) (sqlc.GetInvitationByTokenRow, error)
	GetInvitationByID(ctx context.Context, id pgtype.UUID) (sqlc.GroupInvitation, error)
	UpdateInvitationStatus(ctx context.Context, params sqlc.UpdateInvitationStatusParams) (sqlc.GroupInvitation, error)
	ListInvitationsByGroup(ctx context.Context, groupID pgtype.UUID) ([]sqlc.GroupInvitation, error)
	GetPendingInvitationsByEmail(ctx context.Context, email string) ([]sqlc.GetPendingInvitationsByEmailRow, error)
}

type groupInvitationRepository struct {
	pool    *pgxpool.Pool
	queries *sqlc.Queries
}

func NewGroupInvitationRepository(pool *pgxpool.Pool, queries *sqlc.Queries) GroupInvitationRepository {
	return &groupInvitationRepository{pool: pool, queries: queries}
}

func (r *groupInvitationRepository) BeginTx(ctx context.Context) (pgx.Tx, error) {
	return r.pool.Begin(ctx)
}

func (r *groupInvitationRepository) WithTx(tx pgx.Tx) GroupInvitationRepository {
	return &groupInvitationRepository{
		pool:    r.pool,
		queries: r.queries.WithTx(tx),
	}
}

func (r *groupInvitationRepository) CreateInvitation(ctx context.Context, params sqlc.CreateInvitationParams) (sqlc.GroupInvitation, error) {
	return r.queries.CreateInvitation(ctx, params)
}

func (r *groupInvitationRepository) GetInvitationByToken(ctx context.Context, token string) (sqlc.GetInvitationByTokenRow, error) {
	return r.queries.GetInvitationByToken(ctx, token)
}

func (r *groupInvitationRepository) GetInvitationByID(ctx context.Context, id pgtype.UUID) (sqlc.GroupInvitation, error) {
	return r.queries.GetInvitationByID(ctx, id)
}

func (r *groupInvitationRepository) UpdateInvitationStatus(ctx context.Context, params sqlc.UpdateInvitationStatusParams) (sqlc.GroupInvitation, error) {
	return r.queries.UpdateInvitationStatus(ctx, params)
}

func (r *groupInvitationRepository) ListInvitationsByGroup(ctx context.Context, groupID pgtype.UUID) ([]sqlc.GroupInvitation, error) {
	return r.queries.ListInvitationsByGroup(ctx, groupID)
}

func (r *groupInvitationRepository) GetPendingInvitationsByEmail(ctx context.Context, email string) ([]sqlc.GetPendingInvitationsByEmailRow, error) {
	return r.queries.GetPendingInvitationsByEmail(ctx, email)
}
