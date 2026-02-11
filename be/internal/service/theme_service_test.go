package service

import (
	"context"
	"errors"
	"testing"

	"github.com/dhruvsaxena1998/splitplus/internal/db/sqlc"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type mockThemeRepository struct {
	listThemePresetsFunc           func(ctx context.Context) ([]sqlc.ThemePreset, error)
	getThemePresetBySlugFunc       func(ctx context.Context, slug string) (sqlc.ThemePreset, error)
	listUserThemesFunc             func(ctx context.Context, userID pgtype.UUID) ([]sqlc.UserTheme, error)
	getUserThemeByIDFunc           func(ctx context.Context, id pgtype.UUID) (sqlc.UserTheme, error)
	createUserThemeFunc            func(ctx context.Context, params sqlc.CreateUserThemeParams) (sqlc.UserTheme, error)
	updateUserThemeFunc            func(ctx context.Context, params sqlc.UpdateUserThemeParams) (sqlc.UserTheme, error)
	deleteUserThemeFunc            func(ctx context.Context, id pgtype.UUID) error
	getUserThemePreferencesFunc    func(ctx context.Context, userID pgtype.UUID) (sqlc.UserThemePreference, error)
	upsertUserThemePreferencesFunc func(ctx context.Context, params sqlc.UpsertUserThemePreferencesParams) (sqlc.UserThemePreference, error)
}

func (m *mockThemeRepository) ListThemePresets(ctx context.Context) ([]sqlc.ThemePreset, error) {
	if m.listThemePresetsFunc != nil {
		return m.listThemePresetsFunc(ctx)
	}
	return nil, nil
}

func (m *mockThemeRepository) GetThemePresetBySlug(ctx context.Context, slug string) (sqlc.ThemePreset, error) {
	if m.getThemePresetBySlugFunc != nil {
		return m.getThemePresetBySlugFunc(ctx, slug)
	}
	return sqlc.ThemePreset{}, nil
}

func (m *mockThemeRepository) ListUserThemes(ctx context.Context, userID pgtype.UUID) ([]sqlc.UserTheme, error) {
	if m.listUserThemesFunc != nil {
		return m.listUserThemesFunc(ctx, userID)
	}
	return nil, nil
}

func (m *mockThemeRepository) GetUserThemeByID(ctx context.Context, id pgtype.UUID) (sqlc.UserTheme, error) {
	if m.getUserThemeByIDFunc != nil {
		return m.getUserThemeByIDFunc(ctx, id)
	}
	return sqlc.UserTheme{}, nil
}

func (m *mockThemeRepository) CreateUserTheme(ctx context.Context, params sqlc.CreateUserThemeParams) (sqlc.UserTheme, error) {
	if m.createUserThemeFunc != nil {
		return m.createUserThemeFunc(ctx, params)
	}
	return sqlc.UserTheme{}, nil
}

func (m *mockThemeRepository) UpdateUserTheme(ctx context.Context, params sqlc.UpdateUserThemeParams) (sqlc.UserTheme, error) {
	if m.updateUserThemeFunc != nil {
		return m.updateUserThemeFunc(ctx, params)
	}
	return sqlc.UserTheme{}, nil
}

func (m *mockThemeRepository) DeleteUserTheme(ctx context.Context, id pgtype.UUID) error {
	if m.deleteUserThemeFunc != nil {
		return m.deleteUserThemeFunc(ctx, id)
	}
	return nil
}

func (m *mockThemeRepository) GetUserThemePreferences(ctx context.Context, userID pgtype.UUID) (sqlc.UserThemePreference, error) {
	if m.getUserThemePreferencesFunc != nil {
		return m.getUserThemePreferencesFunc(ctx, userID)
	}
	return sqlc.UserThemePreference{}, nil
}

func (m *mockThemeRepository) UpsertUserThemePreferences(ctx context.Context, params sqlc.UpsertUserThemePreferencesParams) (sqlc.UserThemePreference, error) {
	if m.upsertUserThemePreferencesFunc != nil {
		return m.upsertUserThemePreferencesFunc(ctx, params)
	}
	return sqlc.UserThemePreference{}, nil
}

func TestThemeService_GetPreferences_SeedsFromLegacy(t *testing.T) {
	repo := &mockThemeRepository{}
	repo.getUserThemePreferencesFunc = func(ctx context.Context, userID pgtype.UUID) (sqlc.UserThemePreference, error) {
		return sqlc.UserThemePreference{}, pgx.ErrNoRows
	}
	repo.upsertUserThemePreferencesFunc = func(ctx context.Context, params sqlc.UpsertUserThemePreferencesParams) (sqlc.UserThemePreference, error) {
		if params.ActivePresetSlug.String != "minimal" {
			t.Fatalf("expected seeded preset minimal, got %s", params.ActivePresetSlug.String)
		}
		if params.Mode != ThemeModeDark {
			t.Fatalf("expected seeded mode dark, got %s", params.Mode)
		}
		return sqlc.UserThemePreference{
			UserID:           params.UserID,
			ActiveType:       params.ActiveType,
			ActivePresetSlug: params.ActivePresetSlug,
			Mode:             params.Mode,
		}, nil
	}
	repo.getThemePresetBySlugFunc = func(ctx context.Context, slug string) (sqlc.ThemePreset, error) {
		return sqlc.ThemePreset{Slug: slug, IsActive: true}, nil
	}

	svc := NewThemeService(repo)
	view, err := svc.GetPreferences(context.Background(), mustUUID(t, "11111111-1111-1111-1111-111111111111"), "light", "dark")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if view.Active.Type != ThemeActiveTypePreset {
		t.Fatalf("expected active preset, got %s", view.Active.Type)
	}
	if view.Active.PresetSlug != "minimal" {
		t.Fatalf("expected preset minimal, got %s", view.Active.PresetSlug)
	}
	if view.Mode != ThemeModeDark {
		t.Fatalf("expected mode dark, got %s", view.Mode)
	}
}

func TestThemeService_CreateUserTheme_InvalidFont(t *testing.T) {
	svc := NewThemeService(&mockThemeRepository{})
	_, err := svc.CreateUserTheme(context.Background(), mustUUID(t, "11111111-1111-1111-1111-111111111111"), CreateUserThemeInput{
		Name:           "My Theme",
		BasePresetSlug: "minimal",
		FontFamilyKey:  "invalid-font",
		LightTokens:    map[string]string{},
		DarkTokens:     map[string]string{},
	})
	if !errors.Is(err, ErrThemeFontInvalid) {
		t.Fatalf("expected ErrThemeFontInvalid, got %v", err)
	}
}

func TestThemeService_UpdatePreferences_CustomForbidden(t *testing.T) {
	repo := &mockThemeRepository{}
	repo.getUserThemeByIDFunc = func(ctx context.Context, id pgtype.UUID) (sqlc.UserTheme, error) {
		return sqlc.UserTheme{UserID: mustUUID(t, "22222222-2222-2222-2222-222222222222")}, nil
	}

	svc := NewThemeService(repo)
	_, err := svc.UpdatePreferences(context.Background(), mustUUID(t, "11111111-1111-1111-1111-111111111111"), UpdateThemePreferencesInput{
		Mode:              ThemeModeLight,
		ActiveType:        ThemeActiveTypeCustom,
		ActiveUserThemeID: "33333333-3333-3333-3333-333333333333",
	})
	if !errors.Is(err, ErrThemeForbidden) {
		t.Fatalf("expected ErrThemeForbidden, got %v", err)
	}
}

func mustUUID(t *testing.T, value string) pgtype.UUID {
	t.Helper()
	var out pgtype.UUID
	if err := out.Scan(value); err != nil {
		t.Fatalf("invalid test uuid %q: %v", value, err)
	}
	return out
}
