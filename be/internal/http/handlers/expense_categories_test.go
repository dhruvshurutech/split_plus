package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dhruvsaxena1998/splitplus/internal/db/sqlc"
	"github.com/dhruvsaxena1998/splitplus/internal/http/middleware"
	"github.com/dhruvsaxena1998/splitplus/internal/http/response"
	"github.com/dhruvsaxena1998/splitplus/internal/service"
	"github.com/dhruvsaxena1998/splitplus/internal/testutil"
	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/jackc/pgx/v5/pgtype"
)

// MockExpenseCategoryService for testing
type MockExpenseCategoryService struct {
	GetCategoryPresetsFunc          func() []service.CategoryPreset
	ListCategoriesForGroupFunc      func(ctx context.Context, groupID, requesterID pgtype.UUID) ([]sqlc.ExpenseCategory, error)
	CreateCategoryForGroupFunc      func(ctx context.Context, input service.CreateGroupCategoryInput) (sqlc.ExpenseCategory, error)
	UpdateCategoryFunc              func(ctx context.Context, input service.UpdateCategoryInput) (sqlc.ExpenseCategory, error)
	DeleteCategoryFunc              func(ctx context.Context, categoryID, groupID, requesterID pgtype.UUID) error
	CreateCategoriesFromPresetsFunc func(ctx context.Context, groupID, userID pgtype.UUID, presetSlugs []string) ([]sqlc.ExpenseCategory, error)
}

func (m *MockExpenseCategoryService) GetCategoryPresets() []service.CategoryPreset {
	if m.GetCategoryPresetsFunc != nil {
		return m.GetCategoryPresetsFunc()
	}
	return []service.CategoryPreset{}
}

func (m *MockExpenseCategoryService) ListCategoriesForGroup(ctx context.Context, groupID, requesterID pgtype.UUID) ([]sqlc.ExpenseCategory, error) {
	if m.ListCategoriesForGroupFunc != nil {
		return m.ListCategoriesForGroupFunc(ctx, groupID, requesterID)
	}
	return []sqlc.ExpenseCategory{}, nil
}

func (m *MockExpenseCategoryService) CreateCategoryForGroup(ctx context.Context, input service.CreateGroupCategoryInput) (sqlc.ExpenseCategory, error) {
	if m.CreateCategoryForGroupFunc != nil {
		return m.CreateCategoryForGroupFunc(ctx, input)
	}
	return sqlc.ExpenseCategory{}, nil
}

func (m *MockExpenseCategoryService) UpdateCategory(ctx context.Context, input service.UpdateCategoryInput) (sqlc.ExpenseCategory, error) {
	if m.UpdateCategoryFunc != nil {
		return m.UpdateCategoryFunc(ctx, input)
	}
	return sqlc.ExpenseCategory{}, nil
}

func (m *MockExpenseCategoryService) DeleteCategory(ctx context.Context, categoryID, groupID, requesterID pgtype.UUID) error {
	if m.DeleteCategoryFunc != nil {
		return m.DeleteCategoryFunc(ctx, categoryID, groupID, requesterID)
	}
	return nil
}

func (m *MockExpenseCategoryService) CreateCategoriesFromPresets(ctx context.Context, groupID, userID pgtype.UUID, presetSlugs []string) ([]sqlc.ExpenseCategory, error) {
	if m.CreateCategoriesFromPresetsFunc != nil {
		return m.CreateCategoriesFromPresetsFunc(ctx, groupID, userID, presetSlugs)
	}
	return []sqlc.ExpenseCategory{}, nil
}

var _ service.ExpenseCategoryService = (*MockExpenseCategoryService)(nil)

func TestGetCategoryPresetsHandler(t *testing.T) {
	tests := []struct {
		name             string
		mockSetup        func(*MockExpenseCategoryService)
		expectedStatus   int
		validateResponse func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name: "success - get presets",
			mockSetup: func(mock *MockExpenseCategoryService) {
				mock.GetCategoryPresetsFunc = func() []service.CategoryPreset {
					return []service.CategoryPreset{
						{Slug: "food", Name: "Food & Drink", Icon: "🍔", Color: "#FF6B6B"},
						{Slug: "transport", Name: "Transportation", Icon: "🚗", Color: "#96CEB4"},
					}
				}
			},
			expectedStatus: http.StatusOK,
			validateResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var resp response.StandardResponse[[]CategoryPresetResponse, response.ErrorDetail]
				if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
					t.Fatalf("failed to unmarshal response: %v", err)
				}
				if !resp.Status {
					t.Error("expected status true")
				}
				if resp.Data == nil {
					t.Fatal("expected data, got nil")
				}
				if len(*resp.Data) != 2 {
					t.Errorf("expected 2 presets, got %d", len(*resp.Data))
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &MockExpenseCategoryService{}
			if tt.mockSetup != nil {
				tt.mockSetup(mock)
			}

			handler := GetCategoryPresetsHandler(mock)
			req := httptest.NewRequest(http.MethodGet, "/categories/presets", nil)
			w := httptest.NewRecorder()

			handler(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}
			if tt.validateResponse != nil {
				tt.validateResponse(t, w)
			}
		})
	}
}

func TestListGroupCategoriesHandler(t *testing.T) {
	userID := testutil.CreateTestUUID(1)
	groupID := testutil.CreateTestUUID(10)

	tests := []struct {
		name             string
		groupID          string
		userID           pgtype.UUID
		mockSetup        func(*MockExpenseCategoryService)
		expectedStatus   int
		validateResponse func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name:    "success - list categories",
			groupID: groupID.String(),
			userID:  userID,
			mockSetup: func(mock *MockExpenseCategoryService) {
				mock.ListCategoriesForGroupFunc = func(ctx context.Context, groupID, requesterID pgtype.UUID) ([]sqlc.ExpenseCategory, error) {
					return []sqlc.ExpenseCategory{
						{ID: testutil.CreateTestUUID(100), Name: "Food", Slug: "food"},
						{ID: testutil.CreateTestUUID(101), Name: "Transport", Slug: "transport"},
					}, nil
				}
			},
			expectedStatus: http.StatusOK,
			validateResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var resp response.StandardResponse[[]CategoryResponse, response.ErrorDetail]
				if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
					t.Fatalf("failed to unmarshal response: %v", err)
				}
				if resp.Data == nil {
					t.Fatal("expected data, got nil")
				}
				if len(*resp.Data) != 2 {
					t.Errorf("expected 2 categories, got %d", len(*resp.Data))
				}
			},
		},
		{
			name:    "error - not group member",
			groupID: groupID.String(),
			userID:  userID,
			mockSetup: func(mock *MockExpenseCategoryService) {
				mock.ListCategoriesForGroupFunc = func(ctx context.Context, groupID, requesterID pgtype.UUID) ([]sqlc.ExpenseCategory, error) {
					return nil, service.ErrNotGroupMember
				}
			},
			expectedStatus: http.StatusForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &MockExpenseCategoryService{}
			if tt.mockSetup != nil {
				tt.mockSetup(mock)
			}

			handler := ListGroupCategoriesHandler(mock)
			req := createAuthenticatedRequest(http.MethodGet, "/groups/"+tt.groupID+"/categories", nil, tt.userID)

			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("group_id", tt.groupID)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			w := httptest.NewRecorder()
			handler(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}
			if tt.validateResponse != nil {
				tt.validateResponse(t, w)
			}
		})
	}
}

func TestCreateGroupCategoryHandler(t *testing.T) {
	userID := testutil.CreateTestUUID(1)
	groupID := testutil.CreateTestUUID(10)

	tests := []struct {
		name             string
		requestBody      CreateCategoryRequest
		groupID          string
		userID           pgtype.UUID
		mockSetup        func(*MockExpenseCategoryService)
		expectedStatus   int
		validateResponse func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name:        "success - create category",
			requestBody: CreateCategoryRequest{Name: "Custom Category", Icon: "💰", Color: "#FF0000"},
			groupID:     groupID.String(),
			userID:      userID,
			mockSetup: func(mock *MockExpenseCategoryService) {
				mock.CreateCategoryForGroupFunc = func(ctx context.Context, input service.CreateGroupCategoryInput) (sqlc.ExpenseCategory, error) {
					return sqlc.ExpenseCategory{
						ID:      testutil.CreateTestUUID(200),
						GroupID: input.GroupID,
						Name:    input.Name,
						Slug:    service.GenerateSlug(input.Name),
					}, nil
				}
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name:           "error - empty name",
			requestBody:    CreateCategoryRequest{Name: ""},
			groupID:        groupID.String(),
			userID:         userID,
			mockSetup:      func(mock *MockExpenseCategoryService) {},
			expectedStatus: http.StatusUnprocessableEntity,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &MockExpenseCategoryService{}
			if tt.mockSetup != nil {
				tt.mockSetup(mock)
			}

			v := validator.New()
			handler := CreateGroupCategoryHandler(mock)

			body, _ := json.Marshal(tt.requestBody)
			req := createAuthenticatedRequest(http.MethodPost, "/groups/"+tt.groupID+"/categories", body, tt.userID)

			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("group_id", tt.groupID)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			wrappedHandler := middleware.ValidateBody[CreateCategoryRequest](v)(handler)
			w := httptest.NewRecorder()
			wrappedHandler.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}
}
