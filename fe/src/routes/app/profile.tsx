import { useEffect, useState } from 'react'
import { Link, createFileRoute, useNavigate } from '@tanstack/react-router'

import type { MeResponse } from '@/lib/api/users'
import { ApiError } from '@/lib/api/client'
import { Button, buttonVariants } from '@/components/ui/button'
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
import { logout, logoutAll } from '@/lib/api/auth'
import { getMe } from '@/lib/api/users'
import { clearSessionTokens } from '@/lib/session'
import { cn } from '@/lib/utils'

export const Route = createFileRoute('/app/profile')({ component: ProfilePage })

function ProfilePage() {
  const navigate = useNavigate()
  const [me, setMe] = useState<MeResponse | null>(null)
  const [error, setError] = useState<string | null>(null)
  const [isLoading, setIsLoading] = useState(true)
  const [isSubmitting, setIsSubmitting] = useState(false)

  useEffect(() => {
    let mounted = true

    async function loadMe() {
      try {
        const response = await getMe()
        if (mounted) setMe(response)
      } catch (err) {
        if (err instanceof ApiError && err.statusCode === 401) {
          clearSessionTokens()
          navigate({ to: '/login' })
          return
        }

        if (mounted)
          setError(
            err instanceof Error ? err.message : 'Unable to load profile',
          )
      } finally {
        if (mounted) setIsLoading(false)
      }
    }

    loadMe()
    return () => {
      mounted = false
    }
  }, [])

  async function handleLogout() {
    setIsSubmitting(true)
    try {
      await logout()
      navigate({ to: '/login' })
    } finally {
      setIsSubmitting(false)
    }
  }

  async function handleLogoutAll() {
    setIsSubmitting(true)
    try {
      await logoutAll()
      navigate({ to: '/login' })
    } finally {
      setIsSubmitting(false)
    }
  }

  return (
    <div className="space-y-4">
      <header>
        <p className="text-sm text-muted-foreground">Profile</p>
        <h1 className="mt-1 text-xl font-semibold">Account</h1>
      </header>

      <Card className="border-border/70 bg-card">
        <CardHeader>
          <CardTitle>
            {isLoading ? 'Loading profile...' : me?.name || 'Unknown user'}
          </CardTitle>
          <CardDescription>{me?.email || 'No email available'}</CardDescription>
        </CardHeader>
      </Card>

      <Card className="border-border/70 bg-card">
        <CardHeader>
          <CardTitle>Appearance</CardTitle>
          <CardDescription>
            Theme customization is available in the dedicated Themes page.
          </CardDescription>
        </CardHeader>
        <CardContent>
          <Link
            to="/app/themes"
            className={cn(
              buttonVariants({ variant: 'outline' }),
              'h-9 w-full text-sm',
            )}
          >
            Open Themes
          </Link>
        </CardContent>
      </Card>

      <Card className="border-border/70 bg-card">
        <CardHeader>
          <CardTitle>Session</CardTitle>
        </CardHeader>
        <CardContent className="space-y-3">
          {error ? <p className="text-xs text-destructive">{error}</p> : null}
          <Button
            disabled={isSubmitting}
            className="h-10 w-full text-sm"
            onClick={handleLogout}
          >
            Logout
          </Button>
          <Button
            variant="outline"
            disabled={isSubmitting}
            className="h-10 w-full text-sm"
            onClick={handleLogoutAll}
          >
            Logout All Devices
          </Button>
        </CardContent>
      </Card>
    </div>
  )
}
