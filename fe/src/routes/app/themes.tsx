import { useMemo, useState } from 'react'
import { Link, createFileRoute } from '@tanstack/react-router'
import { ArrowLeft, ChevronDown } from 'lucide-react'

import { useTheme } from '@/components/theme-provider'
import { Button, buttonVariants } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { cn } from '@/lib/utils'
import { FONT_LABELS } from '@/lib/theme'
import type { FontFamilyKey, ThemeMode, ThemeTokens } from '@/lib/theme'

const COLOR_OPTIONS_LIGHT = {
  background: [
    { label: 'Use preset', value: '' },
    { label: 'Paper', value: 'oklch(0.99 0.004 95)' },
    { label: 'Cool Mist', value: 'oklch(0.97 0.006 240)' },
    { label: 'Warm Sand', value: 'oklch(0.97 0.01 30)' },
  ],
  foreground: [
    { label: 'Use preset', value: '' },
    { label: 'Soft Ink', value: 'oklch(0.24 0.012 255)' },
    { label: 'Slate', value: 'oklch(0.28 0.014 245)' },
    { label: 'Charcoal', value: 'oklch(0.2 0.01 70)' },
  ],
  primary: [
    { label: 'Use preset', value: '' },
    { label: 'Ocean Blue', value: 'oklch(0.52 0.08 250)' },
    { label: 'Coral', value: 'oklch(0.64 0.12 30)' },
    { label: 'Forest', value: 'oklch(0.56 0.09 145)' },
    { label: 'Graphite', value: 'oklch(0.45 0.01 255)' },
  ],
  border: [
    { label: 'Use preset', value: '' },
    { label: 'Soft Border', value: 'oklch(0.88 0.008 250)' },
    { label: 'Medium Border', value: 'oklch(0.8 0.01 250)' },
    { label: 'Strong Border', value: 'oklch(0.7 0.012 250)' },
  ],
  'dock-background': [
    { label: 'Use preset', value: '' },
    {
      label: 'Glass',
      value: 'color-mix(in oklab, var(--card) 95%, white 5%)',
    },
    {
      label: 'Muted',
      value: 'color-mix(in oklab, var(--muted) 86%, white 14%)',
    },
  ],
} as const

const COLOR_OPTIONS_DARK = {
  background: [
    { label: 'Use preset', value: '' },
    { label: 'Night Blue', value: 'oklch(0.2 0.01 255)' },
    { label: 'Graphite', value: 'oklch(0.18 0.008 255)' },
    { label: 'Deep Warm', value: 'oklch(0.22 0.02 28)' },
  ],
  foreground: [
    { label: 'Use preset', value: '' },
    { label: 'Soft White', value: 'oklch(0.92 0.004 95)' },
    { label: 'Cool White', value: 'oklch(0.9 0.004 255)' },
  ],
  primary: [
    { label: 'Use preset', value: '' },
    { label: 'Sky', value: 'oklch(0.76 0.04 245)' },
    { label: 'Peach', value: 'oklch(0.78 0.07 36)' },
    { label: 'Mint', value: 'oklch(0.78 0.06 155)' },
    { label: 'Silver', value: 'oklch(0.82 0.01 255)' },
  ],
  border: [
    { label: 'Use preset', value: '' },
    { label: 'Soft Border', value: 'oklch(0.36 0.01 255)' },
    { label: 'Medium Border', value: 'oklch(0.42 0.012 255)' },
    { label: 'Strong Border', value: 'oklch(0.5 0.014 255)' },
  ],
  'dock-background': [
    { label: 'Use preset', value: '' },
    {
      label: 'Glass Dark',
      value: 'color-mix(in oklab, var(--card) 88%, black 12%)',
    },
    {
      label: 'Muted Dark',
      value: 'color-mix(in oklab, var(--muted) 78%, black 22%)',
    },
  ],
} as const

const EDITABLE_COLOR_KEYS = [
  'background',
  'foreground',
  'primary',
  'border',
  'dock-background',
] as const

type ThemeFormState = {
  id: string | null
  name: string
  basePresetSlug: string
  fontFamilyKey: FontFamilyKey
  lightTokens: ThemeTokens
  darkTokens: ThemeTokens
}

const FONT_STACKS: Record<FontFamilyKey, string> = {
  avenir: "'Avenir Next', 'Segoe UI', 'Noto Sans', sans-serif",
  'newspaper-serif': "'Charter', 'Iowan Old Style', 'Palatino Linotype', serif",
  'geist-mono': "'Geist Mono Variable', 'SFMono-Regular', Consolas, monospace",
  'charter-serif': "'Charter', 'Iowan Old Style', 'Palatino Linotype', serif",
}

function emptyForm(basePresetSlug: string): ThemeFormState {
  return {
    id: null,
    name: '',
    basePresetSlug,
    fontFamilyKey: 'avenir',
    lightTokens: {},
    darkTokens: {},
  }
}

export const Route = createFileRoute('/app/themes')({
  component: ThemeGalleryPage,
})

function ThemeGalleryPage() {
  const {
    mode,
    theme,
    presets,
    userThemes,
    activeTheme,
    isLoading,
    setTheme,
    setMode,
    setActiveCustomTheme,
    createCustomTheme,
    updateCustomTheme,
    deleteCustomTheme,
  } = useTheme()

  const defaultPreset = presets[0]?.slug ?? 'soft-professional'
  const [form, setForm] = useState<ThemeFormState>(emptyForm(defaultPreset))
  const [isEditorOpen, setIsEditorOpen] = useState(false)
  const [previewMode, setPreviewMode] = useState<ThemeMode>('light')
  const [isSaving, setIsSaving] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const groupedPresets = useMemo(
    () => ({
      palette: presets.filter((entry) => entry.family === 'palette'),
      expressive: presets.filter((entry) => entry.family === 'expressive'),
    }),
    [presets],
  )

  function patchForm(next: Partial<ThemeFormState>) {
    setForm((prev) => ({ ...prev, ...next }))
  }

  function patchModeToken(
    currentMode: ThemeMode,
    key: keyof ThemeTokens,
    value: string,
  ) {
    if (currentMode === 'light') {
      patchForm({
        lightTokens: {
          ...form.lightTokens,
          [key]: value,
        },
      })
      return
    }

    patchForm({
      darkTokens: {
        ...form.darkTokens,
        [key]: value,
      },
    })
  }

  function patchRadius(radius: number) {
    const value = `${Math.round(radius)}%`
    patchForm({
      lightTokens: { ...form.lightTokens, radius: value },
      darkTokens: { ...form.darkTokens, radius: value },
    })
  }

  async function submitForm() {
    setIsSaving(true)
    setError(null)
    const payload = {
      name: form.name.trim(),
      base_preset_slug: form.basePresetSlug as
        | 'minimal'
        | 'soft-professional'
        | 'atelier'
        | 'newspaper'
        | 'mono-slate',
      font_family_key: form.fontFamilyKey,
      light_tokens: compactTokens(form.lightTokens),
      dark_tokens: compactTokens(form.darkTokens),
    }

    try {
      if (form.id) {
        await updateCustomTheme(form.id, payload)
      } else {
        await createCustomTheme(payload)
      }
      setForm(emptyForm(defaultPreset))
      setIsEditorOpen(false)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Unable to save theme.')
    } finally {
      setIsSaving(false)
    }
  }

  function startEdit(themeToEdit: (typeof userThemes)[number]) {
    setForm({
      id: themeToEdit.id,
      name: themeToEdit.name,
      basePresetSlug: themeToEdit.base_preset_slug,
      fontFamilyKey: themeToEdit.font_family_key,
      lightTokens: { ...themeToEdit.light_tokens },
      darkTokens: { ...themeToEdit.dark_tokens },
    })
    setIsEditorOpen(true)
  }

  const selectedModeTokens =
    previewMode === 'light' ? form.lightTokens : form.darkTokens

  const currentRadius = parseRadiusValue(
    form.lightTokens.radius || form.darkTokens.radius,
  )

  return (
    <div className="space-y-4">
      <header>
        <div className="mb-2 flex items-center gap-2">
          <Link
            to="/app/profile"
            aria-label="Back to profile"
            className={cn(
              buttonVariants({ variant: 'outline', size: 'icon' }),
              'size-8 rounded-full',
            )}
          >
            <ArrowLeft className="size-4" />
          </Link>
          <h1 className="text-xl font-semibold">Themes</h1>
        </div>
        <p className="text-sm text-muted-foreground">
          Choose a preset or create your own private theme.
        </p>
      </header>

      <Card className="border-border/70 bg-card">
        <CardHeader>
          <CardTitle className="text-base">Mode</CardTitle>
        </CardHeader>
        <CardContent className="grid grid-cols-2 gap-2">
          {(['light', 'dark'] as const).map((entry) => (
            <Button
              key={entry}
              variant={mode === entry ? 'default' : 'outline'}
              onClick={() => setMode(entry)}
            >
              {entry === 'light' ? 'Light' : 'Dark'}
            </Button>
          ))}
        </CardContent>
      </Card>

      {(['palette', 'expressive'] as const).map((family) => (
        <Card key={family} className="border-border/70 bg-card">
          <CardHeader>
            <CardTitle className="text-base">
              {family === 'palette' ? 'Palette Presets' : 'Expressive Presets'}
            </CardTitle>
          </CardHeader>
          <CardContent className="grid gap-2">
            {groupedPresets[family].map((preset) => {
              const isActive =
                activeTheme.type === 'preset' &&
                activeTheme.preset_slug === preset.slug

              return (
                <button
                  key={preset.id}
                  type="button"
                  onClick={() => setTheme(preset.slug)}
                  className={cn(
                    'flex items-center justify-between rounded-[var(--radius)] border border-border/70 px-3 py-2 text-left text-sm',
                    isActive ? 'bg-muted/50 ring-1 ring-foreground/20' : '',
                  )}
                >
                  <span>{preset.label}</span>
                  {isActive ? (
                    <span className="text-xs text-muted-foreground">Active</span>
                  ) : null}
                </button>
              )
            })}
          </CardContent>
        </Card>
      ))}

      <Card className="border-border/70 bg-card">
        <CardHeader>
          <CardTitle className="text-base">My Themes</CardTitle>
        </CardHeader>
        <CardContent className="space-y-2">
          {userThemes.map((item) => {
            const isActive =
              activeTheme.type === 'custom' &&
              activeTheme.user_theme_id === item.id

            return (
              <div
                key={item.id}
                className="rounded-[var(--radius)] border border-border/70 p-3"
              >
                <div className="flex items-center justify-between">
                  <div>
                    <p className="text-sm font-medium">{item.name}</p>
                    <p className="text-xs text-muted-foreground">
                      Base: {item.base_preset_slug} · {FONT_LABELS[item.font_family_key]}
                    </p>
                  </div>
                  {isActive ? (
                    <span className="text-xs text-muted-foreground">Active</span>
                  ) : null}
                </div>
                <div className="mt-2 flex gap-2">
                  <Button
                    size="sm"
                    variant={isActive ? 'secondary' : 'default'}
                    onClick={() => setActiveCustomTheme(item.id)}
                  >
                    Apply
                  </Button>
                  <Button size="sm" variant="outline" onClick={() => startEdit(item)}>
                    Edit
                  </Button>
                  <Button
                    size="sm"
                    variant="outline"
                    onClick={() => {
                      deleteCustomTheme(item.id).catch((err) => {
                        setError(
                          err instanceof Error
                            ? err.message
                            : 'Unable to delete theme.',
                        )
                      })
                    }}
                  >
                    Delete
                  </Button>
                </div>
              </div>
            )
          })}
          {userThemes.length === 0 && !isLoading ? (
            <p className="text-xs text-muted-foreground">No custom themes yet.</p>
          ) : null}
        </CardContent>
      </Card>

      <Card className="border-border/70 bg-card">
        <CardHeader>
          <button
            type="button"
            onClick={() => setIsEditorOpen((prev) => !prev)}
            className="flex w-full items-center justify-between text-left"
          >
            <CardTitle className="text-base">
              {form.id ? 'Edit Theme' : 'Create Theme'}
            </CardTitle>
            <ChevronDown
              className={cn(
                'size-4 text-muted-foreground transition-transform',
                isEditorOpen ? 'rotate-180' : '',
              )}
            />
          </button>
        </CardHeader>

        {isEditorOpen ? (
          <CardContent className="space-y-4">
            <div className="space-y-1">
              <p className="text-xs text-muted-foreground">Theme name</p>
              <Input
                value={form.name}
                onChange={(event) => patchForm({ name: event.target.value })}
                placeholder="My theme"
              />
            </div>

            <div className="grid grid-cols-2 gap-2">
              <SelectField
                label="Base preset"
                value={form.basePresetSlug}
                onChange={(value) => patchForm({ basePresetSlug: value })}
                options={presets.map((preset) => ({
                  label: preset.label,
                  value: preset.slug,
                }))}
              />

              <SelectField
                label="Font"
                value={form.fontFamilyKey}
                onChange={(value) =>
                  patchForm({ fontFamilyKey: value as FontFamilyKey })
                }
                options={Object.entries(FONT_LABELS).map(([key, label]) => ({
                  label,
                  value: key,
                }))}
              />
            </div>

            <div className="space-y-2 rounded-[var(--radius)] border border-border/70 p-3">
              <div className="flex items-center justify-between">
                <p className="text-xs font-medium text-muted-foreground">Appearance</p>
                <div className="flex gap-1">
                  {(['light', 'dark'] as const).map((entry) => (
                    <Button
                      key={entry}
                      size="sm"
                      variant={previewMode === entry ? 'default' : 'outline'}
                      className="h-7 px-2 text-xs"
                      onClick={() => setPreviewMode(entry)}
                    >
                      {entry === 'light' ? 'Light' : 'Dark'}
                    </Button>
                  ))}
                </div>
              </div>

              <div className="grid grid-cols-2 gap-2">
                {EDITABLE_COLOR_KEYS.map((key) => (
                  <SelectField
                    key={key}
                    label={humanizeTokenName(key)}
                    value={selectedModeTokens[key] ?? ''}
                    onChange={(value) => patchModeToken(previewMode, key, value)}
                    options={
                      previewMode === 'light'
                        ? COLOR_OPTIONS_LIGHT[key]
                        : COLOR_OPTIONS_DARK[key]
                    }
                  />
                ))}
              </div>

              <div className="space-y-1">
                <div className="flex items-center justify-between">
                  <p className="text-xs text-muted-foreground">Corner roundness</p>
                  <span className="text-xs text-muted-foreground">
                    {Math.round(currentRadius)}%
                  </span>
                </div>
                <input
                  type="range"
                  min="0"
                  max="100"
                  step="5"
                  value={currentRadius}
                  onChange={(event) => patchRadius(Number(event.target.value))}
                  className="w-full"
                />
              </div>
            </div>

            <ThemeLivePreview
              basePresetSlug={form.basePresetSlug}
              fontFamilyKey={form.fontFamilyKey}
              mode={previewMode}
              lightTokens={form.lightTokens}
              darkTokens={form.darkTokens}
            />

            {error ? <p className="text-xs text-destructive">{error}</p> : null}

            <div className="flex gap-2">
              <Button disabled={isSaving} onClick={submitForm}>
                {form.id ? 'Save Changes' : 'Create Theme'}
              </Button>
              {form.id ? (
                <Button
                  variant="outline"
                  onClick={() => {
                    setForm(emptyForm(defaultPreset))
                    setIsEditorOpen(false)
                  }}
                >
                  Cancel
                </Button>
              ) : null}
            </div>
          </CardContent>
        ) : null}
      </Card>
    </div>
  )
}

function SelectField({
  label,
  value,
  options,
  onChange,
}: {
  label: string
  value: string
  options: ReadonlyArray<{ label: string; value: string }>
  onChange: (value: string) => void
}) {
  return (
    <label className="space-y-1">
      <span className="text-xs text-muted-foreground">{label}</span>
      <select
        className="h-[var(--control-height)] w-full rounded-[var(--radius)] border border-input bg-transparent px-3 text-sm"
        value={value}
        onChange={(event) => onChange(event.target.value)}
      >
        {options.map((entry) => (
          <option key={`${label}-${entry.value || 'preset'}`} value={entry.value}>
            {entry.label}
          </option>
        ))}
      </select>
    </label>
  )
}

function ThemeLivePreview({
  basePresetSlug,
  fontFamilyKey,
  mode,
  lightTokens,
  darkTokens,
}: {
  basePresetSlug: string
  fontFamilyKey: FontFamilyKey
  mode: ThemeMode
  lightTokens: ThemeTokens
  darkTokens: ThemeTokens
}) {
  const tokens = mode === 'dark' ? darkTokens : lightTokens
  const style = tokensToStyle(tokens, fontFamilyKey)

  return (
    <div className="space-y-2">
      <p className="text-xs font-medium text-muted-foreground">Live preview</p>
      <div
        className={cn(
          'overflow-hidden rounded-[var(--radius)] border border-border/70',
          `theme-${basePresetSlug}`,
          mode === 'dark' ? 'dark' : '',
        )}
        style={style}
      >
        <div className="bg-background px-3 py-2 text-foreground">
          <div className="flex items-center justify-between">
            <p className="text-[11px] text-muted-foreground">Split+</p>
            <p className="text-[11px] text-muted-foreground">Today</p>
          </div>

          <h3 className="mt-1 text-sm font-semibold">Weekend Trip</h3>
          <p className="text-[11px] text-muted-foreground">3 unsettled expenses</p>

          <div className="mt-2 grid grid-cols-2 gap-2">
            <div className="rounded-[var(--radius)] border border-border bg-card p-2">
              <p className="text-[10px] text-muted-foreground">You owe</p>
              <p className="text-xs font-semibold">Rs 240</p>
            </div>
            <div className="rounded-[var(--radius)] border border-border bg-card p-2">
              <p className="text-[10px] text-muted-foreground">You get</p>
              <p className="text-xs font-semibold">Rs 620</p>
            </div>
          </div>

          <div className="mt-2 flex items-center gap-2">
            <button
              type="button"
              className="h-7 rounded-[var(--radius)] bg-primary px-3 text-xs text-primary-foreground"
            >
              Settle Up
            </button>
            <button
              type="button"
              className="h-7 rounded-[var(--radius)] border border-border px-3 text-xs"
            >
              Add Expense
            </button>
          </div>
        </div>
      </div>
    </div>
  )
}

function tokensToStyle(tokens: ThemeTokens, font: FontFamilyKey) {
  const style: Record<string, string> = {
    '--custom-font-family': FONT_STACKS[font],
    fontFamily: FONT_STACKS[font],
  }

  for (const [key, value] of Object.entries(tokens)) {
    if (!value) continue
    style[`--${key}`] = value
  }

  return style as React.CSSProperties
}

function parseRadiusValue(radius: string | undefined): number {
  if (!radius) return 35
  const trimmed = radius.trim()

  const percentMatch = /^([0-9]+(?:\.[0-9]+)?)%$/.exec(trimmed)
  if (percentMatch) {
    return Math.min(100, Math.max(0, Number(percentMatch[1])))
  }

  const remMatch = /^([0-9]+(?:\.[0-9]+)?)rem$/.exec(trimmed)
  if (remMatch) {
    const mapped = Number(remMatch[1]) * 40
    return Math.min(100, Math.max(0, mapped))
  }

  return 35
}

function humanizeTokenName(key: string): string {
  switch (key) {
    case 'background':
      return 'Background'
    case 'foreground':
      return 'Text color'
    case 'primary':
      return 'Primary color'
    case 'border':
      return 'Border style'
    case 'dock-background':
      return 'Bottom bar style'
    default:
      return key
  }
}

function compactTokens(tokens: ThemeTokens): ThemeTokens {
  const out: ThemeTokens = {}
  for (const [key, value] of Object.entries(tokens)) {
    const trimmed = value.trim()
    if (trimmed) {
      out[key as keyof ThemeTokens] = trimmed
    }
  }
  return out
}
