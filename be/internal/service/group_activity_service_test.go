package service

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/dhruvsaxena1998/splitplus/internal/db/sqlc"
	"github.com/dhruvsaxena1998/splitplus/internal/testutil"
	"github.com/jackc/pgx/v5/pgtype"
)

// MockGroupActivityRepository for testing
type MockGroupActivityRepository struct {
	CreateActivityFunc      func(ctx context.Context, params sqlc.CreateGroupActivityParams) (sqlc.GroupActivity, error)
	ListGroupActivitiesFunc func(ctx context.Context, params sqlc.ListGroupActivitiesParams) ([]sqlc.ListGroupActivitiesRow, error)
	GetExpenseHistoryFunc   func(ctx context.Context, expenseID pgtype.UUID) ([]sqlc.GetExpenseHistoryRow, error)
}

func (m *MockGroupActivityRepository) CreateActivity(ctx context.Context, params sqlc.CreateGroupActivityParams) (sqlc.GroupActivity, error) {
	if m.CreateActivityFunc != nil {
		return m.CreateActivityFunc(ctx, params)
	}
	return sqlc.GroupActivity{}, nil
}

func (m *MockGroupActivityRepository) ListGroupActivities(ctx context.Context, params sqlc.ListGroupActivitiesParams) ([]sqlc.ListGroupActivitiesRow, error) {
	if m.ListGroupActivitiesFunc != nil {
		return m.ListGroupActivitiesFunc(ctx, params)
	}
	return []sqlc.ListGroupActivitiesRow{}, nil
}

func (m *MockGroupActivityRepository) GetExpenseHistory(ctx context.Context, expenseID pgtype.UUID) ([]sqlc.GetExpenseHistoryRow, error) {
	if m.GetExpenseHistoryFunc != nil {
		return m.GetExpenseHistoryFunc(ctx, expenseID)
	}
	return []sqlc.GetExpenseHistoryRow{}, nil
}

func TestGroupActivityService_LogActivity(t *testing.T) {
	tests := []struct {
		name          string
		input         LogActivityInput
		mockSetup     func(*MockGroupActivityRepository)
		expectedError error
		validate      func(*testing.T, LogActivityInput)
	}{
		{
			name: "success - log activity with metadata",
			input: LogActivityInput{
				GroupID:    testutil.CreateTestUUID(1),
				UserID:     testutil.CreateTestUUID(10),
				Action:     "expense_created",
				EntityType: "expense",
				EntityID:   testutil.CreateTestUUID(100),
				Metadata: map[string]interface{}{
					"amount": "25.50",
					"title":  "Lunch",
				},
			},
			mockSetup: func(repo *MockGroupActivityRepository) {
				repo.CreateActivityFunc = func(ctx context.Context, params sqlc.CreateGroupActivityParams) (sqlc.GroupActivity, error) {
					return sqlc.GroupActivity{
						ID:         testutil.CreateTestUUID(1000),
						GroupID:    params.GroupID,
						UserID:     params.UserID,
						Action:     params.Action,
						EntityType: params.EntityType,
						EntityID:   params.EntityID,
						Metadata:   params.Metadata,
					}, nil
				}
			},
			expectedError: nil,
			validate: func(t *testing.T, input LogActivityInput) {
				if input.Action != "expense_created" {
					t.Errorf("expected action 'expense_created', got '%s'", input.Action)
				}
			},
		},
		{
			name: "success - log activity with empty metadata",
			input: LogActivityInput{
				GroupID:    testutil.CreateTestUUID(1),
				UserID:     testutil.CreateTestUUID(10),
				Action:     "member_joined",
				EntityType: "group",
				EntityID:   testutil.CreateTestUUID(1),
				Metadata:   map[string]interface{}{},
			},
			mockSetup: func(repo *MockGroupActivityRepository) {
				repo.CreateActivityFunc = func(ctx context.Context, params sqlc.CreateGroupActivityParams) (sqlc.GroupActivity, error) {
					return sqlc.GroupActivity{
						ID:     testutil.CreateTestUUID(1000),
						Action: params.Action,
					}, nil
				}
			},
			expectedError: nil,
		},
		{
			name: "success - log activity with nil metadata",
			input: LogActivityInput{
				GroupID:    testutil.CreateTestUUID(1),
				UserID:     testutil.CreateTestUUID(10),
				Action:     "settlement_created",
				EntityType: "settlement",
				EntityID:   testutil.CreateTestUUID(200),
				Metadata:   nil,
			},
			mockSetup: func(repo *MockGroupActivityRepository) {
				repo.CreateActivityFunc = func(ctx context.Context, params sqlc.CreateGroupActivityParams) (sqlc.GroupActivity, error) {
					// Verify nil metadata marshals to "null"
					var m interface{}
					if err := json.Unmarshal(params.Metadata, &m); err != nil {
						t.Errorf("metadata should be valid JSON: %v", err)
					}
					return sqlc.GroupActivity{ID: testutil.CreateTestUUID(1000)}, nil
				}
			},
			expectedError: nil,
		},
		{
			name: "success - log activity with complex metadata",
			input: LogActivityInput{
				GroupID:    testutil.CreateTestUUID(1),
				UserID:     testutil.CreateTestUUID(10),
				Action:     "expense_updated",
				EntityType: "expense",
				EntityID:   testutil.CreateTestUUID(100),
				Metadata: map[string]interface{}{
					"previous_amount": "20.00",
					"new_amount":      "25.50",
					"changes": map[string]interface{}{
						"title": "Updated title",
					},
				},
			},
			mockSetup: func(repo *MockGroupActivityRepository) {
				repo.CreateActivityFunc = func(ctx context.Context, params sqlc.CreateGroupActivityParams) (sqlc.GroupActivity, error) {
					// Verify complex metadata can be marshaled
					var m map[string]interface{}
					if err := json.Unmarshal(params.Metadata, &m); err != nil {
						return sqlc.GroupActivity{}, err
					}
					if changes, ok := m["changes"].(map[string]interface{}); !ok || changes["title"] != "Updated title" {
						t.Error("nested metadata not preserved")
					}
					return sqlc.GroupActivity{ID: testutil.CreateTestUUID(1000)}, nil
				}
			},
			expectedError: nil,
		},
		{
			name: "error - repository fails",
			input: LogActivityInput{
				GroupID:    testutil.CreateTestUUID(1),
				UserID:     testutil.CreateTestUUID(10),
				Action:     "test_action",
				EntityType: "test",
				EntityID:   testutil.CreateTestUUID(100),
				Metadata:   map[string]interface{}{},
			},
			mockSetup: func(repo *MockGroupActivityRepository) {
				repo.CreateActivityFunc = func(ctx context.Context, params sqlc.CreateGroupActivityParams) (sqlc.GroupActivity, error) {
					return sqlc.GroupActivity{}, errors.New("database error")
				}
			},
			expectedError: errors.New("database error"),
		},
		{
			name: "error - invalid metadata (circular reference would fail in real scenario)",
			input: LogActivityInput{
				GroupID:    testutil.CreateTestUUID(1),
				UserID:     testutil.CreateTestUUID(10),
				Action:     "test",
				EntityType: "test",
				EntityID:   testutil.CreateTestUUID(100),
				Metadata: map[string]interface{}{
					"channel": make(chan int), // channels can't be marshaled to JSON
				},
			},
			mockSetup:     func(repo *MockGroupActivityRepository) {},
			expectedError: errors.New("json: unsupported type: chan int"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &MockGroupActivityRepository{}
			if tt.mockSetup != nil {
				tt.mockSetup(repo)
			}

			service := NewGroupActivityService(repo)
			err := service.LogActivity(context.Background(), tt.input)

			if tt.expectedError != nil {
				if err == nil {
					t.Fatalf("expected error %v, got nil", tt.expectedError)
				}
				if err.Error() != tt.expectedError.Error() {
					t.Errorf("expected error '%v', got '%v'", tt.expectedError, err)
				}
			} else {
				if err != nil {
					t.Fatalf("expected no error, got %v", err)
				}
				if tt.validate != nil {
					tt.validate(t, tt.input)
				}
			}
		})
	}
}

func TestGroupActivityService_ListGroupActivities(t *testing.T) {
	tests := []struct {
		name          string
		groupID       pgtype.UUID
		limit         int32
		offset        int32
		mockSetup     func(*MockGroupActivityRepository)
		expectedError error
		validate      func(*testing.T, []sqlc.ListGroupActivitiesRow)
	}{
		{
			name:    "success - list activities",
			groupID: testutil.CreateTestUUID(1),
			limit:   10,
			offset:  0,
			mockSetup: func(repo *MockGroupActivityRepository) {
				repo.ListGroupActivitiesFunc = func(ctx context.Context, params sqlc.ListGroupActivitiesParams) ([]sqlc.ListGroupActivitiesRow, error) {
					return []sqlc.ListGroupActivitiesRow{
						{
							ID:     testutil.CreateTestUUID(100),
							Action: "expense_created",
						},
						{
							ID:     testutil.CreateTestUUID(101),
							Action: "comment_added",
						},
						{
							ID:     testutil.CreateTestUUID(102),
							Action: "settlement_created",
						},
					}, nil
				}
			},
			expectedError: nil,
			validate: func(t *testing.T, activities []sqlc.ListGroupActivitiesRow) {
				if len(activities) != 3 {
					t.Errorf("expected 3 activities, got %d", len(activities))
				}
				if activities[0].Action != "expense_created" {
					t.Errorf("expected first action to be 'expense_created', got '%s'", activities[0].Action)
				}
			},
		},
		{
			name:    "success - empty list",
			groupID: testutil.CreateTestUUID(1),
			limit:   10,
			offset:  0,
			mockSetup: func(repo *MockGroupActivityRepository) {
				repo.ListGroupActivitiesFunc = func(ctx context.Context, params sqlc.ListGroupActivitiesParams) ([]sqlc.ListGroupActivitiesRow, error) {
					return []sqlc.ListGroupActivitiesRow{}, nil
				}
			},
			expectedError: nil,
			validate: func(t *testing.T, activities []sqlc.ListGroupActivitiesRow) {
				if len(activities) != 0 {
					t.Errorf("expected 0 activities, got %d", len(activities))
				}
			},
		},
		{
			name:    "success - pagination",
			groupID: testutil.CreateTestUUID(1),
			limit:   5,
			offset:  10,
			mockSetup: func(repo *MockGroupActivityRepository) {
				repo.ListGroupActivitiesFunc = func(ctx context.Context, params sqlc.ListGroupActivitiesParams) ([]sqlc.ListGroupActivitiesRow, error) {
					// Verify pagination params are passed correctly
					if params.Limit != 5 {
						t.Errorf("expected limit 5, got %d", params.Limit)
					}
					if params.Offset != 10 {
						t.Errorf("expected offset 10, got %d", params.Offset)
					}
					return []sqlc.ListGroupActivitiesRow{
						{ID: testutil.CreateTestUUID(110)},
						{ID: testutil.CreateTestUUID(111)},
					}, nil
				}
			},
			expectedError: nil,
			validate: func(t *testing.T, activities []sqlc.ListGroupActivitiesRow) {
				if len(activities) != 2 {
					t.Errorf("expected 2 activities, got %d", len(activities))
				}
			},
		},
		{
			name:    "error - repository fails",
			groupID: testutil.CreateTestUUID(1),
			limit:   10,
			offset:  0,
			mockSetup: func(repo *MockGroupActivityRepository) {
				repo.ListGroupActivitiesFunc = func(ctx context.Context, params sqlc.ListGroupActivitiesParams) ([]sqlc.ListGroupActivitiesRow, error) {
					return nil, errors.New("database error")
				}
			},
			expectedError: errors.New("database error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &MockGroupActivityRepository{}
			if tt.mockSetup != nil {
				tt.mockSetup(repo)
			}

			service := NewGroupActivityService(repo)
			result, err := service.ListGroupActivities(context.Background(), tt.groupID, tt.limit, tt.offset)

			if tt.expectedError != nil {
				if err == nil {
					t.Fatalf("expected error %v, got nil", tt.expectedError)
				}
				if err.Error() != tt.expectedError.Error() {
					t.Errorf("expected error '%v', got '%v'", tt.expectedError, err)
				}
			} else {
				if err != nil {
					t.Fatalf("expected no error, got %v", err)
				}
				if tt.validate != nil {
					tt.validate(t, result)
				}
			}
		})
	}
}

func TestGroupActivityService_GetExpenseHistory(t *testing.T) {
	tests := []struct {
		name          string
		expenseID     pgtype.UUID
		mockSetup     func(*MockGroupActivityRepository)
		expectedError error
		validate      func(*testing.T, []sqlc.GetExpenseHistoryRow)
	}{
		{
			name:      "success - get expense history",
			expenseID: testutil.CreateTestUUID(100),
			mockSetup: func(repo *MockGroupActivityRepository) {
				repo.GetExpenseHistoryFunc = func(ctx context.Context, expenseID pgtype.UUID) ([]sqlc.GetExpenseHistoryRow, error) {
					return []sqlc.GetExpenseHistoryRow{
						{
							ID:     testutil.CreateTestUUID(200),
							Action: "expense_created",
						},
						{
							ID:     testutil.CreateTestUUID(201),
							Action: "expense_updated",
						},
						{
							ID:     testutil.CreateTestUUID(202),
							Action: "comment_added",
						},
					}, nil
				}
			},
			expectedError: nil,
			validate: func(t *testing.T, history []sqlc.GetExpenseHistoryRow) {
				if len(history) != 3 {
					t.Errorf("expected 3 history items, got %d", len(history))
				}
				if history[0].Action != "expense_created" {
					t.Errorf("expected first action to be 'expense_created', got '%s'", history[0].Action)
				}
			},
		},
		{
			name:      "success - empty history",
			expenseID: testutil.CreateTestUUID(100),
			mockSetup: func(repo *MockGroupActivityRepository) {
				repo.GetExpenseHistoryFunc = func(ctx context.Context, expenseID pgtype.UUID) ([]sqlc.GetExpenseHistoryRow, error) {
					return []sqlc.GetExpenseHistoryRow{}, nil
				}
			},
			expectedError: nil,
			validate: func(t *testing.T, history []sqlc.GetExpenseHistoryRow) {
				if len(history) != 0 {
					t.Errorf("expected 0 history items, got %d", len(history))
				}
			},
		},
		{
			name:      "success - multiple activity types for one expense",
			expenseID: testutil.CreateTestUUID(100),
			mockSetup: func(repo *MockGroupActivityRepository) {
				repo.GetExpenseHistoryFunc = func(ctx context.Context, expenseID pgtype.UUID) ([]sqlc.GetExpenseHistoryRow, error) {
					// Verify expenseID is passed correctly
					if expenseID.Bytes != testutil.CreateTestUUID(100).Bytes {
						t.Error("expense ID not passed correctly")
					}
					return []sqlc.GetExpenseHistoryRow{
						{ID: testutil.CreateTestUUID(200), Action: "expense_created"},
						{ID: testutil.CreateTestUUID(201), Action: "comment_added"},
						{ID: testutil.CreateTestUUID(202), Action: "comment_added"},
						{ID: testutil.CreateTestUUID(203), Action: "expense_updated"},
						{ID: testutil.CreateTestUUID(204), Action: "comment_added"},
					}, nil
				}
			},
			expectedError: nil,
			validate: func(t *testing.T, history []sqlc.GetExpenseHistoryRow) {
				if len(history) != 5 {
					t.Errorf("expected 5 history items, got %d", len(history))
				}
			},
		},
		{
			name:      "error - repository fails",
			expenseID: testutil.CreateTestUUID(100),
			mockSetup: func(repo *MockGroupActivityRepository) {
				repo.GetExpenseHistoryFunc = func(ctx context.Context, expenseID pgtype.UUID) ([]sqlc.GetExpenseHistoryRow, error) {
					return nil, errors.New("database error")
				}
			},
			expectedError: errors.New("database error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &MockGroupActivityRepository{}
			if tt.mockSetup != nil {
				tt.mockSetup(repo)
			}

			service := NewGroupActivityService(repo)
			result, err := service.GetExpenseHistory(context.Background(), tt.expenseID)

			if tt.expectedError != nil {
				if err == nil {
					t.Fatalf("expected error %v, got nil", tt.expectedError)
				}
				if err.Error() != tt.expectedError.Error() {
					t.Errorf("expected error '%v', got '%v'", tt.expectedError, err)
				}
			} else {
				if err != nil {
					t.Fatalf("expected no error, got %v", err)
				}
				if tt.validate != nil {
					tt.validate(t, result)
				}
			}
		})
	}
}
