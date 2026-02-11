import { useEffect, useState } from 'react'
import { Link, createFileRoute, useNavigate } from '@tanstack/react-router'
import { Plus } from 'lucide-react'

import type { OverallUserBalance } from '@/lib/api/balances'
import type { UserGroup } from '@/lib/api/groups'
import { Button } from '@/components/ui/button'
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Textarea } from '@/components/ui/textarea'
import { listOverallUserBalances } from '@/lib/api/balances'
import { createGroup, listUserGroups } from '@/lib/api/groups'

const CURRENCIES = [
  { code: 'USD', label: 'US Dollar' },
  { code: 'INR', label: 'Indian Rupee' }
]

export const Route = createFileRoute('/app/groups/')({ component: GroupsPage })

function GroupsPage() {
  const navigate = useNavigate()
  const [groups, setGroups] = useState<Array<UserGroup>>([])
  const [groupBalanceById, setGroupBalanceById] = useState<
    Partial<Record<string, OverallUserBalance>>
  >({})
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  const [showCreate, setShowCreate] = useState(false)
  const [isCreating, setIsCreating] = useState(false)
  const [createError, setCreateError] = useState<string | null>(null)
  const [name, setName] = useState('')
  const [description, setDescription] = useState('')
  const [currencyCode, setCurrencyCode] = useState('USD')
  const trimmedName = name.trim()

  useEffect(() => {
    let mounted = true

    async function loadGroups() {
      try {
        const [data, overallBalances] = await Promise.all([
          Promise.resolve(listUserGroups()),
          Promise.resolve(listOverallUserBalances()).catch(() => []),
        ])
        if (mounted) {
          setGroups(data)
          setGroupBalanceById(
            Object.fromEntries(
              overallBalances.map((row) => [row.groupId, row]),
            ) as Partial<Record<string, OverallUserBalance>>,
          )
        }
      } catch (err) {
        if (mounted)
          setError(
            err instanceof Error ? err.message : 'Unable to load your groups.',
          )
      } finally {
        if (mounted) setIsLoading(false)
      }
    }

    loadGroups()
    return () => {
      mounted = false
    }
  }, [])

  async function handleCreateGroup(e: React.FormEvent<HTMLFormElement>) {
    e.preventDefault()
    if (!trimmedName) return

    setCreateError(null)
    setIsCreating(true)

    try {
      const created = await createGroup({
        name: trimmedName,
        description: description.trim(),
        currencyCode,
      })
      setGroups((prev) => [
        {
          id: created.id,
          name: created.name,
          description: created.description,
          currencyCode: created.currencyCode,
          createdAt: created.createdAt,
          membershipId: '',
          memberRole: created.role,
          memberStatus: 'active',
          memberJoinedAt: created.createdAt,
        },
        ...prev,
      ])
      setName('')
      setDescription('')
      setCurrencyCode('USD')
      setShowCreate(false)
      navigate({ to: '/app/groups/$groupId', params: { groupId: created.id } })
    } catch (err) {
      setCreateError(
        err instanceof Error ? err.message : 'Unable to create group.',
      )
    } finally {
      setIsCreating(false)
    }
  }

  return (
    <div className="space-y-4">
      <header className="flex items-center justify-between">
        <div>
          <p className="text-sm text-muted-foreground">Groups</p>
          <h1 className="mt-1 text-xl font-semibold">Your circles</h1>
          <p className="mt-1 text-xs text-muted-foreground">
            {groups.length} {groups.length === 1 ? 'group' : 'groups'}
          </p>
        </div>
        <Button
          aria-label={showCreate ? 'Close create group form' : 'Create group'}
          onClick={() => {
            setShowCreate((prev) => !prev)
            setCreateError(null)
          }}
        >
          <Plus className="size-4" />
          {showCreate ? 'Close' : 'Create'}
        </Button>
      </header>

      {showCreate ? (
        <Card className="border-border/70 bg-card">
          <CardHeader>
            <CardTitle>Create group</CardTitle>
            <CardDescription>
              Add a name and invite your friends next.
            </CardDescription>
          </CardHeader>
          <CardContent>
            <form className="space-y-3" onSubmit={handleCreateGroup}>
              <Input
                required
                placeholder="Group name"
                value={name}
                onChange={(e) => setName(e.target.value)}
              />
              <Textarea
                placeholder="Description (optional)"
                value={description}
                onChange={(e) => setDescription(e.target.value)}
              />
              <label className="grid gap-1 text-xs text-muted-foreground">
                Currency
                <select
                  required
                  value={currencyCode}
                  onChange={(e) => setCurrencyCode(e.target.value)}
                  className="border-input bg-background text-foreground h-8 rounded-none border px-2 text-sm outline-none"
                >
                  {CURRENCIES.map((currency) => (
                    <option key={currency.code} value={currency.code}>
                      {currency.code} - {currency.label}
                    </option>
                  ))}
                </select>
              </label>
              {createError ? (
                <p className="text-xs text-destructive">{createError}</p>
              ) : null}
              <div className="flex gap-2">
                <Button
                  type="button"
                  variant="outline"
                  className="flex-1"
                  disabled={isCreating}
                  onClick={() => {
                    setShowCreate(false)
                    setCreateError(null)
                  }}
                >
                  Cancel
                </Button>
                <Button
                  type="submit"
                  disabled={isCreating || !trimmedName}
                  className="flex-1"
                >
                  {isCreating ? 'Creating...' : 'Create group'}
                </Button>
              </div>
            </form>
          </CardContent>
        </Card>
      ) : null}

      {isLoading ? (
        <Card className="border-border/70 bg-card">
          <CardContent className="py-6 text-sm text-muted-foreground">
            Loading groups...
          </CardContent>
        </Card>
      ) : null}

      {error ? (
        <Card className="border-border/70 bg-card">
          <CardContent className="py-6 text-sm text-destructive">
            {error}
          </CardContent>
        </Card>
      ) : null}

      <div className="space-y-3">
        {!isLoading && !error && groups.length === 0 ? (
          <Card className="border-border/70 bg-card">
            <CardContent className="py-6 text-sm text-muted-foreground">
              You are not in any groups yet. Create your first group.
            </CardContent>
          </Card>
        ) : null}

        {groups.map((group) => (
          <Link
            key={group.id}
            to="/app/groups/$groupId"
            params={{ groupId: group.id }}
            className="block"
          >
            <Card className="border-border/70 bg-card transition-colors hover:bg-muted/30">
              <CardHeader>
                <CardTitle>{group.name}</CardTitle>
                <CardDescription>
                  {group.description || 'No description yet'}
                </CardDescription>
              </CardHeader>
              <CardContent className="flex items-center justify-between">
                <div className="space-y-0.5">
                  <span className="block text-xs text-muted-foreground">
                    Role: {group.memberRole}
                  </span>
                  <span className="block text-[11px] text-muted-foreground">
                    {group.currencyCode}
                  </span>
                </div>
                <GroupNetAmount
                  amount={groupBalanceById[group.id]?.balance || '0'}
                  currencyCode={group.currencyCode}
                />
              </CardContent>
            </Card>
          </Link>
        ))}
      </div>
    </div>
  )
}

function GroupNetAmount({
  amount,
  currencyCode,
}: {
  amount: string
  currencyCode: string
}) {
  const numericAmount = Number(amount)
  if (!Number.isFinite(numericAmount)) {
    return <span className="text-sm font-semibold">{amount}</span>
  }

  const absolute = Math.abs(numericAmount)
  const formatted = new Intl.NumberFormat('en', {
    style: 'currency',
    currency: currencyCode,
    maximumFractionDigits: 2,
  }).format(absolute)

  if (numericAmount > 0) {
    return (
      <span className="text-sm font-semibold text-green-600">
        + {formatted}
      </span>
    )
  }

  if (numericAmount < 0) {
    return (
      <span className="text-sm font-semibold text-destructive">
        - {formatted}
      </span>
    )
  }

  return (
    <span className="text-sm font-semibold text-muted-foreground">
      {formatted}
    </span>
  )
}
