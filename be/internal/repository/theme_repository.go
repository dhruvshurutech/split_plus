package repository

import (
	"context"

	"github.com/dhruvsaxena1998/splitplus/internal/db/sqlc"
	"github.com/jackc/pgx/v5/pgtype"
)

type ThemeRepository interface {
	ListThemePresets(ctx context.Context) ([]sqlc.ThemePreset, error)
	GetThemePresetBySlug(ctx context.Context, slug string) (sqlc.ThemePreset, error)

	ListUserThemes(ctx context.Context, userID pgtype.UUID) ([]sqlc.UserTheme, error)
	GetUserThemeByID(ctx context.Context, id pgtype.UUID) (sqlc.UserTheme, error)
	CreateUserTheme(ctx context.Context, params sqlc.CreateUserThemeParams) (sqlc.UserTheme, error)
	UpdateUserTheme(ctx context.Context, params sqlc.UpdateUserThemeParams) (sqlc.UserTheme, error)
	DeleteUserTheme(ctx context.Context, id pgtype.UUID) error

	GetUserThemePreferences(ctx context.Context, userID pgtype.UUID) (sqlc.UserThemePreference, error)
	UpsertUserThemePreferences(ctx context.Context, params sqlc.UpsertUserThemePreferencesParams) (sqlc.UserThemePreference, error)
}

type themeRepository struct {
	queries *sqlc.Queries
}

func NewThemeRepository(queries *sqlc.Queries) ThemeRepository {
	return &themeRepository{queries: queries}
}

func (r *themeRepository) ListThemePresets(ctx context.Context) ([]sqlc.ThemePreset, error) {
	return r.queries.ListThemePresets(ctx)
}

func (r *themeRepository) GetThemePresetBySlug(ctx context.Context, slug string) (sqlc.ThemePreset, error) {
	return r.queries.GetThemePresetBySlug(ctx, slug)
}

func (r *themeRepository) ListUserThemes(ctx context.Context, userID pgtype.UUID) ([]sqlc.UserTheme, error) {
	return r.queries.ListUserThemes(ctx, userID)
}

func (r *themeRepository) GetUserThemeByID(ctx context.Context, id pgtype.UUID) (sqlc.UserTheme, error) {
	return r.queries.GetUserThemeByID(ctx, id)
}

func (r *themeRepository) CreateUserTheme(ctx context.Context, params sqlc.CreateUserThemeParams) (sqlc.UserTheme, error) {
	return r.queries.CreateUserTheme(ctx, params)
}

func (r *themeRepository) UpdateUserTheme(ctx context.Context, params sqlc.UpdateUserThemeParams) (sqlc.UserTheme, error) {
	return r.queries.UpdateUserTheme(ctx, params)
}

func (r *themeRepository) DeleteUserTheme(ctx context.Context, id pgtype.UUID) error {
	return r.queries.DeleteUserTheme(ctx, id)
}

func (r *themeRepository) GetUserThemePreferences(ctx context.Context, userID pgtype.UUID) (sqlc.UserThemePreference, error) {
	return r.queries.GetUserThemePreferences(ctx, userID)
}

func (r *themeRepository) UpsertUserThemePreferences(ctx context.Context, params sqlc.UpsertUserThemePreferencesParams) (sqlc.UserThemePreference, error) {
	return r.queries.UpsertUserThemePreferences(ctx, params)
}
