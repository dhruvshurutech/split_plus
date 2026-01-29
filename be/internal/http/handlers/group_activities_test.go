package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dhruvsaxena1998/splitplus/internal/db/sqlc"
	"github.com/dhruvsaxena1998/splitplus/internal/http/response"
	"github.com/dhruvsaxena1998/splitplus/internal/service"
	"github.com/dhruvsaxena1998/splitplus/internal/testutil"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

// MockGroupActivityService for testing
type MockGroupActivityService struct {
	ListGroupActivitiesFunc func(ctx context.Context, groupID pgtype.UUID, limit, offset int32) ([]sqlc.ListGroupActivitiesRow, error)
	GetExpenseHistoryFunc   func(ctx context.Context, expenseID pgtype.UUID) ([]sqlc.GetExpenseHistoryRow, error)
}

func (m *MockGroupActivityService) LogActivity(ctx context.Context, input service.LogActivityInput) error {
	// Not needed for handler tests
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

var _ service.GroupActivityService = (*MockGroupActivityService)(nil)

func TestListGroupActivitiesHandler(t *testing.T) {
	groupID := testutil.CreateTestUUID(10)

	tests := []struct {
		name             string
		groupID          string
		queryParams      map[string]string
		mockSetup        func(*MockGroupActivityService)
		expectedStatus   int
		validateResponse func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name:    "success - list activities",
			groupID: groupID.String(),
			mockSetup: func(mock *MockGroupActivityService) {
				mock.ListGroupActivitiesFunc = func(ctx context.Context, gID pgtype.UUID, limit, offset int32) ([]sqlc.ListGroupActivitiesRow, error) {
					metadata, _ := json.Marshal(map[string]interface{}{"amount": "25.50"})
					return []sqlc.ListGroupActivitiesRow{
						{
							ID:         testutil.CreateTestUUID(100),
							GroupID:    gID,
							Action:     "expense_created",
							EntityType: "expense",
							Metadata:   metadata,
							UserEmail:  "user1@example.com",
						},
						{
							ID:         testutil.CreateTestUUID(101),
							GroupID:    gID,
							Action:     "comment_added",
							EntityType: "expense",
							Metadata:   []byte("{}"),
							UserEmail:  "user2@example.com",
						},
					}, nil
				}
			},
			expectedStatus: http.StatusOK,
			validateResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var resp response.StandardResponse[[]ActivityResponse, response.ErrorDetail]
				if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
					t.Fatalf("failed to unmarshal response: %v", err)
				}
				if resp.Data == nil {
					t.Fatal("expected data, got nil")
				}
				if len(*resp.Data) != 2 {
					t.Errorf("expected 2 activities, got %d", len(*resp.Data))
				}
				if (*resp.Data)[0].Action != "expense_created" {
					t.Errorf("expected action 'expense_created', got '%s'", (*resp.Data)[0].Action)
				}
			},
		},
		{
			name:    "success - empty list",
			groupID: groupID.String(),
			mockSetup: func(mock *MockGroupActivityService) {
				mock.ListGroupActivitiesFunc = func(ctx context.Context, gID pgtype.UUID, limit, offset int32) ([]sqlc.ListGroupActivitiesRow, error) {
					return []sqlc.ListGroupActivitiesRow{}, nil
				}
			},
			expectedStatus: http.StatusOK,
			validateResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var resp response.StandardResponse[[]ActivityResponse, response.ErrorDetail]
				if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
					t.Fatalf("failed to unmarshal response: %v", err)
				}
				if resp.Data == nil {
					t.Fatal("expected data, got nil")
				}
				if len(*resp.Data) != 0 {
					t.Errorf("expected 0 activities, got %d", len(*resp.Data))
				}
			},
		},
		{
			name:        "success - pagination",
			groupID:     groupID.String(),
			queryParams: map[string]string{"limit": "5", "offset": "10"},
			mockSetup: func(mock *MockGroupActivityService) {
				mock.ListGroupActivitiesFunc = func(ctx context.Context, gID pgtype.UUID, limit, offset int32) ([]sqlc.ListGroupActivitiesRow, error) {
					// Verify pagination params
					if limit != 5 {
						t.Errorf("expected limit 5, got %d", limit)
					}
					if offset != 10 {
						t.Errorf("expected offset 10, got %d", offset)
					}
					return []sqlc.ListGroupActivitiesRow{
						{
							ID:         testutil.CreateTestUUID(100),
							GroupID:    gID,
							Action:     "test_action",
							EntityType: "test",
							UserEmail:  "test@example.com",
						},
					}, nil
				}
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:        "success - default pagination",
			groupID:     groupID.String(),
			queryParams: map[string]string{},
			mockSetup: func(mock *MockGroupActivityService) {
				mock.ListGroupActivitiesFunc = func(ctx context.Context, gID pgtype.UUID, limit, offset int32) ([]sqlc.ListGroupActivitiesRow, error) {
					// Verify default values
					if limit != 20 {
						t.Errorf("expected default limit 20, got %d", limit)
					}
					if offset != 0 {
						t.Errorf("expected default offset 0, got %d", offset)
					}
					return []sqlc.ListGroupActivitiesRow{}, nil
				}
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "error - invalid group_id",
			groupID:        "invalid-uuid",
			mockSetup:      func(mock *MockGroupActivityService) {},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &MockGroupActivityService{}
			if tt.mockSetup != nil {
				tt.mockSetup(mock)
			}

			handler := ListGroupActivitiesHandler(mock)

			url := "/groups/" + tt.groupID + "/activity"
			if len(tt.queryParams) > 0 {
				url += "?"
				first := true
				for k, v := range tt.queryParams {
					if !first {
						url += "&"
					}
					url += k + "=" + v
					first = false
				}
			}

			req := httptest.NewRequest(http.MethodGet, url, nil)

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

func TestGetExpenseHistoryHandler(t *testing.T) {
	expenseID := testutil.CreateTestUUID(100)

	tests := []struct {
		name             string
		expenseID        string
		mockSetup        func(*MockGroupActivityService)
		expectedStatus   int
		validateResponse func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name:      "success - get expense history",
			expenseID: expenseID.String(),
			mockSetup: func(mock *MockGroupActivityService) {
				mock.GetExpenseHistoryFunc = func(ctx context.Context, eID pgtype.UUID) ([]sqlc.GetExpenseHistoryRow, error) {
					metadata1, _ := json.Marshal(map[string]interface{}{"amount": "20.00"})
					metadata2, _ := json.Marshal(map[string]interface{}{"new_amount": "25.50"})
					return []sqlc.GetExpenseHistoryRow{
						{
							ID:         testutil.CreateTestUUID(200),
							Action:     "expense_created",
							EntityType: "expense",
							EntityID:   eID,
							Metadata:   metadata1,
							UserEmail:  "user1@example.com",
						},
						{
							ID:         testutil.CreateTestUUID(201),
							Action:     "expense_updated",
							EntityType: "expense",
							EntityID:   eID,
							Metadata:   metadata2,
							UserEmail:  "user1@example.com",
						},
						{
							ID:         testutil.CreateTestUUID(202),
							Action:     "comment_added",
							EntityType: "expense",
							EntityID:   eID,
							Metadata:   []byte("{}"),
							UserEmail:  "user2@example.com",
						},
					}, nil
				}
			},
			expectedStatus: http.StatusOK,
			validateResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var resp response.StandardResponse[[]ActivityResponse, response.ErrorDetail]
				if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
					t.Fatalf("failed to unmarshal response: %v", err)
				}
				if resp.Data == nil {
					t.Fatal("expected data, got nil")
				}
				if len(*resp.Data) != 3 {
					t.Errorf("expected 3 history items, got %d", len(*resp.Data))
				}
				if (*resp.Data)[0].Action != "expense_created" {
					t.Errorf("expected first action 'expense_created', got '%s'", (*resp.Data)[0].Action)
				}
			},
		},
		{
			name:      "success - empty history",
			expenseID: expenseID.String(),
			mockSetup: func(mock *MockGroupActivityService) {
				mock.GetExpenseHistoryFunc = func(ctx context.Context, eID pgtype.UUID) ([]sqlc.GetExpenseHistoryRow, error) {
					return []sqlc.GetExpenseHistoryRow{}, nil
				}
			},
			expectedStatus: http.StatusOK,
			validateResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var resp response.StandardResponse[[]ActivityResponse, response.ErrorDetail]
				if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
					t.Fatalf("failed to unmarshal response: %v", err)
				}
				if resp.Data == nil {
					t.Fatal("expected data, got nil")
				}
				if len(*resp.Data) != 0 {
					t.Errorf("expected 0 history items, got %d", len(*resp.Data))
				}
			},
		},
		{
			name:           "error - invalid expense_id",
			expenseID:      "invalid-uuid",
			mockSetup:      func(mock *MockGroupActivityService) {},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &MockGroupActivityService{}
			if tt.mockSetup != nil {
				tt.mockSetup(mock)
			}

			handler := GetExpenseHistoryHandler(mock)
			req := httptest.NewRequest(http.MethodGet, "/groups/10/expenses/"+tt.expenseID+"/history", nil)

			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("expense_id", tt.expenseID)
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
