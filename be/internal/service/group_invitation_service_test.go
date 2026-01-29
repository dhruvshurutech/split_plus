package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/dhruvsaxena1998/splitplus/internal/db/sqlc"
	"github.com/dhruvsaxena1998/splitplus/internal/repository"
	"github.com/dhruvsaxena1998/splitplus/internal/testutil"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)



func TestGroupInvitationService_CreateInvitation(t *testing.T) {
	groupID := testutil.CreateTestUUID(10)
	inviterID := testutil.CreateTestUUID(1)
	invitationID := testutil.CreateTestUUID(100)
	
	tests := []struct {
		name          string
		input         CreateInvitationInput
		mockSetup     func(*testutil.MockGroupInvitationRepository, *testutil.MockPendingUserRepository, *testutil.MockGroupRepository, *testutil.MockUserRepository)
		expectedError error
		validate      func(*testing.T, string)
	}{
		{
			name: "successful invitation - existing user",
			input: CreateInvitationInput{
				GroupID:   groupID,
				Email:     "user@example.com",
				InvitedBy: inviterID,
			},
			mockSetup: func(mockInv *testutil.MockGroupInvitationRepository, mockPending *testutil.MockPendingUserRepository, mockGroup *testutil.MockGroupRepository, mockUser *testutil.MockUserRepository) {
				// Mock permissions check
				mockGroup.GetGroupMemberFunc = func(ctx context.Context, params sqlc.GetGroupMemberParams) (sqlc.GroupMember, error) {
					// Identify call by args
					if params.UserID == inviterID {
						return testutil.CreateTestGroupMember(testutil.CreateTestUUID(50), groupID, inviterID, "owner", "active"), nil
					}
					return sqlc.GroupMember{}, pgx.ErrNoRows
				}
				
				// Mock creating pending user - We just allow it.
				mockPending.CreatePendingUserFunc = func(ctx context.Context, params sqlc.CreatePendingUserParams) (sqlc.PendingUser, error) {
					return sqlc.PendingUser{ID: testutil.CreateTestUUID(200), Email: params.Email}, nil
				}
				
				// Mock creating invitation
				mockInv.CreateInvitationFunc = func(ctx context.Context, params sqlc.CreateInvitationParams) (sqlc.GroupInvitation, error) {
					return sqlc.GroupInvitation{
						ID:        invitationID,
						GroupID:   groupID,
						Email:     params.Email,
						Token:     params.Token,
						Status:    "pending",
						ExpiresAt: params.ExpiresAt,
					}, nil
				}
			},
			expectedError: nil,
			validate: func(t *testing.T, token string) {
				if token == "" {
					t.Errorf("expected token, got empty string")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockInv := &testutil.MockGroupInvitationRepository{}
			mockPending := &testutil.MockPendingUserRepository{}
			mockGroup := &testutil.MockGroupRepository{}
			mockUser := &testutil.MockUserRepository{}
			
			tt.mockSetup(mockInv, mockPending, mockGroup, mockUser)
			
			mockSvc := &testutil.MockUserService{}
			svc := NewGroupInvitationService(mockInv, mockPending, mockGroup, mockUser, mockSvc)
			result, err := svc.CreateInvitation(context.Background(), tt.input)

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

func TestGroupInvitationService_AcceptInvitation(t *testing.T) {
	groupID := testutil.CreateTestUUID(10)
	invitationID := testutil.CreateTestUUID(100)
	userID := testutil.CreateTestUUID(2)
	pendingUserID := testutil.CreateTestUUID(3)
	
	tests := []struct {
		name          string
		input         AcceptInvitationInput
		mockSetup     func(*testutil.MockGroupInvitationRepository, *testutil.MockPendingUserRepository, *testutil.MockGroupRepository, *testutil.MockUserRepository)
		expectedError error
	}{
		{
			name: "successful acceptance",
			input: AcceptInvitationInput{
				Token:  "valid-token",
				UserID: userID,
			},
			mockSetup: func(mockInv *testutil.MockGroupInvitationRepository, mockPending *testutil.MockPendingUserRepository, mockGroup *testutil.MockGroupRepository, mockUser *testutil.MockUserRepository) {
				// Get invitation
				mockInv.GetInvitationByTokenFunc = func(ctx context.Context, token string) (sqlc.GetInvitationByTokenRow, error) {
					return sqlc.GetInvitationByTokenRow{
						ID:        invitationID,
						GroupID:   groupID,
						Email:     "user@example.com",
						Token:     "valid-token",
						Status:    "pending",
						Role:      "member",
						ExpiresAt: pgtype.Timestamptz{Time: time.Now().Add(time.Hour), Valid: true},
					}, nil
				}
				
				// Check membership - not member
				mockGroup.GetGroupMemberFunc = func(ctx context.Context, params sqlc.GetGroupMemberParams) (sqlc.GroupMember, error) {
					return sqlc.GroupMember{}, pgx.ErrNoRows
				}
				
				// Transaction
				mockGroup.BeginTxFunc = func(ctx context.Context) (pgx.Tx, error) {
					return &testutil.MockTx{}, nil
				}
				mockGroup.WithTxFunc = func(tx pgx.Tx) repository.GroupRepository { return mockGroup }
				mockInv.WithTxFunc = func(tx pgx.Tx) repository.GroupInvitationRepository { return mockInv }
				mockPending.WithTxFunc = func(tx pgx.Tx) repository.PendingUserRepository { return mockPending }
				
				// Create member
				mockGroup.CreateGroupMemberFunc = func(ctx context.Context, params sqlc.CreateGroupMemberParams) (sqlc.GroupMember, error) {
					return testutil.CreateTestGroupMember(testutil.CreateTestUUID(200), groupID, userID, "member", "active"), nil
				}
				
				// Update invitation
				mockInv.UpdateInvitationStatusFunc = func(ctx context.Context, params sqlc.UpdateInvitationStatusParams) (sqlc.GroupInvitation, error) {
					return sqlc.GroupInvitation{}, nil
				}
				
				// Get pending user
				mockPending.GetPendingUserByEmailFunc = func(ctx context.Context, email string) (sqlc.PendingUser, error) {
					return sqlc.PendingUser{ID: pendingUserID, Email: email}, nil
				}
				
				// Claim expenses
				mockPending.UpdatePendingPaymentUserIDFunc = func(ctx context.Context, params sqlc.UpdatePendingPaymentUserIDParams) error {
					return nil
				}
				mockPending.UpdatePendingSplitUserIDFunc = func(ctx context.Context, params sqlc.UpdatePendingSplitUserIDParams) error {
					return nil
				}
			},
			expectedError: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockInv := &testutil.MockGroupInvitationRepository{}
			mockPending := &testutil.MockPendingUserRepository{}
			mockGroup := &testutil.MockGroupRepository{}
			mockUser := &testutil.MockUserRepository{}
			
			tt.mockSetup(mockInv, mockPending, mockGroup, mockUser)
			
			mockSvc := &testutil.MockUserService{}
			svc := NewGroupInvitationService(mockInv, mockPending, mockGroup, mockUser, mockSvc)
			_, err := svc.AcceptInvitation(context.Background(), tt.input)

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
			}
		})
	}
}

func TestGroupInvitationService_JoinGroup(t *testing.T) {
	groupID := testutil.CreateTestUUID(10)
	invitationID := testutil.CreateTestUUID(100)
	userID := testutil.CreateTestUUID(2)
	email := "user@example.com"
	token := "valid-token"
	
	tests := []struct {
		name          string
		input         JoinGroupInput
		mockSetup     func(*testutil.MockGroupInvitationRepository, *testutil.MockPendingUserRepository, *testutil.MockGroupRepository, *testutil.MockUserRepository, *testutil.MockUserService)
		expectedError error
	}{
		{
			name: "successful join - new user",
			input: JoinGroupInput{
				Token:    token,
				Password: "password123",
				Name:     "New User",
			},
			mockSetup: func(mockInv *testutil.MockGroupInvitationRepository, mockPending *testutil.MockPendingUserRepository, mockGroup *testutil.MockGroupRepository, mockUser *testutil.MockUserRepository, mockSvc *testutil.MockUserService) {
				// Get invitation
				mockInv.GetInvitationByTokenFunc = func(ctx context.Context, t string) (sqlc.GetInvitationByTokenRow, error) {
					return sqlc.GetInvitationByTokenRow{
						ID:        invitationID,
						GroupID:   groupID,
						Email:     email,
						Token:     token,
						Status:    "pending",
						Role:      "member",
						ExpiresAt: pgtype.Timestamptz{Time: time.Now().Add(time.Hour), Valid: true},
					}, nil
				}
				
				// User not found by email
				mockUser.GetUserByEmailFunc = func(ctx context.Context, e string) (sqlc.User, error) {
					return sqlc.User{}, pgx.ErrNoRows
				}
				
				// Create user
				mockSvc.CreateUserFunc = func(ctx context.Context, e, p string) (sqlc.User, error) {
					return testutil.CreateTestUser(userID, e), nil
				}
				
				// AcceptInvitation mocks...
				mockGroup.GetGroupMemberFunc = func(ctx context.Context, params sqlc.GetGroupMemberParams) (sqlc.GroupMember, error) {
					return sqlc.GroupMember{}, pgx.ErrNoRows
				}
				mockGroup.BeginTxFunc = func(ctx context.Context) (pgx.Tx, error) { return &testutil.MockTx{}, nil }
				mockGroup.WithTxFunc = func(tx pgx.Tx) repository.GroupRepository { return mockGroup }
				mockInv.WithTxFunc = func(tx pgx.Tx) repository.GroupInvitationRepository { return mockInv }
				mockPending.WithTxFunc = func(tx pgx.Tx) repository.PendingUserRepository { return mockPending }
				mockGroup.CreateGroupMemberFunc = func(ctx context.Context, params sqlc.CreateGroupMemberParams) (sqlc.GroupMember, error) {
					return testutil.CreateTestGroupMember(testutil.CreateTestUUID(200), groupID, userID, "member", "active"), nil
				}
				mockInv.UpdateInvitationStatusFunc = func(ctx context.Context, params sqlc.UpdateInvitationStatusParams) (sqlc.GroupInvitation, error) {
					return sqlc.GroupInvitation{}, nil
				}
				mockPending.GetPendingUserByEmailFunc = func(ctx context.Context, e string) (sqlc.PendingUser, error) {
					return sqlc.PendingUser{}, pgx.ErrNoRows
				}
			},
			expectedError: nil,
		},
		{
			name: "successful join - logged in user",
			input: JoinGroupInput{
				Token:               token,
				AuthenticatedUserID: userID,
			},
			mockSetup: func(mockInv *testutil.MockGroupInvitationRepository, mockPending *testutil.MockPendingUserRepository, mockGroup *testutil.MockGroupRepository, mockUser *testutil.MockUserRepository, mockSvc *testutil.MockUserService) {
				mockInv.GetInvitationByTokenFunc = func(ctx context.Context, t string) (sqlc.GetInvitationByTokenRow, error) {
					return sqlc.GetInvitationByTokenRow{
						ID: invitationID, GroupID: groupID, Email: email, Token: token, Status: "pending", Role: "member",
						ExpiresAt: pgtype.Timestamptz{Time: time.Now().Add(time.Hour), Valid: true},
					}, nil
				}
				mockUser.GetUserByIDFunc = func(ctx context.Context, id pgtype.UUID) (sqlc.User, error) {
					return testutil.CreateTestUser(userID, email), nil
				}
				// AcceptInvitation mocks...
				mockGroup.GetGroupMemberFunc = func(ctx context.Context, params sqlc.GetGroupMemberParams) (sqlc.GroupMember, error) {
					return sqlc.GroupMember{}, pgx.ErrNoRows
				}
				mockGroup.BeginTxFunc = func(ctx context.Context) (pgx.Tx, error) { return &testutil.MockTx{}, nil }
				mockGroup.WithTxFunc = func(tx pgx.Tx) repository.GroupRepository { return mockGroup }
				mockInv.WithTxFunc = func(tx pgx.Tx) repository.GroupInvitationRepository { return mockInv }
				mockPending.WithTxFunc = func(tx pgx.Tx) repository.PendingUserRepository { return mockPending }
				mockGroup.CreateGroupMemberFunc = func(ctx context.Context, params sqlc.CreateGroupMemberParams) (sqlc.GroupMember, error) {
					return testutil.CreateTestGroupMember(testutil.CreateTestUUID(200), groupID, userID, "member", "active"), nil
				}
				mockInv.UpdateInvitationStatusFunc = func(ctx context.Context, params sqlc.UpdateInvitationStatusParams) (sqlc.GroupInvitation, error) {
					return sqlc.GroupInvitation{}, nil
				}
				mockPending.GetPendingUserByEmailFunc = func(ctx context.Context, e string) (sqlc.PendingUser, error) {
					return sqlc.PendingUser{}, pgx.ErrNoRows
				}
			},
			expectedError: nil,
		},
		{
			name: "email mismatch - forbidden",
			input: JoinGroupInput{
				Token:               token,
				AuthenticatedUserID: userID,
			},
			mockSetup: func(mockInv *testutil.MockGroupInvitationRepository, mockPending *testutil.MockPendingUserRepository, mockGroup *testutil.MockGroupRepository, mockUser *testutil.MockUserRepository, mockSvc *testutil.MockUserService) {
				mockInv.GetInvitationByTokenFunc = func(ctx context.Context, t string) (sqlc.GetInvitationByTokenRow, error) {
					return sqlc.GetInvitationByTokenRow{
						ID: invitationID, GroupID: groupID, Email: "different@example.com", Token: token, Status: "pending", Role: "member",
						ExpiresAt: pgtype.Timestamptz{Time: time.Now().Add(time.Hour), Valid: true},
					}, nil
				}
				mockUser.GetUserByIDFunc = func(ctx context.Context, id pgtype.UUID) (sqlc.User, error) {
					return testutil.CreateTestUser(userID, email), nil
				}
			},
			expectedError: errors.New("logged in user email does not match invitation email"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockInv := &testutil.MockGroupInvitationRepository{}
			mockPending := &testutil.MockPendingUserRepository{}
			mockGroup := &testutil.MockGroupRepository{}
			mockUser := &testutil.MockUserRepository{}
			mockSvc := &testutil.MockUserService{}
			
			tt.mockSetup(mockInv, mockPending, mockGroup, mockUser, mockSvc)
			
			svc := NewGroupInvitationService(mockInv, mockPending, mockGroup, mockUser, mockSvc)
			_, _, err := svc.JoinGroup(context.Background(), tt.input)

			if tt.expectedError != nil {
				if err == nil {
					t.Errorf("expected error %v, got nil", tt.expectedError)
					return
				}
				if err.Error() != tt.expectedError.Error() {
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
