import { useEffect, useState } from 'react'
import { Link, createFileRoute, useNavigate } from '@tanstack/react-router'

import type { Invitation } from '@/lib/api/invitations'
import { Button } from '@/components/ui/button'
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import {
  acceptInvitation,
  getInvitation,
  joinGroupViaInvitation,
} from '@/lib/api/invitations'
import { login } from '@/lib/api/auth'
import { ApiError } from '@/lib/api/client'
import { clearSessionTokens, hasSession } from '@/lib/session'

export const Route = createFileRoute('/invite/$token')({
  component: InviteTokenPage,
})

function InviteTokenPage() {
  const navigate = useNavigate()
  const { token } = Route.useParams()
  const [invitation, setInvitation] = useState<Invitation | null>(null)
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [isSubmitting, setIsSubmitting] = useState(false)
  const [name, setName] = useState('')
  const [password, setPassword] = useState('')
  const [successMessage, setSuccessMessage] = useState<string | null>(null)

  const isAuthenticated = hasSession()

  useEffect(() => {
    let mounted = true

    async function loadInvitation() {
      try {
        const response = await getInvitation(token)
        if (mounted) setInvitation(response)
      } catch (err) {
        if (mounted)
          setError(
            err instanceof Error
              ? err.message
              : 'Unable to load invitation details.',
          )
      } finally {
        if (mounted) setIsLoading(false)
      }
    }

    loadInvitation()
    return () => {
      mounted = false
    }
  }, [token])

  async function handleAccept() {
    if (!invitation) return
    setError(null)
    setSuccessMessage(null)
    setIsSubmitting(true)
    try {
      await acceptInvitation(token)
      setSuccessMessage('Invitation accepted. Redirecting to your group...')
      navigate({
        to: '/app/groups/$groupId',
        params: { groupId: invitation.groupId },
      })
    } catch (err) {
      if (err instanceof ApiError && err.statusCode === 401) {
        clearSessionTokens()
        navigate({ to: '/login' })
        return
      }

      setError(
        err instanceof Error ? err.message : 'Unable to accept invitation.',
      )
    } finally {
      setIsSubmitting(false)
    }
  }

  async function handleJoin(e: React.FormEvent<HTMLFormElement>) {
    e.preventDefault()
    if (!invitation) return
    setError(null)
    setSuccessMessage(null)
    setIsSubmitting(true)
    try {
      await joinGroupViaInvitation(token, { name, password })
      await login(invitation.email, password)
      setSuccessMessage('Joined successfully. Redirecting to your group...')
      navigate({
        to: '/app/groups/$groupId',
        params: { groupId: invitation.groupId },
      })
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Unable to join group.')
    } finally {
      setIsSubmitting(false)
    }
  }

  return (
    <main className="mx-auto flex min-h-screen w-full max-w-md items-center px-6 py-10">
      <Card className="w-full border-border/70 bg-card">
        <CardHeader>
          <CardTitle className="text-lg">Invitation</CardTitle>
          <CardDescription>
            Join the group and start splitting expenses.
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          {isLoading ? (
            <p className="text-sm text-muted-foreground">
              Loading invitation...
            </p>
          ) : null}

          {error ? <p className="text-sm text-destructive">{error}</p> : null}
          {successMessage ? (
            <p className="text-sm text-foreground">{successMessage}</p>
          ) : null}

          {invitation ? (
            <div className="space-y-3">
              <div className="rounded-[var(--radius)] border border-border/70 bg-muted/20 p-4">
                <p className="text-base font-medium">
                  {(invitation.inviterName ||
                    invitation.inviterEmail ||
                    'Someone') + ' invited you to '}
                  <span className="font-semibold">
                    {invitation.groupName || 'this group'}
                  </span>
                </p>
                {invitation.email ? (
                  <p className="mt-1 text-xs text-muted-foreground">
                    Invitation for {invitation.email}
                  </p>
                ) : null}
              </div>

              {isAuthenticated ? (
                <Button
                  disabled={isSubmitting}
                  onClick={handleAccept}
                  className="w-full"
                >
                  {isSubmitting ? 'Accepting...' : 'Accept invitation'}
                </Button>
              ) : (
                <form className="space-y-3" onSubmit={handleJoin}>
                  <Input
                    placeholder="Name (optional)"
                    value={name}
                    onChange={(e) => setName(e.target.value)}
                  />
                  <Input
                    type="password"
                    required
                    placeholder="Password"
                    value={password}
                    onChange={(e) => setPassword(e.target.value)}
                  />
                  <p className="text-xs text-muted-foreground">
                    Set a password to accept this invitation.
                  </p>
                  <Button
                    type="submit"
                    disabled={isSubmitting}
                    className="w-full"
                  >
                    {isSubmitting ? 'Joining...' : 'Join group'}
                  </Button>
                </form>
              )}
            </div>
          ) : null}

          <div className="grid gap-2">
            {!isAuthenticated ? (
              <Link
                to="/login"
                className="text-center text-xs text-muted-foreground underline-offset-4 hover:underline"
              >
                Already have an account? Sign in first
              </Link>
            ) : null}
            <Link
              to="/"
              className="text-center text-xs text-muted-foreground underline-offset-4 hover:underline"
            >
              Back to home
            </Link>
          </div>
        </CardContent>
      </Card>
    </main>
  )
}
