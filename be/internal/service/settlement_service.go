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

var (
	ErrSettlementNotFound = errors.New("settlement not found")
	ErrInvalidSettlement  = errors.New("invalid settlement")
	ErrInvalidStatus      = errors.New("invalid status")
)

type CreateSettlementInput struct {
	GroupID              pgtype.UUID
	PayerID              pgtype.UUID
	PayeeID              pgtype.UUID
	Amount               string
	CurrencyCode         string
	Status               string
	PaymentMethod        string
	TransactionReference string
	Notes                string
	CreatedBy            pgtype.UUID
}

type UpdateSettlementInput struct {
	SettlementID         pgtype.UUID
	Amount               string
	CurrencyCode         string
	Status               string
	PaymentMethod        string
	TransactionReference string
	Notes                string
	UpdatedBy            pgtype.UUID
}

type UpdateSettlementStatusInput struct {
	SettlementID pgtype.UUID
	Status       string
	UpdatedBy    pgtype.UUID
}

type SettlementService interface {
	CreateSettlement(ctx context.Context, input CreateSettlementInput) (sqlc.Settlement, error)
	GetSettlementByID(ctx context.Context, settlementID, requesterID pgtype.UUID) (sqlc.Settlement, error)
	ListSettlementsByGroup(ctx context.Context, groupID, requesterID pgtype.UUID) ([]sqlc.ListSettlementsByGroupRow, error)
	ListSettlementsByUser(ctx context.Context, userID pgtype.UUID) ([]sqlc.ListSettlementsByUserRow, error)
	UpdateSettlement(ctx context.Context, input UpdateSettlementInput) (sqlc.Settlement, error)
	UpdateSettlementStatus(ctx context.Context, input UpdateSettlementStatusInput) (sqlc.Settlement, error)
	DeleteSettlement(ctx context.Context, settlementID, requesterID pgtype.UUID) error
}

type settlementService struct {
	repo            repository.SettlementRepository
	activityService GroupActivityService
}

func NewSettlementService(repo repository.SettlementRepository, activityService GroupActivityService) SettlementService {
	return &settlementService{
		repo:            repo,
		activityService: activityService,
	}
}

func (s *settlementService) validateGroupMembership(ctx context.Context, groupID, userID pgtype.UUID) error {
	_, err := s.repo.GetGroupMember(ctx, sqlc.GetGroupMemberParams{
		GroupID: groupID,
		UserID:  userID,
	})
	if err != nil {
		return ErrNotGroupMember
	}
	return nil
}

func (s *settlementService) validateStatus(status string) error {
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

func (s *settlementService) CreateSettlement(ctx context.Context, input CreateSettlementInput) (sqlc.Settlement, error) {
	// Validate group exists
	group, err := s.repo.GetGroupByID(ctx, input.GroupID)
	if err != nil {
		return sqlc.Settlement{}, ErrGroupNotFound
	}

	// Validate requester is group member
	if err := s.validateGroupMembership(ctx, input.GroupID, input.CreatedBy); err != nil {
		return sqlc.Settlement{}, err
	}

	// Validate payer and payee are group members
	if err := s.validateGroupMembership(ctx, input.GroupID, input.PayerID); err != nil {
		return sqlc.Settlement{}, errors.New("payer is not a member of this group")
	}
	if err := s.validateGroupMembership(ctx, input.GroupID, input.PayeeID); err != nil {
		return sqlc.Settlement{}, errors.New("payee is not a member of this group")
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

	// Use group currency if not provided
	currencyCode := strings.TrimSpace(input.CurrencyCode)
	if currencyCode == "" {
		currencyCode = group.CurrencyCode
	}

	// Convert amount to numeric
	amountNumeric, err := stringToNumeric(input.Amount)
	if err != nil {
		return sqlc.Settlement{}, ErrInvalidAmount
	}

	// Create settlement
	settlement, err := s.repo.CreateSettlement(ctx, sqlc.CreateSettlementParams{
		GroupID:              input.GroupID,
		Type:                 "group",
		PayerID:              input.PayerID,
		PayeeID:              input.PayeeID,
		Amount:               amountNumeric,
		CurrencyCode:         currencyCode,
		Status:               status,
		PaymentMethod:        pgtype.Text{String: input.PaymentMethod, Valid: input.PaymentMethod != ""},
		TransactionReference: pgtype.Text{String: input.TransactionReference, Valid: input.TransactionReference != ""},
		Notes:                pgtype.Text{String: input.Notes, Valid: input.Notes != ""},
		CreatedBy:            input.CreatedBy,
	})
	if err != nil {
		return sqlc.Settlement{}, err
	}

	// Log activity
	_ = s.activityService.LogActivity(ctx, LogActivityInput{
		GroupID:    input.GroupID,
		UserID:     input.CreatedBy,
		Action:     "settlement_created",
		EntityType: "settlement",
		EntityID:   settlement.ID,
		Metadata: map[string]interface{}{
			"amount":   input.Amount,
			"payer_id": input.PayerID,
			"payee_id": input.PayeeID,
		},
	})

	return settlement, nil
}

func (s *settlementService) GetSettlementByID(ctx context.Context, settlementID, requesterID pgtype.UUID) (sqlc.Settlement, error) {
	settlement, err := s.repo.GetSettlementByID(ctx, settlementID)
	if err != nil {
		return sqlc.Settlement{}, ErrSettlementNotFound
	}

	// Validate requester is group member
	if err := s.validateGroupMembership(ctx, settlement.GroupID, requesterID); err != nil {
		return sqlc.Settlement{}, err
	}

	return settlement, nil
}

func (s *settlementService) ListSettlementsByGroup(ctx context.Context, groupID, requesterID pgtype.UUID) ([]sqlc.ListSettlementsByGroupRow, error) {
	// Validate group exists
	_, err := s.repo.GetGroupByID(ctx, groupID)
	if err != nil {
		return nil, ErrGroupNotFound
	}

	// Validate requester is group member
	if err := s.validateGroupMembership(ctx, groupID, requesterID); err != nil {
		return nil, err
	}

	return s.repo.ListSettlementsByGroup(ctx, groupID)
}

func (s *settlementService) ListSettlementsByUser(ctx context.Context, userID pgtype.UUID) ([]sqlc.ListSettlementsByUserRow, error) {
	return s.repo.ListSettlementsByUser(ctx, userID)
}

func (s *settlementService) UpdateSettlement(ctx context.Context, input UpdateSettlementInput) (sqlc.Settlement, error) {
	// Get existing settlement
	settlement, err := s.repo.GetSettlementByID(ctx, input.SettlementID)
	if err != nil {
		return sqlc.Settlement{}, ErrSettlementNotFound
	}

	// Validate requester is group member
	if err := s.validateGroupMembership(ctx, settlement.GroupID, input.UpdatedBy); err != nil {
		return sqlc.Settlement{}, err
	}

	// Validate amount
	amount, err := decimal.NewFromString(input.Amount)
	if err != nil || amount.LessThanOrEqual(decimal.Zero) {
		return sqlc.Settlement{}, ErrInvalidAmount
	}

	// Validate status
	if err := s.validateStatus(input.Status); err != nil {
		return sqlc.Settlement{}, err
	}

	// Get group for currency
	group, err := s.repo.GetGroupByID(ctx, settlement.GroupID)
	if err != nil {
		return sqlc.Settlement{}, ErrGroupNotFound
	}

	// Use group currency if not provided
	currencyCode := strings.TrimSpace(input.CurrencyCode)
	if currencyCode == "" {
		currencyCode = group.CurrencyCode
	}

	// Convert amount to numeric
	amountNumeric, err := stringToNumeric(input.Amount)
	if err != nil {
		return sqlc.Settlement{}, ErrInvalidAmount
	}

	// Update settlement
	updatedSettlement, err := s.repo.UpdateSettlement(ctx, sqlc.UpdateSettlementParams{
		ID:                   input.SettlementID,
		Amount:               amountNumeric,
		CurrencyCode:         currencyCode,
		Status:               input.Status,
		PaymentMethod:        pgtype.Text{String: input.PaymentMethod, Valid: input.PaymentMethod != ""},
		TransactionReference: pgtype.Text{String: input.TransactionReference, Valid: input.TransactionReference != ""},
		Notes:                pgtype.Text{String: input.Notes, Valid: input.Notes != ""},
		UpdatedBy:            input.UpdatedBy,
	})
	if err != nil {
		return sqlc.Settlement{}, err
	}

	// Log activity
	_ = s.activityService.LogActivity(ctx, LogActivityInput{
		GroupID:    settlement.GroupID,
		UserID:     input.UpdatedBy,
		Action:     "settlement_updated",
		EntityType: "settlement",
		EntityID:   updatedSettlement.ID,
		Metadata: map[string]interface{}{
			"amount": input.Amount,
			"status": input.Status,
		},
	})

	return updatedSettlement, nil
}

func (s *settlementService) UpdateSettlementStatus(ctx context.Context, input UpdateSettlementStatusInput) (sqlc.Settlement, error) {
	// Get existing settlement
	settlement, err := s.repo.GetSettlementByID(ctx, input.SettlementID)
	if err != nil {
		return sqlc.Settlement{}, ErrSettlementNotFound
	}

	// Validate requester is group member
	if err := s.validateGroupMembership(ctx, settlement.GroupID, input.UpdatedBy); err != nil {
		return sqlc.Settlement{}, err
	}

	// Validate status
	if err := s.validateStatus(input.Status); err != nil {
		return sqlc.Settlement{}, err
	}

	// Update status
	updatedSettlement, err := s.repo.UpdateSettlementStatus(ctx, sqlc.UpdateSettlementStatusParams{
		ID:        input.SettlementID,
		Status:    input.Status,
		UpdatedBy: input.UpdatedBy,
	})
	if err != nil {
		return sqlc.Settlement{}, err
	}

	// Log activity if completed
	if input.Status == "completed" && settlement.Status != "completed" {
		_ = s.activityService.LogActivity(ctx, LogActivityInput{
			GroupID:    settlement.GroupID,
			UserID:     input.UpdatedBy,
			Action:     "settlement_completed",
			EntityType: "settlement",
			EntityID:   updatedSettlement.ID,
			Metadata: map[string]interface{}{
				"status": "completed",
			},
		})
	} else if settlement.Status != input.Status {
		// Log generic status update
		_ = s.activityService.LogActivity(ctx, LogActivityInput{
			GroupID:    settlement.GroupID,
			UserID:     input.UpdatedBy,
			Action:     "settlement_status_updated",
			EntityType: "settlement",
			EntityID:   updatedSettlement.ID,
			Metadata: map[string]interface{}{
				"old_status": settlement.Status,
				"new_status": input.Status,
			},
		})
	}

	return updatedSettlement, nil
}

func (s *settlementService) DeleteSettlement(ctx context.Context, settlementID, requesterID pgtype.UUID) error {
	settlement, err := s.repo.GetSettlementByID(ctx, settlementID)
	if err != nil {
		return ErrSettlementNotFound
	}

	// Validate requester is group member
	if err := s.validateGroupMembership(ctx, settlement.GroupID, requesterID); err != nil {
		return err
	}

	if err := s.repo.DeleteSettlement(ctx, settlementID); err != nil {
		return err
	}

	// Log activity
	_ = s.activityService.LogActivity(ctx, LogActivityInput{
		GroupID:    settlement.GroupID,
		UserID:     requesterID,
		Action:     "settlement_deleted",
		EntityType: "settlement",
		EntityID:   settlementID,
		Metadata:   nil,
	})

	return nil
}
