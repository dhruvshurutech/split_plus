package repository

import (
	"context"

	"github.com/dhruvsaxena1998/splitplus/internal/db/sqlc"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type GroupRepository interface {
	// Transaction support
	BeginTx(ctx context.Context) (pgx.Tx, error)
	WithTx(tx pgx.Tx) GroupRepository

	// Group operations
	CreateGroup(ctx context.Context, params sqlc.CreateGroupParams) (sqlc.Group, error)
	GetGroupByID(ctx context.Context, id pgtype.UUID) (sqlc.Group, error)

	// Group member operations
	CreateGroupMember(ctx context.Context, params sqlc.CreateGroupMemberParams) (sqlc.GroupMember, error)
	GetGroupMember(ctx context.Context, params sqlc.GetGroupMemberParams) (sqlc.GroupMember, error)
	UpdateGroupMemberStatus(ctx context.Context, params sqlc.UpdateGroupMemberStatusParams) (sqlc.GroupMember, error)
	ListGroupMembers(ctx context.Context, groupID pgtype.UUID) ([]sqlc.ListGroupMembersRow, error)
	GetGroupsByUserID(ctx context.Context, userID pgtype.UUID) ([]sqlc.GetGroupsByUserIDRow, error)

	// User lookup (for validation)
	GetUserByID(ctx context.Context, id pgtype.UUID) (sqlc.User, error)
}

type groupRepository struct {
	pool    *pgxpool.Pool
	queries *sqlc.Queries
}

func NewGroupRepository(pool *pgxpool.Pool, queries *sqlc.Queries) GroupRepository {
	return &groupRepository{pool: pool, queries: queries}
}

func (r *groupRepository) BeginTx(ctx context.Context) (pgx.Tx, error) {
	return r.pool.Begin(ctx)
}

func (r *groupRepository) WithTx(tx pgx.Tx) GroupRepository {
	return &groupRepository{
		pool:    r.pool,
		queries: r.queries.WithTx(tx),
	}
}

func (r *groupRepository) CreateGroup(ctx context.Context, params sqlc.CreateGroupParams) (sqlc.Group, error) {
	return r.queries.CreateGroup(ctx, params)
}

func (r *groupRepository) GetGroupByID(ctx context.Context, id pgtype.UUID) (sqlc.Group, error) {
	return r.queries.GetGroupByID(ctx, id)
}

func (r *groupRepository) CreateGroupMember(ctx context.Context, params sqlc.CreateGroupMemberParams) (sqlc.GroupMember, error) {
	return r.queries.CreateGroupMember(ctx, params)
}

func (r *groupRepository) GetGroupMember(ctx context.Context, params sqlc.GetGroupMemberParams) (sqlc.GroupMember, error) {
	return r.queries.GetGroupMember(ctx, params)
}

func (r *groupRepository) UpdateGroupMemberStatus(ctx context.Context, params sqlc.UpdateGroupMemberStatusParams) (sqlc.GroupMember, error) {
	return r.queries.UpdateGroupMemberStatus(ctx, params)
}

func (r *groupRepository) ListGroupMembers(ctx context.Context, groupID pgtype.UUID) ([]sqlc.ListGroupMembersRow, error) {
	return r.queries.ListGroupMembers(ctx, groupID)
}

func (r *groupRepository) GetGroupsByUserID(ctx context.Context, userID pgtype.UUID) ([]sqlc.GetGroupsByUserIDRow, error) {
	return r.queries.GetGroupsByUserID(ctx, userID)
}

func (r *groupRepository) GetUserByID(ctx context.Context, id pgtype.UUID) (sqlc.User, error) {
	return r.queries.GetUserByID(ctx, id)
}
