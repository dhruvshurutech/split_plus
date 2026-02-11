import { useTheme } from '@/components/theme-provider'

const modeOptions = [
  { label: 'Light', value: 'light' },
  { label: 'Dark', value: 'dark' },
] as const

export function ThemeSettings() {
  const { theme, mode, presets, setTheme, setMode } = useTheme()

  const presetOptions = presets.map((preset) => ({
    label: preset.label,
    value: preset.slug,
  }))

  return (
    <div className="space-y-3">
      <OptionRow
        label="Preset"
        options={presetOptions}
        active={theme}
        onSelect={(value) => {
          setTheme(value).catch(() => {
            // UI state is reverted in provider.
          })
        }}
      />
      <OptionRow
        label="Mode"
        options={modeOptions}
        active={mode}
        onSelect={(value) => {
          setMode(value).catch(() => {
            // UI state is reverted in provider.
          })
        }}
      />
    </div>
  )
}

function OptionRow<T extends string>({
  label,
  options,
  active,
  onSelect,
}: {
  label: string
  options: ReadonlyArray<{ label: string; value: T }>
  active: T
  onSelect: (value: T) => void
}) {
  return (
    <div className="space-y-1.5">
      <p className="text-sm text-muted-foreground">{label}</p>
      <div className="flex w-full flex-wrap items-center gap-1 rounded-[var(--radius)] border border-border/70 bg-card p-1">
        {options.map((option) => {
          const isActive = active === option.value
          return (
            <button
              key={option.value}
              type="button"
              onClick={() => onSelect(option.value)}
              className={`min-w-[30%] flex-1 rounded-[calc(var(--radius)-4px)] px-2 py-1.5 text-xs transition ${
                isActive
                  ? 'bg-foreground text-background'
                  : 'text-muted-foreground hover:text-foreground'
              }`}
            >
              {option.label}
            </button>
          )
        })}
      </div>
    </div>
  )
}

export function ThemeSwitcher() {
  return <ThemeSettings />
}

export function ModeToggle() {
  return <ThemeSettings />
}
