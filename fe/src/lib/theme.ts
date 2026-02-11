import { createServerFn } from '@tanstack/react-start'
import { getCookie } from '@tanstack/react-start/server'

export type ThemeName =
  | 'minimal'
  | 'newspaper'
  | 'soft-professional'
  | 'atelier'
  | 'mono-slate'

export type ThemeMode = 'light' | 'dark'

export type FontFamilyKey =
  | 'avenir'
  | 'newspaper-serif'
  | 'geist-mono'
  | 'charter-serif'

export type ThemeTokenKey =
  | 'background'
  | 'foreground'
  | 'card'
  | 'card-foreground'
  | 'primary'
  | 'primary-foreground'
  | 'secondary'
  | 'secondary-foreground'
  | 'muted'
  | 'muted-foreground'
  | 'accent'
  | 'accent-foreground'
  | 'border'
  | 'input'
  | 'ring'
  | 'radius'
  | 'dock-background'
  | 'dock-active'

export type ThemeTokens = Partial<Record<ThemeTokenKey, string>>

export type ThemeRef =
  | { type: 'preset'; preset_slug: ThemeName }
  | { type: 'custom'; user_theme_id: string }

export type ThemeResolved = {
  base_preset_slug: ThemeName
  font_family_key: FontFamilyKey
  tokens: ThemeTokens
}

export type ThemePreferences = {
  mode: ThemeMode
  active: ThemeRef
  resolved_theme: ThemeResolved
}

export type PresetTheme = {
  id: string
  slug: ThemeName
  label: string
  family: 'palette' | 'expressive'
  is_active: boolean
}

export type UserTheme = {
  id: string
  name: string
  base_preset_slug: ThemeName
  font_family_key: FontFamilyKey
  light_tokens: ThemeTokens
  dark_tokens: ThemeTokens
  created_at: string
  updated_at: string
}

export type LegacyThemePreferences = {
  theme: ThemeName
  mode: ThemeMode
}

export const THEME_STORAGE_KEY = '_preferred-theme'
export const MODE_STORAGE_KEY = '_preferred-mode'

export const defaultTheme: ThemeName = 'soft-professional'
export const defaultMode: ThemeMode = 'light'

export function resolveLegacyTheme(theme: string | undefined): ThemeName {
  if (
    theme === 'minimal' ||
    theme === 'newspaper' ||
    theme === 'soft-professional' ||
    theme === 'atelier' ||
    theme === 'mono-slate'
  ) {
    return theme
  }

  if (theme === 'light' || theme === 'text-minimal') return 'minimal'
  if (theme === 'dark') return 'soft-professional'

  return defaultTheme
}

export function resolveLegacyMode(mode: string | undefined): ThemeMode {
  if (mode === 'light' || mode === 'dark') return mode
  return defaultMode
}

function readCookie(cookieString: string, key: string): string | undefined {
  const parts = cookieString.split(';')
  for (const part of parts) {
    const [rawKey, rawVal] = part.trim().split('=')
    if (rawKey === key) return decodeURIComponent(rawVal || '')
  }
  return undefined
}

export function getLegacyThemePreferencesClient(): LegacyThemePreferences {
  if (typeof document === 'undefined') {
    return { theme: defaultTheme, mode: defaultMode }
  }

  return {
    theme: resolveLegacyTheme(readCookie(document.cookie, THEME_STORAGE_KEY)),
    mode: resolveLegacyMode(readCookie(document.cookie, MODE_STORAGE_KEY)),
  }
}

export const getLegacyThemePreferencesServerFn = createServerFn().handler(
  () => ({
    theme: resolveLegacyTheme(getCookie(THEME_STORAGE_KEY)),
    mode: resolveLegacyMode(getCookie(MODE_STORAGE_KEY)),
  }),
)

export const FONT_LABELS: Record<FontFamilyKey, string> = {
  avenir: 'Avenir Sans',
  'newspaper-serif': 'Newspaper Serif',
  'geist-mono': 'Geist Mono',
  'charter-serif': 'Charter Serif',
}

export function defaultFontForPreset(theme: ThemeName): FontFamilyKey {
  switch (theme) {
    case 'newspaper':
      return 'newspaper-serif'
    case 'mono-slate':
      return 'geist-mono'
    default:
      return 'avenir'
  }
}
