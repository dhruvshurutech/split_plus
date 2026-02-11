import type {
  FontFamilyKey,
  PresetTheme,
  ThemeMode,
  ThemePreferences,
  ThemeTokens,
  UserTheme,
} from '@/lib/theme'
import { apiRequest } from '@/lib/api/client'

export type UpsertThemePreferencesInput = {
  mode: ThemeMode
  active_type: 'preset' | 'custom'
  active_preset_slug?: string
  active_user_theme_id?: string
}

export type CreateUserThemeInput = {
  name: string
  base_preset_slug: PresetTheme['slug']
  font_family_key: FontFamilyKey
  light_tokens: ThemeTokens
  dark_tokens: ThemeTokens
}

export type UpdateUserThemeInput = CreateUserThemeInput

export async function getThemePreferences(options?: {
  legacyTheme?: string
  legacyMode?: ThemeMode
}) {
  const params = new URLSearchParams()
  if (options?.legacyTheme) params.set('legacy_theme', options.legacyTheme)
  if (options?.legacyMode) params.set('legacy_mode', options.legacyMode)
  const suffix = params.toString() ? `?${params.toString()}` : ''

  return apiRequest<ThemePreferences>(`/users/me/theme/preferences${suffix}`, {
    method: 'GET',
    auth: true,
  })
}

export async function updateThemePreferences(input: UpsertThemePreferencesInput) {
  return apiRequest<ThemePreferences>('/users/me/theme/preferences', {
    method: 'PUT',
    auth: true,
    body: JSON.stringify(input),
  })
}

export async function listPresetThemes() {
  return apiRequest<Array<PresetTheme>>('/themes/presets', {
    method: 'GET',
    auth: true,
  })
}

export async function listUserThemes() {
  return apiRequest<Array<UserTheme>>('/users/me/themes', {
    method: 'GET',
    auth: true,
  })
}

export async function createUserTheme(input: CreateUserThemeInput) {
  return apiRequest<UserTheme>('/users/me/themes', {
    method: 'POST',
    auth: true,
    body: JSON.stringify(input),
  })
}

export async function getUserTheme(themeId: string) {
  return apiRequest<UserTheme>(`/users/me/themes/${themeId}`, {
    method: 'GET',
    auth: true,
  })
}

export async function updateUserTheme(themeId: string, input: UpdateUserThemeInput) {
  return apiRequest<UserTheme>(`/users/me/themes/${themeId}`, {
    method: 'PATCH',
    auth: true,
    body: JSON.stringify(input),
  })
}

export async function deleteUserTheme(themeId: string) {
  return apiRequest<{ message: string }>(`/users/me/themes/${themeId}`, {
    method: 'DELETE',
    auth: true,
  })
}
