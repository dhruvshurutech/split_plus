import { useEffect, useMemo, useState } from 'react'
import { Link, createFileRoute } from '@tanstack/react-router'
import { ArrowLeft } from 'lucide-react'

import type { GroupExpenseDetail } from '@/lib/api/expenses'
import { getGroupExpenseById } from '@/lib/api/expenses'
import { buttonVariants } from '@/components/ui/button'
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
import { cn } from '@/lib/utils'

export const Route = createFileRoute(
  '/app/groups/$groupId/expenses/$expenseId',
)({ component: ExpenseDetailPage })

function ExpenseDetailPage() {
  const { groupId, expenseId } = Route.useParams()
  const [detail, setDetail] = useState<GroupExpenseDetail | null>(null)
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  const paidByLabel = useMemo(() => {
    const first = detail?.payments[0]
    if (!first) return 'Unknown'
    return (
      first.user.name ||
      first.user.email ||
      first.pendingUser?.name ||
      first.pendingUser?.email ||
      'Unknown'
    )
  }, [detail?.payments])

  useEffect(() => {
    let mounted = true
    setIsLoading(true)
    setError(null)

    async function loadExpense() {
      try {
        const loaded = await getGroupExpenseById(groupId, expenseId)
        if (!mounted) return
        setDetail(loaded)
      } catch (err) {
        if (!mounted) return
        setError(
          err instanceof Error ? err.message : 'Unable to load expense detail.',
        )
      } finally {
        if (mounted) setIsLoading(false)
      }
    }

    loadExpense()
    return () => {
      mounted = false
    }
  }, [groupId, expenseId])

  return (
    <div className="space-y-4">
      <header className="space-y-1">
        <div className="mb-2 flex items-center gap-2">
          <Link
            to="/app/groups/$groupId"
            params={{ groupId }}
            aria-label="Back to group"
            className={cn(
              buttonVariants({ variant: 'outline', size: 'icon' }),
              'size-8 rounded-full',
            )}
          >
            <ArrowLeft className="size-4" />
          </Link>
          <h1 className="text-sm text-muted-foreground">Expense detail</h1>
        </div>
        <p className="text-2xl font-semibold leading-tight">
          {detail?.expense.title || `Expense #${expenseId}`}
        </p>
      </header>

      <Card className="border-border/70 bg-card">
        <CardHeader>
          <CardTitle>Summary</CardTitle>
          <CardDescription>Breakdown and participants</CardDescription>
        </CardHeader>
        <CardContent className="space-y-2 text-xs">
          {isLoading ? (
            <p className="text-muted-foreground">Loading expense...</p>
          ) : null}
          {error ? <p className="text-destructive">{error}</p> : null}

          {!isLoading && !error && detail ? (
            <>
              <div className="flex items-center justify-between">
                <span>Total</span>
                <span className="font-medium">
                  {formatAmount(
                    detail.expense.amount,
                    detail.expense.currencyCode,
                  )}
                </span>
              </div>
              <div className="flex items-center justify-between">
                <span>Date</span>
                <span>{formatDate(detail.expense.date)}</span>
              </div>
              <div className="flex items-center justify-between">
                <span>Paid by</span>
                <span>{paidByLabel}</span>
              </div>
              <div className="flex items-center justify-between">
                <span>Splits</span>
                <span>{detail.splits.length}</span>
              </div>
              {detail.expense.notes ? (
                <div className="rounded-[var(--radius)] border border-border/70 p-2 text-[11px] text-muted-foreground">
                  {detail.expense.notes}
                </div>
              ) : null}
            </>
          ) : null}
        </CardContent>
      </Card>

      <Card className="border-border/70 bg-card">
        <CardHeader>
          <CardTitle>Payments</CardTitle>
        </CardHeader>
        <CardContent className="space-y-2 text-xs">
          {!isLoading && !error && detail?.payments.length === 0 ? (
            <p className="text-muted-foreground">No payment rows found.</p>
          ) : null}

          {detail?.payments.map((payment) => (
            <div
              key={payment.id}
              className="flex items-center justify-between rounded-[var(--radius)] border border-border/60 p-2"
            >
              <span>
                {payment.user.name ||
                  payment.user.email ||
                  payment.pendingUser?.name ||
                  payment.pendingUser?.email ||
                  'Unknown user'}
              </span>
              <span className="font-medium">
                {formatAmount(payment.amount, detail.expense.currencyCode)}
              </span>
            </div>
          ))}
        </CardContent>
      </Card>

      <Card className="border-border/70 bg-card">
        <CardHeader>
          <CardTitle>Split ownership</CardTitle>
        </CardHeader>
        <CardContent className="space-y-2 text-xs">
          {!isLoading && !error && detail?.splits.length === 0 ? (
            <p className="text-muted-foreground">No split rows found.</p>
          ) : null}

          {detail?.splits.map((split) => (
            <div
              key={split.id}
              className="flex items-center justify-between rounded-[var(--radius)] border border-border/60 p-2"
            >
              <div>
                <p>
                  {split.user.name ||
                    split.user.email ||
                    split.pendingUser?.name ||
                    split.pendingUser?.email ||
                    'Unknown user'}
                </p>
                <p className="text-[10px] capitalize text-muted-foreground">
                  {split.splitType}
                </p>
              </div>
              <span className="font-medium">
                {formatAmount(
                  split.amountOwned,
                  detail.expense.currencyCode || 'USD',
                )}
              </span>
            </div>
          ))}
        </CardContent>
      </Card>

      <Link
        to="/app/groups/$groupId"
        params={{ groupId }}
        className="inline-flex h-10 w-full items-center justify-center rounded-[var(--radius)] border border-input px-4 text-sm font-medium transition-colors hover:bg-muted/70"
      >
        Back to group
      </Link>
    </div>
  )
}

function formatAmount(amount: string, currencyCode: string) {
  const numericAmount = Number(amount)
  if (!Number.isFinite(numericAmount)) return `${amount} ${currencyCode}`

  return new Intl.NumberFormat('en', {
    style: 'currency',
    currency: currencyCode,
    maximumFractionDigits: 2,
  }).format(numericAmount)
}

function formatDate(value: string) {
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return value

  return new Intl.DateTimeFormat('en', {
    month: 'short',
    day: 'numeric',
    year: 'numeric',
  }).format(date)
}
