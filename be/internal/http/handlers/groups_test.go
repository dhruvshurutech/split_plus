package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/dhruvsaxena1998/splitplus/internal/db/sqlc"
	"github.com/dhruvsaxena1998/splitplus/internal/http/middleware"
	"github.com/dhruvsaxena1998/splitplus/internal/http/response"
	"github.com/dhruvsaxena1998/splitplus/internal/service"
	"github.com/dhruvsaxena1998/splitplus/internal/testutil"
)

// MockGroupService is a mock implementation of GroupService for testing
type MockGroupService struct {
	CreateGroupFunc      func(ctx context.Context, input service.CreateGroupInput) (service.CreateGroupResult, error)

	ListGroupMembersFunc func(ctx context.Context, groupID, requesterID pgtype.UUID) ([]sqlc.ListGroupMembersRow, error)
	ListUserGroupsFunc   func(ctx context.Context, userID pgtype.UUID) ([]sqlc.GetGroupsByUserIDRow, error)
}

func (m *MockGroupService) CreateGroup(ctx context.Context, input service.CreateGroupInput) (service.CreateGroupResult, error) {
	if m.CreateGroupFunc != nil {
		return m.CreateGroupFunc(ctx, input)
	}
	return service.CreateGroupResult{}, nil
}



func (m *MockGroupService) ListGroupMembers(ctx context.Context, groupID, requesterID pgtype.UUID) ([]sqlc.ListGroupMembersRow, error) {
	if m.ListGroupMembersFunc != nil {
		return m.ListGroupMembersFunc(ctx, groupID, requesterID)
	}
	return []sqlc.ListGroupMembersRow{}, nil
}

func (m *MockGroupService) ListUserGroups(ctx context.Context, userID pgtype.UUID) ([]sqlc.GetGroupsByUserIDRow, error) {
	if m.ListUserGroupsFunc != nil {
		return m.ListUserGroupsFunc(ctx, userID)
	}
	return []sqlc.GetGroupsByUserIDRow{}, nil
}

var _ service.GroupService = (*MockGroupService)(nil)

// Helper to create request with auth context
func createAuthenticatedRequest(method, path string, body []byte, userID pgtype.UUID) *http.Request {
	var req *http.Request
	if body != nil {
		req = httptest.NewRequest(method, path, bytes.NewReader(body))
	} else {
		req = httptest.NewRequest(method, path, nil)
	}
	req.Header.Set("Content-Type", "application/json")
	ctx := middleware.SetUserID(req.Context(), userID)
	return req.WithContext(ctx)
}

func TestCreateGroupHandler(t *testing.T) {
	userID := testutil.CreateTestUUID(1)
	groupID := testutil.CreateTestUUID(10)

	tests := []struct {
		name             string
		requestBody      CreateGroupRequest
		userID           pgtype.UUID
		mockSetup        func(*MockGroupService)
		expectedStatus   int
		validateResponse func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name: "successful group creation",
			requestBody: CreateGroupRequest{
				Name:         "Test Group",
				Description:  "A test group",
				CurrencyCode: "EUR",
			},
			userID: userID,
			mockSetup: func(mock *MockGroupService) {
				mock.CreateGroupFunc = func(ctx context.Context, input service.CreateGroupInput) (service.CreateGroupResult, error) {
					return service.CreateGroupResult{
						Group:      testutil.CreateTestGroup(groupID, input.Name, input.CreatedBy),
						Membership: testutil.CreateTestGroupMember(testutil.CreateTestUUID(100), groupID, input.CreatedBy, "owner", "active"),
					}, nil
				}
			},
			expectedStatus: http.StatusCreated,
			validateResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var resp response.StandardResponse[CreateGroupResponse, response.ErrorDetail]
				if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
					t.Fatalf("failed to unmarshal response: %v", err)
				}
				if !resp.Status {
					t.Errorf("expected status true, got false")
				}
				if resp.Data == nil {
					t.Fatal("expected data, got nil")
				}
				if resp.Data.Name != "Test Group" {
					t.Errorf("expected name 'Test Group', got '%s'", resp.Data.Name)
				}
				if resp.Data.Role != "owner" {
					t.Errorf("expected role 'owner', got '%s'", resp.Data.Role)
				}
			},
		},
		{
			name: "invalid group name error",
			requestBody: CreateGroupRequest{
				Name: "Test",
			},
			userID: userID,
			mockSetup: func(mock *MockGroupService) {
				mock.CreateGroupFunc = func(ctx context.Context, input service.CreateGroupInput) (service.CreateGroupResult, error) {
					return service.CreateGroupResult{}, service.ErrInvalidGroupName
				}
			},
			expectedStatus: http.StatusUnprocessableEntity,
		},
		{
			name: "service error",
			requestBody: CreateGroupRequest{
				Name: "Test Group",
			},
			userID: userID,
			mockSetup: func(mock *MockGroupService) {
				mock.CreateGroupFunc = func(ctx context.Context, input service.CreateGroupInput) (service.CreateGroupResult, error) {
					return service.CreateGroupResult{}, errors.New("database error")
				}
			},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &MockGroupService{}
			if tt.mockSetup != nil {
				tt.mockSetup(mockService)
			}

			handler := CreateGroupHandler(mockService)
			validate := validator.New()

			body, _ := json.Marshal(tt.requestBody)
			req := createAuthenticatedRequest(http.MethodPost, "/groups", body, tt.userID)

			wrappedHandler := middleware.ValidateBody[CreateGroupRequest](validate)(handler)
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



func TestListGroupMembersHandler(t *testing.T) {
	groupID := testutil.CreateTestUUID(10)
	userID := testutil.CreateTestUUID(1)

	tests := []struct {
		name             string
		groupID          string
		userID           pgtype.UUID
		mockSetup        func(*MockGroupService)
		expectedStatus   int
		validateResponse func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name:    "successful list",
			groupID: uuidToString(groupID),
			userID:  userID,
			mockSetup: func(mock *MockGroupService) {
				mock.ListGroupMembersFunc = func(ctx context.Context, gID, rID pgtype.UUID) ([]sqlc.ListGroupMembersRow, error) {
					return []sqlc.ListGroupMembersRow{
						{ID: testutil.CreateTestUUID(100), UserEmail: "owner@example.com", Role: "owner", Status: "active"},
						{ID: testutil.CreateTestUUID(101), UserEmail: "member@example.com", Role: "member", Status: "active"},
					}, nil
				}
			},
			expectedStatus: http.StatusOK,
			validateResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var resp response.StandardResponse[[]GroupMemberWithUserResponse, response.ErrorDetail]
				if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
					t.Fatalf("failed to unmarshal response: %v", err)
				}
				if !resp.Status {
					t.Errorf("expected status true, got false")
				}
				if len(*resp.Data) != 2 {
					t.Errorf("expected 2 members, got %d", len(*resp.Data))
				}
			},
		},
		{
			name:    "group not found",
			groupID: uuidToString(groupID),
			userID:  userID,
			mockSetup: func(mock *MockGroupService) {
				mock.ListGroupMembersFunc = func(ctx context.Context, gID, rID pgtype.UUID) ([]sqlc.ListGroupMembersRow, error) {
					return nil, service.ErrGroupNotFound
				}
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:    "not a member",
			groupID: uuidToString(groupID),
			userID:  userID,
			mockSetup: func(mock *MockGroupService) {
				mock.ListGroupMembersFunc = func(ctx context.Context, gID, rID pgtype.UUID) ([]sqlc.ListGroupMembersRow, error) {
					return nil, service.ErrNotGroupMember
				}
			},
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "invalid group_id",
			groupID:        "invalid-uuid",
			userID:         userID,
			mockSetup:      func(mock *MockGroupService) {},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &MockGroupService{}
			if tt.mockSetup != nil {
				tt.mockSetup(mockService)
			}

			handler := ListGroupMembersHandler(mockService)

			req := createAuthenticatedRequest(http.MethodGet, "/groups/"+tt.groupID+"/members", nil, tt.userID)

			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("group_id", tt.groupID)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			w := httptest.NewRecorder()
			handler.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d. Body: %s", tt.expectedStatus, w.Code, w.Body.String())
			}

			if tt.validateResponse != nil {
				tt.validateResponse(t, w)
			}
		})
	}
}

func TestListUserGroupsHandler(t *testing.T) {
	userID := testutil.CreateTestUUID(1)
	groupID1 := testutil.CreateTestUUID(10)
	groupID2 := testutil.CreateTestUUID(11)
	membershipID1 := testutil.CreateTestUUID(100)
	membershipID2 := testutil.CreateTestUUID(101)

	tests := []struct {
		name             string
		userID           pgtype.UUID
		mockSetup        func(*MockGroupService)
		expectedStatus   int
		validateResponse func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name:   "successful list",
			userID: userID,
			mockSetup: func(mock *MockGroupService) {
				mock.ListUserGroupsFunc = func(ctx context.Context, uID pgtype.UUID) ([]sqlc.GetGroupsByUserIDRow, error) {
					return []sqlc.GetGroupsByUserIDRow{
						{
							ID:             groupID1,
							Name:           "Group 1",
							Description:    pgtype.Text{String: "First group", Valid: true},
							CurrencyCode:   "USD",
							CreatedAt:      pgtype.Timestamptz{Time: parseTestTime("2024-01-01T00:00:00Z"), Valid: true},
							MembershipID:   membershipID1,
							MemberRole:     "owner",
							MemberStatus:   "active",
							MemberJoinedAt: pgtype.Timestamptz{Time: parseTestTime("2024-01-01T00:00:00Z"), Valid: true},
						},
						{
							ID:             groupID2,
							Name:           "Group 2",
							Description:    pgtype.Text{String: "Second group", Valid: true},
							CurrencyCode:   "EUR",
							CreatedAt:      pgtype.Timestamptz{Time: parseTestTime("2024-01-02T00:00:00Z"), Valid: true},
							MembershipID:   membershipID2,
							MemberRole:     "member",
							MemberStatus:   "active",
							MemberJoinedAt: pgtype.Timestamptz{Time: parseTestTime("2024-01-02T00:00:00Z"), Valid: true},
						},
					}, nil
				}
			},
			expectedStatus: http.StatusOK,
			validateResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var resp response.StandardResponse[[]UserGroupResponse, response.ErrorDetail]
				if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
					t.Fatalf("failed to unmarshal response: %v", err)
				}
				if !resp.Status {
					t.Errorf("expected status true, got false")
				}
				if resp.Data == nil {
					t.Fatal("expected data, got nil")
				}
				if len(*resp.Data) != 2 {
					t.Errorf("expected 2 groups, got %d", len(*resp.Data))
				}
				if (*resp.Data)[0].Name != "Group 1" {
					t.Errorf("expected first group name 'Group 1', got '%s'", (*resp.Data)[0].Name)
				}
				if (*resp.Data)[0].MemberRole != "owner" {
					t.Errorf("expected first group role 'owner', got '%s'", (*resp.Data)[0].MemberRole)
				}
				if (*resp.Data)[1].Name != "Group 2" {
					t.Errorf("expected second group name 'Group 2', got '%s'", (*resp.Data)[1].Name)
				}
				if (*resp.Data)[1].MemberRole != "member" {
					t.Errorf("expected second group role 'member', got '%s'", (*resp.Data)[1].MemberRole)
				}
			},
		},
		{
			name:   "empty list",
			userID: userID,
			mockSetup: func(mock *MockGroupService) {
				mock.ListUserGroupsFunc = func(ctx context.Context, uID pgtype.UUID) ([]sqlc.GetGroupsByUserIDRow, error) {
					return []sqlc.GetGroupsByUserIDRow{}, nil
				}
			},
			expectedStatus: http.StatusOK,
			validateResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var resp response.StandardResponse[[]UserGroupResponse, response.ErrorDetail]
				if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
					t.Fatalf("failed to unmarshal response: %v", err)
				}
				if !resp.Status {
					t.Errorf("expected status true, got false")
				}
				if resp.Data == nil {
					t.Fatal("expected data, got nil")
				}
				if len(*resp.Data) != 0 {
					t.Errorf("expected 0 groups, got %d", len(*resp.Data))
				}
			},
		},
		{
			name:   "service error",
			userID: userID,
			mockSetup: func(mock *MockGroupService) {
				mock.ListUserGroupsFunc = func(ctx context.Context, uID pgtype.UUID) ([]sqlc.GetGroupsByUserIDRow, error) {
					return nil, errors.New("database error")
				}
			},
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:           "unauthorized - no user ID",
			userID:         pgtype.UUID{Valid: false},
			mockSetup:      func(mock *MockGroupService) {},
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &MockGroupService{}
			if tt.mockSetup != nil {
				tt.mockSetup(mockService)
			}

			handler := ListUserGroupsHandler(mockService)

			var req *http.Request
			if !tt.userID.Valid {
				// For unauthorized test, create request without auth context
				req = httptest.NewRequest(http.MethodGet, "/groups", nil)
			} else {
				req = createAuthenticatedRequest(http.MethodGet, "/groups", nil, tt.userID)
			}

			w := httptest.NewRecorder()
			handler.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d. Body: %s", tt.expectedStatus, w.Code, w.Body.String())
			}

			if tt.validateResponse != nil {
				tt.validateResponse(t, w)
			}
		})
	}
}

// Helper to convert UUID to string
func uuidToString(uuid pgtype.UUID) string {
	if !uuid.Valid {
		return ""
	}
	return formatUUID(uuid.Bytes)
}

func formatUUID(b [16]byte) string {
	return string([]byte{
		hexDigit(b[0] >> 4), hexDigit(b[0]),
		hexDigit(b[1] >> 4), hexDigit(b[1]),
		hexDigit(b[2] >> 4), hexDigit(b[2]),
		hexDigit(b[3] >> 4), hexDigit(b[3]),
		'-',
		hexDigit(b[4] >> 4), hexDigit(b[4]),
		hexDigit(b[5] >> 4), hexDigit(b[5]),
		'-',
		hexDigit(b[6] >> 4), hexDigit(b[6]),
		hexDigit(b[7] >> 4), hexDigit(b[7]),
		'-',
		hexDigit(b[8] >> 4), hexDigit(b[8]),
		hexDigit(b[9] >> 4), hexDigit(b[9]),
		'-',
		hexDigit(b[10] >> 4), hexDigit(b[10]),
		hexDigit(b[11] >> 4), hexDigit(b[11]),
		hexDigit(b[12] >> 4), hexDigit(b[12]),
		hexDigit(b[13] >> 4), hexDigit(b[13]),
		hexDigit(b[14] >> 4), hexDigit(b[14]),
		hexDigit(b[15] >> 4), hexDigit(b[15]),
	})
}

func hexDigit(b byte) byte {
	b = b & 0x0f
	if b < 10 {
		return '0' + b
	}
	return 'a' + b - 10
}

func parseTestTime(timeStr string) time.Time {
	t, _ := time.Parse(time.RFC3339, timeStr)
	return t
}
