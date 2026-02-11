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
import { signup } from '@/lib/api/auth'
import { hasSession } from '@/lib/session'

export const Route = createFileRoute('/signup')({
  beforeLoad: () => {
    if (typeof window === 'undefined') return
    if (hasSession()) {
      throw redirect({ to: '/app' })
    }
  },
  component: SignupPage,
})

function SignupPage() {
  const navigate = useNavigate()
  const [isSubmitting, setIsSubmitting] = useState(false)
  const [error, setError] = useState<string | null>(null)

  async function handleSubmit(e: FormEvent<HTMLFormElement>) {
    e.preventDefault()
    const formData = new FormData(e.currentTarget)
    const name = String(formData.get('name') || '')
    const email = String(formData.get('email') || '')
    const password = String(formData.get('password') || '')

    setError(null)
    setIsSubmitting(true)
    try {
      await signup(name, email, password)
      navigate({ to: '/app' })
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Unable to create account')
    } finally {
      setIsSubmitting(false)
    }
  }

  return (
    <main className="mx-auto flex min-h-screen w-full max-w-md items-center px-6 py-10">
      <Card className="w-full border-border/70 bg-card">
        <CardHeader>
          <CardTitle className="text-lg">Create your account</CardTitle>
          <CardDescription>
            Start splitting expenses in under a minute.
          </CardDescription>
        </CardHeader>
        <CardContent>
          <form className="space-y-3" onSubmit={handleSubmit}>
            <Input type="text" name="name" required placeholder="Full name" />
            <Input type="email" name="email" required placeholder="Email" />
            <Input
              type="password"
              name="password"
              required
              placeholder="Password"
            />
            {error ? <p className="text-xs text-destructive">{error}</p> : null}
            <Button type="submit" disabled={isSubmitting} className="w-full">
              Create account
            </Button>
          </form>
          <p className="mt-4 text-center text-xs text-muted-foreground">
            Already registered?{' '}
            <Link
              to="/login"
              className="text-foreground underline-offset-4 hover:underline"
            >
              Sign in
            </Link>
          </p>
        </CardContent>
      </Card>
    </main>
  )
}
