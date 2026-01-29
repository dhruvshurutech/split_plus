package handlers

import (
	"context"
	"encoding/json"
	"errors"
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

// MockExpenseCommentService for testing
type MockExpenseCommentService struct {
	CreateCommentFunc func(ctx context.Context, expenseID, userID pgtype.UUID, comment string) (sqlc.ExpenseComment, error)
	ListCommentsFunc  func(ctx context.Context, expenseID pgtype.UUID) ([]sqlc.ListExpenseCommentsRow, error)
	UpdateCommentFunc func(ctx context.Context, commentID, userID pgtype.UUID, comment string) (sqlc.ExpenseComment, error)
	DeleteCommentFunc func(ctx context.Context, commentID, userID pgtype.UUID) error
}

func (m *MockExpenseCommentService) CreateComment(ctx context.Context, expenseID, userID pgtype.UUID, comment string) (sqlc.ExpenseComment, error) {
	if m.CreateCommentFunc != nil {
		return m.CreateCommentFunc(ctx, expenseID, userID, comment)
	}
	return sqlc.ExpenseComment{}, nil
}

func (m *MockExpenseCommentService) ListComments(ctx context.Context, expenseID pgtype.UUID) ([]sqlc.ListExpenseCommentsRow, error) {
	if m.ListCommentsFunc != nil {
		return m.ListCommentsFunc(ctx, expenseID)
	}
	return []sqlc.ListExpenseCommentsRow{}, nil
}

func (m *MockExpenseCommentService) UpdateComment(ctx context.Context, commentID, userID pgtype.UUID, comment string) (sqlc.ExpenseComment, error) {
	if m.UpdateCommentFunc != nil {
		return m.UpdateCommentFunc(ctx, commentID, userID, comment)
	}
	return sqlc.ExpenseComment{}, nil
}

func (m *MockExpenseCommentService) DeleteComment(ctx context.Context, commentID, userID pgtype.UUID) error {
	if m.DeleteCommentFunc != nil {
		return m.DeleteCommentFunc(ctx, commentID, userID)
	}
	return nil
}

var _ service.ExpenseCommentService = (*MockExpenseCommentService)(nil)

func TestCreateCommentHandler(t *testing.T) {
	userID := testutil.CreateTestUUID(1)
	expenseID := testutil.CreateTestUUID(100)
	groupID := testutil.CreateTestUUID(10)

	tests := []struct {
		name             string
		requestBody      CommentRequest
		expenseID        string
		userID           pgtype.UUID
		mockSetup        func(*MockExpenseCommentService)
		expectedStatus   int
		validateResponse func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name:        "success - create comment",
			requestBody: CommentRequest{Comment: "Great expense!"},
			expenseID:   expenseID.String(),
			userID:      userID,
			mockSetup: func(mock *MockExpenseCommentService) {
				mock.CreateCommentFunc = func(ctx context.Context, eID, uID pgtype.UUID, comment string) (sqlc.ExpenseComment, error) {
					return sqlc.ExpenseComment{
						ID:        testutil.CreateTestUUID(200),
						ExpenseID: eID,
						UserID:    uID,
						Comment:   comment,
					}, nil
				}
			},
			expectedStatus: http.StatusCreated,
			validateResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var resp response.StandardResponse[CommentResponse, response.ErrorDetail]
				if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
					t.Fatalf("failed to unmarshal response: %v", err)
				}
				if !resp.Status {
					t.Error("expected status true")
				}
				if resp.Data == nil {
					t.Fatal("expected data, got nil")
				}
				if resp.Data.Comment != "Great expense!" {
					t.Errorf("expected comment 'Great expense!', got '%s'", resp.Data.Comment)
				}
			},
		},
		{
			name:           "error - empty comment",
			requestBody:    CommentRequest{Comment: ""},
			expenseID:      expenseID.String(),
			userID:         userID,
			mockSetup:      func(mock *MockExpenseCommentService) {},
			expectedStatus: http.StatusUnprocessableEntity,
		},
		{
			name:        "error - expense not found",
			requestBody: CommentRequest{Comment: "Test comment"},
			expenseID:   expenseID.String(),
			userID:      userID,
			mockSetup: func(mock *MockExpenseCommentService) {
				mock.CreateCommentFunc = func(ctx context.Context, eID, uID pgtype.UUID, comment string) (sqlc.ExpenseComment, error) {
					return sqlc.ExpenseComment{}, service.ErrExpenseNotFound
				}
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:        "error - not group member",
			requestBody: CommentRequest{Comment: "Test comment"},
			expenseID:   expenseID.String(),
			userID:      userID,
			mockSetup: func(mock *MockExpenseCommentService) {
				mock.CreateCommentFunc = func(ctx context.Context, eID, uID pgtype.UUID, comment string) (sqlc.ExpenseComment, error) {
					return sqlc.ExpenseComment{}, service.ErrNotGroupMember
				}
			},
			expectedStatus: http.StatusForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &MockExpenseCommentService{}
			if tt.mockSetup != nil {
				tt.mockSetup(mock)
			}

			v := validator.New()
			handler := CreateCommentHandler(mock)

			body, _ := json.Marshal(tt.requestBody)
			req := createAuthenticatedRequest(http.MethodPost, "/groups/"+groupID.String()+"/expenses/"+tt.expenseID+"/comments", body, tt.userID)

			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("expense_id", tt.expenseID)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			wrappedHandler := middleware.ValidateBody[CommentRequest](v)(handler)
			w := httptest.NewRecorder()
			wrappedHandler.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d. Body: %s", tt.expectedStatus, w.Code, w.Body.String())
			}
			if tt.validateResponse != nil {
				tt.validateResponse(t, w)
			}
		})
	}
}

func TestListCommentsHandler(t *testing.T) {
	expenseID := testutil.CreateTestUUID(100)

	tests := []struct {
		name             string
		expenseID        string
		mockSetup        func(*MockExpenseCommentService)
		expectedStatus   int
		validateResponse func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name:      "success - list comments",
			expenseID: expenseID.String(),
			mockSetup: func(mock *MockExpenseCommentService) {
				mock.ListCommentsFunc = func(ctx context.Context, eID pgtype.UUID) ([]sqlc.ListExpenseCommentsRow, error) {
					return []sqlc.ListExpenseCommentsRow{
						{
							ID:        testutil.CreateTestUUID(200),
							Comment:   "First comment",
							UserEmail: "user1@example.com",
						},
						{
							ID:        testutil.CreateTestUUID(201),
							Comment:   "Second comment",
							UserEmail: "user2@example.com",
						},
					}, nil
				}
			},
			expectedStatus: http.StatusOK,
			validateResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var resp response.StandardResponse[[]CommentResponse, response.ErrorDetail]
				if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
					t.Fatalf("failed to unmarshal response: %v", err)
				}
				if resp.Data == nil {
					t.Fatal("expected data, got nil")
				}
				if len(*resp.Data) != 2 {
					t.Errorf("expected 2 comments, got %d", len(*resp.Data))
				}
			},
		},
		{
			name:      "success - empty list",
			expenseID: expenseID.String(),
			mockSetup: func(mock *MockExpenseCommentService) {
				mock.ListCommentsFunc = func(ctx context.Context, eID pgtype.UUID) ([]sqlc.ListExpenseCommentsRow, error) {
					return []sqlc.ListExpenseCommentsRow{}, nil
				}
			},
			expectedStatus: http.StatusOK,
			validateResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var resp response.StandardResponse[[]CommentResponse, response.ErrorDetail]
				if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
					t.Fatalf("failed to unmarshal response: %v", err)
				}
				if resp.Data == nil {
					t.Fatal("expected data, got nil")
				}
				if len(*resp.Data) != 0 {
					t.Errorf("expected 0 comments, got %d", len(*resp.Data))
				}
			},
		},
		{
			name:           "error - invalid expense_id",
			expenseID:      "invalid-uuid",
			mockSetup:      func(mock *MockExpenseCommentService) {},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &MockExpenseCommentService{}
			if tt.mockSetup != nil {
				tt.mockSetup(mock)
			}

			handler := ListCommentsHandler(mock)
			req := httptest.NewRequest(http.MethodGet, "/groups/10/expenses/"+tt.expenseID+"/comments", nil)

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

func TestUpdateCommentHandler(t *testing.T) {
	userID := testutil.CreateTestUUID(1)
	commentID := testutil.CreateTestUUID(200)

	tests := []struct {
		name           string
		requestBody    CommentRequest
		commentID      string
		userID         pgtype.UUID
		mockSetup      func(*MockExpenseCommentService)
		expectedStatus int
	}{
		{
			name:        "success - update comment",
			requestBody: CommentRequest{Comment: "Updated comment"},
			commentID:   commentID.String(),
			userID:      userID,
			mockSetup: func(mock *MockExpenseCommentService) {
				mock.UpdateCommentFunc = func(ctx context.Context, cID, uID pgtype.UUID, comment string) (sqlc.ExpenseComment, error) {
					return sqlc.ExpenseComment{
						ID:      cID,
						UserID:  uID,
						Comment: comment,
					}, nil
				}
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "error - empty comment",
			requestBody:    CommentRequest{Comment: ""},
			commentID:      commentID.String(),
			userID:         userID,
			mockSetup:      func(mock *MockExpenseCommentService) {},
			expectedStatus: http.StatusUnprocessableEntity,
		},
		{
			name:        "error - permission denied",
			requestBody: CommentRequest{Comment: "Updated comment"},
			commentID:   commentID.String(),
			userID:      userID,
			mockSetup: func(mock *MockExpenseCommentService) {
				mock.UpdateCommentFunc = func(ctx context.Context, cID, uID pgtype.UUID, comment string) (sqlc.ExpenseComment, error) {
					return sqlc.ExpenseComment{}, service.ErrCommentPermissioDenied
				}
			},
			expectedStatus: http.StatusForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &MockExpenseCommentService{}
			if tt.mockSetup != nil {
				tt.mockSetup(mock)
			}

			v := validator.New()
			handler := UpdateCommentHandler(mock)

			body, _ := json.Marshal(tt.requestBody)
			req := createAuthenticatedRequest(http.MethodPut, "/groups/10/expenses/100/comments/"+tt.commentID, body, tt.userID)

			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("comment_id", tt.commentID)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			wrappedHandler := middleware.ValidateBody[CommentRequest](v)(handler)
			w := httptest.NewRecorder()
			wrappedHandler.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}
}

func TestDeleteCommentHandler(t *testing.T) {
	userID := testutil.CreateTestUUID(1)
	commentID := testutil.CreateTestUUID(200)

	tests := []struct {
		name           string
		commentID      string
		userID         pgtype.UUID
		mockSetup      func(*MockExpenseCommentService)
		expectedStatus int
	}{
		{
			name:      "success - delete comment",
			commentID: commentID.String(),
			userID:    userID,
			mockSetup: func(mock *MockExpenseCommentService) {
				mock.DeleteCommentFunc = func(ctx context.Context, cID, uID pgtype.UUID) error {
					return nil
				}
			},
			expectedStatus: http.StatusNoContent,
		},
		{
			name:      "error - permission denied",
			commentID: commentID.String(),
			userID:    userID,
			mockSetup: func(mock *MockExpenseCommentService) {
				mock.DeleteCommentFunc = func(ctx context.Context, cID, uID pgtype.UUID) error {
					return service.ErrCommentPermissioDenied
				}
			},
			expectedStatus: http.StatusForbidden,
		},
		{
			name:      "error - comment not found",
			commentID: commentID.String(),
			userID:    userID,
			mockSetup: func(mock *MockExpenseCommentService) {
				mock.DeleteCommentFunc = func(ctx context.Context, cID, uID pgtype.UUID) error {
					return errors.New("not found")
				}
			},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &MockExpenseCommentService{}
			if tt.mockSetup != nil {
				tt.mockSetup(mock)
			}

			handler := DeleteCommentHandler(mock)
			req := createAuthenticatedRequest(http.MethodDelete, "/groups/10/expenses/100/comments/"+tt.commentID, nil, tt.userID)

			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("comment_id", tt.commentID)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			w := httptest.NewRecorder()
			handler(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}
}
