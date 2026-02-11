import { Link, createFileRoute } from '@tanstack/react-router'
import { ArrowRight, Check, Sparkles, Users } from 'lucide-react'
import type { ReactNode } from 'react'
import { Card, CardContent } from '@/components/ui/card'

export const Route = createFileRoute('/')({ component: LandingPage })

function LandingPage() {
  return (
    <main className="mx-auto flex min-h-screen w-full max-w-md flex-col justify-between px-6 py-8">
      <div className="space-y-6">
        <div className="inline-flex items-center gap-2 rounded-full border border-border bg-card px-3 py-1 text-xs text-muted-foreground">
          <Check className="size-3" /> Split bills without awkwardness
        </div>

        <h1 className="text-4xl font-semibold leading-tight text-foreground">
          Shared expenses,
          <br />
          without the noise.
        </h1>
        <p className="max-w-sm text-base leading-7 text-muted-foreground">
          Keep trips, dinners, and shared costs in one place. Everyone can see
          who paid, who owes, and what is left to settle.
        </p>

        <div className="grid gap-2">
          <FeatureRow
            icon={<Users className="size-4" />}
            title="Made for real groups"
            detail="Great for friends, roommates, and travel plans."
          />
          <FeatureRow
            icon={<Sparkles className="size-4" />}
            title="Clear and stress-free"
            detail="Simple summaries so money talks stay easy."
          />
        </div>

        <Card className="border-border/70 bg-card">
          <CardContent className="pt-1">
            <p className="text-sm text-muted-foreground">
              Pick the look you like in Profile:
            </p>
            <p className="mt-2 text-sm">
              Minimal • Newspaper • Soft Pro • Atelier • Mono Slate • Text
              Minimal
            </p>
          </CardContent>
        </Card>
      </div>

      <div className="space-y-3">
        <Link
          to="/signup"
          className="flex h-12 items-center justify-between rounded-[var(--radius)] border border-border bg-primary px-4 text-sm font-medium text-primary-foreground"
        >
          Create account
          <ArrowRight className="size-4" />
        </Link>
        <Link
          to="/login"
          className="flex h-12 items-center justify-center rounded-[var(--radius)] border border-border bg-card text-sm text-foreground"
        >
          I already have an account
        </Link>
        <Link
          to="/app"
          className="block text-center text-sm text-muted-foreground underline-offset-4 hover:underline"
        >
          Continue in demo mode
        </Link>
      </div>
    </main>
  )
}

function FeatureRow({
  icon,
  title,
  detail,
}: {
  icon: ReactNode
  title: string
  detail: string
}) {
  return (
    <div className="flex items-start gap-3 rounded-[var(--radius)] border border-border/70 bg-card px-3 py-2.5">
      <span className="mt-0.5 text-muted-foreground">{icon}</span>
      <div>
        <p className="text-sm font-medium">{title}</p>
        <p className="text-sm text-muted-foreground">{detail}</p>
      </div>
    </div>
  )
}
