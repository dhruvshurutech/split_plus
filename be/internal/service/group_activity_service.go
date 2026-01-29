package service

import (
	"context"
	"encoding/json"

	"github.com/dhruvsaxena1998/splitplus/internal/db/sqlc"
	"github.com/dhruvsaxena1998/splitplus/internal/repository"
	"github.com/jackc/pgx/v5/pgtype"
)

type LogActivityInput struct {
	GroupID    pgtype.UUID
	UserID     pgtype.UUID
	Action     string
	EntityType string
	EntityID   pgtype.UUID
	Metadata   map[string]interface{}
}

type GroupActivityService interface {
	LogActivity(ctx context.Context, input LogActivityInput) error
	ListGroupActivities(ctx context.Context, groupID pgtype.UUID, limit, offset int32) ([]sqlc.ListGroupActivitiesRow, error)
	GetExpenseHistory(ctx context.Context, expenseID pgtype.UUID) ([]sqlc.GetExpenseHistoryRow, error)
}

type groupActivityService struct {
	repo repository.GroupActivityRepository
}

func NewGroupActivityService(repo repository.GroupActivityRepository) GroupActivityService {
	return &groupActivityService{repo: repo}
}

func (s *groupActivityService) LogActivity(ctx context.Context, input LogActivityInput) error {
	metadataJSON, err := json.Marshal(input.Metadata)
	if err != nil {
		return err
	}

	_, err = s.repo.CreateActivity(ctx, sqlc.CreateGroupActivityParams{
		GroupID:    input.GroupID,
		UserID:     input.UserID,
		Action:     input.Action,
		EntityType: input.EntityType,
		EntityID:   input.EntityID,
		Metadata:   metadataJSON,
	})

	return err
}

func (s *groupActivityService) ListGroupActivities(ctx context.Context, groupID pgtype.UUID, limit, offset int32) ([]sqlc.ListGroupActivitiesRow, error) {
	return s.repo.ListGroupActivities(ctx, sqlc.ListGroupActivitiesParams{
		GroupID: groupID,
		Limit:   limit,
		Offset:  offset,
	})
}

func (s *groupActivityService) GetExpenseHistory(ctx context.Context, expenseID pgtype.UUID) ([]sqlc.GetExpenseHistoryRow, error) {
	return s.repo.GetExpenseHistory(ctx, expenseID)
}
