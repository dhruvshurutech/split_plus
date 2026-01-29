package service

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/dhruvsaxena1998/splitplus/internal/db/sqlc"
	"github.com/dhruvsaxena1998/splitplus/internal/repository"
)

var (
	ErrFriendRequestExists = errors.New("friend request already exists or users are already friends")
	ErrFriendNotFound      = errors.New("friendship not found")
	ErrInvalidFriendAction = errors.New("invalid friend action")
)

type FriendSummary struct {
	ID          pgtype.UUID `json:"id"`
	UserID      pgtype.UUID `json:"user_id"`
	FriendID    pgtype.UUID `json:"friend_id"`
	FriendEmail string      `json:"friend_email"`
	FriendName  string      `json:"friend_name,omitempty"`
	AvatarURL   string      `json:"avatar_url,omitempty"`
}

type FriendService interface {
	SendFriendRequest(ctx context.Context, requesterID, friendID pgtype.UUID) (sqlc.Friendship, error)
	ListFriends(ctx context.Context, userID pgtype.UUID) ([]FriendSummary, error)
	ListIncomingRequests(ctx context.Context, userID pgtype.UUID) ([]sqlc.Friendship, error)
	ListOutgoingRequests(ctx context.Context, userID pgtype.UUID) ([]sqlc.Friendship, error)
	AcceptFriendRequest(ctx context.Context, requesterID, friendshipID pgtype.UUID) (sqlc.Friendship, error)
	DeclineFriendRequest(ctx context.Context, requesterID, friendshipID pgtype.UUID) error
	RemoveFriend(ctx context.Context, requesterID, friendID pgtype.UUID) error
}

type friendService struct {
	repo repository.FriendRepository
}

func NewFriendService(repo repository.FriendRepository) FriendService {
	return &friendService{repo: repo}
}

// canonicalPair ensures (a,b) is always stored with a < b for symmetry.
func canonicalPair(a, b pgtype.UUID) (pgtype.UUID, pgtype.UUID) {
	if lessOrEqualUUID(a, b) {
		return a, b
	}
	return b, a
}

func lessOrEqualUUID(a, b pgtype.UUID) bool {
	for i := 0; i < len(a.Bytes); i++ {
		if a.Bytes[i] < b.Bytes[i] {
			return true
		}
		if a.Bytes[i] > b.Bytes[i] {
			return false
		}
	}
	return true
}

func (s *friendService) SendFriendRequest(ctx context.Context, requesterID, friendID pgtype.UUID) (sqlc.Friendship, error) {
	if requesterID == friendID {
		return sqlc.Friendship{}, ErrInvalidFriendAction
	}

	// ensure friend exists
	if _, err := s.repo.GetUserByID(ctx, friendID); err != nil {
		return sqlc.Friendship{}, err
	}

	a, b := canonicalPair(requesterID, friendID)

	// check existing friendship
	if existing, err := s.repo.GetFriendship(ctx, a, b); err == nil && !existing.DeletedAt.Valid {
		return sqlc.Friendship{}, ErrFriendRequestExists
	}

	params := sqlc.CreateFriendshipParams{
		UserID:       a,
		FriendUserID: b,
		Status:       "pending",
	}
	return s.repo.CreateFriendship(ctx, params)
}

func (s *friendService) ListFriends(ctx context.Context, userID pgtype.UUID) ([]FriendSummary, error) {
	rows, err := s.repo.ListFriends(ctx, userID)
	if err != nil {
		return nil, err
	}
	res := make([]FriendSummary, 0, len(rows))
	for _, r := range rows {
		fs := FriendSummary{
			ID:          r.ID,
			UserID:      userID,
			FriendID:    r.FriendID,
			FriendEmail: r.FriendEmail,
			FriendName:  r.FriendName.String,
			AvatarURL:   r.FriendAvatarUrl.String,
		}
		res = append(res, fs)
	}
	return res, nil
}

func (s *friendService) ListIncomingRequests(ctx context.Context, userID pgtype.UUID) ([]sqlc.Friendship, error) {
	return s.repo.ListIncomingFriendRequests(ctx, userID)
}

func (s *friendService) ListOutgoingRequests(ctx context.Context, userID pgtype.UUID) ([]sqlc.Friendship, error) {
	return s.repo.ListOutgoingFriendRequests(ctx, userID)
}

func (s *friendService) AcceptFriendRequest(ctx context.Context, requesterID, friendshipID pgtype.UUID) (sqlc.Friendship, error) {
	fr, err := s.repo.GetFriendshipByID(ctx, friendshipID)
	if err != nil {
		return sqlc.Friendship{}, ErrFriendNotFound
	}
	// requester must be one of the two users in the friendship
	if requesterID != fr.UserID && requesterID != fr.FriendUserID {
		return sqlc.Friendship{}, ErrInvalidFriendAction
	}
	if fr.Status != "pending" {
		return sqlc.Friendship{}, ErrInvalidFriendAction
	}
	return s.repo.UpdateFriendshipStatus(ctx, sqlc.UpdateFriendshipStatusParams{
		ID:     fr.ID,
		Status: "accepted",
	})
}

func (s *friendService) DeclineFriendRequest(ctx context.Context, requesterID, friendshipID pgtype.UUID) error {
	fr, err := s.repo.GetFriendshipByID(ctx, friendshipID)
	if err != nil {
		return ErrFriendNotFound
	}
	// requester must be one of the two users in the friendship
	if requesterID != fr.UserID && requesterID != fr.FriendUserID {
		return ErrInvalidFriendAction
	}
	if fr.Status != "pending" {
		return ErrInvalidFriendAction
	}
	return s.repo.DeleteFriendship(ctx, friendshipID)
}

func (s *friendService) RemoveFriend(ctx context.Context, requesterID, friendID pgtype.UUID) error {
	a, b := canonicalPair(requesterID, friendID)
	fr, err := s.repo.GetFriendship(ctx, a, b)
	if err != nil {
		return ErrFriendNotFound
	}
	if fr.Status != "accepted" {
		return ErrInvalidFriendAction
	}
	return s.repo.DeleteFriendship(ctx, fr.ID)
}
