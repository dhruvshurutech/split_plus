package service

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/dhruvsaxena1998/splitplus/internal/db/sqlc"
	"github.com/dhruvsaxena1998/splitplus/internal/repository"
)

var (
	ErrGroupNotFound           = errors.New("group not found")
	ErrNotGroupMember          = errors.New("user is not a member of this group")
	ErrInsufficientPermissions = errors.New("insufficient permissions to perform this action")
	ErrAlreadyMember           = errors.New("user is already a member of this group")
	ErrNoPendingInvitation     = errors.New("no pending invitation found")
	ErrInvalidGroupName        = errors.New("group name is required")
)

type CreateGroupInput struct {
	Name         string
	Description  string
	CurrencyCode string
	CreatedBy    pgtype.UUID
}

type CreateGroupResult struct {
	Group      sqlc.Group
	Membership sqlc.GroupMember
}



type GroupMemberDetail struct {
	ID            pgtype.UUID        `json:"id"`
	GroupID       pgtype.UUID        `json:"group_id"`
	UserID        pgtype.UUID        `json:"user_id"`
	Role          string             `json:"role"`
	Status        string             `json:"status"`
	InvitedBy     pgtype.UUID        `json:"invited_by"`
	InvitedAt     pgtype.Timestamptz `json:"invited_at"`
	JoinedAt      pgtype.Timestamptz `json:"joined_at"`
	UserEmail     string             `json:"user_email"`
	UserName      pgtype.Text        `json:"user_name"`
	UserAvatarUrl pgtype.Text        `json:"user_avatar_url"`
	IsPending     bool               `json:"is_pending"`
	PendingUserID pgtype.UUID        `json:"pending_user_id"`
}

type GroupService interface {
	CreateGroup(ctx context.Context, input CreateGroupInput) (CreateGroupResult, error)
	ListGroupMembers(ctx context.Context, groupID, requesterID pgtype.UUID) ([]GroupMemberDetail, error)
	ListUserGroups(ctx context.Context, userID pgtype.UUID) ([]sqlc.GetGroupsByUserIDRow, error)
}

type groupService struct {
	repo            repository.GroupRepository
	invRepo         repository.GroupInvitationRepository
	activityService GroupActivityService
}

func NewGroupService(
	repo repository.GroupRepository,
	invRepo repository.GroupInvitationRepository,
	activityService GroupActivityService,
) GroupService {
	return &groupService{
		repo:            repo,
		invRepo:         invRepo,
		activityService: activityService,
	}
}

func (s *groupService) CreateGroup(ctx context.Context, input CreateGroupInput) (CreateGroupResult, error) {
	name := strings.TrimSpace(input.Name)
	if name == "" {
		return CreateGroupResult{}, ErrInvalidGroupName
	}

	currencyCode := strings.TrimSpace(input.CurrencyCode)
	if currencyCode == "" {
		currencyCode = "USD"
	}

	// Start transaction
	tx, err := s.repo.BeginTx(ctx)
	if err != nil {
		return CreateGroupResult{}, err
	}
	defer tx.Rollback(ctx)

	txRepo := s.repo.WithTx(tx)

	// Create group
	group, err := txRepo.CreateGroup(ctx, sqlc.CreateGroupParams{
		Name:         name,
		Description:  pgtype.Text{String: input.Description, Valid: input.Description != ""},
		CurrencyCode: currencyCode,
		CreatedBy:    input.CreatedBy,
	})
	if err != nil {
		return CreateGroupResult{}, err
	}

	// Create owner membership
	now := pgtype.Timestamptz{Time: time.Now(), Valid: true}
	membership, err := txRepo.CreateGroupMember(ctx, sqlc.CreateGroupMemberParams{
		GroupID:   group.ID,
		UserID:    input.CreatedBy,
		Role:      "owner",
		Status:    "active",
		InvitedBy: pgtype.UUID{},        // null - self-joined as creator
		InvitedAt: pgtype.Timestamptz{}, // null
		JoinedAt:  now,
	})
	if err != nil {
		return CreateGroupResult{}, err
	}

	// Commit transaction
	if err := tx.Commit(ctx); err != nil {
		return CreateGroupResult{}, err
	}

	// Log activity - group created
	_ = s.activityService.LogActivity(ctx, LogActivityInput{
		GroupID:    group.ID,
		UserID:     input.CreatedBy,
		Action:     "group_created",
		EntityType: "group",
		EntityID:   group.ID,
		Metadata: map[string]interface{}{
			"name": group.Name,
		},
	})

	return CreateGroupResult{
		Group:      group,
		Membership: membership,
	}, nil
}

func (s *groupService) ListGroupMembers(ctx context.Context, groupID, requesterID pgtype.UUID) ([]GroupMemberDetail, error) {
	// Verify group exists
	_, err := s.repo.GetGroupByID(ctx, groupID)
	if err != nil {
		return nil, ErrGroupNotFound
	}

	// Verify requester is a member
	_, err = s.repo.GetGroupMember(ctx, sqlc.GetGroupMemberParams{
		GroupID: groupID,
		UserID:  requesterID,
	})
	if err != nil {
		return nil, ErrNotGroupMember
	}

	// Fetch actual members
	members, err := s.repo.ListGroupMembers(ctx, groupID)
	if err != nil {
		return nil, err
	}

	// Fetch pending invitations
	invitations, err := s.invRepo.ListInvitationsByGroup(ctx, groupID)
	if err != nil {
		// Just log error or ignore? Safer to return partial data?
		// But for now let's return error if DB fails
		return nil, err
	}

	// Merge results
	result := make([]GroupMemberDetail, 0, len(members)+len(invitations))

	// Add members
	for _, m := range members {
		result = append(result, GroupMemberDetail{
			ID:            m.ID,
			GroupID:       m.GroupID,
			UserID:        m.UserID,
			Role:          m.Role,
			Status:        m.Status,
			InvitedBy:     m.InvitedBy,
			InvitedAt:     m.InvitedAt,
			JoinedAt:      m.JoinedAt,
			UserEmail:     m.UserEmail,
			UserName:      m.UserName,
			UserAvatarUrl: m.UserAvatarUrl,
			IsPending:     false,
		})
	}

	// Add pending invitations
	// Only those with status 'pending'
	for _, inv := range invitations {
		if inv.Status == "pending" {
			// Note: UserID is empty/null for invitations
			result = append(result, GroupMemberDetail{
				ID:        inv.ID, // Invitation ID used as member ID for display
				GroupID:   inv.GroupID,
				UserID:    pgtype.UUID{}, // No user ID yet
				Role:      inv.Role,
				Status:    "pending",
				InvitedBy: inv.InvitedBy,
				InvitedAt: inv.CreatedAt, // Use creation time as invited_at
				// JoinedAt is null
				UserEmail: inv.Email,
				UserName:  inv.PendingUserName, // Use populated PendingUserName
				IsPending:     true,
				PendingUserID: inv.PendingUserID,
			})
		}
	}

	return result, nil
}

func (s *groupService) ListUserGroups(ctx context.Context, userID pgtype.UUID) ([]sqlc.GetGroupsByUserIDRow, error) {
	return s.repo.GetGroupsByUserID(ctx, userID)
}
