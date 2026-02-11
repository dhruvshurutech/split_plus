package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/dhruvsaxena1998/splitplus/internal/db/sqlc"
	"github.com/dhruvsaxena1998/splitplus/internal/repository"
	"github.com/jackc/pgx/v5/pgtype"
)

var (
	ErrInvitationNotFound                 = errors.New("invitation not found or expired")
	ErrInvitationExpired                  = errors.New("invitation expired")
	ErrInvitationEmailMismatch            = errors.New("logged in user email does not match invitation email")
	ErrPasswordRequiredForExistingAccount = errors.New("password required for existing account")
	ErrPasswordRequiredToCreateAccount    = errors.New("password required to create account")
)

type CreateInvitationInput struct {
	GroupID   pgtype.UUID
	InvitedBy pgtype.UUID
	Email     string
	Role      string // "admin" or "member"
	Name      string // Optional name for the invited person
}

type AcceptInvitationInput struct {
	Token  string
	UserID pgtype.UUID
}

type JoinGroupInput struct {
	Token               string
	Password            string      // Required if not authenticated
	Name                string      // Optional for new users
	AuthenticatedUserID pgtype.UUID // Present if user is already logged in
}

type InvitationDetails struct {
	Invitation   sqlc.GetInvitationByTokenRow
	GroupName    string
	InviterName  string
	InviterEmail string
}

type GroupInvitationService interface {
	CreateInvitation(ctx context.Context, input CreateInvitationInput) (string, error)
	GetInvitation(ctx context.Context, token string) (InvitationDetails, error)
	AcceptInvitation(ctx context.Context, input AcceptInvitationInput) (sqlc.GroupMember, error)
	JoinGroup(ctx context.Context, input JoinGroupInput) (sqlc.User, sqlc.GroupMember, error)
	ListPendingInvitations(ctx context.Context, email string) ([]sqlc.GetPendingInvitationsByEmailRow, error)
}

type groupInvitationService struct {
	invRepo     repository.GroupInvitationRepository
	pendingRepo repository.PendingUserRepository
	groupRepo   repository.GroupRepository
	userRepo    repository.UserRepository
	userService UserService
}

func NewGroupInvitationService(
	invRepo repository.GroupInvitationRepository,
	pendingRepo repository.PendingUserRepository,
	groupRepo repository.GroupRepository,
	userRepo repository.UserRepository,
	userService UserService,
) GroupInvitationService {
	return &groupInvitationService{
		invRepo:     invRepo,
		pendingRepo: pendingRepo,
		groupRepo:   groupRepo,
		userRepo:    userRepo,
		userService: userService,
	}
}

func (s *groupInvitationService) CreateInvitation(ctx context.Context, input CreateInvitationInput) (string, error) {
	// 1. Check permissions (inviter must be member)
	_, err := s.groupRepo.GetGroupMember(ctx, sqlc.GetGroupMemberParams{
		GroupID: input.GroupID,
		UserID:  input.InvitedBy,
	})
	if err != nil {
		return "", ErrNotGroupMember
	}
	// TODO: restrict role assignment based on inviter's role? (e.g. only admin can invite admins)
	// For now, allow any member to invite others as 'member'.

	email := strings.ToLower(strings.TrimSpace(input.Email))

	// 2. Ensure PendingUser exists
	// We create/update pending user to ensure we have an ID references by expenses later.
	// Even if user exists in `users` table, we might use `pending_users` for initial expense assignment via email?
	// Actually, if user exists in `users`, we should probably use `users.id` directly if we knew it.
	// But invitation is by email. We don't know if they have an account yet or what their ID is until they accept.
	// So standardized on pending_user for email-based stuff is fine, or we check if user exists.
	// Let's just create pending_user record for simplicity of referencing by email.
	_, err = s.pendingRepo.CreatePendingUser(ctx, sqlc.CreatePendingUserParams{
		Email: email,
		Name:  pgtype.Text{String: input.Name, Valid: input.Name != ""},
	})
	if err != nil {
		return "", fmt.Errorf("failed to create pending user: %w", err)
	}

	// 3. Generate Token
	token, err := generateToken(32)
	if err != nil {
		return "", err
	}

	// 4. Create Invitation
	expiresAt := time.Now().Add(7 * 24 * time.Hour) // 7 days expiry
	_, err = s.invRepo.CreateInvitation(ctx, sqlc.CreateInvitationParams{
		GroupID:   input.GroupID,
		Email:     email,
		Token:     token,
		Role:      input.Role,
		Status:    "pending",
		InvitedBy: input.InvitedBy,
		ExpiresAt: pgtype.Timestamptz{Time: expiresAt, Valid: true},
	})
	if err != nil {
		return "", fmt.Errorf("failed to create invitation: %w", err)
	}

	// 5. Send Email (Mock)
	// In real app, call email service here.
	// For now, just logging or doing nothing.
	return token, nil
}

func (s *groupInvitationService) GetInvitation(ctx context.Context, token string) (InvitationDetails, error) {
	inv, err := s.invRepo.GetInvitationByToken(ctx, token)
	if err != nil {
		return InvitationDetails{}, ErrInvitationNotFound
	}

	details := InvitationDetails{
		Invitation: inv,
		GroupName:  inv.GroupName,
	}

	inviter, err := s.userRepo.GetUserByID(ctx, inv.InvitedBy)
	if err == nil {
		details.InviterEmail = inviter.Email
		if inviter.Name.Valid {
			details.InviterName = strings.TrimSpace(inviter.Name.String)
		}
	}

	return details, nil
}

func (s *groupInvitationService) AcceptInvitation(ctx context.Context, input AcceptInvitationInput) (sqlc.GroupMember, error) {
	// 1. Get Invitation
	inv, err := s.invRepo.GetInvitationByToken(ctx, input.Token)
	if err != nil {
		return sqlc.GroupMember{}, ErrInvitationNotFound
	}

	// 2. Check if user is already member
	_, err = s.groupRepo.GetGroupMember(ctx, sqlc.GetGroupMemberParams{
		GroupID: inv.GroupID,
		UserID:  input.UserID,
	})
	if err == nil {
		// Already member
		// Update invitation to accepted anyway? Or return error?
		// If return error, frontend handles "You are already a member".
		// But we should probably mark invitation as accepted to clean it up.
		s.invRepo.UpdateInvitationStatus(ctx, sqlc.UpdateInvitationStatusParams{
			ID:     inv.ID,
			Status: "accepted",
		})
		return sqlc.GroupMember{}, ErrAlreadyMember
	}

	// 3. Transaction: Create Member + Update Invitation
	tx, err := s.groupRepo.BeginTx(ctx)
	if err != nil {
		return sqlc.GroupMember{}, err
	}
	defer tx.Rollback(ctx)

	qTx := s.groupRepo.WithTx(tx)
	invTx := s.invRepo.WithTx(tx)

	// Add Member
	now := pgtype.Timestamptz{Time: time.Now(), Valid: true}
	member, err := qTx.CreateGroupMember(ctx, sqlc.CreateGroupMemberParams{
		GroupID:   inv.GroupID,
		UserID:    input.UserID,
		Role:      inv.Role,
		Status:    "active",
		InvitedBy: inv.InvitedBy,
		InvitedAt: inv.CreatedAt,
		JoinedAt:  now,
	})
	if err != nil {
		return sqlc.GroupMember{}, fmt.Errorf("failed to add member: %w", err)
	}

	// Update Invitation
	_, err = invTx.UpdateInvitationStatus(ctx, sqlc.UpdateInvitationStatusParams{
		ID:     inv.ID,
		Status: "accepted",
	})
	if err != nil {
		return sqlc.GroupMember{}, fmt.Errorf("failed to update invitation: %w", err)
	}

	// 4. Claim pending expenses
	// We do this within the transaction to ensure consistency.
	// Find pending user by email
	pendingUser, err := s.pendingRepo.GetPendingUserByEmail(ctx, inv.Email)
	if err == nil {
		pendingTx := s.pendingRepo.WithTx(tx)

		// Update payments
		err = pendingTx.UpdatePendingPaymentUserID(ctx, sqlc.UpdatePendingPaymentUserIDParams{
			UserID:        input.UserID,
			PendingUserID: pendingUser.ID,
		})
		if err != nil {
			return sqlc.GroupMember{}, fmt.Errorf("failed to claim pending payments: %w", err)
		}

		// Update splits
		err = pendingTx.UpdatePendingSplitUserID(ctx, sqlc.UpdatePendingSplitUserIDParams{
			UserID:        input.UserID,
			PendingUserID: pendingUser.ID,
		})
		if err != nil {
			return sqlc.GroupMember{}, fmt.Errorf("failed to claim pending splits: %w", err)
		}

		// Update settlement payer references.
		err = pendingTx.UpdatePendingSettlementPayerUserID(ctx, sqlc.UpdatePendingSettlementPayerUserIDParams{
			PayerID:            input.UserID,
			PayerPendingUserID: pendingUser.ID,
		})
		if err != nil {
			return sqlc.GroupMember{}, fmt.Errorf("failed to claim pending settlement payer refs: %w", err)
		}

		// Update settlement payee references.
		err = pendingTx.UpdatePendingSettlementPayeeUserID(ctx, sqlc.UpdatePendingSettlementPayeeUserIDParams{
			PayeeID:            input.UserID,
			PayeePendingUserID: pendingUser.ID,
		})
		if err != nil {
			return sqlc.GroupMember{}, fmt.Errorf("failed to claim pending settlement payee refs: %w", err)
		}

		// Best effort cleanup: pending users can be referenced by other pending invites.
		// Accepting this invitation must not fail just because cleanup cannot be done now.
		_ = pendingTx.DeletePendingUserByID(ctx, pendingUser.ID)
	}
	// If pending user not found, it means no expenses were assigned to this email yet, which is fine.

	if err := tx.Commit(ctx); err != nil {
		return sqlc.GroupMember{}, err
	}

	return member, nil
}

func (s *groupInvitationService) ListPendingInvitations(ctx context.Context, email string) ([]sqlc.GetPendingInvitationsByEmailRow, error) {
	return s.invRepo.GetPendingInvitationsByEmail(ctx, strings.ToLower(strings.TrimSpace(email)))
}

func (s *groupInvitationService) JoinGroup(ctx context.Context, input JoinGroupInput) (sqlc.User, sqlc.GroupMember, error) {
	// 1. Get Invitation
	inv, err := s.invRepo.GetInvitationByToken(ctx, input.Token)
	if err != nil {
		return sqlc.User{}, sqlc.GroupMember{}, ErrInvitationNotFound
	}

	var user sqlc.User

	// 2. Identify the user
	if input.AuthenticatedUserID.Valid {
		// Scenario: User is already logged in
		user, err = s.userRepo.GetUserByID(ctx, input.AuthenticatedUserID)
		if err != nil {
			return sqlc.User{}, sqlc.GroupMember{}, ErrUserNotFound
		}

		// Verify email match (Security check)
		if strings.ToLower(user.Email) != strings.ToLower(inv.Email) {
			return sqlc.User{}, sqlc.GroupMember{}, ErrInvitationEmailMismatch
		}
	} else {
		// Scenario: Not logged in. Check if user exists.
		_, err := s.userRepo.GetUserByEmail(ctx, inv.Email)
		if err == nil {
			// User exists, must authenticate
			if input.Password == "" {
				return sqlc.User{}, sqlc.GroupMember{}, ErrPasswordRequiredForExistingAccount
			}
			user, err = s.userService.AuthenticateUser(ctx, inv.Email, input.Password)
			if err != nil {
				return sqlc.User{}, sqlc.GroupMember{}, err
			}
		} else {
			// New user, must register
			if input.Password == "" {
				return sqlc.User{}, sqlc.GroupMember{}, ErrPasswordRequiredToCreateAccount
			}
			// Use name from input or fall back to invitation email prefix
			name := input.Name
			if name == "" {
				name = strings.Split(inv.Email, "@")[0]
			}
			user, err = s.userService.CreateUser(ctx, name, inv.Email, input.Password)
			if err != nil {
				return sqlc.User{}, sqlc.GroupMember{}, err
			}
			// TODO: Update user name if we add a name field to users table later
		}
	}

	// 3. Accept Invitation (Handles merging and membership)
	member, err := s.AcceptInvitation(ctx, AcceptInvitationInput{
		Token:  input.Token,
		UserID: user.ID,
	})
	if err != nil {
		if errors.Is(err, ErrAlreadyMember) {
			// If already a member, we still return the user info as a "success" for the join flow
			// Fetch the member info to return it
			m, _ := s.groupRepo.GetGroupMember(ctx, sqlc.GetGroupMemberParams{
				GroupID: inv.GroupID,
				UserID:  user.ID,
			})
			return user, m, nil
		}
		return sqlc.User{}, sqlc.GroupMember{}, err
	}

	return user, member, nil
}

func generateToken(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}
