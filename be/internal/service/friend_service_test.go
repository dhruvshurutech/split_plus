package service

import (
	"context"
	"errors"
	"testing"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/dhruvsaxena1998/splitplus/internal/db/sqlc"
	"github.com/dhruvsaxena1998/splitplus/internal/repository"
)

type MockFriendRepository struct {
	CreateFriendshipFunc           func(ctx context.Context, params sqlc.CreateFriendshipParams) (sqlc.Friendship, error)
	GetFriendshipFunc              func(ctx context.Context, userID, friendUserID pgtype.UUID) (sqlc.Friendship, error)
	GetFriendshipByIDFunc          func(ctx context.Context, id pgtype.UUID) (sqlc.Friendship, error)
	ListFriendsFunc                func(ctx context.Context, userID pgtype.UUID) ([]sqlc.ListFriendsRow, error)
	ListIncomingFriendRequestsFunc func(ctx context.Context, userID pgtype.UUID) ([]sqlc.Friendship, error)
	ListOutgoingFriendRequestsFunc func(ctx context.Context, userID pgtype.UUID) ([]sqlc.Friendship, error)
	UpdateFriendshipStatusFunc     func(ctx context.Context, params sqlc.UpdateFriendshipStatusParams) (sqlc.Friendship, error)
	DeleteFriendshipFunc           func(ctx context.Context, id pgtype.UUID) error
	GetUserByIDFunc                func(ctx context.Context, id pgtype.UUID) (sqlc.User, error)
}

func (m *MockFriendRepository) CreateFriendship(ctx context.Context, params sqlc.CreateFriendshipParams) (sqlc.Friendship, error) {
	if m.CreateFriendshipFunc != nil {
		return m.CreateFriendshipFunc(ctx, params)
	}
	return sqlc.Friendship{}, nil
}

func (m *MockFriendRepository) GetFriendship(ctx context.Context, userID, friendUserID pgtype.UUID) (sqlc.Friendship, error) {
	if m.GetFriendshipFunc != nil {
		return m.GetFriendshipFunc(ctx, userID, friendUserID)
	}
	return sqlc.Friendship{}, errors.New("not found")
}

func (m *MockFriendRepository) GetFriendshipByID(ctx context.Context, id pgtype.UUID) (sqlc.Friendship, error) {
	if m.GetFriendshipByIDFunc != nil {
		return m.GetFriendshipByIDFunc(ctx, id)
	}
	return sqlc.Friendship{}, errors.New("not found")
}

func (m *MockFriendRepository) ListFriends(ctx context.Context, userID pgtype.UUID) ([]sqlc.ListFriendsRow, error) {
	if m.ListFriendsFunc != nil {
		return m.ListFriendsFunc(ctx, userID)
	}
	return nil, nil
}

func (m *MockFriendRepository) ListIncomingFriendRequests(ctx context.Context, userID pgtype.UUID) ([]sqlc.Friendship, error) {
	if m.ListIncomingFriendRequestsFunc != nil {
		return m.ListIncomingFriendRequestsFunc(ctx, userID)
	}
	return nil, nil
}

func (m *MockFriendRepository) ListOutgoingFriendRequests(ctx context.Context, userID pgtype.UUID) ([]sqlc.Friendship, error) {
	if m.ListOutgoingFriendRequestsFunc != nil {
		return m.ListOutgoingFriendRequestsFunc(ctx, userID)
	}
	return nil, nil
}

func (m *MockFriendRepository) UpdateFriendshipStatus(ctx context.Context, params sqlc.UpdateFriendshipStatusParams) (sqlc.Friendship, error) {
	if m.UpdateFriendshipStatusFunc != nil {
		return m.UpdateFriendshipStatusFunc(ctx, params)
	}
	return sqlc.Friendship{}, nil
}

func (m *MockFriendRepository) DeleteFriendship(ctx context.Context, id pgtype.UUID) error {
	if m.DeleteFriendshipFunc != nil {
		return m.DeleteFriendshipFunc(ctx, id)
	}
	return nil
}

func (m *MockFriendRepository) GetUserByID(ctx context.Context, id pgtype.UUID) (sqlc.User, error) {
	if m.GetUserByIDFunc != nil {
		return m.GetUserByIDFunc(ctx, id)
	}
	return sqlc.User{}, nil
}

var _ repository.FriendRepository = (*MockFriendRepository)(nil)

func TestFriendService_SendFriendRequest_Self(t *testing.T) {
	repo := &MockFriendRepository{}
	svc := NewFriendService(repo)

	id := pgtype.UUID{}
	_ = id.Scan("00000000-0000-0000-0000-000000000001")

	_, err := svc.SendFriendRequest(context.Background(), id, id)
	if err != ErrInvalidFriendAction {
		t.Fatalf("expected ErrInvalidFriendAction, got %v", err)
	}
}
