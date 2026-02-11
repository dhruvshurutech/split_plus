import { Outlet, createFileRoute, redirect } from '@tanstack/react-router'

import { MobileShell } from '@/components/mobile-shell'
import { ApiError } from '@/lib/api/client'
import { getMe } from '@/lib/api/users'
import { clearSessionTokens, hasSession } from '@/lib/session'

export const Route = createFileRoute('/app')({
  beforeLoad: async () => {
    if (typeof window === 'undefined') return

    if (!hasSession()) {
      throw redirect({ to: '/login' })
    }

    try {
      await getMe()
    } catch (error) {
      if (error instanceof ApiError) {
        if (
          error.statusCode === 401 ||
          error.code === 'resource.user.not_found'
        ) {
          clearSessionTokens()
          throw redirect({ to: '/login' })
        }
      }

      throw error
    }
  },
  component: AppLayout,
})

function AppLayout() {
  return (
    <MobileShell>
      <Outlet />
    </MobileShell>
  )
}
