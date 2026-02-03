package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/dhruvsaxena1998/splitplus/internal/db/sqlc"
	"github.com/dhruvsaxena1998/splitplus/internal/testutil"
)

func TestGroupService_CreateGroup(t *testing.T) {
	userID := testutil.CreateTestUUID(1)
	groupID := testutil.CreateTestUUID(10)
	memberID := testutil.CreateTestUUID(100)

	tests := []struct {
		name          string
		input         CreateGroupInput
		mockSetup     func(*testutil.MockGroupRepository)
		expectedError error
		validate      func(*testing.T, CreateGroupResult)
	}{
		{
			name: "successful group creation",
			input: CreateGroupInput{
				Name:         "Test Group",
				Description:  "A test group",
				CurrencyCode: "EUR",
				CreatedBy:    userID,
			},
			mockSetup: func(mock *testutil.MockGroupRepository) {
				mock.CreateGroupFunc = func(ctx context.Context, params sqlc.CreateGroupParams) (sqlc.Group, error) {
					return testutil.CreateTestGroup(groupID, params.Name, params.CreatedBy), nil
				}
				mock.CreateGroupMemberFunc = func(ctx context.Context, params sqlc.CreateGroupMemberParams) (sqlc.GroupMember, error) {
					return testutil.CreateTestGroupMember(memberID, params.GroupID, params.UserID, params.Role, params.Status), nil
				}
			},
			expectedError: nil,
			validate: func(t *testing.T, result CreateGroupResult) {
				if result.Group.Name != "Test Group" {
					t.Errorf("expected group name 'Test Group', got '%s'", result.Group.Name)
				}
				if result.Membership.Role != "owner" {
					t.Errorf("expected role 'owner', got '%s'", result.Membership.Role)
				}
				if result.Membership.Status != "active" {
					t.Errorf("expected status 'active', got '%s'", result.Membership.Status)
				}
			},
		},
		{
			name: "empty name",
			input: CreateGroupInput{
				Name:      "",
				CreatedBy: userID,
			},
			mockSetup:     func(mock *testutil.MockGroupRepository) {},
			expectedError: ErrInvalidGroupName,
		},
		{
			name: "whitespace only name",
			input: CreateGroupInput{
				Name:      "   ",
				CreatedBy: userID,
			},
			mockSetup:     func(mock *testutil.MockGroupRepository) {},
			expectedError: ErrInvalidGroupName,
		},
		{
			name: "default currency code",
			input: CreateGroupInput{
				Name:         "Test Group",
				CurrencyCode: "", // empty should default to USD
				CreatedBy:    userID,
			},
			mockSetup: func(mock *testutil.MockGroupRepository) {
				mock.CreateGroupFunc = func(ctx context.Context, params sqlc.CreateGroupParams) (sqlc.Group, error) {
					if params.CurrencyCode != "USD" {
						t.Errorf("expected currency code 'USD', got '%s'", params.CurrencyCode)
					}
					return testutil.CreateTestGroup(groupID, params.Name, params.CreatedBy), nil
				}
				mock.CreateGroupMemberFunc = func(ctx context.Context, params sqlc.CreateGroupMemberParams) (sqlc.GroupMember, error) {
					return testutil.CreateTestGroupMember(memberID, params.GroupID, params.UserID, params.Role, params.Status), nil
				}
			},
			expectedError: nil,
		},
		{
			name: "create group fails",
			input: CreateGroupInput{
				Name:      "Test Group",
				CreatedBy: userID,
			},
			mockSetup: func(mock *testutil.MockGroupRepository) {
				mock.CreateGroupFunc = func(ctx context.Context, params sqlc.CreateGroupParams) (sqlc.Group, error) {
					return sqlc.Group{}, errors.New("database error")
				}
			},
			expectedError: errors.New("database error"),
		},
		{
			name: "create member fails",
			input: CreateGroupInput{
				Name:      "Test Group",
				CreatedBy: userID,
			},
			mockSetup: func(mock *testutil.MockGroupRepository) {
				mock.CreateGroupFunc = func(ctx context.Context, params sqlc.CreateGroupParams) (sqlc.Group, error) {
					return testutil.CreateTestGroup(groupID, params.Name, params.CreatedBy), nil
				}
				mock.CreateGroupMemberFunc = func(ctx context.Context, params sqlc.CreateGroupMemberParams) (sqlc.GroupMember, error) {
					return sqlc.GroupMember{}, errors.New("member creation failed")
				}
			},
			expectedError: errors.New("member creation failed"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &testutil.MockGroupRepository{}
			tt.mockSetup(mockRepo)

			mockActivitySvc := &MockGroupActivityService{}
			svc := NewGroupService(mockRepo, &testutil.MockGroupInvitationRepository{}, mockActivitySvc)
			result, err := svc.CreateGroup(context.Background(), tt.input)

			if tt.expectedError != nil {
				if err == nil {
					t.Errorf("expected error %v, got nil", tt.expectedError)
					return
				}
				if !errors.Is(err, tt.expectedError) && err.Error() != tt.expectedError.Error() {
					t.Errorf("expected error %v, got %v", tt.expectedError, err)
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if tt.validate != nil {
				tt.validate(t, result)
			}
		})
	}
}





func TestGroupService_ListGroupMembers(t *testing.T) {
	groupID := testutil.CreateTestUUID(10)
	ownerID := testutil.CreateTestUUID(1)
	requesterID := testutil.CreateTestUUID(2)
	membershipID := testutil.CreateTestUUID(100)

	tests := []struct {
		name         string
		groupID      pgtype.UUID
		requesterID  pgtype.UUID
		mockSetup    func(*testutil.MockGroupRepository)
		mockInvSetup func(*testutil.MockGroupInvitationRepository)
		expectedError error
		expectedLen   int
	}{
		{
			name:        "successful list",
			groupID:     groupID,
			requesterID: requesterID,
			mockSetup: func(mockRepo *testutil.MockGroupRepository) {
				mockRepo.GetGroupByIDFunc = func(ctx context.Context, id pgtype.UUID) (sqlc.Group, error) {
					return testutil.CreateTestGroup(groupID, "Test Group", ownerID), nil
				}
				mockRepo.GetGroupMemberFunc = func(ctx context.Context, params sqlc.GetGroupMemberParams) (sqlc.GroupMember, error) {
					return testutil.CreateTestGroupMember(membershipID, groupID, requesterID, "member", "active"), nil
				}
				mockRepo.ListGroupMembersFunc = func(ctx context.Context, gID pgtype.UUID) ([]sqlc.ListGroupMembersRow, error) {
					return []sqlc.ListGroupMembersRow{
						{ID: testutil.CreateTestUUID(100), UserEmail: "owner@example.com", Role: "owner", Status: "active"},
						{ID: testutil.CreateTestUUID(101), UserEmail: "member@example.com", Role: "member", Status: "active"},
					}, nil
				}
			},
			mockInvSetup: func(mockInv *testutil.MockGroupInvitationRepository) {
				mockInv.ListInvitationsByGroupFunc = func(ctx context.Context, gID pgtype.UUID) ([]sqlc.ListInvitationsByGroupRow, error) {
					return []sqlc.ListInvitationsByGroupRow{
						{
							ID:              testutil.CreateTestUUID(102),
							GroupID:         groupID,
							Email:           "invited@example.com",
							Status:          "pending",
							Role:            "member",
							CreatedAt:       pgtype.Timestamptz{Time: time.Now(), Valid: true},
							PendingUserName: pgtype.Text{String: "Invited User", Valid: true},
							PendingUserID:   testutil.CreateTestUUID(200),
						},
					}, nil
				}
			},
			expectedError: nil,
			expectedLen:   2,
		},
		{
			name:        "group not found",
			groupID:     groupID,
			requesterID: requesterID,
			mockSetup: func(mock *testutil.MockGroupRepository) {
				mock.GetGroupByIDFunc = func(ctx context.Context, id pgtype.UUID) (sqlc.Group, error) {
					return sqlc.Group{}, errors.New("not found")
				}
			},
			expectedError: ErrGroupNotFound,
		},
		{
			name:        "requester not a member",
			groupID:     groupID,
			requesterID: requesterID,
			mockSetup: func(mock *testutil.MockGroupRepository) {
				mock.GetGroupByIDFunc = func(ctx context.Context, id pgtype.UUID) (sqlc.Group, error) {
					return testutil.CreateTestGroup(groupID, "Test Group", ownerID), nil
				}
				mock.GetGroupMemberFunc = func(ctx context.Context, params sqlc.GetGroupMemberParams) (sqlc.GroupMember, error) {
					return sqlc.GroupMember{}, errors.New("not found")
				}
			},
			expectedError: ErrNotGroupMember,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &testutil.MockGroupRepository{}
			tt.mockSetup(mockRepo)

			mockActivitySvc := &MockGroupActivityService{}
			mockInvRepo := &testutil.MockGroupInvitationRepository{}
			svc := NewGroupService(mockRepo, mockInvRepo, mockActivitySvc)
			result, err := svc.ListGroupMembers(context.Background(), tt.groupID, tt.requesterID)

			if tt.expectedError != nil {
				if err == nil {
					t.Errorf("expected error %v, got nil", tt.expectedError)
					return
				}
				if !errors.Is(err, tt.expectedError) && err.Error() != tt.expectedError.Error() {
					t.Errorf("expected error %v, got %v", tt.expectedError, err)
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if len(result) != tt.expectedLen {
				t.Errorf("expected %d members, got %d", tt.expectedLen, len(result))
			}
		})
	}
}

func TestGroupService_ListUserGroups(t *testing.T) {
	userID := testutil.CreateTestUUID(1)
	groupID1 := testutil.CreateTestUUID(10)
	groupID2 := testutil.CreateTestUUID(11)

	tests := []struct {
		name          string
		userID        pgtype.UUID
		mockSetup     func(*testutil.MockGroupRepository)
		expectedError error
		expectedLen   int
		validate      func(*testing.T, []sqlc.GetGroupsByUserIDRow)
	}{
		{
			name:   "successful list",
			userID: userID,
			mockSetup: func(mock *testutil.MockGroupRepository) {
				mock.GetGroupsByUserIDFunc = func(ctx context.Context, uID pgtype.UUID) ([]sqlc.GetGroupsByUserIDRow, error) {
					parseTime := func(s string) time.Time {
						t, _ := time.Parse(time.RFC3339, s)
						return t
					}
					return []sqlc.GetGroupsByUserIDRow{
						{
							ID:             groupID1,
							Name:           "Group 1",
							Description:    pgtype.Text{String: "First group", Valid: true},
							CurrencyCode:   "USD",
							CreatedAt:      pgtype.Timestamptz{Time: parseTime("2024-01-01T00:00:00Z"), Valid: true},
							MembershipID:   testutil.CreateTestUUID(100),
							MemberRole:     "owner",
							MemberStatus:   "active",
							MemberJoinedAt: pgtype.Timestamptz{Time: parseTime("2024-01-01T00:00:00Z"), Valid: true},
						},
						{
							ID:             groupID2,
							Name:           "Group 2",
							Description:    pgtype.Text{String: "Second group", Valid: true},
							CurrencyCode:   "EUR",
							CreatedAt:      pgtype.Timestamptz{Time: parseTime("2024-01-02T00:00:00Z"), Valid: true},
							MembershipID:   testutil.CreateTestUUID(101),
							MemberRole:     "member",
							MemberStatus:   "active",
							MemberJoinedAt: pgtype.Timestamptz{Time: parseTime("2024-01-02T00:00:00Z"), Valid: true},
						},
					}, nil
				}
			},
			expectedError: nil,
			expectedLen:   2,
			validate: func(t *testing.T, result []sqlc.GetGroupsByUserIDRow) {
				if len(result) != 2 {
					t.Errorf("expected 2 groups, got %d", len(result))
				}
				if result[0].Name != "Group 1" {
					t.Errorf("expected first group name 'Group 1', got '%s'", result[0].Name)
				}
				if result[0].MemberRole != "owner" {
					t.Errorf("expected first group role 'owner', got '%s'", result[0].MemberRole)
				}
				if result[1].Name != "Group 2" {
					t.Errorf("expected second group name 'Group 2', got '%s'", result[1].Name)
				}
				if result[1].MemberRole != "member" {
					t.Errorf("expected second group role 'member', got '%s'", result[1].MemberRole)
				}
			},
		},
		{
			name:   "empty list",
			userID: userID,
			mockSetup: func(mock *testutil.MockGroupRepository) {
				mock.GetGroupsByUserIDFunc = func(ctx context.Context, uID pgtype.UUID) ([]sqlc.GetGroupsByUserIDRow, error) {
					return []sqlc.GetGroupsByUserIDRow{}, nil
				}
			},
			expectedError: nil,
			expectedLen:   0,
		},
		{
			name:   "repository error",
			userID: userID,
			mockSetup: func(mock *testutil.MockGroupRepository) {
				mock.GetGroupsByUserIDFunc = func(ctx context.Context, uID pgtype.UUID) ([]sqlc.GetGroupsByUserIDRow, error) {
					return nil, errors.New("database error")
				}
			},
			expectedError: errors.New("database error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &testutil.MockGroupRepository{}
			tt.mockSetup(mockRepo)

			mockActivitySvc := &MockGroupActivityService{}
			svc := NewGroupService(mockRepo, &testutil.MockGroupInvitationRepository{}, mockActivitySvc)
			result, err := svc.ListUserGroups(context.Background(), tt.userID)

			if tt.expectedError != nil {
				if err == nil {
					t.Errorf("expected error %v, got nil", tt.expectedError)
					return
				}
				if !errors.Is(err, tt.expectedError) && err.Error() != tt.expectedError.Error() {
					t.Errorf("expected error %v, got %v", tt.expectedError, err)
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if len(result) != tt.expectedLen {
				t.Errorf("expected %d groups, got %d", tt.expectedLen, len(result))
			}

			if tt.validate != nil {
				tt.validate(t, result)
			}
		})
	}
}
