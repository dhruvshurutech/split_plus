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



type GroupService interface {
	CreateGroup(ctx context.Context, input CreateGroupInput) (CreateGroupResult, error)
	ListGroupMembers(ctx context.Context, groupID, requesterID pgtype.UUID) ([]sqlc.ListGroupMembersRow, error)
	ListUserGroups(ctx context.Context, userID pgtype.UUID) ([]sqlc.GetGroupsByUserIDRow, error)
}

type groupService struct {
	repo            repository.GroupRepository
	activityService GroupActivityService
}

func NewGroupService(repo repository.GroupRepository, activityService GroupActivityService) GroupService {
	return &groupService{
		repo:            repo,
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
	// (Optional: maybe not needed for feed inside group, but good for completeness)
	// Actually, feed is PER group, so logging "group created" inside the group feed is logical as first event.
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



func (s *groupService) ListGroupMembers(ctx context.Context, groupID, requesterID pgtype.UUID) ([]sqlc.ListGroupMembersRow, error) {
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

	// Fetch members with user details
	return s.repo.ListGroupMembers(ctx, groupID)
}

func (s *groupService) ListUserGroups(ctx context.Context, userID pgtype.UUID) ([]sqlc.GetGroupsByUserIDRow, error) {
	return s.repo.GetGroupsByUserID(ctx, userID)
}
