import { useState } from 'react'
import {
  Link,
  createFileRoute,
  redirect,
  useNavigate,
} from '@tanstack/react-router'
import type { FormEvent } from 'react'

import { Button } from '@/components/ui/button'
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { login } from '@/lib/api/auth'
import { hasSession } from '@/lib/session'

export const Route = createFileRoute('/login')({
  beforeLoad: () => {
    if (typeof window === 'undefined') return
    if (hasSession()) {
      throw redirect({ to: '/app' })
    }
  },
  component: LoginPage,
})

function LoginPage() {
  const navigate = useNavigate()
  const [isSubmitting, setIsSubmitting] = useState(false)
  const [error, setError] = useState<string | null>(null)

  async function handleSubmit(e: FormEvent<HTMLFormElement>) {
    e.preventDefault()

    const formData = new FormData(e.currentTarget)
    const email = String(formData.get('email') || '')
    const password = String(formData.get('password') || '')

    setError(null)
    setIsSubmitting(true)

    try {
      await login(email, password)
      navigate({ to: '/app' })
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Unable to sign in')
    } finally {
      setIsSubmitting(false)
    }
  }

  return (
    <main className="mx-auto flex min-h-screen w-full max-w-md items-center px-6 py-10">
      <Card className="w-full border-border/70 bg-card">
        <CardHeader>
          <CardTitle className="text-lg">Welcome back</CardTitle>
          <CardDescription>
            Sign in to continue to your groups and balances.
          </CardDescription>
        </CardHeader>
        <CardContent>
          <form className="space-y-3" onSubmit={handleSubmit}>
            <Input type="email" name="email" required placeholder="Email" />
            <Input
              type="password"
              name="password"
              required
              placeholder="Password"
            />
            {error ? <p className="text-xs text-destructive">{error}</p> : null}
            <Button type="submit" disabled={isSubmitting} className="w-full">
              Sign in
            </Button>
          </form>
          <p className="mt-4 text-center text-xs text-muted-foreground">
            New here?{' '}
            <Link
              to="/signup"
              className="text-foreground underline-offset-4 hover:underline"
            >
              Create an account
            </Link>
          </p>
        </CardContent>
      </Card>
    </main>
  )
}
