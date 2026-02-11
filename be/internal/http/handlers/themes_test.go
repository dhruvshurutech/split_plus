package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dhruvsaxena1998/splitplus/internal/db/sqlc"
	"github.com/dhruvsaxena1998/splitplus/internal/http/middleware"
	"github.com/dhruvsaxena1998/splitplus/internal/http/response"
	"github.com/dhruvsaxena1998/splitplus/internal/service"
	"github.com/go-playground/validator/v10"
	"github.com/jackc/pgx/v5/pgtype"
)

type stubThemeService struct {
	createUserThemeFunc func(ctx context.Context, userID pgtype.UUID, input service.CreateUserThemeInput) (service.ThemeUserTheme, error)
}

func (s *stubThemeService) ListThemePresets(ctx context.Context) ([]sqlc.ThemePreset, error) {
	return nil, nil
}
func (s *stubThemeService) ListUserThemes(ctx context.Context, userID pgtype.UUID) ([]service.ThemeUserTheme, error) {
	return nil, nil
}
func (s *stubThemeService) GetUserTheme(ctx context.Context, userID, themeID pgtype.UUID) (service.ThemeUserTheme, error) {
	return service.ThemeUserTheme{}, nil
}
func (s *stubThemeService) CreateUserTheme(ctx context.Context, userID pgtype.UUID, input service.CreateUserThemeInput) (service.ThemeUserTheme, error) {
	if s.createUserThemeFunc != nil {
		return s.createUserThemeFunc(ctx, userID, input)
	}
	return service.ThemeUserTheme{}, nil
}
func (s *stubThemeService) UpdateUserTheme(ctx context.Context, userID, themeID pgtype.UUID, input service.UpdateUserThemeInput) (service.ThemeUserTheme, error) {
	return service.ThemeUserTheme{}, nil
}
func (s *stubThemeService) DeleteUserTheme(ctx context.Context, userID, themeID pgtype.UUID) error {
	return nil
}
func (s *stubThemeService) GetPreferences(ctx context.Context, userID pgtype.UUID, legacyTheme, legacyMode string) (service.ThemePreferencesView, error) {
	return service.ThemePreferencesView{}, nil
}
func (s *stubThemeService) UpdatePreferences(ctx context.Context, userID pgtype.UUID, input service.UpdateThemePreferencesInput) (service.ThemePreferencesView, error) {
	return service.ThemePreferencesView{}, nil
}

func TestGetThemePreferencesHandler_Unauthorized(t *testing.T) {
	h := GetThemePreferencesHandler(&stubThemeService{})
	req := httptest.NewRequest(http.MethodGet, "/users/me/theme/preferences", nil)
	w := httptest.NewRecorder()

	h.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}

func TestCreateUserThemeHandler_InvalidPayloadFromService(t *testing.T) {
	mockSvc := &stubThemeService{
		createUserThemeFunc: func(ctx context.Context, userID pgtype.UUID, input service.CreateUserThemeInput) (service.ThemeUserTheme, error) {
			return service.ThemeUserTheme{}, service.ErrThemeTokensInvalid
		},
	}
	handler := CreateUserThemeHandler(mockSvc)
	validate := validator.New()

	body, _ := json.Marshal(CreateThemeRequest{
		Name:           "Theme",
		BasePresetSlug: "minimal",
		FontFamilyKey:  "avenir",
	})

	req := httptest.NewRequest(http.MethodPost, "/users/me/themes", bytes.NewReader(body))
	req = req.WithContext(middleware.SetUserID(req.Context(), mustParseUUID(t, "11111111-1111-1111-1111-111111111111")))
	w := httptest.NewRecorder()

	wrapped := middleware.ValidateBodyWithScope[CreateThemeRequest](validate, "theme")(handler)
	wrapped.ServeHTTP(w, req)

	if w.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected 422, got %d", w.Code)
	}

	var resp response.StandardResponse[interface{}, response.ErrorDetail]
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	if resp.Error == nil || resp.Error.Code != "validation.theme.invalid" {
		t.Fatalf("expected validation.theme.invalid, got %+v", resp.Error)
	}
}

func mustParseUUID(t *testing.T, val string) pgtype.UUID {
	t.Helper()
	var out pgtype.UUID
	if err := out.Scan(val); err != nil {
		t.Fatalf("invalid uuid: %v", err)
	}
	return out
}
