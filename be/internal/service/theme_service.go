package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/dhruvsaxena1998/splitplus/internal/db/sqlc"
	"github.com/dhruvsaxena1998/splitplus/internal/repository"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

const (
	ThemeActiveTypePreset = "preset"
	ThemeActiveTypeCustom = "custom"
	ThemeModeLight        = "light"
	ThemeModeDark         = "dark"
	ThemeDefaultPreset    = "soft-professional"
	ThemeDefaultMode      = ThemeModeLight
	ThemeCustomLimit      = 25
)

var (
	ErrThemeNotFound          = errors.New("theme not found")
	ErrThemePresetNotFound    = errors.New("theme preset not found")
	ErrThemePreferencesNotSet = errors.New("theme preferences not set")
	ErrThemeForbidden         = errors.New("forbidden")
	ErrThemeNameRequired      = errors.New("theme name is required")
	ErrThemeNameExists        = errors.New("theme name already exists")
	ErrThemeModeInvalid       = errors.New("invalid theme mode")
	ErrThemeActiveInvalid     = errors.New("invalid active theme")
	ErrThemeFontInvalid       = errors.New("invalid font family")
	ErrThemeTokensInvalid     = errors.New("invalid theme tokens")
	ErrThemeLimitReached      = errors.New("theme limit reached")
)

var (
	themeRadiusPattern = regexp.MustCompile(`^\d+(\.\d+)?(rem|px|%)$`)
	themeColorPattern  = regexp.MustCompile(`^(#[0-9a-fA-F]{3,8}|oklch\([^)]+\)|rgb(a)?\([^)]+\)|hsl(a)?\([^)]+\)|var\(--[a-z0-9-]+\))$`)
)

var allowedThemeTokenKeys = map[string]bool{
	"background":           true,
	"foreground":           true,
	"card":                 true,
	"card-foreground":      true,
	"primary":              true,
	"primary-foreground":   true,
	"secondary":            true,
	"secondary-foreground": true,
	"muted":                true,
	"muted-foreground":     true,
	"accent":               true,
	"accent-foreground":    true,
	"border":               true,
	"input":                true,
	"ring":                 true,
	"radius":               true,
	"dock-background":      true,
	"dock-active":          true,
}

var allowedFontFamilyKeys = map[string]bool{
	"avenir":          true,
	"newspaper-serif": true,
	"geist-mono":      true,
	"charter-serif":   true,
}

var presetDefaultFont = map[string]string{
	"minimal":           "avenir",
	"soft-professional": "avenir",
	"atelier":           "avenir",
	"newspaper":         "newspaper-serif",
	"mono-slate":        "geist-mono",
}

type ThemeService interface {
	ListThemePresets(ctx context.Context) ([]sqlc.ThemePreset, error)
	ListUserThemes(ctx context.Context, userID pgtype.UUID) ([]ThemeUserTheme, error)
	GetUserTheme(ctx context.Context, userID, themeID pgtype.UUID) (ThemeUserTheme, error)
	CreateUserTheme(ctx context.Context, userID pgtype.UUID, input CreateUserThemeInput) (ThemeUserTheme, error)
	UpdateUserTheme(ctx context.Context, userID, themeID pgtype.UUID, input UpdateUserThemeInput) (ThemeUserTheme, error)
	DeleteUserTheme(ctx context.Context, userID, themeID pgtype.UUID) error
	GetPreferences(ctx context.Context, userID pgtype.UUID, legacyTheme, legacyMode string) (ThemePreferencesView, error)
	UpdatePreferences(ctx context.Context, userID pgtype.UUID, input UpdateThemePreferencesInput) (ThemePreferencesView, error)
}

type themeService struct {
	repo repository.ThemeRepository
}

func NewThemeService(repo repository.ThemeRepository) ThemeService {
	return &themeService{repo: repo}
}

type ThemeUserTheme struct {
	ID             string            `json:"id"`
	Name           string            `json:"name"`
	BasePresetSlug string            `json:"base_preset_slug"`
	FontFamilyKey  string            `json:"font_family_key"`
	LightTokens    map[string]string `json:"light_tokens"`
	DarkTokens     map[string]string `json:"dark_tokens"`
	CreatedAt      string            `json:"created_at"`
	UpdatedAt      string            `json:"updated_at"`
}

type ThemeRef struct {
	Type        string `json:"type"`
	PresetSlug  string `json:"preset_slug,omitempty"`
	UserThemeID string `json:"user_theme_id,omitempty"`
}

type ThemeResolved struct {
	BasePresetSlug string            `json:"base_preset_slug"`
	FontFamilyKey  string            `json:"font_family_key"`
	Tokens         map[string]string `json:"tokens"`
}

type ThemePreferencesView struct {
	Mode          string        `json:"mode"`
	Active        ThemeRef      `json:"active"`
	ResolvedTheme ThemeResolved `json:"resolved_theme"`
}

type CreateUserThemeInput struct {
	Name           string            `json:"name"`
	BasePresetSlug string            `json:"base_preset_slug"`
	FontFamilyKey  string            `json:"font_family_key"`
	LightTokens    map[string]string `json:"light_tokens"`
	DarkTokens     map[string]string `json:"dark_tokens"`
}

type UpdateUserThemeInput struct {
	Name           string            `json:"name"`
	BasePresetSlug string            `json:"base_preset_slug"`
	FontFamilyKey  string            `json:"font_family_key"`
	LightTokens    map[string]string `json:"light_tokens"`
	DarkTokens     map[string]string `json:"dark_tokens"`
}

type UpdateThemePreferencesInput struct {
	Mode              string `json:"mode"`
	ActiveType        string `json:"active_type"`
	ActivePresetSlug  string `json:"active_preset_slug"`
	ActiveUserThemeID string `json:"active_user_theme_id"`
}

func (s *themeService) ListThemePresets(ctx context.Context) ([]sqlc.ThemePreset, error) {
	return s.repo.ListThemePresets(ctx)
}

func (s *themeService) ListUserThemes(ctx context.Context, userID pgtype.UUID) ([]ThemeUserTheme, error) {
	themes, err := s.repo.ListUserThemes(ctx, userID)
	if err != nil {
		return nil, err
	}
	resp := make([]ThemeUserTheme, 0, len(themes))
	for _, t := range themes {
		mapped, mapErr := mapUserTheme(t)
		if mapErr != nil {
			return nil, mapErr
		}
		resp = append(resp, mapped)
	}
	return resp, nil
}

func (s *themeService) GetUserTheme(ctx context.Context, userID, themeID pgtype.UUID) (ThemeUserTheme, error) {
	theme, err := s.repo.GetUserThemeByID(ctx, themeID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ThemeUserTheme{}, ErrThemeNotFound
		}
		return ThemeUserTheme{}, err
	}
	if theme.UserID != userID {
		return ThemeUserTheme{}, ErrThemeForbidden
	}
	return mapUserTheme(theme)
}

func (s *themeService) CreateUserTheme(ctx context.Context, userID pgtype.UUID, input CreateUserThemeInput) (ThemeUserTheme, error) {
	if err := validateThemeInput(input.Name, input.BasePresetSlug, input.FontFamilyKey, input.LightTokens, input.DarkTokens); err != nil {
		return ThemeUserTheme{}, err
	}
	if err := s.ensurePresetExists(ctx, input.BasePresetSlug); err != nil {
		return ThemeUserTheme{}, err
	}
	existing, err := s.repo.ListUserThemes(ctx, userID)
	if err != nil {
		return ThemeUserTheme{}, err
	}
	if len(existing) >= ThemeCustomLimit {
		return ThemeUserTheme{}, ErrThemeLimitReached
	}
	lightRaw, err := json.Marshal(input.LightTokens)
	if err != nil {
		return ThemeUserTheme{}, err
	}
	darkRaw, err := json.Marshal(input.DarkTokens)
	if err != nil {
		return ThemeUserTheme{}, err
	}
	theme, err := s.repo.CreateUserTheme(ctx, sqlc.CreateUserThemeParams{
		UserID:         userID,
		Name:           strings.TrimSpace(input.Name),
		BasePresetSlug: input.BasePresetSlug,
		FontFamilyKey:  input.FontFamilyKey,
		LightTokens:    lightRaw,
		DarkTokens:     darkRaw,
	})
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return ThemeUserTheme{}, ErrThemeNameExists
		}
		if strings.Contains(err.Error(), "duplicate") {
			return ThemeUserTheme{}, ErrThemeNameExists
		}
		return ThemeUserTheme{}, err
	}
	return mapUserTheme(theme)
}

func (s *themeService) UpdateUserTheme(ctx context.Context, userID, themeID pgtype.UUID, input UpdateUserThemeInput) (ThemeUserTheme, error) {
	if err := validateThemeInput(input.Name, input.BasePresetSlug, input.FontFamilyKey, input.LightTokens, input.DarkTokens); err != nil {
		return ThemeUserTheme{}, err
	}
	if err := s.ensurePresetExists(ctx, input.BasePresetSlug); err != nil {
		return ThemeUserTheme{}, err
	}
	existing, err := s.repo.GetUserThemeByID(ctx, themeID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ThemeUserTheme{}, ErrThemeNotFound
		}
		return ThemeUserTheme{}, err
	}
	if existing.UserID != userID {
		return ThemeUserTheme{}, ErrThemeForbidden
	}
	lightRaw, err := json.Marshal(input.LightTokens)
	if err != nil {
		return ThemeUserTheme{}, err
	}
	darkRaw, err := json.Marshal(input.DarkTokens)
	if err != nil {
		return ThemeUserTheme{}, err
	}
	updated, err := s.repo.UpdateUserTheme(ctx, sqlc.UpdateUserThemeParams{
		ID:             themeID,
		Name:           strings.TrimSpace(input.Name),
		BasePresetSlug: input.BasePresetSlug,
		FontFamilyKey:  input.FontFamilyKey,
		LightTokens:    lightRaw,
		DarkTokens:     darkRaw,
	})
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return ThemeUserTheme{}, ErrThemeNameExists
		}
		return ThemeUserTheme{}, err
	}
	return mapUserTheme(updated)
}

func (s *themeService) DeleteUserTheme(ctx context.Context, userID, themeID pgtype.UUID) error {
	theme, err := s.repo.GetUserThemeByID(ctx, themeID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrThemeNotFound
		}
		return err
	}
	if theme.UserID != userID {
		return ErrThemeForbidden
	}
	if err := s.repo.DeleteUserTheme(ctx, themeID); err != nil {
		return err
	}

	pref, err := s.repo.GetUserThemePreferences(ctx, userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil
		}
		return err
	}
	if pref.ActiveType == ThemeActiveTypeCustom && pref.ActiveUserThemeID == themeID {
		_, err = s.repo.UpsertUserThemePreferences(ctx, sqlc.UpsertUserThemePreferencesParams{
			UserID:            userID,
			ActiveType:        ThemeActiveTypePreset,
			ActivePresetSlug:  pgtype.Text{String: ThemeDefaultPreset, Valid: true},
			ActiveUserThemeID: pgtype.UUID{},
			Mode:              pref.Mode,
		})
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *themeService) GetPreferences(ctx context.Context, userID pgtype.UUID, legacyTheme, legacyMode string) (ThemePreferencesView, error) {
	pref, err := s.repo.GetUserThemePreferences(ctx, userID)
	if err != nil {
		if !errors.Is(err, pgx.ErrNoRows) {
			return ThemePreferencesView{}, err
		}
		seedPreset := normalizeLegacyTheme(legacyTheme)
		seedMode := normalizeLegacyMode(legacyMode)
		pref, err = s.repo.UpsertUserThemePreferences(ctx, sqlc.UpsertUserThemePreferencesParams{
			UserID:            userID,
			ActiveType:        ThemeActiveTypePreset,
			ActivePresetSlug:  pgtype.Text{String: seedPreset, Valid: true},
			ActiveUserThemeID: pgtype.UUID{},
			Mode:              seedMode,
		})
		if err != nil {
			return ThemePreferencesView{}, err
		}
	}

	view, resolveErr := s.resolvePreferenceView(ctx, userID, pref)
	if resolveErr == nil {
		return view, nil
	}
	if !errors.Is(resolveErr, ErrThemeNotFound) && !errors.Is(resolveErr, ErrThemePresetNotFound) {
		return ThemePreferencesView{}, resolveErr
	}

	fallback, err := s.repo.UpsertUserThemePreferences(ctx, sqlc.UpsertUserThemePreferencesParams{
		UserID:            userID,
		ActiveType:        ThemeActiveTypePreset,
		ActivePresetSlug:  pgtype.Text{String: ThemeDefaultPreset, Valid: true},
		ActiveUserThemeID: pgtype.UUID{},
		Mode:              normalizeLegacyMode(pref.Mode),
	})
	if err != nil {
		return ThemePreferencesView{}, err
	}
	return s.resolvePreferenceView(ctx, userID, fallback)
}

func (s *themeService) UpdatePreferences(ctx context.Context, userID pgtype.UUID, input UpdateThemePreferencesInput) (ThemePreferencesView, error) {
	mode := normalizeLegacyMode(input.Mode)
	if mode != input.Mode && input.Mode != "" {
		if input.Mode != ThemeModeLight && input.Mode != ThemeModeDark {
			return ThemePreferencesView{}, ErrThemeModeInvalid
		}
	}
	activeType := strings.TrimSpace(input.ActiveType)
	if activeType != ThemeActiveTypePreset && activeType != ThemeActiveTypeCustom {
		return ThemePreferencesView{}, ErrThemeActiveInvalid
	}

	params := sqlc.UpsertUserThemePreferencesParams{
		UserID:            userID,
		ActiveType:        activeType,
		ActivePresetSlug:  pgtype.Text{},
		ActiveUserThemeID: pgtype.UUID{},
		Mode:              mode,
	}

	if activeType == ThemeActiveTypePreset {
		slug := strings.TrimSpace(input.ActivePresetSlug)
		if slug == "" {
			return ThemePreferencesView{}, ErrThemeActiveInvalid
		}
		if err := s.ensurePresetExists(ctx, slug); err != nil {
			return ThemePreferencesView{}, err
		}
		params.ActivePresetSlug = pgtype.Text{String: slug, Valid: true}
	} else {
		if strings.TrimSpace(input.ActiveUserThemeID) == "" {
			return ThemePreferencesView{}, ErrThemeActiveInvalid
		}
		themeID, err := parseUUIDString(input.ActiveUserThemeID)
		if err != nil {
			return ThemePreferencesView{}, ErrThemeActiveInvalid
		}
		theme, err := s.repo.GetUserThemeByID(ctx, themeID)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return ThemePreferencesView{}, ErrThemeNotFound
			}
			return ThemePreferencesView{}, err
		}
		if theme.UserID != userID {
			return ThemePreferencesView{}, ErrThemeForbidden
		}
		params.ActiveUserThemeID = themeID
	}

	pref, err := s.repo.UpsertUserThemePreferences(ctx, params)
	if err != nil {
		return ThemePreferencesView{}, err
	}
	return s.resolvePreferenceView(ctx, userID, pref)
}

func (s *themeService) resolvePreferenceView(ctx context.Context, userID pgtype.UUID, pref sqlc.UserThemePreference) (ThemePreferencesView, error) {
	view := ThemePreferencesView{
		Mode:   normalizeLegacyMode(pref.Mode),
		Active: ThemeRef{Type: pref.ActiveType},
	}

	if pref.ActiveType == ThemeActiveTypePreset {
		slug := pref.ActivePresetSlug.String
		if slug == "" {
			return ThemePreferencesView{}, ErrThemePresetNotFound
		}
		if err := s.ensurePresetExists(ctx, slug); err != nil {
			return ThemePreferencesView{}, err
		}
		view.Active.PresetSlug = slug
		view.ResolvedTheme = ThemeResolved{
			BasePresetSlug: slug,
			FontFamilyKey:  defaultFontForPreset(slug),
			Tokens:         map[string]string{},
		}
		return view, nil
	}

	if pref.ActiveType == ThemeActiveTypeCustom {
		theme, err := s.repo.GetUserThemeByID(ctx, pref.ActiveUserThemeID)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return ThemePreferencesView{}, ErrThemeNotFound
			}
			return ThemePreferencesView{}, err
		}
		if theme.UserID != userID {
			return ThemePreferencesView{}, ErrThemeForbidden
		}
		if err := s.ensurePresetExists(ctx, theme.BasePresetSlug); err != nil {
			return ThemePreferencesView{}, err
		}
		tokens, err := parseTokensByMode(theme, view.Mode)
		if err != nil {
			return ThemePreferencesView{}, err
		}
		view.Active.UserThemeID = themeUUIDString(theme.ID)
		view.ResolvedTheme = ThemeResolved{
			BasePresetSlug: theme.BasePresetSlug,
			FontFamilyKey:  theme.FontFamilyKey,
			Tokens:         tokens,
		}
		return view, nil
	}

	return ThemePreferencesView{}, ErrThemeActiveInvalid
}

func (s *themeService) ensurePresetExists(ctx context.Context, slug string) error {
	_, err := s.repo.GetThemePresetBySlug(ctx, slug)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrThemePresetNotFound
		}
		return err
	}
	return nil
}

func mapUserTheme(theme sqlc.UserTheme) (ThemeUserTheme, error) {
	lightTokens, err := parseTokenBytes(theme.LightTokens)
	if err != nil {
		return ThemeUserTheme{}, err
	}
	darkTokens, err := parseTokenBytes(theme.DarkTokens)
	if err != nil {
		return ThemeUserTheme{}, err
	}
	return ThemeUserTheme{
		ID:             themeUUIDString(theme.ID),
		Name:           theme.Name,
		BasePresetSlug: theme.BasePresetSlug,
		FontFamilyKey:  theme.FontFamilyKey,
		LightTokens:    lightTokens,
		DarkTokens:     darkTokens,
		CreatedAt:      theme.CreatedAt.Time.String(),
		UpdatedAt:      theme.UpdatedAt.Time.String(),
	}, nil
}

func parseTokensByMode(theme sqlc.UserTheme, mode string) (map[string]string, error) {
	if mode == ThemeModeDark {
		return parseTokenBytes(theme.DarkTokens)
	}
	return parseTokenBytes(theme.LightTokens)
}

func parseTokenBytes(raw []byte) (map[string]string, error) {
	if len(raw) == 0 {
		return map[string]string{}, nil
	}
	var tokens map[string]string
	if err := json.Unmarshal(raw, &tokens); err != nil {
		return nil, err
	}
	if tokens == nil {
		return map[string]string{}, nil
	}
	return tokens, nil
}

func validateThemeInput(name, basePresetSlug, fontFamilyKey string, lightTokens, darkTokens map[string]string) error {
	trimmedName := strings.TrimSpace(name)
	if trimmedName == "" {
		return ErrThemeNameRequired
	}
	if len(trimmedName) > 60 {
		return ErrThemeNameRequired
	}
	if strings.TrimSpace(basePresetSlug) == "" {
		return ErrThemePresetNotFound
	}
	if !allowedFontFamilyKeys[fontFamilyKey] {
		return ErrThemeFontInvalid
	}
	if err := validateThemeTokens(lightTokens); err != nil {
		return err
	}
	if err := validateThemeTokens(darkTokens); err != nil {
		return err
	}
	return nil
}

func validateThemeTokens(tokens map[string]string) error {
	for k, v := range tokens {
		if !allowedThemeTokenKeys[k] {
			return fmt.Errorf("%w: key %s", ErrThemeTokensInvalid, k)
		}
		val := strings.TrimSpace(v)
		if val == "" || len(val) > 120 {
			return fmt.Errorf("%w: value %s", ErrThemeTokensInvalid, k)
		}
		if k == "radius" {
			if !themeRadiusPattern.MatchString(val) {
				return fmt.Errorf("%w: radius", ErrThemeTokensInvalid)
			}
			continue
		}
		if strings.HasPrefix(val, "color-mix(") && strings.HasSuffix(val, ")") {
			continue
		}
		if !themeColorPattern.MatchString(val) {
			return fmt.Errorf("%w: color", ErrThemeTokensInvalid)
		}
	}
	return nil
}

func normalizeLegacyTheme(theme string) string {
	theme = strings.TrimSpace(theme)
	switch theme {
	case "minimal", "soft-professional", "atelier", "newspaper", "mono-slate":
		return theme
	case "light", "text-minimal":
		return "minimal"
	case "dark":
		return ThemeDefaultPreset
	default:
		return ThemeDefaultPreset
	}
}

func normalizeLegacyMode(mode string) string {
	if mode == ThemeModeDark {
		return ThemeModeDark
	}
	return ThemeModeLight
}

func parseUUIDString(id string) (pgtype.UUID, error) {
	var parsed pgtype.UUID
	err := parsed.Scan(id)
	return parsed, err
}

func themeUUIDString(id pgtype.UUID) string {
	if !id.Valid {
		return ""
	}
	b := id.Bytes
	buf := make([]byte, 36)
	hex := "0123456789abcdef"
	j := 0
	for i := 0; i < 16; i++ {
		if i == 4 || i == 6 || i == 8 || i == 10 {
			buf[j] = '-'
			j++
		}
		buf[j] = hex[b[i]>>4]
		j++
		buf[j] = hex[b[i]&0x0f]
		j++
	}
	return string(buf)
}

func defaultFontForPreset(slug string) string {
	font, ok := presetDefaultFont[slug]
	if ok {
		return font
	}
	return "avenir"
}
