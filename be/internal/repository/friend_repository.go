package repository

import (
	"context"

	"github.com/dhruvsaxena1998/splitplus/internal/db/sqlc"
	"github.com/jackc/pgx/v5/pgtype"
)

type FriendRepository interface {
	CreateFriendship(ctx context.Context, params sqlc.CreateFriendshipParams) (sqlc.Friendship, error)
	GetFriendship(ctx context.Context, userID, friendUserID pgtype.UUID) (sqlc.Friendship, error)
	GetFriendshipByID(ctx context.Context, id pgtype.UUID) (sqlc.Friendship, error)
	ListFriends(ctx context.Context, userID pgtype.UUID) ([]sqlc.ListFriendsRow, error)
	ListIncomingFriendRequests(ctx context.Context, userID pgtype.UUID) ([]sqlc.Friendship, error)
	ListOutgoingFriendRequests(ctx context.Context, userID pgtype.UUID) ([]sqlc.Friendship, error)
	UpdateFriendshipStatus(ctx context.Context, params sqlc.UpdateFriendshipStatusParams) (sqlc.Friendship, error)
	DeleteFriendship(ctx context.Context, id pgtype.UUID) error

	GetUserByID(ctx context.Context, id pgtype.UUID) (sqlc.User, error)
}

type friendRepository struct {
	queries *sqlc.Queries
}

func NewFriendRepository(queries *sqlc.Queries) FriendRepository {
	return &friendRepository{queries: queries}
}

func (r *friendRepository) CreateFriendship(ctx context.Context, params sqlc.CreateFriendshipParams) (sqlc.Friendship, error) {
	return r.queries.CreateFriendship(ctx, params)
}

func (r *friendRepository) GetFriendship(ctx context.Context, userID, friendUserID pgtype.UUID) (sqlc.Friendship, error) {
	return r.queries.GetFriendship(ctx, sqlc.GetFriendshipParams{
		UserID:       userID,
		FriendUserID: friendUserID,
	})
}

func (r *friendRepository) GetFriendshipByID(ctx context.Context, id pgtype.UUID) (sqlc.Friendship, error) {
	return r.queries.GetFriendshipByID(ctx, id)
}

func (r *friendRepository) ListFriends(ctx context.Context, userID pgtype.UUID) ([]sqlc.ListFriendsRow, error) {
	return r.queries.ListFriends(ctx, userID)
}

func (r *friendRepository) ListIncomingFriendRequests(ctx context.Context, userID pgtype.UUID) ([]sqlc.Friendship, error) {
	return r.queries.ListIncomingFriendRequests(ctx, userID)
}

func (r *friendRepository) ListOutgoingFriendRequests(ctx context.Context, userID pgtype.UUID) ([]sqlc.Friendship, error) {
	return r.queries.ListOutgoingFriendRequests(ctx, userID)
}

func (r *friendRepository) UpdateFriendshipStatus(ctx context.Context, params sqlc.UpdateFriendshipStatusParams) (sqlc.Friendship, error) {
	return r.queries.UpdateFriendshipStatus(ctx, params)
}

func (r *friendRepository) DeleteFriendship(ctx context.Context, id pgtype.UUID) error {
	return r.queries.DeleteFriendship(ctx, id)
}

func (r *friendRepository) GetUserByID(ctx context.Context, id pgtype.UUID) (sqlc.User, error) {
	return r.queries.GetUserByID(ctx, id)
}
