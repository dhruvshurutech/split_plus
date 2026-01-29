package service

import (
	"context"
	"errors"
	"testing"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/dhruvsaxena1998/splitplus/internal/db/sqlc"
	"github.com/dhruvsaxena1998/splitplus/internal/testutil"
)

// MockExpenseCategoryRepository for testing
type MockExpenseCategoryRepository struct {
	CreateCategoryFunc         func(ctx context.Context, params sqlc.CreateGroupCategoryParams) (sqlc.ExpenseCategory, error)
	GetCategoryByIDFunc        func(ctx context.Context, id pgtype.UUID) (sqlc.ExpenseCategory, error)
	GetCategoryBySlugFunc      func(ctx context.Context, groupID pgtype.UUID, slug string) (sqlc.ExpenseCategory, error)
	ListCategoriesForGroupFunc func(ctx context.Context, groupID pgtype.UUID) ([]sqlc.ExpenseCategory, error)
	UpdateCategoryFunc         func(ctx context.Context, params sqlc.UpdateGroupCategoryParams) (sqlc.ExpenseCategory, error)
	DeleteCategoryFunc         func(ctx context.Context, id, updatedBy pgtype.UUID) error
	IsGroupMemberFunc          func(ctx context.Context, groupID, userID pgtype.UUID) (bool, error)
}

func (m *MockExpenseCategoryRepository) CreateCategory(ctx context.Context, params sqlc.CreateGroupCategoryParams) (sqlc.ExpenseCategory, error) {
	if m.CreateCategoryFunc != nil {
		return m.CreateCategoryFunc(ctx, params)
	}
	return sqlc.ExpenseCategory{}, nil
}

func (m *MockExpenseCategoryRepository) GetCategoryByID(ctx context.Context, id pgtype.UUID) (sqlc.ExpenseCategory, error) {
	if m.GetCategoryByIDFunc != nil {
		return m.GetCategoryByIDFunc(ctx, id)
	}
	return sqlc.ExpenseCategory{}, errors.New("not found")
}

func (m *MockExpenseCategoryRepository) GetCategoryBySlug(ctx context.Context, groupID pgtype.UUID, slug string) (sqlc.ExpenseCategory, error) {
	if m.GetCategoryBySlugFunc != nil {
		return m.GetCategoryBySlugFunc(ctx, groupID, slug)
	}
	return sqlc.ExpenseCategory{}, errors.New("not found")
}

func (m *MockExpenseCategoryRepository) ListCategoriesForGroup(ctx context.Context, groupID pgtype.UUID) ([]sqlc.ExpenseCategory, error) {
	if m.ListCategoriesForGroupFunc != nil {
		return m.ListCategoriesForGroupFunc(ctx, groupID)
	}
	return []sqlc.ExpenseCategory{}, nil
}

func (m *MockExpenseCategoryRepository) UpdateCategory(ctx context.Context, params sqlc.UpdateGroupCategoryParams) (sqlc.ExpenseCategory, error) {
	if m.UpdateCategoryFunc != nil {
		return m.UpdateCategoryFunc(ctx, params)
	}
	return sqlc.ExpenseCategory{}, nil
}

func (m *MockExpenseCategoryRepository) DeleteCategory(ctx context.Context, id, updatedBy pgtype.UUID) error {
	if m.DeleteCategoryFunc != nil {
		return m.DeleteCategoryFunc(ctx, id, updatedBy)
	}
	return nil
}

func (m *MockExpenseCategoryRepository) IsGroupMember(ctx context.Context, groupID, userID pgtype.UUID) (bool, error) {
	if m.IsGroupMemberFunc != nil {
		return m.IsGroupMemberFunc(ctx, groupID, userID)
	}
	return true, nil
}

func TestExpenseCategoryService_ListCategoriesForGroup(t *testing.T) {
	tests := []struct {
		name          string
		groupID       pgtype.UUID
		requesterID   pgtype.UUID
		mockSetup     func(*MockExpenseCategoryRepository)
		expectedError error
		validate      func(*testing.T, []sqlc.ExpenseCategory)
	}{
		{
			name:        "success - list categories",
			groupID:     testutil.CreateTestUUID(10),
			requesterID: testutil.CreateTestUUID(1),
			mockSetup: func(m *MockExpenseCategoryRepository) {
				m.IsGroupMemberFunc = func(ctx context.Context, groupID, userID pgtype.UUID) (bool, error) {
					return true, nil
				}
				m.ListCategoriesForGroupFunc = func(ctx context.Context, groupID pgtype.UUID) ([]sqlc.ExpenseCategory, error) {
					return []sqlc.ExpenseCategory{
						{ID: testutil.CreateTestUUID(100), GroupID: groupID, Name: "Food", Slug: "food"},
						{ID: testutil.CreateTestUUID(101), GroupID: groupID, Name: "Transport", Slug: "transport"},
					}, nil
				}
			},
			validate: func(t *testing.T, categories []sqlc.ExpenseCategory) {
				if len(categories) != 2 {
					t.Errorf("expected 2 categories, got %d", len(categories))
				}
			},
		},
		{
			name:        "error - not group member",
			groupID:     testutil.CreateTestUUID(10),
			requesterID: testutil.CreateTestUUID(1),
			mockSetup: func(m *MockExpenseCategoryRepository) {
				m.IsGroupMemberFunc = func(ctx context.Context, groupID, userID pgtype.UUID) (bool, error) {
					return false, nil
				}
			},
			expectedError: ErrNotGroupMember,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &MockExpenseCategoryRepository{}
			tt.mockSetup(mockRepo)

			service := NewExpenseCategoryService(mockRepo)
			categories, err := service.ListCategoriesForGroup(context.Background(), tt.groupID, tt.requesterID)

			if tt.expectedError != nil {
				if err == nil {
					t.Errorf("expected error %v, got nil", tt.expectedError)
				} else if !errors.Is(err, tt.expectedError) {
					t.Errorf("expected error %v, got %v", tt.expectedError, err)
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if tt.validate != nil {
				tt.validate(t, categories)
			}
		})
	}
}

func TestExpenseCategoryService_CreateCategoryForGroup(t *testing.T) {
	tests := []struct {
		name          string
		input         CreateGroupCategoryInput
		mockSetup     func(*MockExpenseCategoryRepository)
		expectedError error
		validate      func(*testing.T, sqlc.ExpenseCategory)
	}{
		{
			name: "success - create category",
			input: CreateGroupCategoryInput{
				GroupID:   testutil.CreateTestUUID(10),
				Name:      "Food & Drink",
				Icon:      "🍔",
				Color:     "#FF6B6B",
				CreatedBy: testutil.CreateTestUUID(1),
			},
			mockSetup: func(m *MockExpenseCategoryRepository) {
				m.IsGroupMemberFunc = func(ctx context.Context, groupID, userID pgtype.UUID) (bool, error) {
					return true, nil
				}
				m.CreateCategoryFunc = func(ctx context.Context, params sqlc.CreateGroupCategoryParams) (sqlc.ExpenseCategory, error) {
					return sqlc.ExpenseCategory{
						ID:        testutil.CreateTestUUID(100),
						GroupID:   params.GroupID,
						Slug:      params.Slug,
						Name:      params.Name,
						Icon:      params.Icon,
						Color:     params.Color,
						CreatedBy: params.CreatedBy,
						UpdatedBy: params.CreatedBy, // UpdatedBy is same as CreatedBy in CreateGroupCategoryParams
					}, nil
				}
			},
			validate: func(t *testing.T, category sqlc.ExpenseCategory) {
				if category.Name != "Food & Drink" {
					t.Errorf("expected name 'Food & Drink', got '%s'", category.Name)
				}
				if category.Slug != "food-drink" {
					t.Errorf("expected slug 'food-drink', got '%s'", category.Slug)
				}
			},
		},
		{
			name: "error - empty name",
			input: CreateGroupCategoryInput{
				GroupID:   testutil.CreateTestUUID(10),
				Name:      "",
				CreatedBy: testutil.CreateTestUUID(1),
			},
			mockSetup: func(m *MockExpenseCategoryRepository) {
				m.IsGroupMemberFunc = func(ctx context.Context, groupID, userID pgtype.UUID) (bool, error) {
					return true, nil
				}
			},
			expectedError: ErrInvalidCategoryName,
		},
		{
			name: "error - not group member",
			input: CreateGroupCategoryInput{
				GroupID:   testutil.CreateTestUUID(10),
				Name:      "Food",
				CreatedBy: testutil.CreateTestUUID(1),
			},
			mockSetup: func(m *MockExpenseCategoryRepository) {
				m.IsGroupMemberFunc = func(ctx context.Context, groupID, userID pgtype.UUID) (bool, error) {
					return false, nil
				}
			},
			expectedError: ErrNotGroupMember,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &MockExpenseCategoryRepository{}
			tt.mockSetup(mockRepo)

			service := NewExpenseCategoryService(mockRepo)
			category, err := service.CreateCategoryForGroup(context.Background(), tt.input)

			if tt.expectedError != nil {
				if err == nil {
					t.Errorf("expected error %v, got nil", tt.expectedError)
				} else if !errors.Is(err, tt.expectedError) {
					t.Errorf("expected error %v, got %v", tt.expectedError, err)
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if tt.validate != nil {
				tt.validate(t, category)
			}
		})
	}
}

func TestExpenseCategoryService_UpdateCategory(t *testing.T) {
	tests := []struct {
		name          string
		input         UpdateCategoryInput
		mockSetup     func(*MockExpenseCategoryRepository)
		expectedError error
		validate      func(*testing.T, sqlc.ExpenseCategory)
	}{
		{
			name: "success - update category",
			input: UpdateCategoryInput{
				CategoryID: testutil.CreateTestUUID(100),
				GroupID:    testutil.CreateTestUUID(10),
				Name:       "Updated Food",
				Icon:       "🍕",
				Color:      "#FF0000",
				UpdatedBy:  testutil.CreateTestUUID(1),
			},
			mockSetup: func(m *MockExpenseCategoryRepository) {
				m.GetCategoryByIDFunc = func(ctx context.Context, id pgtype.UUID) (sqlc.ExpenseCategory, error) {
					return sqlc.ExpenseCategory{
						ID:      id,
						GroupID: testutil.CreateTestUUID(10),
						Name:    "Food",
						Slug:    "food",
					}, nil
				}
				m.IsGroupMemberFunc = func(ctx context.Context, groupID, userID pgtype.UUID) (bool, error) {
					return true, nil
				}
				m.UpdateCategoryFunc = func(ctx context.Context, params sqlc.UpdateGroupCategoryParams) (sqlc.ExpenseCategory, error) {
					return sqlc.ExpenseCategory{
						ID:        params.ID,
						GroupID:   testutil.CreateTestUUID(10),
						Slug:      params.Slug,
						Name:      params.Name,
						Icon:      params.Icon,
						Color:     params.Color,
						UpdatedBy: params.UpdatedBy,
					}, nil
				}
			},
			validate: func(t *testing.T, category sqlc.ExpenseCategory) {
				if category.Name != "Updated Food" {
					t.Errorf("expected name 'Updated Food', got '%s'", category.Name)
				}
			},
		},
		{
			name: "error - category not found",
			input: UpdateCategoryInput{
				CategoryID: testutil.CreateTestUUID(100),
				GroupID:    testutil.CreateTestUUID(10),
				Name:       "Updated Food",
				UpdatedBy:  testutil.CreateTestUUID(1),
			},
			mockSetup: func(m *MockExpenseCategoryRepository) {
				m.GetCategoryByIDFunc = func(ctx context.Context, id pgtype.UUID) (sqlc.ExpenseCategory, error) {
					return sqlc.ExpenseCategory{}, errors.New("not found")
				}
			},
			expectedError: ErrCategoryNotFound,
		},
		{
			name: "error - category belongs to different group",
			input: UpdateCategoryInput{
				CategoryID: testutil.CreateTestUUID(100),
				GroupID:    testutil.CreateTestUUID(10),
				Name:       "Updated Food",
				UpdatedBy:  testutil.CreateTestUUID(1),
			},
			mockSetup: func(m *MockExpenseCategoryRepository) {
				m.GetCategoryByIDFunc = func(ctx context.Context, id pgtype.UUID) (sqlc.ExpenseCategory, error) {
					return sqlc.ExpenseCategory{
						ID:      id,
						GroupID: testutil.CreateTestUUID(20), // Different group
					}, nil
				}
			},
			expectedError: ErrCategoryNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &MockExpenseCategoryRepository{}
			tt.mockSetup(mockRepo)

			service := NewExpenseCategoryService(mockRepo)
			category, err := service.UpdateCategory(context.Background(), tt.input)

			if tt.expectedError != nil {
				if err == nil {
					t.Errorf("expected error %v, got nil", tt.expectedError)
				} else if !errors.Is(err, tt.expectedError) {
					t.Errorf("expected error %v, got %v", tt.expectedError, err)
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if tt.validate != nil {
				tt.validate(t, category)
			}
		})
	}
}

func TestExpenseCategoryService_DeleteCategory(t *testing.T) {
	tests := []struct {
		name          string
		categoryID    pgtype.UUID
		groupID       pgtype.UUID
		requesterID   pgtype.UUID
		mockSetup     func(*MockExpenseCategoryRepository)
		expectedError error
	}{
		{
			name:        "success - delete category",
			categoryID:  testutil.CreateTestUUID(100),
			groupID:     testutil.CreateTestUUID(10),
			requesterID: testutil.CreateTestUUID(1),
			mockSetup: func(m *MockExpenseCategoryRepository) {
				m.GetCategoryByIDFunc = func(ctx context.Context, id pgtype.UUID) (sqlc.ExpenseCategory, error) {
					return sqlc.ExpenseCategory{
						ID:      id,
						GroupID: testutil.CreateTestUUID(10),
					}, nil
				}
				m.IsGroupMemberFunc = func(ctx context.Context, groupID, userID pgtype.UUID) (bool, error) {
					return true, nil
				}
				m.DeleteCategoryFunc = func(ctx context.Context, id, updatedBy pgtype.UUID) error {
					return nil
				}
			},
		},
		{
			name:        "error - category not found",
			categoryID:  testutil.CreateTestUUID(100),
			groupID:     testutil.CreateTestUUID(10),
			requesterID: testutil.CreateTestUUID(1),
			mockSetup: func(m *MockExpenseCategoryRepository) {
				m.GetCategoryByIDFunc = func(ctx context.Context, id pgtype.UUID) (sqlc.ExpenseCategory, error) {
					return sqlc.ExpenseCategory{}, errors.New("not found")
				}
			},
			expectedError: ErrCategoryNotFound,
		},
		{
			name:        "error - not group member",
			categoryID:  testutil.CreateTestUUID(100),
			groupID:     testutil.CreateTestUUID(10),
			requesterID: testutil.CreateTestUUID(1),
			mockSetup: func(m *MockExpenseCategoryRepository) {
				m.GetCategoryByIDFunc = func(ctx context.Context, id pgtype.UUID) (sqlc.ExpenseCategory, error) {
					return sqlc.ExpenseCategory{
						ID:      id,
						GroupID: testutil.CreateTestUUID(10),
					}, nil
				}
				m.IsGroupMemberFunc = func(ctx context.Context, groupID, userID pgtype.UUID) (bool, error) {
					return false, nil
				}
			},
			expectedError: ErrNotGroupMember,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &MockExpenseCategoryRepository{}
			tt.mockSetup(mockRepo)

			service := NewExpenseCategoryService(mockRepo)
			err := service.DeleteCategory(context.Background(), tt.categoryID, tt.groupID, tt.requesterID)

			if tt.expectedError != nil {
				if err == nil {
					t.Errorf("expected error %v, got nil", tt.expectedError)
				} else if !errors.Is(err, tt.expectedError) {
					t.Errorf("expected error %v, got %v", tt.expectedError, err)
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}
