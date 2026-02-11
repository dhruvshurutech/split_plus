import { useEffect, useMemo, useState } from 'react'
import { Link, createFileRoute, useNavigate } from '@tanstack/react-router'
import { ArrowLeft } from 'lucide-react'

import type { GroupMember } from '@/lib/api/groups'
import { invalidateGroupActivityCache } from '@/lib/api/activities'
import {
  invalidateGroupBalancesCache,
  invalidateGroupDebtsCache,
} from '@/lib/api/balances'
import { listGroupMembers } from '@/lib/api/groups'
import { invalidateGroupExpensesCache } from '@/lib/api/expenses'
import {
  createGroupSettlement,
  invalidateGroupSettlementsCache,
} from '@/lib/api/settlements'
import { Button, buttonVariants } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Textarea } from '@/components/ui/textarea'
import { cn } from '@/lib/utils'

export const Route = createFileRoute('/app/groups/$groupId/settlements/new')({
  component: NewSettlementPage,
})

function NewSettlementPage() {
  const { groupId } = Route.useParams()
  const navigate = useNavigate()
  const [members, setMembers] = useState<Array<GroupMember>>([])
  const [isLoading, setIsLoading] = useState(true)
  const [loadError, setLoadError] = useState<string | null>(null)

  const [payerId, setPayerId] = useState('')
  const [payeeId, setPayeeId] = useState('')
  const [amount, setAmount] = useState('')
  const [status, setStatus] = useState<'pending' | 'completed' | 'cancelled'>(
    'pending',
  )
  const [paymentMethod, setPaymentMethod] = useState('')
  const [transactionReference, setTransactionReference] = useState('')
  const [notes, setNotes] = useState('')
  const [submitError, setSubmitError] = useState<string | null>(null)
  const [isSubmitting, setIsSubmitting] = useState(false)
  const prefill = useMemo(() => {
    if (typeof window === 'undefined') {
      return { payerId: '', payeeId: '', amount: '' }
    }
    const params = new URLSearchParams(window.location.search)
    return {
      payerId: params.get('payerId') || '',
      payeeId: params.get('payeeId') || '',
      amount: params.get('amount') || '',
    }
  }, [])

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

  useEffect(() => {
    let mounted = true
    setIsLoading(true)
    setLoadError(null)

    async function loadMembers() {
      try {
        const rows = await listGroupMembers(groupId, { force: true })
        if (!mounted) return
        const participants = rows.filter(
          (member) =>
            (member.status === 'active' || member.status === 'pending') &&
            member.userId &&
            (member.user.name || member.user.email),
        )
        setMembers(rows)
        const participantIDs = new Set(
          participants.map((member) => member.userId),
        )
        const prefillPayerId = participantIDs.has(prefill.payerId)
          ? prefill.payerId
          : ''
        const prefillPayeeId = participantIDs.has(prefill.payeeId)
          ? prefill.payeeId
          : ''
        setPayerId(prefillPayerId || participants[0]?.userId || '')
        setPayeeId(
          prefillPayeeId || participants[1]?.userId || participants[0]?.userId || '',
        )
        if (prefill.amount) setAmount(prefill.amount)
      } catch (err) {
        if (!mounted) return
        setLoadError(
          err instanceof Error ? err.message : 'Unable to load group members.',
        )
      } finally {
        if (mounted) setIsLoading(false)
      }
    }

    loadMembers()
    return () => {
      mounted = false
    }
  }, [groupId, prefill.amount, prefill.payeeId, prefill.payerId])

  async function handleSubmit(e: React.FormEvent<HTMLFormElement>) {
    e.preventDefault()
    setSubmitError(null)

    const totalCents = parseAmountToCents(amount)
    if (totalCents === null || totalCents <= 0) {
      setSubmitError('Please enter an amount greater than 0.')
      return
    }

    if (!payerId || !payeeId) {
      setSubmitError('Please select both payer and payee.')
      return
    }

    if (payerId === payeeId) {
      setSubmitError('Payer and payee must be different members.')
      return
    }

    const payerMember = participantMembers.find((member) => member.userId === payerId)
    const payeeMember = participantMembers.find((member) => member.userId === payeeId)
    if (!payerMember || !payeeMember) {
      setSubmitError('Selected members are invalid for this group.')
      return
    }

    setIsSubmitting(true)
    try {
      await createGroupSettlement(groupId, {
        ...(payerMember.status === 'pending'
          ? { payerPendingUserId: payerId }
          : { payerId }),
        ...(payeeMember.status === 'pending'
          ? { payeePendingUserId: payeeId }
          : { payeeId }),
        amount: centsToAmount(totalCents),
        status,
        paymentMethod: paymentMethod.trim(),
        transactionReference: transactionReference.trim(),
        notes: notes.trim(),
      })

      invalidateGroupSettlementsCache(groupId)
      invalidateGroupBalancesCache(groupId)
      invalidateGroupDebtsCache(groupId)
      invalidateGroupActivityCache(groupId)
      invalidateGroupExpensesCache(groupId)
      navigate({ to: '/app/groups/$groupId', params: { groupId } })
    } catch (err) {
      setSubmitError(
        err instanceof Error ? err.message : 'Unable to create settlement.',
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
          <h1 className="text-xl font-semibold">Record settlement</h1>
        </div>
        <p className="text-sm text-muted-foreground">
          Settle dues in this group
        </p>
      </header>

      <Card className="border-border/70 bg-card">
        <CardHeader>
          <CardTitle>Settlement details</CardTitle>
        </CardHeader>
        <CardContent className="space-y-3">
          {isLoading ? (
            <p className="text-xs text-muted-foreground">Loading members...</p>
          ) : null}
          {loadError ? (
            <p className="text-xs text-destructive">{loadError}</p>
          ) : null}

          {!isLoading && !loadError ? (
            <form className="space-y-3" onSubmit={handleSubmit}>
              <label className="grid gap-1 text-xs text-muted-foreground">
                Payer
                <select
                  required
                  value={payerId}
                  onChange={(e) => setPayerId(e.target.value)}
                  className="border-input bg-background text-foreground h-8 rounded-none border px-2 text-sm outline-none"
                >
                  {participantMembers.map((member) => (
                    <option key={member.userId} value={member.userId}>
                      {member.user.name || member.user.email}
                    </option>
                  ))}
                </select>
              </label>

              <label className="grid gap-1 text-xs text-muted-foreground">
                Payee
                <select
                  required
                  value={payeeId}
                  onChange={(e) => setPayeeId(e.target.value)}
                  className="border-input bg-background text-foreground h-8 rounded-none border px-2 text-sm outline-none"
                >
                  {participantMembers.map((member) => (
                    <option key={member.userId} value={member.userId}>
                      {member.user.name || member.user.email}
                    </option>
                  ))}
                </select>
              </label>

              <Input
                required
                type="number"
                min="0.01"
                step="0.01"
                placeholder="Amount"
                value={amount}
                onChange={(e) => setAmount(e.target.value)}
              />

              <label className="grid gap-1 text-xs text-muted-foreground">
                Status
                <select
                  required
                  value={status}
                  onChange={(e) =>
                    setStatus(
                      e.target.value as 'pending' | 'completed' | 'cancelled',
                    )
                  }
                  className="border-input bg-background text-foreground h-8 rounded-none border px-2 text-sm outline-none"
                >
                  <option value="pending">pending</option>
                  <option value="completed">completed</option>
                  <option value="cancelled">cancelled</option>
                </select>
              </label>

              <Input
                placeholder="Payment method (optional)"
                value={paymentMethod}
                onChange={(e) => setPaymentMethod(e.target.value)}
              />

              <Input
                placeholder="Transaction reference (optional)"
                value={transactionReference}
                onChange={(e) => setTransactionReference(e.target.value)}
              />

              <Textarea
                placeholder="Notes (optional)"
                value={notes}
                onChange={(e) => setNotes(e.target.value)}
              />

              {submitError ? (
                <p className="text-xs text-destructive">{submitError}</p>
              ) : null}

              <Button
                type="submit"
                className="h-10 w-full text-sm"
                disabled={isSubmitting || participantMembers.length < 2}
              >
                {isSubmitting ? 'Saving settlement...' : 'Save settlement'}
              </Button>
            </form>
          ) : null}
        </CardContent>
      </Card>
    </div>
  )
}

function parseAmountToCents(value: string) {
  const numeric = Number(value)
  if (!Number.isFinite(numeric)) return null
  return Math.round(numeric * 100)
}

function centsToAmount(cents: number) {
  return (cents / 100).toFixed(2)
}
