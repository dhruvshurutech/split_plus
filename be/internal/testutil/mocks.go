package testutil

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/dhruvsaxena1998/splitplus/internal/db/sqlc"
	"github.com/dhruvsaxena1998/splitplus/internal/repository"
)

// ============================================================================
// MockUserRepository
// ============================================================================

type MockUserRepository struct {
	CreateUserFunc     func(ctx context.Context, params sqlc.CreateUserParams) (sqlc.User, error)
	GetUserByEmailFunc func(ctx context.Context, email string) (sqlc.User, error)
	GetUserByIDFunc    func(ctx context.Context, id pgtype.UUID) (sqlc.User, error)
}

func (m *MockUserRepository) CreateUser(ctx context.Context, params sqlc.CreateUserParams) (sqlc.User, error) {
	if m.CreateUserFunc != nil {
		return m.CreateUserFunc(ctx, params)
	}
	return sqlc.User{}, nil
}

func (m *MockUserRepository) GetUserByEmail(ctx context.Context, email string) (sqlc.User, error) {
	if m.GetUserByEmailFunc != nil {
		return m.GetUserByEmailFunc(ctx, email)
	}
	return sqlc.User{}, nil
}

func (m *MockUserRepository) GetUserByID(ctx context.Context, id pgtype.UUID) (sqlc.User, error) {
	if m.GetUserByIDFunc != nil {
		return m.GetUserByIDFunc(ctx, id)
	}
	return sqlc.User{}, nil
}

var _ repository.UserRepository = (*MockUserRepository)(nil)

// ============================================================================
// MockUserService
// ============================================================================

type MockUserService struct {
	CreateUserFunc       func(ctx context.Context, name, email, password string) (sqlc.User, error)
	AuthenticateUserFunc func(ctx context.Context, email, password string) (sqlc.User, error)
	GetUserFunc          func(ctx context.Context, id pgtype.UUID) (sqlc.User, error)
}

func (m *MockUserService) CreateUser(ctx context.Context, name, email, password string) (sqlc.User, error) {
	if m.CreateUserFunc != nil {
		return m.CreateUserFunc(ctx, name, email, password)
	}
	return sqlc.User{}, nil
}

func (m *MockUserService) GetUser(ctx context.Context, id pgtype.UUID) (sqlc.User, error) {
	if m.GetUserFunc != nil {
		return m.GetUserFunc(ctx, id)
	}
	return sqlc.User{}, nil
}

func (m *MockUserService) AuthenticateUser(ctx context.Context, email, password string) (sqlc.User, error) {
	if m.AuthenticateUserFunc != nil {
		return m.AuthenticateUserFunc(ctx, email, password)
	}
	return sqlc.User{}, nil
}

// ============================================================================
// MockGroupRepository
// ============================================================================

type MockGroupRepository struct {
	BeginTxFunc                 func(ctx context.Context) (pgx.Tx, error)
	WithTxFunc                  func(tx pgx.Tx) repository.GroupRepository
	CreateGroupFunc             func(ctx context.Context, params sqlc.CreateGroupParams) (sqlc.Group, error)
	GetGroupByIDFunc            func(ctx context.Context, id pgtype.UUID) (sqlc.Group, error)
	CreateGroupMemberFunc       func(ctx context.Context, params sqlc.CreateGroupMemberParams) (sqlc.GroupMember, error)
	GetGroupMemberFunc          func(ctx context.Context, params sqlc.GetGroupMemberParams) (sqlc.GroupMember, error)
	UpdateGroupMemberStatusFunc func(ctx context.Context, params sqlc.UpdateGroupMemberStatusParams) (sqlc.GroupMember, error)
	ListGroupMembersFunc        func(ctx context.Context, groupID pgtype.UUID) ([]sqlc.ListGroupMembersRow, error)
	GetGroupsByUserIDFunc       func(ctx context.Context, userID pgtype.UUID) ([]sqlc.GetGroupsByUserIDRow, error)
	GetUserByIDFunc             func(ctx context.Context, id pgtype.UUID) (sqlc.User, error)
}

func (m *MockGroupRepository) BeginTx(ctx context.Context) (pgx.Tx, error) {
	if m.BeginTxFunc != nil {
		return m.BeginTxFunc(ctx)
	}
	return &MockTx{}, nil
}

func (m *MockGroupRepository) WithTx(tx pgx.Tx) repository.GroupRepository {
	if m.WithTxFunc != nil {
		return m.WithTxFunc(tx)
	}
	return m
}

func (m *MockGroupRepository) CreateGroup(ctx context.Context, params sqlc.CreateGroupParams) (sqlc.Group, error) {
	if m.CreateGroupFunc != nil {
		return m.CreateGroupFunc(ctx, params)
	}
	return sqlc.Group{}, nil
}

func (m *MockGroupRepository) GetGroupByID(ctx context.Context, id pgtype.UUID) (sqlc.Group, error) {
	if m.GetGroupByIDFunc != nil {
		return m.GetGroupByIDFunc(ctx, id)
	}
	return sqlc.Group{}, nil
}

func (m *MockGroupRepository) CreateGroupMember(ctx context.Context, params sqlc.CreateGroupMemberParams) (sqlc.GroupMember, error) {
	if m.CreateGroupMemberFunc != nil {
		return m.CreateGroupMemberFunc(ctx, params)
	}
	return sqlc.GroupMember{}, nil
}

func (m *MockGroupRepository) GetGroupMember(ctx context.Context, params sqlc.GetGroupMemberParams) (sqlc.GroupMember, error) {
	if m.GetGroupMemberFunc != nil {
		return m.GetGroupMemberFunc(ctx, params)
	}
	return sqlc.GroupMember{}, nil
}

func (m *MockGroupRepository) UpdateGroupMemberStatus(ctx context.Context, params sqlc.UpdateGroupMemberStatusParams) (sqlc.GroupMember, error) {
	if m.UpdateGroupMemberStatusFunc != nil {
		return m.UpdateGroupMemberStatusFunc(ctx, params)
	}
	return sqlc.GroupMember{}, nil
}

func (m *MockGroupRepository) ListGroupMembers(ctx context.Context, groupID pgtype.UUID) ([]sqlc.ListGroupMembersRow, error) {
	if m.ListGroupMembersFunc != nil {
		return m.ListGroupMembersFunc(ctx, groupID)
	}
	return []sqlc.ListGroupMembersRow{}, nil
}

func (m *MockGroupRepository) GetGroupsByUserID(ctx context.Context, userID pgtype.UUID) ([]sqlc.GetGroupsByUserIDRow, error) {
	if m.GetGroupsByUserIDFunc != nil {
		return m.GetGroupsByUserIDFunc(ctx, userID)
	}
	return []sqlc.GetGroupsByUserIDRow{}, nil
}

func (m *MockGroupRepository) GetUserByID(ctx context.Context, id pgtype.UUID) (sqlc.User, error) {
	if m.GetUserByIDFunc != nil {
		return m.GetUserByIDFunc(ctx, id)
	}
	return sqlc.User{}, nil
}

var _ repository.GroupRepository = (*MockGroupRepository)(nil)

// ============================================================================
// MockTx - Mock transaction for testing
// ============================================================================

type MockTx struct {
	CommitFunc   func(ctx context.Context) error
	RollbackFunc func(ctx context.Context) error
}

func (m *MockTx) Begin(ctx context.Context) (pgx.Tx, error) {
	return m, nil
}

func (m *MockTx) Commit(ctx context.Context) error {
	if m.CommitFunc != nil {
		return m.CommitFunc(ctx)
	}
	return nil
}

func (m *MockTx) Rollback(ctx context.Context) error {
	if m.RollbackFunc != nil {
		return m.RollbackFunc(ctx)
	}
	return nil
}

func (m *MockTx) CopyFrom(ctx context.Context, tableName pgx.Identifier, columnNames []string, rowSrc pgx.CopyFromSource) (int64, error) {
	return 0, nil
}

func (m *MockTx) SendBatch(ctx context.Context, b *pgx.Batch) pgx.BatchResults {
	return nil
}

func (m *MockTx) LargeObjects() pgx.LargeObjects {
	return pgx.LargeObjects{}
}

func (m *MockTx) Prepare(ctx context.Context, name, sql string) (*pgconn.StatementDescription, error) {
	return nil, nil
}

func (m *MockTx) Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error) {
	return pgconn.CommandTag{}, nil
}

func (m *MockTx) Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
	return nil, nil
}

func (m *MockTx) QueryRow(ctx context.Context, sql string, args ...any) pgx.Row {
	return nil
}

func (m *MockTx) Conn() *pgx.Conn {
	return nil
}

// ============================================================================
// MockGroupInvitationRepository
// ============================================================================

type MockGroupInvitationRepository struct {
	BeginTxFunc                  func(ctx context.Context) (pgx.Tx, error)
	WithTxFunc                   func(tx pgx.Tx) repository.GroupInvitationRepository
	CreateInvitationFunc         func(ctx context.Context, params sqlc.CreateInvitationParams) (sqlc.GroupInvitation, error)
	GetInvitationByTokenFunc     func(ctx context.Context, token string) (sqlc.GetInvitationByTokenRow, error)
	GetInvitationByIDFunc        func(ctx context.Context, id pgtype.UUID) (sqlc.GroupInvitation, error)
	UpdateInvitationStatusFunc   func(ctx context.Context, params sqlc.UpdateInvitationStatusParams) (sqlc.GroupInvitation, error)
	ListInvitationsByGroupFunc   func(ctx context.Context, groupID pgtype.UUID) ([]sqlc.ListInvitationsByGroupRow, error)
	GetPendingInvitationsByEmailFunc func(ctx context.Context, email string) ([]sqlc.GetPendingInvitationsByEmailRow, error)
}

func (m *MockGroupInvitationRepository) BeginTx(ctx context.Context) (pgx.Tx, error) {
	if m.BeginTxFunc != nil {
		return m.BeginTxFunc(ctx)
	}
	return &MockTx{}, nil
}

func (m *MockGroupInvitationRepository) WithTx(tx pgx.Tx) repository.GroupInvitationRepository {
	if m.WithTxFunc != nil {
		return m.WithTxFunc(tx)
	}
	return m
}

func (m *MockGroupInvitationRepository) CreateInvitation(ctx context.Context, params sqlc.CreateInvitationParams) (sqlc.GroupInvitation, error) {
	if m.CreateInvitationFunc != nil {
		return m.CreateInvitationFunc(ctx, params)
	}
	return sqlc.GroupInvitation{}, nil
}

func (m *MockGroupInvitationRepository) GetInvitationByToken(ctx context.Context, token string) (sqlc.GetInvitationByTokenRow, error) {
	if m.GetInvitationByTokenFunc != nil {
		return m.GetInvitationByTokenFunc(ctx, token)
	}
	return sqlc.GetInvitationByTokenRow{}, nil
}

func (m *MockGroupInvitationRepository) GetInvitationByID(ctx context.Context, id pgtype.UUID) (sqlc.GroupInvitation, error) {
	if m.GetInvitationByIDFunc != nil {
		return m.GetInvitationByIDFunc(ctx, id)
	}
	return sqlc.GroupInvitation{}, nil
}

func (m *MockGroupInvitationRepository) UpdateInvitationStatus(ctx context.Context, params sqlc.UpdateInvitationStatusParams) (sqlc.GroupInvitation, error) {
	if m.UpdateInvitationStatusFunc != nil {
		return m.UpdateInvitationStatusFunc(ctx, params)
	}
	return sqlc.GroupInvitation{}, nil
}

func (m *MockGroupInvitationRepository) ListInvitationsByGroup(ctx context.Context, groupID pgtype.UUID) ([]sqlc.ListInvitationsByGroupRow, error) {
	if m.ListInvitationsByGroupFunc != nil {
		return m.ListInvitationsByGroupFunc(ctx, groupID)
	}
	return []sqlc.ListInvitationsByGroupRow{}, nil
}

func (m *MockGroupInvitationRepository) GetPendingInvitationsByEmail(ctx context.Context, email string) ([]sqlc.GetPendingInvitationsByEmailRow, error) {
	if m.GetPendingInvitationsByEmailFunc != nil {
		return m.GetPendingInvitationsByEmailFunc(ctx, email)
	}
	return []sqlc.GetPendingInvitationsByEmailRow{}, nil
}

var _ repository.GroupInvitationRepository = (*MockGroupInvitationRepository)(nil)

// ============================================================================
// MockPendingUserRepository
// ============================================================================

type MockPendingUserRepository struct {
	BeginTxFunc                func(ctx context.Context) (pgx.Tx, error)
	WithTxFunc                 func(tx pgx.Tx) repository.PendingUserRepository
	CreatePendingUserFunc      func(ctx context.Context, params sqlc.CreatePendingUserParams) (sqlc.PendingUser, error)
	GetPendingUserByEmailFunc  func(ctx context.Context, email string) (sqlc.PendingUser, error)
	GetPendingUserByIDFunc     func(ctx context.Context, id pgtype.UUID) (sqlc.PendingUser, error)
	UpdatePendingPaymentUserIDFunc func(ctx context.Context, params sqlc.UpdatePendingPaymentUserIDParams) error
	UpdatePendingSplitUserIDFunc   func(ctx context.Context, params sqlc.UpdatePendingSplitUserIDParams) error
}

func (m *MockPendingUserRepository) BeginTx(ctx context.Context) (pgx.Tx, error) {
	if m.BeginTxFunc != nil {
		return m.BeginTxFunc(ctx)
	}
	return &MockTx{}, nil
}

func (m *MockPendingUserRepository) WithTx(tx pgx.Tx) repository.PendingUserRepository {
	if m.WithTxFunc != nil {
		return m.WithTxFunc(tx)
	}
	return m
}

func (m *MockPendingUserRepository) CreatePendingUser(ctx context.Context, params sqlc.CreatePendingUserParams) (sqlc.PendingUser, error) {
	if m.CreatePendingUserFunc != nil {
		return m.CreatePendingUserFunc(ctx, params)
	}
	return sqlc.PendingUser{}, nil
}

func (m *MockPendingUserRepository) GetPendingUserByEmail(ctx context.Context, email string) (sqlc.PendingUser, error) {
	if m.GetPendingUserByEmailFunc != nil {
		return m.GetPendingUserByEmailFunc(ctx, email)
	}
	return sqlc.PendingUser{}, nil
}

func (m *MockPendingUserRepository) GetPendingUserByID(ctx context.Context, id pgtype.UUID) (sqlc.PendingUser, error) {
	if m.GetPendingUserByIDFunc != nil {
		return m.GetPendingUserByIDFunc(ctx, id)
	}
	return sqlc.PendingUser{}, nil
}

func (m *MockPendingUserRepository) UpdatePendingPaymentUserID(ctx context.Context, params sqlc.UpdatePendingPaymentUserIDParams) error {
	if m.UpdatePendingPaymentUserIDFunc != nil {
		return m.UpdatePendingPaymentUserIDFunc(ctx, params)
	}
	return nil
}

func (m *MockPendingUserRepository) UpdatePendingSplitUserID(ctx context.Context, params sqlc.UpdatePendingSplitUserIDParams) error {
	if m.UpdatePendingSplitUserIDFunc != nil {
		return m.UpdatePendingSplitUserIDFunc(ctx, params)
	}
	return nil
}

var _ repository.PendingUserRepository = (*MockPendingUserRepository)(nil)

// ============================================================================
// Test Data Helpers
// ============================================================================

func CreateTestUser(id pgtype.UUID, email string) sqlc.User {
	return sqlc.User{
		ID:    id,
		Email: email,
		CreatedAt: pgtype.Timestamptz{
			Time:  parseTime("2024-01-01T00:00:00Z"),
			Valid: true,
		},
		UpdatedAt: pgtype.Timestamptz{
			Time:  parseTime("2024-01-01T00:00:00Z"),
			Valid: true,
		},
		DeletedAt: pgtype.Timestamptz{Valid: false},
	}
}

func CreateTestGroup(id pgtype.UUID, name string, createdBy pgtype.UUID) sqlc.Group {
	return sqlc.Group{
		ID:           id,
		Name:         name,
		Description:  pgtype.Text{String: "Test group description", Valid: true},
		CurrencyCode: "USD",
		CreatedBy:    createdBy,
		UpdatedBy:    createdBy,
		CreatedAt: pgtype.Timestamptz{
			Time:  parseTime("2024-01-01T00:00:00Z"),
			Valid: true,
		},
		UpdatedAt: pgtype.Timestamptz{
			Time:  parseTime("2024-01-01T00:00:00Z"),
			Valid: true,
		},
		DeletedAt: pgtype.Timestamptz{Valid: false},
	}
}

func CreateTestGroupMember(id, groupID, userID pgtype.UUID, role, status string) sqlc.GroupMember {
	now := pgtype.Timestamptz{Time: parseTime("2024-01-01T00:00:00Z"), Valid: true}
	return sqlc.GroupMember{
		ID:        id,
		GroupID:   groupID,
		UserID:    userID,
		Role:      role,
		Status:    status,
		InvitedBy: pgtype.UUID{},
		InvitedAt: pgtype.Timestamptz{},
		JoinedAt:  now,
		CreatedAt: now,
		UpdatedAt: now,
		DeletedAt: pgtype.Timestamptz{Valid: false},
	}
}

func CreateTestUUID(n int) pgtype.UUID {
	uuid := pgtype.UUID{Valid: true}
	uuid.Bytes[15] = byte(n)
	return uuid
}

func parseTime(timeStr string) time.Time {
	t, _ := time.Parse(time.RFC3339, timeStr)
	return t
}
