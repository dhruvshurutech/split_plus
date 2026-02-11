import { HeadContent, Scripts, createRootRoute } from '@tanstack/react-router'
import { TanStackRouterDevtoolsPanel } from '@tanstack/react-router-devtools'
import { TanStackDevtools } from '@tanstack/react-devtools'

import appCss from '../styles.css?url'
import type { ThemeMode, ThemeName } from '@/lib/theme'
import { getLegacyThemePreferencesServerFn } from '@/lib/theme'
import { ThemeProvider } from '@/components/theme-provider'

const fallbackPreferences = {
  theme: 'soft-professional',
  mode: 'light',
} as const

export const Route = createRootRoute({
  loader: () => getLegacyThemePreferencesServerFn(),
  notFoundComponent: () => (
    <div className="p-4 text-sm text-muted-foreground">Page not found.</div>
  ),
  head: () => ({
    meta: [
      { charSet: 'utf-8' },
      { name: 'viewport', content: 'width=device-width, initial-scale=1' },
      { title: 'Split+ | Mobile Expense Splitting' },
    ],
    links: [{ rel: 'stylesheet', href: appCss }],
  }),

  shellComponent: RootDocument,
})

function RootDocument({ children }: { children: React.ReactNode }) {
  const loaderData = Route.useLoaderData() as unknown
  const preferences: { theme: ThemeName; mode: ThemeMode } =
    typeof loaderData === 'object' &&
    loaderData !== null &&
    'theme' in loaderData &&
    'mode' in loaderData
      ? (loaderData as { theme: ThemeName; mode: ThemeMode })
      : fallbackPreferences

  return (
    <html
      lang="en"
      className={`theme-${preferences.theme} ${preferences.mode === 'dark' ? 'dark' : ''}`}
      suppressHydrationWarning
    >
      <head>
        <HeadContent />
      </head>
      <body>
        <ThemeProvider preferences={preferences}>{children}</ThemeProvider>
        <TanStackDevtools
          config={{ position: 'bottom-right' }}
          plugins={[
            {
              name: 'Tanstack Router',
              render: <TanStackRouterDevtoolsPanel />,
            },
          ]}
        />
        <Scripts />
      </body>
    </html>
  )
}
