package service

import (
	"context"
	"errors"
	"strings"

	"github.com/dhruvsaxena1998/splitplus/internal/db/sqlc"
	"github.com/dhruvsaxena1998/splitplus/internal/repository"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
)

var (
	ErrCategoryNotFound      = errors.New("category not found")
	ErrInvalidCategoryName   = errors.New("category name is required")
	ErrCategoryAlreadyExists = errors.New("category with this name already exists in the group")
)

type CreateGroupCategoryInput struct {
	GroupID   pgtype.UUID
	Name      string
	Icon      string
	Color     string
	CreatedBy pgtype.UUID
}

type UpdateCategoryInput struct {
	CategoryID pgtype.UUID
	GroupID    pgtype.UUID
	Name       string
	Icon       string
	Color      string
	UpdatedBy  pgtype.UUID
}

type ExpenseCategoryService interface {
	GetCategoryPresets() []CategoryPreset
	ListCategoriesForGroup(ctx context.Context, groupID, requesterID pgtype.UUID) ([]sqlc.ExpenseCategory, error)
	CreateCategoryForGroup(ctx context.Context, input CreateGroupCategoryInput) (sqlc.ExpenseCategory, error)
	UpdateCategory(ctx context.Context, input UpdateCategoryInput) (sqlc.ExpenseCategory, error)
	DeleteCategory(ctx context.Context, categoryID, groupID, requesterID pgtype.UUID) error
	CreateCategoriesFromPresets(ctx context.Context, groupID, userID pgtype.UUID, presetSlugs []string) ([]sqlc.ExpenseCategory, error)
}

type expenseCategoryService struct {
	repo repository.ExpenseCategoryRepository
}

func NewExpenseCategoryService(repo repository.ExpenseCategoryRepository) ExpenseCategoryService {
	return &expenseCategoryService{repo: repo}
}

func (s *expenseCategoryService) GetCategoryPresets() []CategoryPreset {
	return SystemCategoryPresets
}

func (s *expenseCategoryService) ListCategoriesForGroup(ctx context.Context, groupID, requesterID pgtype.UUID) ([]sqlc.ExpenseCategory, error) {
	// Validate user is group member
	isMember, err := s.repo.IsGroupMember(ctx, groupID, requesterID)
	if err != nil {
		return nil, err
	}
	if !isMember {
		return nil, ErrNotGroupMember
	}

	return s.repo.ListCategoriesForGroup(ctx, groupID)
}

func (s *expenseCategoryService) CreateCategoryForGroup(ctx context.Context, input CreateGroupCategoryInput) (sqlc.ExpenseCategory, error) {
	// 1. Validate user is group member
	isMember, err := s.repo.IsGroupMember(ctx, input.GroupID, input.CreatedBy)
	if err != nil {
		return sqlc.ExpenseCategory{}, err
	}
	if !isMember {
		return sqlc.ExpenseCategory{}, ErrNotGroupMember
	}

	// 2. Validate and generate slug
	name := strings.TrimSpace(input.Name)
	if name == "" {
		return sqlc.ExpenseCategory{}, ErrInvalidCategoryName
	}

	slug := GenerateSlug(name)
	if slug == "" {
		return sqlc.ExpenseCategory{}, ErrInvalidCategoryName
	}

	// 3. Create category
	category, err := s.repo.CreateCategory(ctx, sqlc.CreateGroupCategoryParams{
		GroupID:   input.GroupID,
		Slug:      slug,
		Name:      name,
		Icon:      pgtype.Text{String: input.Icon, Valid: input.Icon != ""},
		Color:     pgtype.Text{String: input.Color, Valid: input.Color != ""},
		CreatedBy: input.CreatedBy,
	})
	if err != nil {
		// Check for unique constraint violation
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return sqlc.ExpenseCategory{}, ErrCategoryAlreadyExists
		}
		return sqlc.ExpenseCategory{}, err
	}

	return category, nil
}

func (s *expenseCategoryService) UpdateCategory(ctx context.Context, input UpdateCategoryInput) (sqlc.ExpenseCategory, error) {
	// 1. Get existing category
	category, err := s.repo.GetCategoryByID(ctx, input.CategoryID)
	if err != nil {
		return sqlc.ExpenseCategory{}, ErrCategoryNotFound
	}

	// 2. Validate category belongs to the group
	if category.GroupID.Bytes != input.GroupID.Bytes {
		return sqlc.ExpenseCategory{}, ErrCategoryNotFound
	}

	// 3. Validate user is group member
	isMember, err := s.repo.IsGroupMember(ctx, input.GroupID, input.UpdatedBy)
	if err != nil {
		return sqlc.ExpenseCategory{}, err
	}
	if !isMember {
		return sqlc.ExpenseCategory{}, ErrNotGroupMember
	}

	// 4. Validate and generate slug
	name := strings.TrimSpace(input.Name)
	if name == "" {
		return sqlc.ExpenseCategory{}, ErrInvalidCategoryName
	}

	slug := GenerateSlug(name)
	if slug == "" {
		return sqlc.ExpenseCategory{}, ErrInvalidCategoryName
	}

	// 5. Update category
	updated, err := s.repo.UpdateCategory(ctx, sqlc.UpdateGroupCategoryParams{
		ID:        input.CategoryID,
		Slug:      slug,
		Name:      name,
		Icon:      pgtype.Text{String: input.Icon, Valid: input.Icon != ""},
		Color:     pgtype.Text{String: input.Color, Valid: input.Color != ""},
		UpdatedBy: input.UpdatedBy,
	})
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return sqlc.ExpenseCategory{}, ErrCategoryAlreadyExists
		}
		return sqlc.ExpenseCategory{}, err
	}

	return updated, nil
}

func (s *expenseCategoryService) DeleteCategory(ctx context.Context, categoryID, groupID, requesterID pgtype.UUID) error {
	// 1. Get existing category
	category, err := s.repo.GetCategoryByID(ctx, categoryID)
	if err != nil {
		return ErrCategoryNotFound
	}

	// 2. Validate category belongs to the group
	if category.GroupID.Bytes != groupID.Bytes {
		return ErrCategoryNotFound
	}

	// 3. Validate user is group member
	isMember, err := s.repo.IsGroupMember(ctx, groupID, requesterID)
	if err != nil {
		return err
	}
	if !isMember {
		return ErrNotGroupMember
	}

	// 4. Delete category
	return s.repo.DeleteCategory(ctx, categoryID, requesterID)
}

func (s *expenseCategoryService) CreateCategoriesFromPresets(ctx context.Context, groupID, userID pgtype.UUID, presetSlugs []string) ([]sqlc.ExpenseCategory, error) {
	// Validate user is group member
	isMember, err := s.repo.IsGroupMember(ctx, groupID, userID)
	if err != nil {
		return nil, err
	}
	if !isMember {
		return nil, ErrNotGroupMember
	}

	categories := make([]sqlc.ExpenseCategory, 0, len(presetSlugs))

	for _, presetSlug := range presetSlugs {
		// Find preset by slug
		preset := GetPresetBySlug(presetSlug)
		if preset == nil {
			// Skip unknown preset slugs
			continue
		}

		// Check if category already exists in this group
		existing, err := s.repo.GetCategoryBySlug(ctx, groupID, preset.Slug)
		if err == nil && existing.ID.Valid {
			// Category already exists, add to result
			categories = append(categories, existing)
			continue
		}

		// Create category from preset
		cat, err := s.CreateCategoryForGroup(ctx, CreateGroupCategoryInput{
			GroupID:   groupID,
			Name:      preset.Name,
			Icon:      preset.Icon,
			Color:     preset.Color,
			CreatedBy: userID,
		})
		if err != nil {
			// Log error but continue with other presets
			continue
		}

		categories = append(categories, cat)
	}

	return categories, nil
}
