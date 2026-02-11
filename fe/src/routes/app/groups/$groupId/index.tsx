import { useCallback, useEffect, useMemo, useRef, useState } from 'react'
import { Link, createFileRoute } from '@tanstack/react-router'
import { ArrowLeft, ArrowRight, Copy, ReceiptText, Scale, Users } from 'lucide-react'

import type { GroupActivity } from '@/lib/api/activities'
import type { GroupBalance, GroupDebt } from '@/lib/api/balances'
import type { GroupExpense } from '@/lib/api/expenses'
import type { GroupMember, UserGroup } from '@/lib/api/groups'
import type { GroupSettlementWithUsers } from '@/lib/api/settlements'
import { listGroupActivities } from '@/lib/api/activities'
import { listGroupBalances, listGroupDebts } from '@/lib/api/balances'
import { listRecentGroupExpenses } from '@/lib/api/expenses'
import {
  buildInvitationLink,
  createGroupInvitation,
  listGroupMembers,
  listUserGroups,
} from '@/lib/api/groups'
import { listGroupSettlements } from '@/lib/api/settlements'
import { getMe } from '@/lib/api/users'
import { Button, buttonVariants } from '@/components/ui/button'
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { cn } from '@/lib/utils'

const EXPENSE_PAGE_SIZE = 12
const TOTALS_FETCH_PAGE_SIZE = 100
const EXPENSE_TIME_FORMATTER = new Intl.DateTimeFormat('en', {
  month: 'short',
  day: 'numeric',
  hour: 'numeric',
  minute: '2-digit',
})

type GroupTab = 'transactions' | 'members' | 'totals' | 'activity'

export const Route = createFileRoute('/app/groups/$groupId/')({
  component: GroupDetailPage,
})

function GroupDetailPage() {
  const { groupId } = Route.useParams()
  const [activeTab, setActiveTab] = useState<GroupTab>('transactions')

  const activityLoaderRef = useRef<HTMLDivElement | null>(null)
  const isExpensesLoadingRef = useRef(false)

  const [group, setGroup] = useState<UserGroup | null>(null)
  const [members, setMembers] = useState<Array<GroupMember>>([])
  const [balances, setBalances] = useState<Array<GroupBalance>>([])
  const [debts, setDebts] = useState<Array<GroupDebt>>([])
  const [settlements, setSettlements] = useState<Array<GroupSettlementWithUsers>>(
    [],
  )
  const [activities, setActivities] = useState<Array<GroupActivity>>([])

  const [isPageLoading, setIsPageLoading] = useState(true)
  const [pageError, setPageError] = useState<string | null>(null)

  const [expenses, setExpenses] = useState<Array<GroupExpense>>([])
  const [expenseOffset, setExpenseOffset] = useState(0)
  const [hasMoreExpenses, setHasMoreExpenses] = useState(true)
  const [isExpensesLoading, setIsExpensesLoading] = useState(false)
  const [expensesError, setExpensesError] = useState<string | null>(null)

  const [groupTotalSpent, setGroupTotalSpent] = useState('0')
  const [isTotalsLoading, setIsTotalsLoading] = useState(true)

  const [inviteEmail, setInviteEmail] = useState('')
  const [inviteName, setInviteName] = useState('')
  const [inviteRole, setInviteRole] = useState<'member' | 'admin'>('member')
  const [inviteError, setInviteError] = useState<string | null>(null)
  const [inviteLink, setInviteLink] = useState<string | null>(null)
  const [isInviting, setIsInviting] = useState(false)
  const [copyState, setCopyState] = useState<'idle' | 'copied' | 'failed'>(
    'idle',
  )
  const [actionNotice, setActionNotice] = useState<string | null>(null)

  const [meId, setMeId] = useState('')

  const activeMembersCount = useMemo(
    () => members.filter((member) => member.status === 'active').length,
    [members],
  )

  const pendingMembersCount = useMemo(
    () => members.filter((member) => member.status === 'pending').length,
    [members],
  )

  const myBalance = useMemo(
    () => balances.find((balance) => balance.userId === meId) || null,
    [balances, meId],
  )

  const netBalancePreview = useMemo(() => {
    const netByEntity = new Map<string, { label: string; balance: number }>()

    function upsertEntity(key: string, label: string, delta: number) {
      const existing = netByEntity.get(key)
      if (existing) {
        existing.balance += delta
        return
      }
      netByEntity.set(key, { label, balance: delta })
    }

    for (const debt of debts) {
      const amount = Number(debt.amount)
      if (!Number.isFinite(amount) || amount <= 0) continue

      const debtorKey = debt.debtorPendingUserId
        ? `pending:${debt.debtorPendingUserId}`
        : `user:${debt.debtorId}`
      const debtorLabel = debt.debtorName || debt.debtorEmail || 'Unknown'
      upsertEntity(debtorKey, debtorLabel, -amount)

      const creditorKey = debt.creditorPendingUserId
        ? `pending:${debt.creditorPendingUserId}`
        : `user:${debt.creditorId}`
      const creditorLabel = debt.creditorName || debt.creditorEmail || 'Unknown'
      upsertEntity(creditorKey, creditorLabel, amount)
    }

    return [...netByEntity.entries()]
      .map(([key, value]) => ({
        key,
        label: value.label,
        numericBalance: value.balance,
      }))
      .sort((a, b) => Math.abs(b.numericBalance) - Math.abs(a.numericBalance))
      .slice(0, 6)
  }, [debts])

  const memberNetByKey = useMemo(() => {
    const netByEntity = new Map<string, number>()

    for (const debt of debts) {
      const amount = Number(debt.amount)
      if (!Number.isFinite(amount) || amount <= 0) continue

      const debtorKey = debt.debtorPendingUserId
        ? `pending:${debt.debtorPendingUserId}`
        : `user:${debt.debtorId}`
      const creditorKey = debt.creditorPendingUserId
        ? `pending:${debt.creditorPendingUserId}`
        : `user:${debt.creditorId}`

      netByEntity.set(debtorKey, (netByEntity.get(debtorKey) || 0) - amount)
      netByEntity.set(
        creditorKey,
        (netByEntity.get(creditorKey) || 0) + amount,
      )
    }

    return netByEntity
  }, [debts])

  const memberTopDebtByKey = useMemo(() => {
    const topDebtByKey = new Map<
      string,
      {
        creditorLabel: string
        amount: number
        creditorId: string
        creditorIsPending: boolean
      }
    >()

    for (const debt of debts) {
      const amount = Number(debt.amount)
      if (!Number.isFinite(amount) || amount <= 0) continue

      const debtorKey = debt.debtorPendingUserId
        ? `pending:${debt.debtorPendingUserId}`
        : `user:${debt.debtorId}`
      const existing = topDebtByKey.get(debtorKey)
      if (existing && existing.amount >= amount) continue

      topDebtByKey.set(debtorKey, {
        creditorLabel: debt.creditorName || debt.creditorEmail || 'Unknown',
        amount,
        creditorId: debt.creditorPendingUserId || debt.creditorId,
        creditorIsPending: Boolean(debt.creditorPendingUserId),
      })
    }

    return topDebtByKey
  }, [debts])

  const memberTopCreditByKey = useMemo(() => {
    const topCreditByKey = new Map<
      string,
      {
        debtorLabel: string
        amount: number
        debtorId: string
        debtorIsPending: boolean
      }
    >()

    for (const debt of debts) {
      const amount = Number(debt.amount)
      if (!Number.isFinite(amount) || amount <= 0) continue

      const creditorKey = debt.creditorPendingUserId
        ? `pending:${debt.creditorPendingUserId}`
        : `user:${debt.creditorId}`
      const existing = topCreditByKey.get(creditorKey)
      if (existing && existing.amount >= amount) continue

      topCreditByKey.set(creditorKey, {
        debtorLabel: debt.debtorName || debt.debtorEmail || 'Unknown',
        amount,
        debtorId: debt.debtorPendingUserId || debt.debtorId,
        debtorIsPending: Boolean(debt.debtorPendingUserId),
      })
    }

    return topCreditByKey
  }, [debts])

  const totalOutstandingAmount = useMemo(
    () =>
      debts.reduce((sum, debt) => {
        const amount = Number(debt.amount)
        if (!Number.isFinite(amount)) return sum
        return sum + amount
      }, 0),
    [debts],
  )

  const settlementSuggestions = useMemo(() => {
    const summaryByPair = new Map<
      string,
      {
        key: string
        debtorLabel: string
        creditorLabel: string
        debtorId: string
        creditorId: string
        debtorIsPending: boolean
        creditorIsPending: boolean
        amount: number
      }
    >()

    for (const debt of debts) {
      const amount = Number(debt.amount)
      if (!Number.isFinite(amount) || amount <= 0) continue

      const debtorId = debt.debtorPendingUserId || debt.debtorId
      const creditorId = debt.creditorPendingUserId || debt.creditorId
      const pairKey = `${debtorId}:${creditorId}`
      const existing = summaryByPair.get(pairKey)
      if (existing) {
        existing.amount += amount
        continue
      }
      summaryByPair.set(pairKey, {
        key: pairKey,
        debtorLabel: debt.debtorName || debt.debtorEmail || 'Unknown',
        creditorLabel: debt.creditorName || debt.creditorEmail || 'Unknown',
        debtorId,
        creditorId,
        debtorIsPending: Boolean(debt.debtorPendingUserId),
        creditorIsPending: Boolean(debt.creditorPendingUserId),
        amount,
      })
    }

    return [...summaryByPair.values()].sort((a, b) => b.amount - a.amount)
  }, [debts])

  const totalsChartMax = useMemo(() => {
    const values = [
      Number(groupTotalSpent),
      Number(myBalance?.totalPaid || '0'),
      Number(myBalance?.totalOwed || '0'),
      Math.abs(Number(myBalance?.balance || '0')),
    ].filter((value) => Number.isFinite(value))

    return Math.max(...values, 1)
  }, [groupTotalSpent, myBalance?.balance, myBalance?.totalOwed, myBalance?.totalPaid])

  const transactionFeed = useMemo(() => {
    const expenseItems = expenses.map((row) => ({
      id: row.expense.id,
      type: 'expense' as const,
      timestamp: row.expense.date,
      title: row.expense.title,
      subtitle:
        row.payments[0]?.user?.name ||
        row.payments[0]?.user?.email ||
        row.payments[0]?.pendingUser?.name ||
        row.payments[0]?.pendingUser?.email
          ? `Paid by ${
              row.payments[0].user.name ||
              row.payments[0].user.email ||
              row.payments[0].pendingUser?.name ||
              row.payments[0].pendingUser?.email
            }`
          : '',
      amount: row.expense.amount,
      currencyCode: row.expense.currencyCode,
      expenseId: row.expense.id,
    }))

    const settlementItems = settlements.map((settlement) => ({
      id: settlement.id,
      type: 'settlement' as const,
      timestamp: settlement.updatedAt,
      title: `${settlement.payer.name || settlement.payer.email || 'Unknown'} paid ${
        settlement.payee.name || settlement.payee.email || 'Unknown'
      }`,
      subtitle: settlement.status,
      amount: settlement.amount,
      currencyCode: settlement.currencyCode || group?.currencyCode || 'USD',
    }))

    return [...expenseItems, ...settlementItems].sort((a, b) => {
      const aTime = new Date(a.timestamp).getTime()
      const bTime = new Date(b.timestamp).getTime()
      return bTime - aTime
    })
  }, [expenses, group?.currencyCode, settlements])

  useEffect(() => {
    let mounted = true
    setIsPageLoading(true)
    setPageError(null)

    async function loadPageData() {
      try {
        const [
          groupList,
          me,
          groupMembers,
          groupBalances,
          groupDebts,
          groupSettlements,
          groupActivities,
        ] = await Promise.all([
          Promise.resolve(listUserGroups()),
          Promise.resolve(getMe()),
          Promise.resolve(listGroupMembers(groupId, { force: true })),
          Promise.resolve(listGroupBalances(groupId)).catch(() => []),
          Promise.resolve(listGroupDebts(groupId)).catch(() => []),
          Promise.resolve(listGroupSettlements(groupId)).catch(() => []),
          Promise.resolve(listGroupActivities(groupId, { limit: 24 })).catch(
            () => [],
          ),
        ])

        if (!mounted) return

        setGroup(groupList.find((entry) => entry.id === groupId) || null)
        setMeId(me.id)
        setMembers(groupMembers)
        setBalances(groupBalances)
        setDebts(groupDebts)
        setSettlements(groupSettlements)
        setActivities(groupActivities)
      } catch (err) {
        if (!mounted) return
        setPageError(
          err instanceof Error ? err.message : 'Unable to load group details.',
        )
      } finally {
        if (mounted) setIsPageLoading(false)
      }
    }

    loadPageData()
    return () => {
      mounted = false
    }
  }, [groupId])

  const loadExpensePage = useCallback(
    async (offset: number, replace: boolean) => {
      if (isExpensesLoadingRef.current) return

      isExpensesLoadingRef.current = true
      setIsExpensesLoading(true)
      setExpensesError(null)

      try {
        const rows = await listRecentGroupExpenses(groupId, {
          limit: EXPENSE_PAGE_SIZE,
          offset,
        })

        setExpenses((prev) => (replace ? rows : [...prev, ...rows]))
        setExpenseOffset(offset + rows.length)
        setHasMoreExpenses(rows.length === EXPENSE_PAGE_SIZE)
      } catch (err) {
        setExpensesError(
          err instanceof Error
            ? err.message
            : 'Unable to load recent transactions.',
        )
      } finally {
        isExpensesLoadingRef.current = false
        setIsExpensesLoading(false)
      }
    },
    [groupId],
  )

  useEffect(() => {
    setExpenses([])
    setExpenseOffset(0)
    setHasMoreExpenses(true)
    void loadExpensePage(0, true)
  }, [groupId, loadExpensePage])

  useEffect(() => {
    let mounted = true
    setIsTotalsLoading(true)

    async function loadTotals() {
      try {
        const allExpenses: Array<GroupExpense> = []
        let offset = 0

        for (;;) {
          const batch = await listRecentGroupExpenses(groupId, {
            limit: TOTALS_FETCH_PAGE_SIZE,
            offset,
            force: true,
          })
          allExpenses.push(...batch)

          if (batch.length < TOTALS_FETCH_PAGE_SIZE) break
          offset += batch.length

          if (offset > 5000) break
        }

        if (!mounted) return

        const total = allExpenses.reduce((sum, row) => {
          const amount = Number(row.expense.amount)
          if (!Number.isFinite(amount)) return sum
          return sum + amount
        }, 0)

        setGroupTotalSpent(total.toFixed(2))
      } catch {
        if (!mounted) return
        setGroupTotalSpent('0')
      } finally {
        if (mounted) setIsTotalsLoading(false)
      }
    }

    void loadTotals()
    return () => {
      mounted = false
    }
  }, [groupId])

  const loadMoreExpenses = useCallback(() => {
    if (!hasMoreExpenses || isExpensesLoading) return
    void loadExpensePage(expenseOffset, false)
  }, [expenseOffset, hasMoreExpenses, isExpensesLoading, loadExpensePage])

  useEffect(() => {
    const node = activityLoaderRef.current
    if (!node || !hasMoreExpenses || activeTab !== 'transactions') return

    const observer = new IntersectionObserver(
      (entries) => {
        if (entries[0]?.isIntersecting) {
          loadMoreExpenses()
        }
      },
      { rootMargin: '140px' },
    )

    observer.observe(node)
    return () => observer.disconnect()
  }, [activeTab, hasMoreExpenses, loadMoreExpenses])

  async function handleInvite(e: React.FormEvent<HTMLFormElement>) {
    e.preventDefault()
    setInviteError(null)
    setCopyState('idle')
    setIsInviting(true)

    try {
      const result = await createGroupInvitation(groupId, {
        email: inviteEmail,
        name: inviteName,
        role: inviteRole,
      })

      setInviteLink(buildInvitationLink(result.token))
      setInviteEmail('')
      setInviteName('')
      setInviteRole('member')

      const refreshedMembers = await listGroupMembers(groupId, { force: true })
      setMembers(refreshedMembers)
    } catch (err) {
      setInviteError(
        err instanceof Error ? err.message : 'Unable to create invite link.',
      )
    } finally {
      setIsInviting(false)
    }
  }

  async function handleCopyInviteLink() {
    if (!inviteLink) return
    try {
      await navigator.clipboard.writeText(inviteLink)
      setCopyState('copied')
    } catch {
      setCopyState('failed')
    }
  }

  async function handleCopyPendingInviteLink(member: GroupMember) {
    if (!member.invitationToken) return
    const link = buildInvitationLink(member.invitationToken)
    await navigator.clipboard.writeText(link)
  }

  async function handleRemindMember(member: GroupMember) {
    const memberKey =
      member.status === 'pending' ? `pending:${member.userId}` : `user:${member.userId}`
    const topDebt = memberTopDebtByKey.get(memberKey)
    if (!topDebt || !group) return

    const message = `${member.user.name || member.user.email}, please settle ${formatExpenseAmount(
      topDebt.amount.toFixed(2),
      group.currencyCode || 'USD',
    )} to ${topDebt.creditorLabel} in ${group.name}.`

    try {
      await navigator.clipboard.writeText(message)
      setActionNotice('Reminder copied')
    } catch {
      setActionNotice('Could not copy reminder')
    }
    window.setTimeout(() => setActionNotice(null), 1600)
  }

  async function handleRemindPayer(
    payerLabel: string,
    amount: number,
    receiverLabel: string,
  ) {
    if (!group) return
    const message = `${payerLabel}, please settle ${formatExpenseAmount(
      amount.toFixed(2),
      group.currencyCode || 'USD',
    )} to ${receiverLabel} in ${group.name}.`

    try {
      await navigator.clipboard.writeText(message)
      setActionNotice('Reminder copied')
    } catch {
      setActionNotice('Could not copy reminder')
    }
    window.setTimeout(() => setActionNotice(null), 1600)
  }

  return (
    <div className="space-y-3">
      <header className="flex items-center gap-2.5">
        <Link
          to="/app/groups"
          className={cn(
            buttonVariants({ variant: 'outline', size: 'icon' }),
            'size-8 rounded-full shrink-0',
          )}
          aria-label="Back to groups"
        >
          <ArrowLeft className="size-4" />
        </Link>
        <div className="min-w-0 space-y-0.5">
          <p className="text-[11px] text-muted-foreground">Group</p>
          <div className="flex items-center gap-2">
            <Users className="size-4 text-muted-foreground shrink-0" />
            <h1 className="text-base font-semibold leading-tight break-words">
              {group?.name || groupId}
            </h1>
          </div>
          <p className="line-clamp-1 text-[11px] text-muted-foreground">
            {group?.description || 'Track expenses, members, and settlement flow.'}
          </p>
        </div>
      </header>

      <div className="sticky top-14 z-10 space-y-1.5 rounded-[var(--radius)] border border-border/70 bg-background/95 p-1.5 backdrop-blur supports-[backdrop-filter]:bg-background/90">
        <div className="grid grid-cols-2 gap-1.5">
          <Link
            to="/app/groups/$groupId/expenses/new"
            params={{ groupId }}
            className="rounded-[var(--radius)] bg-primary px-3 py-2 text-center text-[11px] font-medium text-primary-foreground"
          >
            Add expense
          </Link>
          <Link
            to="/app/groups/$groupId/settlements/new"
            params={{ groupId }}
            className="rounded-[var(--radius)] border border-border/70 bg-card px-3 py-2 text-center text-[11px] font-medium"
          >
            Settle up
          </Link>
        </div>

        <div className="grid grid-cols-4 gap-1 rounded-[calc(var(--radius)-2px)] bg-muted/50 p-1">
          <TabButton
            isActive={activeTab === 'transactions'}
            label="Transactions"
            onClick={() => setActiveTab('transactions')}
          />
          <TabButton
            isActive={activeTab === 'members'}
            label="Members"
            onClick={() => setActiveTab('members')}
          />
          <TabButton
            isActive={activeTab === 'totals'}
            label="Totals"
            onClick={() => setActiveTab('totals')}
          />
          <TabButton
            isActive={activeTab === 'activity'}
            label="Activity"
            onClick={() => setActiveTab('activity')}
          />
        </div>
      </div>

      {isPageLoading ? (
        <Card className="border-border/70 bg-card">
          <CardContent className="py-5 text-xs text-muted-foreground">
            Loading group data...
          </CardContent>
        </Card>
      ) : null}

      {pageError ? (
        <Card className="border-border/70 bg-card">
          <CardContent className="py-5 text-xs text-destructive">
            {pageError}
          </CardContent>
        </Card>
      ) : null}

      {!isPageLoading && !pageError ? (
        <>
          {activeTab === 'transactions' ? (
            <div className="space-y-2.5">
              <Card className="border-border/70 bg-card">
                <CardHeader className="pb-1.5">
                  <CardTitle className="flex items-center gap-2 text-sm">
                    <Scale className="size-4" /> Settlement suggestions
                  </CardTitle>
                </CardHeader>
                <CardContent className="space-y-1.5 pt-0">
                  <p className="text-[11px] text-muted-foreground">
                    Open transfers: {debts.length} • Outstanding:{' '}
                    {formatExpenseAmount(
                      totalOutstandingAmount.toFixed(2),
                      group?.currencyCode || 'USD',
                    )}
                  </p>
                  {debts.length === 0 ? (
                    <p className="text-[11px] text-muted-foreground">
                      No pending debts. Everyone is settled.
                    </p>
                  ) : null}
                  {settlementSuggestions.slice(0, 8).map((item) => (
                    <div
                      key={item.key}
                      className="grid grid-cols-[minmax(0,1fr)_auto] items-center gap-2 border-b border-border/50 py-1.5 text-xs last:border-b-0"
                    >
                      <div className="min-w-0 inline-flex items-center gap-1.5 text-muted-foreground">
                        <span className="truncate">{item.debtorLabel}</span>
                        <ArrowRight className="size-3 shrink-0" />
                        <span className="truncate">{item.creditorLabel}</span>
                      </div>
                      <div className="inline-flex items-center gap-2">
                        <span className="font-semibold text-destructive">
                          {formatExpenseAmount(
                            item.amount.toFixed(2),
                            group?.currencyCode || 'USD',
                          )}
                        </span>
                        <Link
                          to="/app/groups/$groupId/settlements/new"
                          params={{ groupId }}
                          search={{
                            payerId: item.debtorId,
                            payeeId: item.creditorId,
                            amount: item.amount.toFixed(2),
                          }}
                          className="text-[10px] text-muted-foreground underline-offset-2 hover:text-foreground hover:underline"
                        >
                          Settle
                        </Link>
                      </div>
                    </div>
                  ))}
                </CardContent>
              </Card>

              <Card className="border-border/70 bg-card">
                <CardHeader className="pb-2">
                  <CardTitle className="flex items-center gap-2 text-sm">
                    <ReceiptText className="size-4" /> Transactions
                  </CardTitle>
                  <CardDescription>
                    Expenses and settlements combined in one feed.
                  </CardDescription>
                </CardHeader>
                <CardContent className="space-y-1.5">
                  {expenses.length === 0 && settlements.length === 0 && isExpensesLoading ? (
                    <p className="text-xs text-muted-foreground">
                      Loading transactions...
                    </p>
                  ) : null}

                  {expensesError ? (
                    <div className="space-y-2">
                      <p className="text-xs text-destructive">{expensesError}</p>
                      <button
                        type="button"
                        className="text-xs text-muted-foreground underline underline-offset-2 hover:text-foreground"
                        onClick={() => void loadExpensePage(0, true)}
                      >
                        Retry
                      </button>
                    </div>
                  ) : null}

                  {!isExpensesLoading && !expensesError && transactionFeed.length === 0 ? (
                    <p className="text-xs text-muted-foreground">
                      No transactions yet in this group.
                    </p>
                  ) : null}

                  {transactionFeed.map((item) => (
                    <div
                      key={`${item.type}-${item.id}`}
                      className="rounded-[var(--radius)] border border-border/60 p-2.5"
                    >
                      <div className="flex items-start justify-between gap-2">
                        <div>
                          <p className="inline-flex items-center gap-1.5 text-[11px] font-medium">
                            <span
                              className={cn(
                                'rounded px-1.5 py-0.5 text-[10px] uppercase tracking-wide',
                                item.type === 'expense'
                                  ? 'bg-muted text-muted-foreground'
                                  : 'bg-primary/12 text-primary',
                              )}
                            >
                              {item.type}
                            </span>
                            {item.title}
                          </p>
                          <p className="mt-0.5 text-[10px] text-muted-foreground">
                            {formatExpenseAmount(item.amount, item.currencyCode)}
                            {item.subtitle ? ` • ${item.subtitle}` : ''}
                          </p>
                        </div>
                        <p className="text-[10px] text-muted-foreground">
                          {formatExpenseTime(item.timestamp)}
                        </p>
                      </div>

                      {item.type === 'expense' ? (
                        <Link
                          to="/app/groups/$groupId/expenses/$expenseId"
                          params={{ groupId, expenseId: item.expenseId }}
                          className="mt-1 inline-flex text-[10px] text-muted-foreground underline-offset-2 hover:underline"
                        >
                          View expense
                        </Link>
                      ) : null}
                    </div>
                  ))}

                  <div ref={activityLoaderRef} className="h-4" />

                  {isExpensesLoading && expenses.length > 0 ? (
                    <p className="text-center text-[11px] text-muted-foreground">
                      Loading more...
                    </p>
                  ) : null}

                  {!hasMoreExpenses && transactionFeed.length > 0 ? (
                    <p className="text-center text-[11px] text-muted-foreground">
                      End of transactions
                    </p>
                  ) : null}
                </CardContent>
              </Card>
            </div>
          ) : null}

          {activeTab === 'members' ? (
            <div className="space-y-2.5">
              {actionNotice ? (
                <div className="rounded-[var(--radius)] border border-border/70 bg-card px-3 py-2 text-xs text-muted-foreground">
                  {actionNotice}
                </div>
              ) : null}
              <Card className="border-border/70 bg-card">
                <CardHeader className="pb-2">
                  <CardTitle className="text-sm">Members</CardTitle>
                  <CardDescription>
                    {activeMembersCount} active • {pendingMembersCount} pending
                  </CardDescription>
                </CardHeader>
                <CardContent className="space-y-1.5">
                  {members.length === 0 ? (
                    <p className="text-xs text-muted-foreground">No members found.</p>
                  ) : null}

                  {members.map((member) => {
                    const memberKey =
                      member.status === 'pending'
                        ? `pending:${member.userId}`
                        : `user:${member.userId}`
                    const net = memberNetByKey.get(memberKey) || 0
                    const topDebt = memberTopDebtByKey.get(memberKey)
                    const topCredit = memberTopCreditByKey.get(memberKey)
                    const displayName = member.user.name || member.user.email
                    const balanceLabel =
                      net < 0
                        ? `owes ${formatExpenseAmount(
                            Math.abs(net).toFixed(2),
                            group?.currencyCode || 'USD',
                          )}`
                        : net > 0
                          ? `gets ${formatExpenseAmount(
                              net.toFixed(2),
                              group?.currencyCode || 'USD',
                            )}`
                          : 'settled'
                    const hintLabel =
                      topDebt && net < 0
                        ? `pay ${topDebt.creditorLabel} ${formatExpenseAmount(
                            topDebt.amount.toFixed(2),
                            group?.currencyCode || 'USD',
                          )}`
                        : topCredit && net > 0
                          ? `receive from ${topCredit.debtorLabel} ${formatExpenseAmount(
                              topCredit.amount.toFixed(2),
                              group?.currencyCode || 'USD',
                            )}`
                          : ''

                    return (
                      <div
                        key={
                          member.id ||
                          `${member.userId}-${member.invitationToken}-${member.status}`
                        }
                        className="grid gap-2 rounded-[var(--radius)] border border-border/60 px-3 py-2.5 text-xs sm:grid-cols-[minmax(0,1fr)_auto] sm:items-start"
                      >
                        <div className="min-w-0 space-y-0.5">
                          <p className="truncate font-medium">
                            {displayName}{' '}
                            <span className="text-muted-foreground">
                              ({member.role})
                            </span>
                          </p>
                          <p className="truncate text-[11px] text-muted-foreground">
                            {member.user.email}
                          </p>
                        </div>
                        <div className="space-y-0.5 text-left sm:text-right">
                          <p
                            className={cn(
                              'text-[11px] font-medium',
                              net < 0
                                ? 'text-destructive'
                                : net > 0
                                  ? 'text-foreground'
                                  : 'text-muted-foreground',
                            )}
                          >
                            {balanceLabel}
                          </p>
                          {hintLabel ? (
                            <p className="text-[11px] text-muted-foreground">{hintLabel}</p>
                          ) : null}
                          <div className="flex items-center justify-between gap-2 pt-0.5">
                            <div>
                              {member.status === 'pending' && member.invitationToken ? (
                                <button
                                  type="button"
                                  className="inline-flex items-center gap-1 rounded px-1.5 py-0.5 text-[11px] text-muted-foreground hover:bg-muted/60"
                                  onClick={() => void handleCopyPendingInviteLink(member)}
                                >
                                  <Copy className="size-3" />
                                  Copy invite link
                                </button>
                              ) : null}
                            </div>
                            <div className="flex flex-wrap items-center gap-1.5 sm:justify-end">
                              {(topDebt && net < 0) || (topCredit && net > 0) ? (
                                <button
                                  type="button"
                                  className="rounded px-1.5 py-0.5 text-[11px] text-muted-foreground hover:bg-muted/60"
                                  onClick={() =>
                                    topDebt && net < 0
                                      ? void handleRemindMember(member)
                                      : void handleRemindPayer(
                                          topCredit?.debtorLabel || '',
                                          topCredit?.amount || 0,
                                          member.user.name || member.user.email,
                                        )
                                  }
                                >
                                  Remind
                                </button>
                              ) : null}
                              {topDebt && net < 0 ? (
                                <Link
                                  to="/app/groups/$groupId/settlements/new"
                                  params={{ groupId }}
                                  search={{
                                    payerId: member.userId,
                                    payeeId: topDebt.creditorId,
                                    amount: topDebt.amount.toFixed(2),
                                  }}
                                  className="rounded px-1.5 py-0.5 text-[11px] text-muted-foreground hover:bg-muted/60"
                                >
                                  Settle up
                                </Link>
                              ) : null}
                            </div>
                          </div>
                        </div>
                      </div>
                    )
                  })}
                </CardContent>
              </Card>

              <Card className="border-border/70 bg-card">
                <CardHeader className="pb-2">
                  <CardTitle className="text-sm">Invite member</CardTitle>
                  <CardDescription>
                    Invite by email and share a link instantly.
                  </CardDescription>
                </CardHeader>
                <CardContent className="space-y-2.5">
                  <form className="grid gap-2" onSubmit={handleInvite}>
                    <Input
                      required
                      type="email"
                      placeholder="friend@example.com"
                      value={inviteEmail}
                      onChange={(e) => setInviteEmail(e.target.value)}
                    />
                    <Input
                      placeholder="Name (optional)"
                      value={inviteName}
                      onChange={(e) => setInviteName(e.target.value)}
                    />
                    <select
                      value={inviteRole}
                      onChange={(e) =>
                        setInviteRole(e.target.value as 'member' | 'admin')
                      }
                      className="h-[var(--control-height)] w-full rounded-[var(--radius)] border border-input bg-transparent px-3 text-sm"
                    >
                      <option value="member">member</option>
                      <option value="admin">admin</option>
                    </select>

                    {inviteError ? (
                      <p className="text-xs text-destructive">{inviteError}</p>
                    ) : null}

                    <Button type="submit" disabled={isInviting} className="w-full">
                      {isInviting ? 'Creating invite...' : 'Create invite link'}
                    </Button>
                  </form>

                  {inviteLink ? (
                    <div className="space-y-2 rounded-[var(--radius)] border border-border/70 p-3">
                      <p className="text-xs text-muted-foreground">Share this link</p>
                      <p className="break-all text-xs">{inviteLink}</p>
                      <Button
                        type="button"
                        variant="outline"
                        className="w-full"
                        onClick={handleCopyInviteLink}
                      >
                        <Copy className="size-3.5" />
                        {copyState === 'copied'
                          ? 'Copied'
                          : copyState === 'failed'
                            ? 'Copy failed'
                            : 'Copy link'}
                      </Button>
                    </div>
                  ) : null}
                </CardContent>
              </Card>
            </div>
          ) : null}

          {activeTab === 'totals' ? (
            <div className="space-y-2.5">
              <Card className="border-border/70 bg-card">
                <CardHeader className="pb-2">
                  <CardTitle className="text-sm">Group totals</CardTitle>
                  <CardDescription>
                    Spend, your share, and where you stand.
                  </CardDescription>
                </CardHeader>
                <CardContent className="space-y-2.5">
                  {isTotalsLoading ? (
                    <p className="text-xs text-muted-foreground">Loading totals...</p>
                  ) : null}

                  <MiniBar
                    label="Total group spend"
                    value={Number(groupTotalSpent)}
                    max={totalsChartMax}
                    currencyCode={group?.currencyCode || 'USD'}
                  />
                  <MiniBar
                    label="Your paid"
                    value={Number(myBalance?.totalPaid || '0')}
                    max={totalsChartMax}
                    currencyCode={group?.currencyCode || 'USD'}
                  />
                  <MiniBar
                    label="Your share"
                    value={Number(myBalance?.totalOwed || '0')}
                    max={totalsChartMax}
                    currencyCode={group?.currencyCode || 'USD'}
                  />
                  <MiniBar
                    label="Your net"
                    value={Math.abs(Number(myBalance?.balance || '0'))}
                    max={totalsChartMax}
                    currencyCode={group?.currencyCode || 'USD'}
                  />

                  <div className="border-t border-border/70 pt-2" />

                  <MetricRow
                    label="Total group spend"
                    value={formatExpenseAmount(groupTotalSpent, group?.currencyCode || 'USD')}
                  />

                  <MetricRow
                    label="Your paid"
                    value={formatExpenseAmount(
                      myBalance?.totalPaid || '0',
                      group?.currencyCode || 'USD',
                    )}
                  />

                  <MetricRow
                    label="Your share"
                    value={formatExpenseAmount(
                      myBalance?.totalOwed || '0',
                      group?.currencyCode || 'USD',
                    )}
                  />

                  <div className="mt-2 border-t border-border/70 pt-2">
                    <MetricRow
                      label="Your net"
                      value={formatExpenseAmount(
                        myBalance?.balance || '0',
                        group?.currencyCode || 'USD',
                      )}
                      className={
                        Number(myBalance?.balance || '0') < 0
                          ? 'text-destructive'
                          : 'text-foreground'
                      }
                    />
                  </div>
                </CardContent>
              </Card>

              <Card className="border-border/70 bg-card">
                <CardHeader className="pb-2">
                  <CardTitle className="text-sm">Net balances</CardTitle>
                  <CardDescription>
                    Top creditors and debtors, including pending members.
                  </CardDescription>
                </CardHeader>
                <CardContent className="space-y-1.5">
                  {netBalancePreview.length === 0 ? (
                    <p className="text-xs text-muted-foreground">No balances yet.</p>
                  ) : null}
                  {netBalancePreview.map((balance) => (
                    <div
                      key={balance.key}
                      className="flex items-center justify-between text-xs"
                    >
                      <span className="text-muted-foreground">{balance.label}</span>
                      <span
                        className={cn(
                          'font-medium',
                          balance.numericBalance < 0 ? 'text-destructive' : '',
                        )}
                      >
                        {formatExpenseAmount(
                          balance.numericBalance.toFixed(2),
                          group?.currencyCode || 'USD',
                        )}
                      </span>
                    </div>
                  ))}
                </CardContent>
              </Card>
            </div>
          ) : null}

          {activeTab === 'activity' ? (
            <Card className="border-border/70 bg-card">
              <CardHeader className="pb-2">
                <CardTitle className="text-sm">Group activity</CardTitle>
                <CardDescription>
                  Meaningful updates from expenses and settlements.
                </CardDescription>
              </CardHeader>
              <CardContent className="space-y-1.5">
                {activities.length === 0 ? (
                  <p className="text-xs text-muted-foreground">No activity yet.</p>
                ) : null}

                {activities.map((activity) => {
                  const summary = summarizeActivity(activity, group?.currencyCode || 'USD')
                  return (
                    <div
                      key={activity.id}
                      className="rounded-[var(--radius)] border border-border/60 p-2.5"
                    >
                      <div className="flex items-center justify-between gap-2">
                        <p className="text-xs font-medium">{summary.title}</p>
                        <p className="text-[10px] text-muted-foreground">
                          {formatExpenseTime(activity.createdAt)}
                        </p>
                      </div>
                      <p className="mt-1 text-[11px] text-muted-foreground">
                        {summary.detail}
                      </p>
                      <p className="mt-1 text-[11px] text-muted-foreground">
                        By {activity.user.name || activity.user.email || 'Unknown'}
                      </p>
                    </div>
                  )
                })}
              </CardContent>
            </Card>
          ) : null}
        </>
      ) : null}
    </div>
  )
}

function TabButton({
  isActive,
  label,
  onClick,
}: {
  isActive: boolean
  label: string
  onClick: () => void
}) {
  return (
    <button
      type="button"
      onClick={onClick}
      className={cn(
        'rounded-[calc(var(--radius)-4px)] px-2 py-1.5 text-[10px] font-medium tracking-wide',
        isActive
          ? 'bg-background text-foreground shadow-sm'
          : 'text-muted-foreground hover:text-foreground',
      )}
    >
      {label}
    </button>
  )
}

function MetricRow({
  label,
  value,
  className,
}: {
  label: string
  value: string
  className?: string
}) {
  return (
    <div className="flex items-center justify-between text-xs">
      <span className="text-muted-foreground">{label}</span>
      <span className={cn('font-medium', className)}>{value}</span>
    </div>
  )
}

function MiniBar({
  label,
  value,
  max,
  currencyCode,
}: {
  label: string
  value: number
  max: number
  currencyCode: string
}) {
  const safeValue = Number.isFinite(value) ? value : 0
  const widthPercent = Math.max(4, Math.min(100, (safeValue / Math.max(max, 1)) * 100))
  return (
    <div className="space-y-1">
      <div className="flex items-center justify-between text-xs">
        <span className="text-muted-foreground">{label}</span>
        <span className="font-medium">
          {formatExpenseAmount(safeValue.toFixed(2), currencyCode)}
        </span>
      </div>
      <div className="h-2 rounded-full bg-muted/60">
        <div
          className="h-2 rounded-full bg-primary/70"
          style={{ width: `${widthPercent}%` }}
        />
      </div>
    </div>
  )
}

function summarizeActivity(activity: GroupActivity, fallbackCurrencyCode: string) {
  const action = activity.action.replace(/_/g, ' ').trim()
  const label = action
    ? action.charAt(0).toUpperCase() + action.slice(1)
    : 'Activity update'

  const metadata = activity.metadata
  const summary =
    metadata.summary && typeof metadata.summary === 'object'
      ? (metadata.summary as Record<string, unknown>)
      : null
  const before =
    metadata.before && typeof metadata.before === 'object'
      ? (metadata.before as Record<string, unknown>)
      : null
  const after =
    metadata.after && typeof metadata.after === 'object'
      ? (metadata.after as Record<string, unknown>)
      : null

  const summaryTitle = asText(summary?.title)
  const summaryAmount = asText(summary?.amount)
  const summaryCurrency = asText(summary?.currency_code) || fallbackCurrencyCode

  if (activity.action === 'expense_created' && summaryTitle && summaryAmount) {
    return {
      title: `Expense: ${summaryTitle}`,
      detail: formatExpenseAmount(summaryAmount, summaryCurrency),
    }
  }

  if (
    activity.action === 'expense_updated' &&
    before &&
    after &&
    asText(before.amount) &&
    asText(after.amount)
  ) {
    return {
      title: `Expense updated: ${asText(after.title) || asText(before.title) || 'Expense'}`,
      detail: `${formatExpenseAmount(
        asText(before.amount) || '0',
        asText(before.currency_code) || summaryCurrency,
      )} -> ${formatExpenseAmount(
        asText(after.amount) || '0',
        asText(after.currency_code) || summaryCurrency,
      )}`,
    }
  }

  if (activity.action === 'settlement_created' && asText(metadata.amount)) {
    return {
      title: 'Settlement recorded',
      detail: formatExpenseAmount(
        asText(metadata.amount) || '0',
        fallbackCurrencyCode,
      ),
    }
  }

  return {
    title: label,
    detail: `${activity.entityType} updated`,
  }
}

function asText(value: unknown) {
  return typeof value === 'string' ? value : ''
}

function formatExpenseTime(value: string) {
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return value
  return EXPENSE_TIME_FORMATTER.format(date)
}

function formatExpenseAmount(amount: string, currencyCode: string) {
  const numericAmount = Number(amount)
  if (!Number.isFinite(numericAmount)) return `${amount} ${currencyCode}`

  return new Intl.NumberFormat('en', {
    style: 'currency',
    currency: currencyCode,
    maximumFractionDigits: 2,
  }).format(numericAmount)
}
