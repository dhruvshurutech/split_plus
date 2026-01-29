package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/dhruvsaxena1998/splitplus/internal/db/sqlc"
	"github.com/dhruvsaxena1998/splitplus/internal/http/middleware"
	"github.com/dhruvsaxena1998/splitplus/internal/service"
	"github.com/dhruvsaxena1998/splitplus/internal/testutil"
)

type MockFriendService struct {
	SendFriendRequestFunc    func(ctx context.Context, requesterID, friendID pgtype.UUID) (sqlc.Friendship, error)
	ListFriendsFunc          func(ctx context.Context, userID pgtype.UUID) ([]service.FriendSummary, error)
	ListIncomingRequestsFunc func(ctx context.Context, userID pgtype.UUID) ([]sqlc.Friendship, error)
	ListOutgoingRequestsFunc func(ctx context.Context, userID pgtype.UUID) ([]sqlc.Friendship, error)
	AcceptFriendRequestFunc  func(ctx context.Context, requesterID, friendshipID pgtype.UUID) (sqlc.Friendship, error)
	DeclineFriendRequestFunc func(ctx context.Context, requesterID, friendshipID pgtype.UUID) error
	RemoveFriendFunc         func(ctx context.Context, requesterID, friendID pgtype.UUID) error
}

func (m *MockFriendService) SendFriendRequest(ctx context.Context, requesterID, friendID pgtype.UUID) (sqlc.Friendship, error) {
	if m.SendFriendRequestFunc != nil {
		return m.SendFriendRequestFunc(ctx, requesterID, friendID)
	}
	return sqlc.Friendship{}, nil
}

func (m *MockFriendService) ListFriends(ctx context.Context, userID pgtype.UUID) ([]service.FriendSummary, error) {
	if m.ListFriendsFunc != nil {
		return m.ListFriendsFunc(ctx, userID)
	}
	return nil, nil
}

func (m *MockFriendService) ListIncomingRequests(ctx context.Context, userID pgtype.UUID) ([]sqlc.Friendship, error) {
	if m.ListIncomingRequestsFunc != nil {
		return m.ListIncomingRequestsFunc(ctx, userID)
	}
	return nil, nil
}

func (m *MockFriendService) ListOutgoingRequests(ctx context.Context, userID pgtype.UUID) ([]sqlc.Friendship, error) {
	if m.ListOutgoingRequestsFunc != nil {
		return m.ListOutgoingRequestsFunc(ctx, userID)
	}
	return nil, nil
}

func (m *MockFriendService) AcceptFriendRequest(ctx context.Context, requesterID, friendshipID pgtype.UUID) (sqlc.Friendship, error) {
	if m.AcceptFriendRequestFunc != nil {
		return m.AcceptFriendRequestFunc(ctx, requesterID, friendshipID)
	}
	return sqlc.Friendship{}, nil
}

func (m *MockFriendService) DeclineFriendRequest(ctx context.Context, requesterID, friendshipID pgtype.UUID) error {
	if m.DeclineFriendRequestFunc != nil {
		return m.DeclineFriendRequestFunc(ctx, requesterID, friendshipID)
	}
	return nil
}

func (m *MockFriendService) RemoveFriend(ctx context.Context, requesterID, friendID pgtype.UUID) error {
	if m.RemoveFriendFunc != nil {
		return m.RemoveFriendFunc(ctx, requesterID, friendID)
	}
	return nil
}

var _ service.FriendService = (*MockFriendService)(nil)

func createFriendRequest(method, path string, body interface{}, userID pgtype.UUID) *http.Request {
	var req *http.Request
	if body != nil {
		bodyBytes, _ := json.Marshal(body)
		req = httptest.NewRequest(method, path, bytes.NewReader(bodyBytes))
	} else {
		req = httptest.NewRequest(method, path, nil)
	}
	req.Header.Set("Content-Type", "application/json")
	ctx := middleware.SetUserID(req.Context(), userID)
	return req.WithContext(ctx)
}

func TestSendFriendRequestHandler_ValidationAndErrors(t *testing.T) {
	userID := testutil.CreateTestUUID(1)
	friendID := testutil.CreateTestUUID(2)

	tests := []struct {
		name           string
		body           interface{}
		mockSetup      func(*MockFriendService)
		expectedStatus int
	}{
		{
			name: "invalid friend_id format",
			body: SendFriendRequestRequest{
				FriendID: "not-a-uuid",
			},
			mockSetup: func(m *MockFriendService) {},
			// Fails in validation middleware due to invalid UUID
			expectedStatus: http.StatusUnprocessableEntity,
		},
		{
			name: "service error - invalid action",
			body: SendFriendRequestRequest{
				FriendID: friendID.String(),
			},
			mockSetup: func(m *MockFriendService) {
				m.SendFriendRequestFunc = func(ctx context.Context, requesterID, fID pgtype.UUID) (sqlc.Friendship, error) {
					return sqlc.Friendship{}, service.ErrInvalidFriendAction
				}
			},
			expectedStatus: http.StatusUnprocessableEntity,
		},
		{
			name: "service error - already exists",
			body: SendFriendRequestRequest{
				FriendID: friendID.String(),
			},
			mockSetup: func(m *MockFriendService) {
				m.SendFriendRequestFunc = func(ctx context.Context, requesterID, fID pgtype.UUID) (sqlc.Friendship, error) {
					return sqlc.Friendship{}, service.ErrFriendRequestExists
				}
			},
			expectedStatus: http.StatusConflict,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &MockFriendService{}
			tt.mockSetup(mock)

			v := validator.New()
			base := SendFriendRequestHandler(mock)
			handler := middleware.ValidateBody[SendFriendRequestRequest](v)(base)

			req := createFriendRequest("POST", "/friends/requests", tt.body, userID)
			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Fatalf("expected status %d, got %d", tt.expectedStatus, rr.Code)
			}
		})
	}
}

func TestListFriendsHandler_Success(t *testing.T) {
	userID := testutil.CreateTestUUID(1)
	friendID := testutil.CreateTestUUID(2)

	mock := &MockFriendService{
		ListFriendsFunc: func(ctx context.Context, uid pgtype.UUID) ([]service.FriendSummary, error) {
			return []service.FriendSummary{
				{
					ID:          testutil.CreateTestUUID(10),
					UserID:      uid,
					FriendID:    friendID,
					FriendEmail: "friend@example.com",
					FriendName:  "Friend",
				},
			}, nil
		},
	}

	req := createFriendRequest("GET", "/friends", nil, userID)
	rr := httptest.NewRecorder()

	handler := ListFriendsHandler(mock)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rr.Code)
	}
}

func TestAcceptFriendRequestHandler_ErrorMapping(t *testing.T) {
	userID := testutil.CreateTestUUID(1)
	requestID := testutil.CreateTestUUID(3)

	tests := []struct {
		name           string
		mockSetup      func(*MockFriendService)
		expectedStatus int
	}{
		{
			name: "not found",
			mockSetup: func(m *MockFriendService) {
				m.AcceptFriendRequestFunc = func(ctx context.Context, requesterID, fid pgtype.UUID) (sqlc.Friendship, error) {
					return sqlc.Friendship{}, service.ErrFriendNotFound
				}
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name: "invalid action",
			mockSetup: func(m *MockFriendService) {
				m.AcceptFriendRequestFunc = func(ctx context.Context, requesterID, fid pgtype.UUID) (sqlc.Friendship, error) {
					return sqlc.Friendship{}, service.ErrInvalidFriendAction
				}
			},
			expectedStatus: http.StatusUnprocessableEntity,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &MockFriendService{}
			tt.mockSetup(mock)

			req := createFriendRequest("POST", "/friends/requests/"+requestID.String()+"/accept", nil, userID)
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", requestID.String())
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			rr := httptest.NewRecorder()
			handler := AcceptFriendRequestHandler(mock)
			handler.ServeHTTP(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Fatalf("expected %d, got %d", tt.expectedStatus, rr.Code)
			}
		})
	}
}
