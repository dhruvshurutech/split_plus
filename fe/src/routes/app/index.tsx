import { useEffect, useMemo, useState } from 'react'
import { Link, createFileRoute } from '@tanstack/react-router'
import { ArrowUpRight, Layers3, Wallet } from 'lucide-react'
import type { ReactNode } from 'react'

import type { OverallUserBalance } from '@/lib/api/balances'
import type { UserGroup } from '@/lib/api/groups'
import { FeatureNotice } from '@/components/feature-notice'
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
import { listOverallUserBalances } from '@/lib/api/balances'
import { listUserGroups } from '@/lib/api/groups'
import { getMe } from '@/lib/api/users'

export const Route = createFileRoute('/app/')({ component: AppHome })

function AppHome() {
  const [name, setName] = useState('there')
  const [groups, setGroups] = useState<Array<UserGroup>>([])
  const [overallBalances, setOverallBalances] = useState<Array<OverallUserBalance>>([])
  const [isLoading, setIsLoading] = useState(true)

  useEffect(() => {
    let mounted = true
    async function loadDashboard() {
      try {
        const [me, userGroups, balances] = await Promise.all([
          Promise.resolve(getMe()),
          Promise.resolve(listUserGroups()),
          Promise.resolve(listOverallUserBalances()),
        ])
        if (!mounted) return
        setName(me.name || me.email.split('@')[0] || 'there')
        setGroups(userGroups)
        setOverallBalances(balances)
      } finally {
        if (mounted) setIsLoading(false)
      }
    }
    void loadDashboard()
    return () => {
      mounted = false
    }
  }, [])

  const netSummary = useMemo(() => {
    if (overallBalances.length === 0) {
      return {
        label: '₹0.00',
        tone: 'neutral' as const,
        note: 'No balances yet.',
      }
    }

    const currencies = [...new Set(overallBalances.map((row) => row.currencyCode))]
    if (currencies.length !== 1) {
      return {
        label: `${overallBalances.length} groups`,
        tone: 'neutral' as const,
        note: 'Multi-currency balances are shown per group in Groups.',
      }
    }

    const currencyCode = currencies[0] || 'USD'
    const total = overallBalances.reduce((sum, row) => {
      const value = Number(row.balance)
      return Number.isFinite(value) ? sum + value : sum
    }, 0)
    const absolute = Math.abs(total)
    const formatted = new Intl.NumberFormat('en', {
      style: 'currency',
      currency: currencyCode,
      maximumFractionDigits: 2,
    }).format(absolute)

    if (total > 0) {
      return {
        label: `+ ${formatted}`,
        tone: 'positive' as const,
        note: 'You are owed more than you owe.',
      }
    }
    if (total < 0) {
      return {
        label: `- ${formatted}`,
        tone: 'negative' as const,
        note: 'You owe more than you are owed.',
      }
    }
    return {
      label: formatted,
      tone: 'neutral' as const,
      note: 'You are settled across current groups.',
    }
  }, [overallBalances])

  return (
    <div className="space-y-4">
      <header>
        <p className="text-sm text-muted-foreground">Dashboard</p>
        <h1 className="mt-2 text-2xl font-semibold">
          {isLoading ? 'Loading workspace...' : `Good to see you, ${name}`}
        </h1>
      </header>

      <Card className="border-border/70 bg-card">
        <CardHeader>
          <CardTitle className="flex items-center justify-between text-base">
            Net Balance
          </CardTitle>
          <CardDescription>Across all groups and friends</CardDescription>
        </CardHeader>
        <CardContent>
          <p
            className={`text-3xl font-semibold ${
              netSummary.tone === 'positive'
                ? 'text-green-600'
                : netSummary.tone === 'negative'
                  ? 'text-destructive'
                  : ''
            }`}
          >
            {isLoading ? '...' : netSummary.label}
          </p>
          <p className="mt-1 text-xs text-muted-foreground">{netSummary.note}</p>
        </CardContent>
      </Card>

      <div className="grid grid-cols-2 gap-3">
        <QuickCard
          title="Groups"
          value={isLoading ? '...' : String(groups.length)}
          icon={<Layers3 className="size-4" />}
          to="/app/groups"
        />
        <QuickCard
          title="Friends"
          value="Soon"
          icon={<Wallet className="size-4" />}
          to="/app/friends"
        />
      </div>

      <FeatureNotice
        title="Friends balance is not included yet"
        description="Friend-level settlement and dashboard totals will be enabled in a follow-up alpha update."
      />
    </div>
  )
}

function QuickCard({
  title,
  value,
  icon,
  to,
}: {
  title: string
  value: string
  icon: ReactNode
  to: string
}) {
  return (
    <Link to={to} className="block">
      <Card className="border-border/70 bg-card transition hover:bg-muted/20">
        <CardHeader>
          <CardTitle className="flex items-center justify-between text-sm">
            {title}
            {icon}
          </CardTitle>
        </CardHeader>
        <CardContent className="flex items-end justify-between">
          <span className="text-2xl font-semibold">{value}</span>
          <ArrowUpRight className="size-4 text-muted-foreground" />
        </CardContent>
      </Card>
    </Link>
  )
}
