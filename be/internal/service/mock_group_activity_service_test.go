package service

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/dhruvsaxena1998/splitplus/internal/db/sqlc"
)

// MockGroupActivityService for testing
type MockGroupActivityService struct {
	LogActivityFunc         func(ctx context.Context, input LogActivityInput) error
	ListGroupActivitiesFunc func(ctx context.Context, groupID pgtype.UUID, limit, offset int32) ([]sqlc.ListGroupActivitiesRow, error)
	GetExpenseHistoryFunc   func(ctx context.Context, expenseID pgtype.UUID) ([]sqlc.GetExpenseHistoryRow, error)
}

func (m *MockGroupActivityService) LogActivity(ctx context.Context, input LogActivityInput) error {
	if m.LogActivityFunc != nil {
		return m.LogActivityFunc(ctx, input)
	}
	return nil
}

func (m *MockGroupActivityService) ListGroupActivities(ctx context.Context, groupID pgtype.UUID, limit, offset int32) ([]sqlc.ListGroupActivitiesRow, error) {
	if m.ListGroupActivitiesFunc != nil {
		return m.ListGroupActivitiesFunc(ctx, groupID, limit, offset)
	}
	return []sqlc.ListGroupActivitiesRow{}, nil
}

func (m *MockGroupActivityService) GetExpenseHistory(ctx context.Context, expenseID pgtype.UUID) ([]sqlc.GetExpenseHistoryRow, error) {
	if m.GetExpenseHistoryFunc != nil {
		return m.GetExpenseHistoryFunc(ctx, expenseID)
	}
	return []sqlc.GetExpenseHistoryRow{}, nil
}

var _ GroupActivityService = (*MockGroupActivityService)(nil)
