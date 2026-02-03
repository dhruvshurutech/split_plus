package service

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strconv"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/shopspring/decimal"

	"github.com/dhruvsaxena1998/splitplus/internal/db/sqlc"
	"github.com/dhruvsaxena1998/splitplus/internal/repository"
)

var (
	ErrBalanceNotFound = errors.New("balance not found")
)

// BalanceResponse represents a user's balance in a group
type BalanceResponse struct {
	UserID        pgtype.UUID `json:"user_id"`
	UserEmail     string      `json:"user_email"`
	UserName      string      `json:"user_name"`
	UserAvatarURL string      `json:"user_avatar_url"`
	TotalPaid     string      `json:"total_paid"`
	TotalOwed     string      `json:"total_owed"`
	Balance       string      `json:"balance"`
}

// GroupBalanceResponse represents balance for a group
type GroupBalanceResponse struct {
	GroupID      pgtype.UUID `json:"group_id"`
	GroupName    string      `json:"group_name"`
	CurrencyCode string      `json:"currency_code"`
	TotalPaid    string      `json:"total_paid"`
	TotalOwed    string      `json:"total_owed"`
	Balance      string      `json:"balance"`
}

// DebtResponse represents who owes whom
type DebtResponse struct {
	DebtorID      pgtype.UUID `json:"debtor_id"`
	DebtorPendingUserID *pgtype.UUID `json:"debtor_pending_user_id,omitempty"`
	DebtorEmail   string      `json:"debtor_email"`
	DebtorName    string      `json:"debtor_name"`
	CreditorID    pgtype.UUID `json:"creditor_id"`
	CreditorPendingUserID *pgtype.UUID `json:"creditor_pending_user_id,omitempty"`
	CreditorEmail string      `json:"creditor_email"`
	CreditorName  string      `json:"creditor_name"`
	Amount        string      `json:"amount"`
}

type BalanceService interface {
	GetGroupBalances(ctx context.Context, groupID, requesterID pgtype.UUID) ([]BalanceResponse, error)
	GetUserBalanceInGroup(ctx context.Context, groupID, userID, requesterID pgtype.UUID) (BalanceResponse, error)
	GetOverallUserBalance(ctx context.Context, userID pgtype.UUID) ([]GroupBalanceResponse, error)
	GetSimplifiedDebts(ctx context.Context, groupID, requesterID pgtype.UUID) ([]DebtResponse, error)
}

type balanceService struct {
	repo repository.BalanceRepository
}

func NewBalanceService(repo repository.BalanceRepository) BalanceService {
	return &balanceService{repo: repo}
}

// Helper to convert interface{} (from PostgreSQL numeric) to decimal string
func numericInterfaceToString(val interface{}) string {
	if val == nil {
		return "0"
	}

	// Try different numeric types
	switch v := val.(type) {
	case string:
		return v
	case []byte:
		return string(v)
	case int64:
		return strconv.FormatInt(v, 10)
	case int32:
		return strconv.FormatInt(int64(v), 10)
	case float64:
		return decimal.NewFromFloat(v).String()
	case int:
		return strconv.Itoa(v)
	case pgtype.Numeric:
		// Handle pgtype.Numeric directly
		if !v.Valid {
			return "0"
		}
		if val, err := v.Value(); err == nil {
			if str, ok := val.(string); ok {
				return str
			}
		}
		return "0"
	default:
		// Try to convert via string representation
		if str := fmt.Sprintf("%v", v); str != "" && str != "<nil>" {
			// Try to parse as decimal to validate
			if _, err := decimal.NewFromString(str); err == nil {
				return str
			}
		}
		return "0"
	}
}

// Helper to convert numeric to decimal for calculations
func numericInterfaceToDecimal(val interface{}) decimal.Decimal {
	str := numericInterfaceToString(val)
	d, err := decimal.NewFromString(str)
	if err != nil {
		return decimal.Zero
	}
	return d
}

func (s *balanceService) validateGroupMembership(ctx context.Context, groupID, userID pgtype.UUID) error {
	_, err := s.repo.GetGroupMember(ctx, sqlc.GetGroupMemberParams{
		GroupID: groupID,
		UserID:  userID,
	})
	if err != nil {
		return ErrNotGroupMember
	}
	return nil
}

func (s *balanceService) GetGroupBalances(ctx context.Context, groupID, requesterID pgtype.UUID) ([]BalanceResponse, error) {
	// Validate group exists
	_, err := s.repo.GetGroupByID(ctx, groupID)
	if err != nil {
		return nil, ErrGroupNotFound
	}

	// Validate requester is group member
	if err := s.validateGroupMembership(ctx, groupID, requesterID); err != nil {
		return nil, err
	}

	// Get balances
	rows, err := s.repo.GetGroupBalances(ctx, groupID)
	if err != nil {
		return nil, err
	}

	balances := make([]BalanceResponse, 0, len(rows))
	for _, row := range rows {
		totalPaid := numericInterfaceToDecimal(row.TotalPaid)
		totalOwed := numericInterfaceToDecimal(row.TotalOwed)
		balance := totalPaid.Sub(totalOwed)

		userName := ""
		if row.UserName.Valid {
			userName = row.UserName.String
		}

		userAvatarURL := ""
		if row.UserAvatarUrl.Valid {
			userAvatarURL = row.UserAvatarUrl.String
		}

		balances = append(balances, BalanceResponse{
			UserID:        row.UserID,
			UserEmail:     row.UserEmail,
			UserName:      userName,
			UserAvatarURL: userAvatarURL,
			TotalPaid:     totalPaid.String(),
			TotalOwed:     totalOwed.String(),
			Balance:       balance.String(),
		})
	}

	return balances, nil
}

func (s *balanceService) GetUserBalanceInGroup(ctx context.Context, groupID, userID, requesterID pgtype.UUID) (BalanceResponse, error) {
	// Validate group exists
	_, err := s.repo.GetGroupByID(ctx, groupID)
	if err != nil {
		return BalanceResponse{}, ErrGroupNotFound
	}

	// Validate requester is group member
	if err := s.validateGroupMembership(ctx, groupID, requesterID); err != nil {
		return BalanceResponse{}, err
	}

	// Get balance
	row, err := s.repo.GetUserBalanceInGroup(ctx, sqlc.GetUserBalanceInGroupParams{
		GroupID: groupID,
		UserID:  userID,
	})
	if err != nil {
		return BalanceResponse{}, ErrBalanceNotFound
	}

	totalPaid := numericInterfaceToDecimal(row.TotalPaid)
	totalOwed := numericInterfaceToDecimal(row.TotalOwed)
	balance := totalPaid.Sub(totalOwed)

	userName := ""
	if row.UserName.Valid {
		userName = row.UserName.String
	}

	userAvatarURL := ""
	if row.UserAvatarUrl.Valid {
		userAvatarURL = row.UserAvatarUrl.String
	}

	return BalanceResponse{
		UserID:        row.UserID,
		UserEmail:     row.UserEmail,
		UserName:      userName,
		UserAvatarURL: userAvatarURL,
		TotalPaid:     totalPaid.String(),
		TotalOwed:     totalOwed.String(),
		Balance:       balance.String(),
	}, nil
}

func (s *balanceService) GetOverallUserBalance(ctx context.Context, userID pgtype.UUID) ([]GroupBalanceResponse, error) {
	rows, err := s.repo.GetOverallUserBalance(ctx, userID)
	if err != nil {
		return nil, err
	}

	balances := make([]GroupBalanceResponse, 0, len(rows))
	for _, row := range rows {
		totalPaid := numericInterfaceToDecimal(row.TotalPaid)
		totalOwed := numericInterfaceToDecimal(row.TotalOwed)
		balance := totalPaid.Sub(totalOwed)

		balances = append(balances, GroupBalanceResponse{
			GroupID:      row.GroupID,
			GroupName:    row.GroupName,
			CurrencyCode: row.CurrencyCode,
			TotalPaid:    totalPaid.String(),
			TotalOwed:    totalOwed.String(),
			Balance:      balance.String(),
		})
	}

	return balances, nil
}

func (s *balanceService) GetSimplifiedDebts(ctx context.Context, groupID, requesterID pgtype.UUID) ([]DebtResponse, error) {
	// Validate group exists
	_, err := s.repo.GetGroupByID(ctx, groupID)
	if err != nil {
		return nil, ErrGroupNotFound
	}

	// Validate requester is group member
	if err := s.validateGroupMembership(ctx, groupID, requesterID); err != nil {
		return nil, err
	}

	// Get per-user balances for the group (including pending users)
	rows, err := s.repo.GetGroupBalancesWithPending(ctx, groupID)
	if err != nil {
		return nil, err
	}

	// Build debtor / creditor lists from balances
	type node struct {
		ID            pgtype.UUID
		PendingUserID pgtype.UUID
		Email         string
		Name          string
		Balance       decimal.Decimal
	}

	var debtors, creditors []node
	for _, row := range rows {
		totalPaid := numericInterfaceToDecimal(row.TotalPaid)
		totalOwed := numericInterfaceToDecimal(row.TotalOwed)
		balance := totalPaid.Sub(totalOwed)

		if balance.IsZero() {
			continue
		}

		name := ""
		if row.Name.Valid {
			name = row.Name.String
		}

		n := node{
			ID:            row.UserID,
			PendingUserID: row.PendingUserID,
			Email:         row.Email,
			Name:          name,
			Balance:       balance,
		}

		if balance.GreaterThan(decimal.Zero) {
			creditors = append(creditors, n)
		} else if balance.LessThan(decimal.Zero) {
			debtors = append(debtors, n)
		}
	}

	// Nothing to simplify
	if len(debtors) == 0 || len(creditors) == 0 {
		return []DebtResponse{}, nil
	}

	// Sort debtors (most negative first) and creditors (most positive first)
	sort.Slice(debtors, func(i, j int) bool {
		return debtors[i].Balance.LessThan(debtors[j].Balance)
	})
	sort.Slice(creditors, func(i, j int) bool {
		// greater balance first
		return creditors[i].Balance.GreaterThan(creditors[j].Balance)
	})

	debts := make([]DebtResponse, 0)

	// Greedy min-cash-flow style matching
	i, j := 0, 0
	for i < len(debtors) && j < len(creditors) {
		d := &debtors[i]
		c := &creditors[j]

		// Debtor balances are negative
		debtorOwes := d.Balance.Neg() // positive amount owed
		creditorOwed := c.Balance     // positive amount to receive

		// If either side is effectively settled, move on
		if debtorOwes.IsZero() {
			i++
			continue
		}
		if creditorOwed.IsZero() {
			j++
			continue
		}

		amount := decimal.Min(debtorOwes, creditorOwed)
		if amount.IsZero() {
			break
		}

		debts = append(debts, DebtResponse{
			DebtorID:              d.ID,
			DebtorPendingUserID:   pendingIDPtr(d.PendingUserID),
			DebtorEmail:           d.Email,
			DebtorName:            d.Name,
			CreditorID:            c.ID,
			CreditorPendingUserID: pendingIDPtr(c.PendingUserID),
			CreditorEmail:         c.Email,
			CreditorName:          c.Name,
			Amount:                amount.String(),
		})

		// Update balances
		d.Balance = d.Balance.Add(amount) // less negative
		c.Balance = c.Balance.Sub(amount) // less positive

		// Advance pointers when a side is settled
		if d.Balance.IsZero() {
			i++
		}
		if c.Balance.IsZero() {
			j++
		}
	}

	return debts, nil
}

func pendingIDPtr(id pgtype.UUID) *pgtype.UUID {
	if !id.Valid {
		return nil
	}
	return &id
}
