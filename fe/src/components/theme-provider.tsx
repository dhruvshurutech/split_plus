import { createContext, use, useEffect, useMemo, useState } from 'react'
import type { PropsWithChildren } from 'react'

import type {
  FontFamilyKey,
  PresetTheme,
  ThemeMode,
  ThemeName,
  ThemePreferences,
  ThemeTokens,
  UserTheme,
} from '@/lib/theme'
import {
  defaultFontForPreset,
  getLegacyThemePreferencesClient,
} from '@/lib/theme'
import type {
  CreateUserThemeInput,
  UpdateUserThemeInput,
} from '@/lib/api/themes'
import {
  createUserTheme,
  deleteUserTheme,
  listPresetThemes,
  listUserThemes,
  updateThemePreferences,
  updateUserTheme,
} from '@/lib/api/themes'
import { getThemePreferences } from '@/lib/api/themes'
import { hasSession } from '@/lib/session'

type ThemeContextVal = {
  mode: ThemeMode
  theme: ThemeName
  presets: Array<PresetTheme>
  userThemes: Array<UserTheme>
  activeTheme: ThemePreferences['active']
  resolvedTheme: ThemePreferences['resolved_theme']
  isLoading: boolean
  setTheme: (preset: ThemeName) => Promise<void>
  setMode: (mode: ThemeMode) => Promise<void>
  setActiveCustomTheme: (themeID: string) => Promise<void>
  refreshThemes: () => Promise<void>
  createCustomTheme: (input: CreateUserThemeInput) => Promise<UserTheme>
  updateCustomTheme: (id: string, input: UpdateUserThemeInput) => Promise<UserTheme>
  deleteCustomTheme: (id: string) => Promise<void>
}

type Props = PropsWithChildren<{ preferences: { theme: ThemeName; mode: ThemeMode } }>

const ThemeContext = createContext<ThemeContextVal | null>(null)

const TOKEN_TO_CSS_VARIABLE: Record<string, string> = {
  background: '--background',
  foreground: '--foreground',
  card: '--card',
  'card-foreground': '--card-foreground',
  primary: '--primary',
  'primary-foreground': '--primary-foreground',
  secondary: '--secondary',
  'secondary-foreground': '--secondary-foreground',
  muted: '--muted',
  'muted-foreground': '--muted-foreground',
  accent: '--accent',
  'accent-foreground': '--accent-foreground',
  border: '--border',
  input: '--input',
  ring: '--ring',
  radius: '--radius',
  'dock-background': '--dock-background',
  'dock-active': '--dock-active',
}

const FONT_STACKS: Record<FontFamilyKey, string> = {
  avenir: "'Avenir Next', 'Segoe UI', 'Noto Sans', sans-serif",
  'newspaper-serif': "'Charter', 'Iowan Old Style', 'Palatino Linotype', serif",
  'geist-mono': "'Geist Mono Variable', 'SFMono-Regular', Consolas, monospace",
  'charter-serif': "'Charter', 'Iowan Old Style', 'Palatino Linotype', serif",
}

export function ThemeProvider({ children, preferences }: Props) {
  const [mode, setModeState] = useState<ThemeMode>(preferences.mode)
  const [activeTheme, setActiveTheme] = useState<ThemePreferences['active']>({
    type: 'preset',
    preset_slug: preferences.theme,
  })
  const [resolvedTheme, setResolvedTheme] = useState<ThemePreferences['resolved_theme']>({
    base_preset_slug: preferences.theme,
    font_family_key: defaultFontForPreset(preferences.theme),
    tokens: {},
  })
  const [presets, setPresets] = useState<Array<PresetTheme>>([])
  const [userThemes, setUserThemes] = useState<Array<UserTheme>>([])
  const [isLoading, setIsLoading] = useState(false)

  const theme = resolvedTheme.base_preset_slug

  async function refreshThemes() {
    if (!hasSession()) return

    setIsLoading(true)
    try {
      const legacy = getLegacyThemePreferencesClient()
      const [nextPreferences, nextPresets, nextUserThemes] = await Promise.all([
        getThemePreferences({
          legacyTheme: legacy.theme,
          legacyMode: legacy.mode,
        }),
        listPresetThemes(),
        listUserThemes(),
      ])
      setModeState(nextPreferences.mode)
      setActiveTheme(nextPreferences.active)
      setResolvedTheme(nextPreferences.resolved_theme)
      setPresets(nextPresets)
      setUserThemes(nextUserThemes)
    } finally {
      setIsLoading(false)
    }
  }

  useEffect(() => {
    refreshThemes().catch(() => {
      // Keep legacy bootstrap values when API call fails.
    })
  }, [])

  useEffect(() => {
    if (typeof document === 'undefined') return
    const html = document.documentElement

    html.classList.remove(
      'theme-minimal',
      'theme-newspaper',
      'theme-soft-professional',
      'theme-atelier',
      'theme-mono-slate',
      'theme-custom',
      'dark',
    )

    html.classList.add(`theme-${resolvedTheme.base_preset_slug}`)
    if (mode === 'dark') html.classList.add('dark')

    for (const cssVar of Object.values(TOKEN_TO_CSS_VARIABLE)) {
      html.style.removeProperty(cssVar)
    }
    html.style.removeProperty('--custom-font-family')

    if (activeTheme.type === 'custom') {
      html.classList.add('theme-custom')
      for (const [token, tokenValue] of Object.entries(resolvedTheme.tokens)) {
        const cssVariable = TOKEN_TO_CSS_VARIABLE[token]
        if (cssVariable && tokenValue) {
          html.style.setProperty(cssVariable, tokenValue)
        }
      }
      html.style.setProperty(
        '--custom-font-family',
        FONT_STACKS[resolvedTheme.font_family_key],
      )
    }
  }, [activeTheme, resolvedTheme, mode])

  async function applyPreferenceUpdate(input: {
    mode: ThemeMode
    active_type: 'preset' | 'custom'
    active_preset_slug?: ThemeName
    active_user_theme_id?: string
  }) {
    const result = await updateThemePreferences(input)
    setModeState(result.mode)
    setActiveTheme(result.active)
    setResolvedTheme(result.resolved_theme)
  }

  async function setTheme(nextTheme: ThemeName) {
    const prevTheme = activeTheme
    const prevResolved = resolvedTheme

    setActiveTheme({ type: 'preset', preset_slug: nextTheme })
    setResolvedTheme({
      base_preset_slug: nextTheme,
      font_family_key: defaultFontForPreset(nextTheme),
      tokens: {},
    })

    try {
      await applyPreferenceUpdate({
        mode,
        active_type: 'preset',
        active_preset_slug: nextTheme,
      })
    } catch (error) {
      setActiveTheme(prevTheme)
      setResolvedTheme(prevResolved)
      throw error
    }
  }

  async function setMode(nextMode: ThemeMode) {
    const prevMode = mode
    setModeState(nextMode)

    try {
      if (activeTheme.type === 'preset') {
        await applyPreferenceUpdate({
          mode: nextMode,
          active_type: 'preset',
          active_preset_slug: activeTheme.preset_slug,
        })
      } else {
        await applyPreferenceUpdate({
          mode: nextMode,
          active_type: 'custom',
          active_user_theme_id: activeTheme.user_theme_id,
        })
      }
    } catch (error) {
      setModeState(prevMode)
      throw error
    }
  }

  async function setActiveCustomTheme(themeID: string) {
    const prevTheme = activeTheme
    const prevResolved = resolvedTheme

    const selected = userThemes.find((entry) => entry.id === themeID)
    if (selected) {
      setActiveTheme({ type: 'custom', user_theme_id: themeID })
      setResolvedTheme({
        base_preset_slug: selected.base_preset_slug,
        font_family_key: selected.font_family_key,
        tokens: mode === 'dark' ? selected.dark_tokens : selected.light_tokens,
      })
    }

    try {
      await applyPreferenceUpdate({
        mode,
        active_type: 'custom',
        active_user_theme_id: themeID,
      })
    } catch (error) {
      setActiveTheme(prevTheme)
      setResolvedTheme(prevResolved)
      throw error
    }
  }

  async function createCustomThemeWrapper(input: CreateUserThemeInput) {
    const created = await createUserTheme(input)
    await refreshThemes()
    return created
  }

  async function updateCustomThemeWrapper(
    id: string,
    input: UpdateUserThemeInput,
  ) {
    const updated = await updateUserTheme(id, input)
    await refreshThemes()
    return updated
  }

  async function deleteCustomThemeWrapper(id: string) {
    await deleteUserTheme(id)
    await refreshThemes()
  }

  const contextValue = useMemo<ThemeContextVal>(
    () => ({
      mode,
      theme,
      presets,
      userThemes,
      activeTheme,
      resolvedTheme,
      isLoading,
      setTheme,
      setMode,
      setActiveCustomTheme,
      refreshThemes,
      createCustomTheme: createCustomThemeWrapper,
      updateCustomTheme: updateCustomThemeWrapper,
      deleteCustomTheme: deleteCustomThemeWrapper,
    }),
    [
      mode,
      theme,
      presets,
      userThemes,
      activeTheme,
      resolvedTheme,
      isLoading,
    ],
  )

  return <ThemeContext value={contextValue}>{children}</ThemeContext>
}

export function useTheme() {
  const val = use(ThemeContext)
  if (!val) throw new Error('useTheme called outside of ThemeProvider!')
  return val
}
