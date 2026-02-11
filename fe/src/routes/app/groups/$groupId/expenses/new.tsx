import { useEffect, useMemo, useState } from 'react'
import { Link, createFileRoute, useNavigate } from '@tanstack/react-router'
import { ArrowLeft } from 'lucide-react'

import type { GroupMember, UserGroup } from '@/lib/api/groups'
import { invalidateGroupActivityCache } from '@/lib/api/activities'
import {
  invalidateGroupBalancesCache,
  invalidateGroupDebtsCache,
} from '@/lib/api/balances'
import {
  createGroupExpense,
  invalidateGroupExpensesCache,
} from '@/lib/api/expenses'
import { listGroupMembers, listUserGroups } from '@/lib/api/groups'
import { invalidateGroupSettlementsCache } from '@/lib/api/settlements'
import { Button, buttonVariants } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Textarea } from '@/components/ui/textarea'
import { cn } from '@/lib/utils'

type SplitMode = 'equal' | 'fixed' | 'percentage' | 'shares'

export const Route = createFileRoute('/app/groups/$groupId/expenses/new')({
  component: NewExpensePage,
})

function NewExpensePage() {
  const { groupId } = Route.useParams()
  const navigate = useNavigate()

  const [group, setGroup] = useState<UserGroup | null>(null)
  const [members, setMembers] = useState<Array<GroupMember>>([])
  const [isLoading, setIsLoading] = useState(true)
  const [loadError, setLoadError] = useState<string | null>(null)

  const [title, setTitle] = useState('')
  const [notes, setNotes] = useState('')
  const [amount, setAmount] = useState('')
  const [date, setDate] = useState(defaultDateValue())
  const [payerId, setPayerId] = useState('')
  const [splitMode, setSplitMode] = useState<SplitMode>('equal')
  const [selectedMemberIds, setSelectedMemberIds] = useState<Array<string>>([])
  const [fixedAmounts, setFixedAmounts] = useState<Record<string, string>>({})
  const [percentageAmounts, setPercentageAmounts] = useState<
    Record<string, string>
  >({})
  const [shareValues, setShareValues] = useState<Record<string, string>>({})
  const [submitError, setSubmitError] = useState<string | null>(null)
  const [isSubmitting, setIsSubmitting] = useState(false)

  const participantMembers = useMemo(
    () =>
      members.filter(
        (member) =>
          (member.status === 'active' || member.status === 'pending') &&
          member.userId &&
          (member.user.name || member.user.email),
      ),
    [members],
  )

  const selectedMembers = useMemo(
    () =>
      participantMembers.filter((member) =>
        selectedMemberIds.includes(member.userId),
      ),
    [participantMembers, selectedMemberIds],
  )

  const equalSharePreview = useMemo(() => {
    const totalCents = parseAmountToCents(amount)
    if (
      totalCents === null ||
      totalCents <= 0 ||
      selectedMembers.length === 0
    ) {
      return null
    }
    const shares = splitEvenly(totalCents, selectedMembers.length)
    return centsToAmount(shares[0] || 0)
  }, [amount, selectedMembers.length])

  useEffect(() => {
    let mounted = true
    setIsLoading(true)
    setLoadError(null)

    async function loadData() {
      try {
        const [groupList, groupMembers] = await Promise.all([
          listUserGroups(),
          listGroupMembers(groupId, { force: true }),
        ])
        if (!mounted) return

        const foundGroup =
          groupList.find((entry) => entry.id === groupId) || null
        const participantGroupMembers = groupMembers.filter(
          (member) =>
            (member.status === 'active' || member.status === 'pending') &&
            member.userId &&
            (member.user.name || member.user.email),
        )
        const activeGroupMembers = groupMembers.filter(
          (member) =>
            member.status === 'active' &&
            member.userId &&
            (member.user.name || member.user.email),
        )

        setGroup(foundGroup)
        setMembers(groupMembers)
        setSelectedMemberIds(participantGroupMembers.map((member) => member.userId))
        setPayerId(
          activeGroupMembers[0]?.userId || participantGroupMembers[0]?.userId || '',
        )
      } catch (err) {
        if (!mounted) return
        setLoadError(
          err instanceof Error
            ? err.message
            : 'Unable to load expense form data.',
        )
      } finally {
        if (mounted) setIsLoading(false)
      }
    }

    loadData()
    return () => {
      mounted = false
    }
  }, [groupId])

  useEffect(() => {
    const participantIds = new Set(
      participantMembers.map((member) => member.userId),
    )

    setFixedAmounts((prev) => filterByIds(prev, participantIds))
    setPercentageAmounts((prev) => filterByIds(prev, participantIds))
    setShareValues((prev) => filterByIds(prev, participantIds))

    if (payerId && participantMembers.some((member) => member.userId === payerId))
      return
    setPayerId(participantMembers[0]?.userId || '')
  }, [participantMembers, payerId])

  function toggleMember(memberId: string) {
    setSelectedMemberIds((prev) => {
      if (prev.includes(memberId)) {
        if (prev.length === 1) return prev
        return prev.filter((id) => id !== memberId)
      }
      return [...prev, memberId]
    })
  }

  async function handleSubmit(e: React.FormEvent<HTMLFormElement>) {
    e.preventDefault()
    setSubmitError(null)

    const trimmedTitle = title.trim()
    if (!trimmedTitle) {
      setSubmitError('Please enter a title.')
      return
    }

    if (!/^\d{4}-\d{2}-\d{2}$/.test(date)) {
      setSubmitError('Please enter a valid date in YYYY-MM-DD format.')
      return
    }

    const totalCents = parseAmountToCents(amount)
    if (totalCents === null || totalCents <= 0) {
      setSubmitError('Please enter an amount greater than 0.')
      return
    }

    if (!payerId) {
      setSubmitError('Please select who paid.')
      return
    }

    const payerMember = participantMembers.find((member) => member.userId === payerId)
    if (!payerMember) {
      setSubmitError('Selected payer is no longer available.')
      return
    }

    if (selectedMembers.length === 0) {
      setSubmitError('Select at least one participant.')
      return
    }

    const splitConfig = buildSplitConfig({
      splitMode,
      selectedMembers,
      fixedAmounts,
      percentageAmounts,
      shareValues,
      totalCents,
    })

    if (!splitConfig.ok) {
      setSubmitError(splitConfig.error)
      return
    }

    setIsSubmitting(true)
    try {
      const created = await createGroupExpense(groupId, {
        title: trimmedTitle,
        notes: notes.trim(),
        amount: centsToAmount(totalCents),
        currencyCode: group?.currencyCode || 'USD',
        date,
        payments: [
          payerMember.status === 'pending'
            ? {
                pendingUserId: payerId,
                amount: centsToAmount(totalCents),
              }
            : {
                userId: payerId,
                amount: centsToAmount(totalCents),
              },
        ],
        splits: splitConfig.splits,
      })

      invalidateGroupExpensesCache(groupId)
      invalidateGroupSettlementsCache(groupId)
      invalidateGroupBalancesCache(groupId)
      invalidateGroupDebtsCache(groupId)
      invalidateGroupActivityCache(groupId)

      navigate({
        to: '/app/groups/$groupId/expenses/$expenseId',
        params: { groupId, expenseId: created.expense.id },
      })
    } catch (err) {
      setSubmitError(
        err instanceof Error ? err.message : 'Unable to create expense.',
      )
    } finally {
      setIsSubmitting(false)
    }
  }

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
          <h1 className="text-xl font-semibold">Add expense</h1>
        </div>
        <p className="text-sm text-muted-foreground">
          {group?.name || 'Group expense'}
        </p>
      </header>

      <Card className="border-border/70 bg-card">
        <CardHeader>
          <CardTitle>Expense details</CardTitle>
        </CardHeader>
        <CardContent className="space-y-3">
          {isLoading ? (
            <p className="text-xs text-muted-foreground">
              Loading form data...
            </p>
          ) : null}
          {loadError ? (
            <p className="text-xs text-destructive">{loadError}</p>
          ) : null}

          {!isLoading && !loadError ? (
            <form className="space-y-3" onSubmit={handleSubmit}>
              <Input
                required
                placeholder="Title"
                value={title}
                onChange={(e) => setTitle(e.target.value)}
              />

              <Input
                required
                type="number"
                min="0.01"
                step="0.01"
                placeholder={`Amount (${group?.currencyCode || 'USD'})`}
                value={amount}
                onChange={(e) => setAmount(e.target.value)}
              />

              <Input
                required
                type="date"
                value={date}
                onChange={(e) => setDate(e.target.value)}
              />

              <Textarea
                placeholder="Notes (optional)"
                value={notes}
                onChange={(e) => setNotes(e.target.value)}
              />

              <label className="grid gap-1 text-xs text-muted-foreground">
                Paid by
                <select
                  required
                  value={payerId}
                  onChange={(e) => setPayerId(e.target.value)}
                  className="border-input bg-background text-foreground h-8 rounded-none border px-2 text-sm outline-none"
                >
                  {participantMembers.map((member) => (
                    <option key={member.userId} value={member.userId}>
                      {member.user.name || member.user.email}
                      {member.status === 'pending' ? ' (pending)' : ''}
                    </option>
                  ))}
                </select>
              </label>

              <fieldset className="space-y-2">
                <legend className="text-xs text-muted-foreground">
                  Split mode
                </legend>
                <div className="grid grid-cols-2 gap-2">
                  <Button
                    type="button"
                    variant={splitMode === 'equal' ? 'default' : 'outline'}
                    onClick={() => setSplitMode('equal')}
                  >
                    Equal
                  </Button>
                  <Button
                    type="button"
                    variant={splitMode === 'fixed' ? 'default' : 'outline'}
                    onClick={() => setSplitMode('fixed')}
                  >
                    Fixed
                  </Button>
                  <Button
                    type="button"
                    variant={splitMode === 'percentage' ? 'default' : 'outline'}
                    onClick={() => setSplitMode('percentage')}
                  >
                    Percentage
                  </Button>
                  <Button
                    type="button"
                    variant={splitMode === 'shares' ? 'default' : 'outline'}
                    onClick={() => setSplitMode('shares')}
                  >
                    Shares
                  </Button>
                </div>
              </fieldset>

              <fieldset className="space-y-2">
                <legend className="text-xs text-muted-foreground">
                  Participants
                </legend>
                <div className="space-y-2 rounded-[var(--radius)] border border-border/70 p-2">
                  {participantMembers.map((member) => (
                    <label
                      key={member.userId}
                      className="flex items-center justify-between gap-2 text-xs"
                    >
                      <span className="inline-flex items-center gap-2">
                        <input
                          type="checkbox"
                          checked={selectedMemberIds.includes(member.userId)}
                          onChange={() => toggleMember(member.userId)}
                        />
                        {member.user.name || member.user.email}
                        {member.status === 'pending' ? (
                          <span className="text-[10px] text-muted-foreground">
                            (pending)
                          </span>
                        ) : null}
                      </span>

                      {selectedMemberIds.includes(member.userId) ? (
                        <>
                          {splitMode === 'fixed' ? (
                            <Input
                              type="number"
                              min="0.01"
                              step="0.01"
                              placeholder="0.00"
                              value={fixedAmounts[member.userId] || ''}
                              onChange={(e) =>
                                setFixedAmounts((prev) => ({
                                  ...prev,
                                  [member.userId]: e.target.value,
                                }))
                              }
                              className="h-7 w-28"
                            />
                          ) : null}

                          {splitMode === 'percentage' ? (
                            <Input
                              type="number"
                              min="0"
                              step="0.01"
                              placeholder="%"
                              value={percentageAmounts[member.userId] || ''}
                              onChange={(e) =>
                                setPercentageAmounts((prev) => ({
                                  ...prev,
                                  [member.userId]: e.target.value,
                                }))
                              }
                              className="h-7 w-28"
                            />
                          ) : null}

                          {splitMode === 'shares' ? (
                            <Input
                              type="number"
                              min="1"
                              step="1"
                              placeholder="shares"
                              value={shareValues[member.userId] || ''}
                              onChange={(e) =>
                                setShareValues((prev) => ({
                                  ...prev,
                                  [member.userId]: e.target.value,
                                }))
                              }
                              className="h-7 w-28"
                            />
                          ) : null}
                        </>
                      ) : null}
                    </label>
                  ))}
                </div>

                {splitMode === 'equal' && equalSharePreview ? (
                  <p className="text-[11px] text-muted-foreground">
                    Approx. {group?.currencyCode || 'USD'} {equalSharePreview}{' '}
                    per participant
                  </p>
                ) : null}
              </fieldset>

              {submitError ? (
                <p className="text-xs text-destructive">{submitError}</p>
              ) : null}

              <Button
                type="submit"
                className="h-10 w-full text-sm"
                disabled={isSubmitting || participantMembers.length === 0}
              >
                {isSubmitting ? 'Saving expense...' : 'Save expense'}
              </Button>
            </form>
          ) : null}
        </CardContent>
      </Card>
    </div>
  )
}

function buildSplitConfig({
  splitMode,
  selectedMembers,
  fixedAmounts,
  percentageAmounts,
  shareValues,
  totalCents,
}: {
  splitMode: SplitMode
  selectedMembers: Array<GroupMember>
  fixedAmounts: Record<string, string>
  percentageAmounts: Record<string, string>
  shareValues: Record<string, string>
  totalCents: number
}) {
  if (splitMode === 'equal') {
    return {
      ok: true as const,
      splits: selectedMembers.map((member) => ({
        ...(member.status === 'pending'
          ? { pendingUserId: member.userId }
          : { userId: member.userId }),
        type: 'equal' as const,
      })),
    }
  }

  if (splitMode === 'fixed') {
    let total = 0
    const splits: Array<{
      userId?: string
      pendingUserId?: string
      type: 'fixed'
      amount: string
    }> = []

    for (const member of selectedMembers) {
      const cents = parseAmountToCents(fixedAmounts[member.userId] || '')
      if (cents === null || cents <= 0) {
        return {
          ok: false as const,
          error: 'Enter a valid fixed amount for each selected participant.',
        }
      }
      total += cents
      splits.push({
        ...(member.status === 'pending'
          ? { pendingUserId: member.userId }
          : { userId: member.userId }),
        type: 'fixed',
        amount: centsToAmount(cents),
      })
    }

    if (total !== totalCents) {
      return {
        ok: false as const,
        error: 'Fixed split amounts must add up to total expense amount.',
      }
    }

    return { ok: true as const, splits }
  }

  if (splitMode === 'percentage') {
    let totalPercent = 0
    const splits: Array<{
      userId?: string
      pendingUserId?: string
      type: 'percentage'
      percentage: string
    }> = []

    for (const member of selectedMembers) {
      const value = Number(percentageAmounts[member.userId] || '')
      if (!Number.isFinite(value) || value <= 0) {
        return {
          ok: false as const,
          error: 'Enter a valid percentage for each selected participant.',
        }
      }
      totalPercent += value
      splits.push({
        ...(member.status === 'pending'
          ? { pendingUserId: member.userId }
          : { userId: member.userId }),
        type: 'percentage',
        percentage: value.toFixed(2),
      })
    }

    if (Math.abs(totalPercent - 100) > 0.01) {
      return {
        ok: false as const,
        error: 'Percentages must add up to 100.',
      }
    }

    return { ok: true as const, splits }
  }

  let totalShares = 0
  const splits: Array<{
    userId?: string
    pendingUserId?: string
    type: 'shares'
    shares: number
  }> = []

  for (const member of selectedMembers) {
    const value = Number(shareValues[member.userId] || '')
    if (!Number.isInteger(value) || value <= 0) {
      return {
        ok: false as const,
        error:
          'Enter a valid whole number of shares for each selected participant.',
      }
    }
    totalShares += value
    splits.push({
      ...(member.status === 'pending'
        ? { pendingUserId: member.userId }
        : { userId: member.userId }),
      type: 'shares',
      shares: value,
    })
  }

  if (totalShares <= 0) {
    return {
      ok: false as const,
      error: 'Shares must be greater than 0.',
    }
  }

  return { ok: true as const, splits }
}

function filterByIds(
  source: Record<string, string>,
  ids: Set<string>,
): Record<string, string> {
  const next: Record<string, string> = {}
  for (const [key, value] of Object.entries(source)) {
    if (ids.has(key)) next[key] = value
  }
  return next
}

function parseAmountToCents(value: string) {
  const numeric = Number(value)
  if (!Number.isFinite(numeric)) return null
  return Math.round(numeric * 100)
}

function centsToAmount(cents: number) {
  return (cents / 100).toFixed(2)
}

function splitEvenly(totalCents: number, size: number) {
  if (size <= 0) return []
  const base = Math.floor(totalCents / size)
  const remainder = totalCents % size
  const shares = Array.from({ length: size }, () => base)
  for (let index = 0; index < remainder; index += 1) {
    shares[index] += 1
  }
  return shares
}

function defaultDateValue() {
  return new Date().toISOString().slice(0, 10)
}
