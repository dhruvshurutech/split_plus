package handlers

import (
	"errors"
	"net/http"

	"github.com/dhruvsaxena1998/splitplus/internal/http/middleware"
	"github.com/dhruvsaxena1998/splitplus/internal/http/response"
	"github.com/dhruvsaxena1998/splitplus/internal/service"
	"github.com/go-chi/chi/v5"
)

type ThemePresetResponse struct {
	ID       string `json:"id"`
	Slug     string `json:"slug"`
	Label    string `json:"label"`
	Family   string `json:"family"`
	IsActive bool   `json:"is_active"`
}

type CreateThemeRequest struct {
	Name           string            `json:"name" validate:"required"`
	BasePresetSlug string            `json:"base_preset_slug" validate:"required"`
	FontFamilyKey  string            `json:"font_family_key" validate:"required"`
	LightTokens    map[string]string `json:"light_tokens"`
	DarkTokens     map[string]string `json:"dark_tokens"`
}

type UpdateThemeRequest struct {
	Name           string            `json:"name" validate:"required"`
	BasePresetSlug string            `json:"base_preset_slug" validate:"required"`
	FontFamilyKey  string            `json:"font_family_key" validate:"required"`
	LightTokens    map[string]string `json:"light_tokens"`
	DarkTokens     map[string]string `json:"dark_tokens"`
}

type UpdateThemePreferencesRequest struct {
	Mode              string `json:"mode" validate:"required,oneof=light dark"`
	ActiveType        string `json:"active_type" validate:"required,oneof=preset custom"`
	ActivePresetSlug  string `json:"active_preset_slug"`
	ActiveUserThemeID string `json:"active_user_theme_id"`
}

func ListThemePresetsHandler(themeService service.ThemeService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		presets, err := themeService.ListThemePresets(r.Context())
		if err != nil {
			response.SendErrorWithCode(w, http.StatusInternalServerError, "system.theme.presets_failed", "Unable to load theme presets.")
			return
		}

		resp := make([]ThemePresetResponse, 0, len(presets))
		for _, preset := range presets {
			resp = append(resp, ThemePresetResponse{
				ID:       formatUUID(preset.ID.Bytes),
				Slug:     preset.Slug,
				Label:    preset.Label,
				Family:   preset.Family,
				IsActive: preset.IsActive,
			})
		}

		response.SendSuccess(w, http.StatusOK, resp)
	}
}

func GetThemePreferencesHandler(themeService service.ThemeService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := middleware.GetUserID(r)
		if !ok {
			response.SendErrorWithCode(w, http.StatusUnauthorized, "auth.authorization.unauthorized", "Unauthorized.")
			return
		}

		legacyTheme := r.URL.Query().Get("legacy_theme")
		legacyMode := r.URL.Query().Get("legacy_mode")

		preferences, err := themeService.GetPreferences(r.Context(), userID, legacyTheme, legacyMode)
		if err != nil {
			response.SendErrorWithCode(w, http.StatusInternalServerError, "system.theme.preferences_failed", "Unable to load theme preferences.")
			return
		}

		response.SendSuccess(w, http.StatusOK, preferences)
	}
}

func UpdateThemePreferencesHandler(themeService service.ThemeService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := middleware.GetUserID(r)
		if !ok {
			response.SendErrorWithCode(w, http.StatusUnauthorized, "auth.authorization.unauthorized", "Unauthorized.")
			return
		}

		req, ok := middleware.GetBody[UpdateThemePreferencesRequest](r)
		if !ok {
			response.SendErrorWithCode(w, http.StatusInternalServerError, "system.request.context_invalid", "Invalid request context.")
			return
		}

		preferences, err := themeService.UpdatePreferences(r.Context(), userID, service.UpdateThemePreferencesInput{
			Mode:              req.Mode,
			ActiveType:        req.ActiveType,
			ActivePresetSlug:  req.ActivePresetSlug,
			ActiveUserThemeID: req.ActiveUserThemeID,
		})
		if err != nil {
			switch {
			case errors.Is(err, service.ErrThemeNotFound), errors.Is(err, service.ErrThemePresetNotFound):
				response.SendErrorWithCode(w, http.StatusNotFound, "resource.theme.not_found", "Theme not found.")
			case errors.Is(err, service.ErrThemeForbidden):
				response.SendErrorWithCode(w, http.StatusForbidden, "auth.authorization.forbidden", "You do not have access to this theme.")
			case errors.Is(err, service.ErrThemeModeInvalid), errors.Is(err, service.ErrThemeActiveInvalid):
				response.SendErrorWithCode(w, http.StatusUnprocessableEntity, "validation.theme.preferences.invalid", "Invalid theme preferences payload.")
			default:
				response.SendErrorWithCode(w, http.StatusInternalServerError, "system.theme.preferences_update_failed", "Unable to update theme preferences.")
			}
			return
		}

		response.SendSuccess(w, http.StatusOK, preferences)
	}
}

func ListUserThemesHandler(themeService service.ThemeService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := middleware.GetUserID(r)
		if !ok {
			response.SendErrorWithCode(w, http.StatusUnauthorized, "auth.authorization.unauthorized", "Unauthorized.")
			return
		}

		themes, err := themeService.ListUserThemes(r.Context(), userID)
		if err != nil {
			response.SendErrorWithCode(w, http.StatusInternalServerError, "system.theme.list_failed", "Unable to load themes.")
			return
		}

		response.SendSuccess(w, http.StatusOK, themes)
	}
}

func CreateUserThemeHandler(themeService service.ThemeService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := middleware.GetUserID(r)
		if !ok {
			response.SendErrorWithCode(w, http.StatusUnauthorized, "auth.authorization.unauthorized", "Unauthorized.")
			return
		}

		req, ok := middleware.GetBody[CreateThemeRequest](r)
		if !ok {
			response.SendErrorWithCode(w, http.StatusInternalServerError, "system.request.context_invalid", "Invalid request context.")
			return
		}

		theme, err := themeService.CreateUserTheme(r.Context(), userID, service.CreateUserThemeInput{
			Name:           req.Name,
			BasePresetSlug: req.BasePresetSlug,
			FontFamilyKey:  req.FontFamilyKey,
			LightTokens:    req.LightTokens,
			DarkTokens:     req.DarkTokens,
		})
		if err != nil {
			switch {
			case errors.Is(err, service.ErrThemeNameRequired), errors.Is(err, service.ErrThemeFontInvalid), errors.Is(err, service.ErrThemeTokensInvalid):
				response.SendErrorWithCode(w, http.StatusUnprocessableEntity, "validation.theme.invalid", "Invalid theme payload.")
			case errors.Is(err, service.ErrThemePresetNotFound):
				response.SendErrorWithCode(w, http.StatusNotFound, "resource.theme.preset_not_found", "Preset theme not found.")
			case errors.Is(err, service.ErrThemeNameExists):
				response.SendErrorWithCode(w, http.StatusConflict, "conflict.theme.name_exists", "Theme name already exists.")
			case errors.Is(err, service.ErrThemeLimitReached):
				response.SendErrorWithCode(w, http.StatusConflict, "conflict.theme.limit_reached", "You have reached the maximum number of custom themes.")
			default:
				response.SendErrorWithCode(w, http.StatusInternalServerError, "system.theme.create_failed", "Unable to create theme.")
			}
			return
		}

		response.SendSuccess(w, http.StatusCreated, theme)
	}
}

func GetUserThemeHandler(themeService service.ThemeService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := middleware.GetUserID(r)
		if !ok {
			response.SendErrorWithCode(w, http.StatusUnauthorized, "auth.authorization.unauthorized", "Unauthorized.")
			return
		}

		themeID, err := parseUUID(chi.URLParam(r, "theme_id"))
		if err != nil {
			response.SendErrorWithCode(w, http.StatusBadRequest, "validation.theme.id_invalid", "Invalid theme ID.")
			return
		}

		theme, err := themeService.GetUserTheme(r.Context(), userID, themeID)
		if err != nil {
			switch {
			case errors.Is(err, service.ErrThemeNotFound):
				response.SendErrorWithCode(w, http.StatusNotFound, "resource.theme.not_found", "Theme not found.")
			case errors.Is(err, service.ErrThemeForbidden):
				response.SendErrorWithCode(w, http.StatusForbidden, "auth.authorization.forbidden", "You do not have access to this theme.")
			default:
				response.SendErrorWithCode(w, http.StatusInternalServerError, "system.theme.fetch_failed", "Unable to load theme.")
			}
			return
		}

		response.SendSuccess(w, http.StatusOK, theme)
	}
}

func UpdateUserThemeHandler(themeService service.ThemeService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := middleware.GetUserID(r)
		if !ok {
			response.SendErrorWithCode(w, http.StatusUnauthorized, "auth.authorization.unauthorized", "Unauthorized.")
			return
		}

		themeID, err := parseUUID(chi.URLParam(r, "theme_id"))
		if err != nil {
			response.SendErrorWithCode(w, http.StatusBadRequest, "validation.theme.id_invalid", "Invalid theme ID.")
			return
		}

		req, ok := middleware.GetBody[UpdateThemeRequest](r)
		if !ok {
			response.SendErrorWithCode(w, http.StatusInternalServerError, "system.request.context_invalid", "Invalid request context.")
			return
		}

		theme, err := themeService.UpdateUserTheme(r.Context(), userID, themeID, service.UpdateUserThemeInput{
			Name:           req.Name,
			BasePresetSlug: req.BasePresetSlug,
			FontFamilyKey:  req.FontFamilyKey,
			LightTokens:    req.LightTokens,
			DarkTokens:     req.DarkTokens,
		})
		if err != nil {
			switch {
			case errors.Is(err, service.ErrThemeNotFound):
				response.SendErrorWithCode(w, http.StatusNotFound, "resource.theme.not_found", "Theme not found.")
			case errors.Is(err, service.ErrThemeForbidden):
				response.SendErrorWithCode(w, http.StatusForbidden, "auth.authorization.forbidden", "You do not have access to this theme.")
			case errors.Is(err, service.ErrThemeNameRequired), errors.Is(err, service.ErrThemeFontInvalid), errors.Is(err, service.ErrThemeTokensInvalid):
				response.SendErrorWithCode(w, http.StatusUnprocessableEntity, "validation.theme.invalid", "Invalid theme payload.")
			case errors.Is(err, service.ErrThemePresetNotFound):
				response.SendErrorWithCode(w, http.StatusNotFound, "resource.theme.preset_not_found", "Preset theme not found.")
			case errors.Is(err, service.ErrThemeNameExists):
				response.SendErrorWithCode(w, http.StatusConflict, "conflict.theme.name_exists", "Theme name already exists.")
			default:
				response.SendErrorWithCode(w, http.StatusInternalServerError, "system.theme.update_failed", "Unable to update theme.")
			}
			return
		}

		response.SendSuccess(w, http.StatusOK, theme)
	}
}

func DeleteUserThemeHandler(themeService service.ThemeService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := middleware.GetUserID(r)
		if !ok {
			response.SendErrorWithCode(w, http.StatusUnauthorized, "auth.authorization.unauthorized", "Unauthorized.")
			return
		}

		themeID, err := parseUUID(chi.URLParam(r, "theme_id"))
		if err != nil {
			response.SendErrorWithCode(w, http.StatusBadRequest, "validation.theme.id_invalid", "Invalid theme ID.")
			return
		}

		err = themeService.DeleteUserTheme(r.Context(), userID, themeID)
		if err != nil {
			switch {
			case errors.Is(err, service.ErrThemeNotFound):
				response.SendErrorWithCode(w, http.StatusNotFound, "resource.theme.not_found", "Theme not found.")
			case errors.Is(err, service.ErrThemeForbidden):
				response.SendErrorWithCode(w, http.StatusForbidden, "auth.authorization.forbidden", "You do not have access to this theme.")
			default:
				response.SendErrorWithCode(w, http.StatusInternalServerError, "system.theme.delete_failed", "Unable to delete theme.")
			}
			return
		}

		response.SendSuccess(w, http.StatusOK, map[string]string{"message": "theme deleted"})
	}
}
