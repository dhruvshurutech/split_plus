package service

import (
	"context"
	"errors"
	"testing"

	"github.com/dhruvsaxena1998/splitplus/internal/db/sqlc"
	"github.com/dhruvsaxena1998/splitplus/internal/testutil"
	"github.com/jackc/pgx/v5/pgtype"
)

// MockExpenseCommentRepository for testing
type MockExpenseCommentRepository struct {
	CreateCommentFunc  func(ctx context.Context, params sqlc.CreateExpenseCommentParams) (sqlc.ExpenseComment, error)
	GetCommentByIDFunc func(ctx context.Context, id pgtype.UUID) (sqlc.GetExpenseCommentByIDRow, error)
	ListCommentsFunc   func(ctx context.Context, expenseID pgtype.UUID) ([]sqlc.ListExpenseCommentsRow, error)
	UpdateCommentFunc  func(ctx context.Context, params sqlc.UpdateExpenseCommentParams) (sqlc.ExpenseComment, error)
	DeleteCommentFunc  func(ctx context.Context, id pgtype.UUID) error
	CountCommentsFunc  func(ctx context.Context, expenseID pgtype.UUID) (int64, error)
}

func (m *MockExpenseCommentRepository) CreateComment(ctx context.Context, params sqlc.CreateExpenseCommentParams) (sqlc.ExpenseComment, error) {
	if m.CreateCommentFunc != nil {
		return m.CreateCommentFunc(ctx, params)
	}
	return sqlc.ExpenseComment{}, nil
}

func (m *MockExpenseCommentRepository) GetCommentByID(ctx context.Context, id pgtype.UUID) (sqlc.GetExpenseCommentByIDRow, error) {
	if m.GetCommentByIDFunc != nil {
		return m.GetCommentByIDFunc(ctx, id)
	}
	return sqlc.GetExpenseCommentByIDRow{}, errors.New("not found")
}

func (m *MockExpenseCommentRepository) ListComments(ctx context.Context, expenseID pgtype.UUID) ([]sqlc.ListExpenseCommentsRow, error) {
	if m.ListCommentsFunc != nil {
		return m.ListCommentsFunc(ctx, expenseID)
	}
	return []sqlc.ListExpenseCommentsRow{}, nil
}

func (m *MockExpenseCommentRepository) UpdateComment(ctx context.Context, params sqlc.UpdateExpenseCommentParams) (sqlc.ExpenseComment, error) {
	if m.UpdateCommentFunc != nil {
		return m.UpdateCommentFunc(ctx, params)
	}
	return sqlc.ExpenseComment{}, nil
}

func (m *MockExpenseCommentRepository) DeleteComment(ctx context.Context, id pgtype.UUID) error {
	if m.DeleteCommentFunc != nil {
		return m.DeleteCommentFunc(ctx, id)
	}
	return nil
}

func (m *MockExpenseCommentRepository) CountComments(ctx context.Context, expenseID pgtype.UUID) (int64, error) {
	if m.CountCommentsFunc != nil {
		return m.CountCommentsFunc(ctx, expenseID)
	}
	return 0, nil
}

// MockExpenseService for testing
type MockExpenseService struct {
	GetExpenseByIDFunc func(ctx context.Context, expenseID, userID pgtype.UUID) (sqlc.Expense, error)
}

func (m *MockExpenseService) GetExpenseByID(ctx context.Context, expenseID, userID pgtype.UUID) (sqlc.Expense, error) {
	if m.GetExpenseByIDFunc != nil {
		return m.GetExpenseByIDFunc(ctx, expenseID, userID)
	}
	return sqlc.Expense{}, nil
}

// Implement other ExpenseService methods as no-ops for interface compliance
func (m *MockExpenseService) CreateExpense(ctx context.Context, input CreateExpenseInput) (CreateExpenseResult, error) {
	return CreateExpenseResult{}, nil
}
func (m *MockExpenseService) ListExpensesByGroup(ctx context.Context, groupID, requesterID pgtype.UUID) ([]sqlc.Expense, error) {
	return []sqlc.Expense{}, nil
}
func (m *MockExpenseService) UpdateExpense(ctx context.Context, input UpdateExpenseInput) (CreateExpenseResult, error) {
	return CreateExpenseResult{}, nil
}
func (m *MockExpenseService) DeleteExpense(ctx context.Context, expenseID, requesterID pgtype.UUID) error {
	return nil
}
func (m *MockExpenseService) GetExpensePayments(ctx context.Context, expenseID, requesterID pgtype.UUID) ([]sqlc.ListExpensePaymentsRow, error) {
	return []sqlc.ListExpensePaymentsRow{}, nil
}
func (m *MockExpenseService) GetExpenseSplits(ctx context.Context, expenseID, requesterID pgtype.UUID) ([]sqlc.ListExpenseSplitsRow, error) {
	return []sqlc.ListExpenseSplitsRow{}, nil
}
func (m *MockExpenseService) SearchExpenses(ctx context.Context, input SearchExpensesInput, requesterID pgtype.UUID) ([]sqlc.Expense, error) {
	return []sqlc.Expense{}, nil
}

func TestExpenseCommentService_CreateComment(t *testing.T) {
	tests := []struct {
		name          string
		expenseID     pgtype.UUID
		userID        pgtype.UUID
		comment       string
		mockSetup     func(*MockExpenseCommentRepository, *MockExpenseService, *MockGroupActivityService)
		expectedError error
		validate      func(*testing.T, sqlc.GetExpenseCommentByIDRow)
	}{
		{
			name:      "success - create comment",
			expenseID: testutil.CreateTestUUID(1),
			userID:    testutil.CreateTestUUID(10),
			comment:   "This is a test comment",
			mockSetup: func(repo *MockExpenseCommentRepository, expSvc *MockExpenseService, actSvc *MockGroupActivityService) {
				expSvc.GetExpenseByIDFunc = func(ctx context.Context, expenseID, userID pgtype.UUID) (sqlc.Expense, error) {
					return sqlc.Expense{
						ID:      expenseID,
						GroupID: testutil.CreateTestUUID(100),
						Title:   "Test Expense",
					}, nil
				}
				repo.CreateCommentFunc = func(ctx context.Context, params sqlc.CreateExpenseCommentParams) (sqlc.ExpenseComment, error) {
					return sqlc.ExpenseComment{
						ID:        testutil.CreateTestUUID(1000),
						ExpenseID: params.ExpenseID,
						UserID:    params.UserID,
						Comment:   params.Comment,
					}, nil
				}
				repo.GetCommentByIDFunc = func(ctx context.Context, id pgtype.UUID) (sqlc.GetExpenseCommentByIDRow, error) {
					return sqlc.GetExpenseCommentByIDRow{
						ID:        id,
						ExpenseID: testutil.CreateTestUUID(1),
						UserID:    testutil.CreateTestUUID(10),
						Comment:   "This is a test comment",
						UserName:  pgtype.Text{String: "Test User", Valid: true},
						UserEmail: "test@example.com",
					}, nil
				}
				actSvc.LogActivityFunc = func(ctx context.Context, input LogActivityInput) error {
					return nil
				}
			},
			expectedError: nil,
			validate: func(t *testing.T, comment sqlc.GetExpenseCommentByIDRow) {
				if comment.Comment != "This is a test comment" {
					t.Errorf("expected comment text 'This is a test comment', got '%s'", comment.Comment)
				}
			},
		},
		{
			name:      "error - empty comment",
			expenseID: testutil.CreateTestUUID(1),
			userID:    testutil.CreateTestUUID(10),
			comment:   "",
			mockSetup: func(repo *MockExpenseCommentRepository, expSvc *MockExpenseService, actSvc *MockGroupActivityService) {
				// No setup needed, validation happens before repo call
			},
			expectedError: ErrCommentEmpty,
		},
		{
			name:      "error - expense not found",
			expenseID: testutil.CreateTestUUID(1),
			userID:    testutil.CreateTestUUID(10),
			comment:   "Valid comment",
			mockSetup: func(repo *MockExpenseCommentRepository, expSvc *MockExpenseService, actSvc *MockGroupActivityService) {
				expSvc.GetExpenseByIDFunc = func(ctx context.Context, expenseID, userID pgtype.UUID) (sqlc.Expense, error) {
					return sqlc.Expense{}, errors.New("expense not found")
				}
			},
			expectedError: errors.New("expense not found"),
		},
		{
			name:      "error - repository create fails",
			expenseID: testutil.CreateTestUUID(1),
			userID:    testutil.CreateTestUUID(10),
			comment:   "Valid comment",
			mockSetup: func(repo *MockExpenseCommentRepository, expSvc *MockExpenseService, actSvc *MockGroupActivityService) {
				expSvc.GetExpenseByIDFunc = func(ctx context.Context, expenseID, userID pgtype.UUID) (sqlc.Expense, error) {
					return sqlc.Expense{
						ID:      expenseID,
						GroupID: testutil.CreateTestUUID(100),
					}, nil
				}
				repo.CreateCommentFunc = func(ctx context.Context, params sqlc.CreateExpenseCommentParams) (sqlc.ExpenseComment, error) {
					return sqlc.ExpenseComment{}, errors.New("database error")
				}
			},
			expectedError: errors.New("database error"),
		},
		{
			name:      "success - activity logging failure does not fail comment creation",
			expenseID: testutil.CreateTestUUID(1),
			userID:    testutil.CreateTestUUID(10),
			comment:   "Test comment",
			mockSetup: func(repo *MockExpenseCommentRepository, expSvc *MockExpenseService, actSvc *MockGroupActivityService) {
				expSvc.GetExpenseByIDFunc = func(ctx context.Context, expenseID, userID pgtype.UUID) (sqlc.Expense, error) {
					return sqlc.Expense{
						ID:      expenseID,
						GroupID: testutil.CreateTestUUID(100),
					}, nil
				}
				repo.CreateCommentFunc = func(ctx context.Context, params sqlc.CreateExpenseCommentParams) (sqlc.ExpenseComment, error) {
					return sqlc.ExpenseComment{
						ID:        testutil.CreateTestUUID(1000),
						ExpenseID: params.ExpenseID,
						UserID:    params.UserID,
						Comment:   params.Comment,
					}, nil
				}
				repo.GetCommentByIDFunc = func(ctx context.Context, id pgtype.UUID) (sqlc.GetExpenseCommentByIDRow, error) {
					return sqlc.GetExpenseCommentByIDRow{
						ID:        id,
						ExpenseID: testutil.CreateTestUUID(1),
						UserID:    testutil.CreateTestUUID(10),
						Comment:   "Test comment",
						UserName:  pgtype.Text{String: "Test User", Valid: true},
						UserEmail: "test@example.com",
					}, nil
				}
				actSvc.LogActivityFunc = func(ctx context.Context, input LogActivityInput) error {
					return errors.New("activity service error")
				}
			},
			expectedError: nil,
			validate: func(t *testing.T, comment sqlc.GetExpenseCommentByIDRow) {
				if comment.ID.Bytes != testutil.CreateTestUUID(1000).Bytes {
					t.Error("comment should be created even if activity logging fails")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &MockExpenseCommentRepository{}
			expSvc := &MockExpenseService{}
			actSvc := &MockGroupActivityService{}

			if tt.mockSetup != nil {
				tt.mockSetup(repo, expSvc, actSvc)
			}

			service := NewExpenseCommentService(repo, expSvc, actSvc)
			result, err := service.CreateComment(context.Background(), tt.expenseID, tt.userID, tt.comment)

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

func TestExpenseCommentService_ListComments(t *testing.T) {
	tests := []struct {
		name          string
		expenseID     pgtype.UUID
		mockSetup     func(*MockExpenseCommentRepository)
		expectedError error
		validate      func(*testing.T, []sqlc.ListExpenseCommentsRow)
	}{
		{
			name:      "success - list comments",
			expenseID: testutil.CreateTestUUID(1),
			mockSetup: func(repo *MockExpenseCommentRepository) {
				repo.ListCommentsFunc = func(ctx context.Context, expenseID pgtype.UUID) ([]sqlc.ListExpenseCommentsRow, error) {
					return []sqlc.ListExpenseCommentsRow{
						{
							ID:      testutil.CreateTestUUID(100),
							Comment: "First comment",
						},
						{
							ID:      testutil.CreateTestUUID(101),
							Comment: "Second comment",
						},
					}, nil
				}
			},
			expectedError: nil,
			validate: func(t *testing.T, comments []sqlc.ListExpenseCommentsRow) {
				if len(comments) != 2 {
					t.Errorf("expected 2 comments, got %d", len(comments))
				}
			},
		},
		{
			name:      "success - empty list",
			expenseID: testutil.CreateTestUUID(1),
			mockSetup: func(repo *MockExpenseCommentRepository) {
				repo.ListCommentsFunc = func(ctx context.Context, expenseID pgtype.UUID) ([]sqlc.ListExpenseCommentsRow, error) {
					return []sqlc.ListExpenseCommentsRow{}, nil
				}
			},
			expectedError: nil,
			validate: func(t *testing.T, comments []sqlc.ListExpenseCommentsRow) {
				if len(comments) != 0 {
					t.Errorf("expected 0 comments, got %d", len(comments))
				}
			},
		},
		{
			name:      "error - repository fails",
			expenseID: testutil.CreateTestUUID(1),
			mockSetup: func(repo *MockExpenseCommentRepository) {
				repo.ListCommentsFunc = func(ctx context.Context, expenseID pgtype.UUID) ([]sqlc.ListExpenseCommentsRow, error) {
					return nil, errors.New("database error")
				}
			},
			expectedError: errors.New("database error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &MockExpenseCommentRepository{}
			if tt.mockSetup != nil {
				tt.mockSetup(repo)
			}

			service := NewExpenseCommentService(repo, &MockExpenseService{}, &MockGroupActivityService{})
			result, err := service.ListComments(context.Background(), tt.expenseID)

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

func TestExpenseCommentService_UpdateComment(t *testing.T) {
	tests := []struct {
		name          string
		commentID     pgtype.UUID
		userID        pgtype.UUID
		comment       string
		mockSetup     func(*MockExpenseCommentRepository)
		expectedError error
		validate      func(*testing.T, sqlc.ExpenseComment)
	}{
		{
			name:      "success - update comment",
			commentID: testutil.CreateTestUUID(1),
			userID:    testutil.CreateTestUUID(10),
			comment:   "Updated comment",
			mockSetup: func(repo *MockExpenseCommentRepository) {
				repo.GetCommentByIDFunc = func(ctx context.Context, id pgtype.UUID) (sqlc.GetExpenseCommentByIDRow, error) {
					return sqlc.GetExpenseCommentByIDRow{
						ID:     id,
						UserID: testutil.CreateTestUUID(10),
					}, nil
				}
				repo.UpdateCommentFunc = func(ctx context.Context, params sqlc.UpdateExpenseCommentParams) (sqlc.ExpenseComment, error) {
					return sqlc.ExpenseComment{
						ID:      params.ID,
						Comment: params.Comment,
					}, nil
				}
			},
			expectedError: nil,
			validate: func(t *testing.T, comment sqlc.ExpenseComment) {
				if comment.Comment != "Updated comment" {
					t.Errorf("expected comment 'Updated comment', got '%s'", comment.Comment)
				}
			},
		},
		{
			name:      "error - empty comment",
			commentID: testutil.CreateTestUUID(1),
			userID:    testutil.CreateTestUUID(10),
			comment:   "",
			mockSetup: func(repo *MockExpenseCommentRepository) {
				// No setup needed, validation happens first
			},
			expectedError: ErrCommentEmpty,
		},
		{
			name:      "error - comment not found",
			commentID: testutil.CreateTestUUID(1),
			userID:    testutil.CreateTestUUID(10),
			comment:   "Updated comment",
			mockSetup: func(repo *MockExpenseCommentRepository) {
				repo.GetCommentByIDFunc = func(ctx context.Context, id pgtype.UUID) (sqlc.GetExpenseCommentByIDRow, error) {
					return sqlc.GetExpenseCommentByIDRow{}, errors.New("not found")
				}
			},
			expectedError: errors.New("not found"),
		},
		{
			name:      "error - permission denied (different user)",
			commentID: testutil.CreateTestUUID(1),
			userID:    testutil.CreateTestUUID(10),
			comment:   "Updated comment",
			mockSetup: func(repo *MockExpenseCommentRepository) {
				repo.GetCommentByIDFunc = func(ctx context.Context, id pgtype.UUID) (sqlc.GetExpenseCommentByIDRow, error) {
					return sqlc.GetExpenseCommentByIDRow{
						ID:     id,
						UserID: testutil.CreateTestUUID(99), // Different user
					}, nil
				}
			},
			expectedError: ErrCommentPermissioDenied,
		},
		{
			name:      "error - repository update fails",
			commentID: testutil.CreateTestUUID(1),
			userID:    testutil.CreateTestUUID(10),
			comment:   "Updated comment",
			mockSetup: func(repo *MockExpenseCommentRepository) {
				repo.GetCommentByIDFunc = func(ctx context.Context, id pgtype.UUID) (sqlc.GetExpenseCommentByIDRow, error) {
					return sqlc.GetExpenseCommentByIDRow{
						ID:     id,
						UserID: testutil.CreateTestUUID(10),
					}, nil
				}
				repo.UpdateCommentFunc = func(ctx context.Context, params sqlc.UpdateExpenseCommentParams) (sqlc.ExpenseComment, error) {
					return sqlc.ExpenseComment{}, errors.New("database error")
				}
			},
			expectedError: errors.New("database error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &MockExpenseCommentRepository{}
			if tt.mockSetup != nil {
				tt.mockSetup(repo)
			}

			service := NewExpenseCommentService(repo, &MockExpenseService{}, &MockGroupActivityService{})
			result, err := service.UpdateComment(context.Background(), tt.commentID, tt.userID, tt.comment)

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

func TestExpenseCommentService_DeleteComment(t *testing.T) {
	tests := []struct {
		name          string
		commentID     pgtype.UUID
		userID        pgtype.UUID
		mockSetup     func(*MockExpenseCommentRepository)
		expectedError error
	}{
		{
			name:      "success - delete comment",
			commentID: testutil.CreateTestUUID(1),
			userID:    testutil.CreateTestUUID(10),
			mockSetup: func(repo *MockExpenseCommentRepository) {
				repo.GetCommentByIDFunc = func(ctx context.Context, id pgtype.UUID) (sqlc.GetExpenseCommentByIDRow, error) {
					return sqlc.GetExpenseCommentByIDRow{
						ID:     id,
						UserID: testutil.CreateTestUUID(10),
					}, nil
				}
				repo.DeleteCommentFunc = func(ctx context.Context, id pgtype.UUID) error {
					return nil
				}
			},
			expectedError: nil,
		},
		{
			name:      "error - comment not found",
			commentID: testutil.CreateTestUUID(1),
			userID:    testutil.CreateTestUUID(10),
			mockSetup: func(repo *MockExpenseCommentRepository) {
				repo.GetCommentByIDFunc = func(ctx context.Context, id pgtype.UUID) (sqlc.GetExpenseCommentByIDRow, error) {
					return sqlc.GetExpenseCommentByIDRow{}, errors.New("not found")
				}
			},
			expectedError: errors.New("not found"),
		},
		{
			name:      "error - permission denied (different user)",
			commentID: testutil.CreateTestUUID(1),
			userID:    testutil.CreateTestUUID(10),
			mockSetup: func(repo *MockExpenseCommentRepository) {
				repo.GetCommentByIDFunc = func(ctx context.Context, id pgtype.UUID) (sqlc.GetExpenseCommentByIDRow, error) {
					return sqlc.GetExpenseCommentByIDRow{
						ID:     id,
						UserID: testutil.CreateTestUUID(99), // Different user
					}, nil
				}
			},
			expectedError: ErrCommentPermissioDenied,
		},
		{
			name:      "error - repository delete fails",
			commentID: testutil.CreateTestUUID(1),
			userID:    testutil.CreateTestUUID(10),
			mockSetup: func(repo *MockExpenseCommentRepository) {
				repo.GetCommentByIDFunc = func(ctx context.Context, id pgtype.UUID) (sqlc.GetExpenseCommentByIDRow, error) {
					return sqlc.GetExpenseCommentByIDRow{
						ID:     id,
						UserID: testutil.CreateTestUUID(10),
					}, nil
				}
				repo.DeleteCommentFunc = func(ctx context.Context, id pgtype.UUID) error {
					return errors.New("database error")
				}
			},
			expectedError: errors.New("database error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &MockExpenseCommentRepository{}
			if tt.mockSetup != nil {
				tt.mockSetup(repo)
			}

			service := NewExpenseCommentService(repo, &MockExpenseService{}, &MockGroupActivityService{})
			err := service.DeleteComment(context.Background(), tt.commentID, tt.userID)

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
			}
		})
	}
}

func TestTruncateString(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		max      int
		expected string
	}{
		{
			name:     "string shorter than max",
			input:    "short",
			max:      10,
			expected: "short",
		},
		{
			name:     "string equal to max",
			input:    "exactly10c",
			max:      10,
			expected: "exactly10c",
		},
		{
			name:     "string longer than max",
			input:    "this is a very long string that needs truncation",
			max:      20,
			expected: "this is a very long ...",
		},
		{
			name:     "empty string",
			input:    "",
			max:      10,
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := truncateString(tt.input, tt.max)
			if result != tt.expected {
				t.Errorf("expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}
