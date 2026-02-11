import { Home, Layers3, UserCircle2, Users } from 'lucide-react'
import { Link, useRouterState } from '@tanstack/react-router'

const navItems = [
  { label: 'Home', to: '/app', icon: Home },
  { label: 'Groups', to: '/app/groups', icon: Layers3 },
  { label: 'Friends', to: '/app/friends', icon: Users },
  { label: 'Profile', to: '/app/profile', icon: UserCircle2 },
] as const

export function MobileShell({ children }: { children: React.ReactNode }) {
  const pathname = useRouterState({
    select: (s) => s.location.pathname,
  })

  function isNavItemActive(itemTo: string) {
    if (itemTo === '/app') {
      return pathname === '/app'
    }

    if (itemTo === '/app/profile' && pathname === '/app/themes') {
      return true
    }

    return pathname === itemTo || pathname.startsWith(`${itemTo}/`)
  }

  const activeNavLabel =
    navItems.find((item) => isNavItemActive(item.to))?.label || 'Workspace'

  return (
    <div className="app-shell relative mx-auto min-h-screen w-full max-w-md border-x border-border/60 bg-[radial-gradient(circle_at_top,rgba(148,163,184,0.12),transparent_40%)] pb-24">
      <header className="app-header sticky top-0 z-10 mb-4 border-b border-border/70 bg-background/92 pb-3 pt-4 backdrop-blur supports-[backdrop-filter]:bg-background/85">
        <p className="text-xs text-muted-foreground">Split+</p>
        <p className="text-sm font-medium text-foreground">{activeNavLabel}</p>
      </header>

      <div className="app-content min-h-screen pb-4">{children}</div>

      <nav className="app-nav fixed inset-x-0 bottom-3 z-20 mx-auto w-full max-w-[var(--app-page-max-width)] px-2">
        <div className="app-dock flex items-center justify-between rounded-[calc(var(--radius)+8px)] border border-border/70 p-1 shadow-sm">
          {navItems.map((item) => {
            const Icon = item.icon
            const isActive = isNavItemActive(item.to)

            return (
              <Link
                key={item.to}
                to={item.to}
                className={`flex flex-1 items-center justify-center gap-1 rounded-[calc(var(--radius)-2px)] px-2 py-2 text-xs transition ${
                  isActive
                    ? 'app-dock-item-active bg-muted/70 text-foreground'
                    : 'text-muted-foreground hover:text-foreground'
                }`}
              >
                <Icon className="size-4" />
                {item.label}
              </Link>
            )
          })}
        </div>
      </nav>
    </div>
  )
}
