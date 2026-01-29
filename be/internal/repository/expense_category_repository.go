package repository

import (
	"context"

	"github.com/dhruvsaxena1998/splitplus/internal/db/sqlc"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ExpenseCategoryRepository interface {
	CreateCategory(ctx context.Context, params sqlc.CreateGroupCategoryParams) (sqlc.ExpenseCategory, error)
	GetCategoryByID(ctx context.Context, id pgtype.UUID) (sqlc.ExpenseCategory, error)
	GetCategoryBySlug(ctx context.Context, groupID pgtype.UUID, slug string) (sqlc.ExpenseCategory, error)
	ListCategoriesForGroup(ctx context.Context, groupID pgtype.UUID) ([]sqlc.ExpenseCategory, error)
	UpdateCategory(ctx context.Context, params sqlc.UpdateGroupCategoryParams) (sqlc.ExpenseCategory, error)
	DeleteCategory(ctx context.Context, id, updatedBy pgtype.UUID) error
	IsGroupMember(ctx context.Context, groupID, userID pgtype.UUID) (bool, error)
}

type expenseCategoryRepository struct {
	queries *sqlc.Queries
	pool    *pgxpool.Pool
}

func NewExpenseCategoryRepository(pool *pgxpool.Pool) ExpenseCategoryRepository {
	return &expenseCategoryRepository{
		queries: sqlc.New(pool),
		pool:    pool,
	}
}

func (r *expenseCategoryRepository) CreateCategory(ctx context.Context, params sqlc.CreateGroupCategoryParams) (sqlc.ExpenseCategory, error) {
	return r.queries.CreateGroupCategory(ctx, params)
}

func (r *expenseCategoryRepository) GetCategoryByID(ctx context.Context, id pgtype.UUID) (sqlc.ExpenseCategory, error) {
	return r.queries.GetCategoryByID(ctx, id)
}

func (r *expenseCategoryRepository) GetCategoryBySlug(ctx context.Context, groupID pgtype.UUID, slug string) (sqlc.ExpenseCategory, error) {
	return r.queries.GetCategoryBySlug(ctx, sqlc.GetCategoryBySlugParams{
		GroupID: groupID,
		Slug:    slug,
	})
}

func (r *expenseCategoryRepository) ListCategoriesForGroup(ctx context.Context, groupID pgtype.UUID) ([]sqlc.ExpenseCategory, error) {
	return r.queries.ListCategoriesForGroup(ctx, groupID)
}

func (r *expenseCategoryRepository) UpdateCategory(ctx context.Context, params sqlc.UpdateGroupCategoryParams) (sqlc.ExpenseCategory, error) {
	return r.queries.UpdateGroupCategory(ctx, params)
}

func (r *expenseCategoryRepository) DeleteCategory(ctx context.Context, id, updatedBy pgtype.UUID) error {
	return r.queries.DeleteGroupCategory(ctx, sqlc.DeleteGroupCategoryParams{
		ID:        id,
		UpdatedBy: updatedBy,
	})
}

func (r *expenseCategoryRepository) IsGroupMember(ctx context.Context, groupID, userID pgtype.UUID) (bool, error) {
	// Use existing query from group repository
	member, err := r.queries.GetGroupMember(ctx, sqlc.GetGroupMemberParams{
		GroupID: groupID,
		UserID:  userID,
	})
	if err != nil {
		return false, nil // Not a member
	}
	return member.Status == "active", nil
}
