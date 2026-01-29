package service

import (
	"context"
	"errors"
	"strings"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/shopspring/decimal"

	"github.com/dhruvsaxena1998/splitplus/internal/db/sqlc"
	"github.com/dhruvsaxena1998/splitplus/internal/repository"
)

type FriendSettlementService interface {
	CreateFriendSettlement(ctx context.Context, creatorID, friendID pgtype.UUID, input CreateFriendSettlementInput) (sqlc.Settlement, error)
	ListFriendSettlements(ctx context.Context, userID, friendID pgtype.UUID) ([]sqlc.ListFriendSettlementsRow, error)
	GetFriendSettlementByID(ctx context.Context, settlementID, requesterID, friendID pgtype.UUID) (sqlc.Settlement, error)
	UpdateFriendSettlementStatus(ctx context.Context, settlementID, requesterID, friendID pgtype.UUID, status string) (sqlc.Settlement, error)
}

type CreateFriendSettlementInput struct {
	PayerID              pgtype.UUID
	PayeeID              pgtype.UUID
	Amount               string
	CurrencyCode         string
	Status               string
	PaymentMethod        string
	TransactionReference string
	Notes                string
}

type friendSettlementService struct {
	settlementRepo repository.SettlementRepository
	friendRepo     repository.FriendRepository
}

func NewFriendSettlementService(settlementRepo repository.SettlementRepository, friendRepo repository.FriendRepository) FriendSettlementService {
	return &friendSettlementService{
		settlementRepo: settlementRepo,
		friendRepo:     friendRepo,
	}
}

func canonicalPairForSettlements(a, b pgtype.UUID) (pgtype.UUID, pgtype.UUID) {
	if lessOrEqualUUIDForSettlements(a, b) {
		return a, b
	}
	return b, a
}

func lessOrEqualUUIDForSettlements(a, b pgtype.UUID) bool {
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

func (s *friendSettlementService) validateFriendship(ctx context.Context, userID, friendID pgtype.UUID) error {
	a, b := canonicalPairForSettlements(userID, friendID)
	fr, err := s.friendRepo.GetFriendship(ctx, a, b)
	if err != nil {
		return ErrFriendNotFound
	}
	if fr.Status != "accepted" {
		return ErrInvalidFriendAction
	}
	return nil
}

func (s *friendSettlementService) validateStatus(status string) error {
	validStatuses := map[string]bool{
		"pending":   true,
		"completed": true,
		"cancelled": true,
	}
	if !validStatuses[status] {
		return ErrInvalidStatus
	}
	return nil
}

func (s *friendSettlementService) CreateFriendSettlement(ctx context.Context, creatorID, friendID pgtype.UUID, input CreateFriendSettlementInput) (sqlc.Settlement, error) {
	// Validate friendship
	if err := s.validateFriendship(ctx, creatorID, friendID); err != nil {
		return sqlc.Settlement{}, err
	}

	// Validate payer and payee are the two friends
	if input.PayerID != creatorID && input.PayerID != friendID {
		return sqlc.Settlement{}, errors.New("payer must be one of the two friends")
	}
	if input.PayeeID != creatorID && input.PayeeID != friendID {
		return sqlc.Settlement{}, errors.New("payee must be one of the two friends")
	}

	// Validate payer != payee
	if input.PayerID == input.PayeeID {
		return sqlc.Settlement{}, errors.New("payer and payee cannot be the same user")
	}

	// Validate amount
	amount, err := decimal.NewFromString(input.Amount)
	if err != nil || amount.LessThanOrEqual(decimal.Zero) {
		return sqlc.Settlement{}, ErrInvalidAmount
	}

	// Validate status
	status := strings.TrimSpace(input.Status)
	if status == "" {
		status = "pending"
	}
	if err := s.validateStatus(status); err != nil {
		return sqlc.Settlement{}, err
	}

	// Use default currency if not provided
	currencyCode := strings.TrimSpace(input.CurrencyCode)
	if currencyCode == "" {
		currencyCode = "USD" // Default for friend settlements
	}

	// Convert amount to numeric
	amountNumeric, err := stringToNumeric(input.Amount)
	if err != nil {
		return sqlc.Settlement{}, ErrInvalidAmount
	}

	// Create settlement with type='friend' and group_id=NULL
	settlement, err := s.settlementRepo.CreateSettlement(ctx, sqlc.CreateSettlementParams{
		GroupID:              pgtype.UUID{Valid: false}, // NULL for friend settlements
		Type:                 "friend",
		PayerID:              input.PayerID,
		PayeeID:              input.PayeeID,
		Amount:               amountNumeric,
		CurrencyCode:         currencyCode,
		Status:               status,
		PaymentMethod:        pgtype.Text{String: input.PaymentMethod, Valid: input.PaymentMethod != ""},
		TransactionReference: pgtype.Text{String: input.TransactionReference, Valid: input.TransactionReference != ""},
		Notes:                pgtype.Text{String: input.Notes, Valid: input.Notes != ""},
		CreatedBy:            creatorID,
	})
	if err != nil {
		return sqlc.Settlement{}, err
	}

	return settlement, nil
}

func (s *friendSettlementService) ListFriendSettlements(ctx context.Context, userID, friendID pgtype.UUID) ([]sqlc.ListFriendSettlementsRow, error) {
	// Validate friendship
	if err := s.validateFriendship(ctx, userID, friendID); err != nil {
		return nil, err
	}

	return s.settlementRepo.ListFriendSettlements(ctx, sqlc.ListFriendSettlementsParams{
		PayerID: userID,
		PayeeID: friendID,
	})
}

func (s *friendSettlementService) GetFriendSettlementByID(ctx context.Context, settlementID, requesterID, friendID pgtype.UUID) (sqlc.Settlement, error) {
	// Validate friendship
	if err := s.validateFriendship(ctx, requesterID, friendID); err != nil {
		return sqlc.Settlement{}, err
	}

	settlement, err := s.settlementRepo.GetSettlementByID(ctx, settlementID)
	if err != nil {
		return sqlc.Settlement{}, ErrSettlementNotFound
	}

	// Verify it's a friend settlement and involves these two users
	if settlement.Type != "friend" {
		return sqlc.Settlement{}, ErrSettlementNotFound
	}

	// Verify requester is one of the two friends
	if requesterID != settlement.PayerID && requesterID != settlement.PayeeID {
		return sqlc.Settlement{}, ErrSettlementNotFound
	}

	return settlement, nil
}

func (s *friendSettlementService) UpdateFriendSettlementStatus(ctx context.Context, settlementID, requesterID, friendID pgtype.UUID, status string) (sqlc.Settlement, error) {
	// Validate settlement access
	if _, err := s.GetFriendSettlementByID(ctx, settlementID, requesterID, friendID); err != nil {
		return sqlc.Settlement{}, err
	}

	// Validate status
	if err := s.validateStatus(status); err != nil {
		return sqlc.Settlement{}, err
	}

	return s.settlementRepo.UpdateSettlementStatus(ctx, sqlc.UpdateSettlementStatusParams{
		ID:        settlementID,
		Status:    status,
		UpdatedBy: requesterID,
	})
}
