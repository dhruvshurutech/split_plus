import { apiRequest } from '@/lib/api/client'

type RawUUID = {
  Bytes?: Array<number>
  Valid?: boolean
}

type RawExpense = {
  id: string | RawUUID
  group_id: string | RawUUID
  title: string
  notes?: string
  amount: string
  currency_code: string
  date: string
  category_id?: string | RawUUID
  tags?: Array<string>
  created_at: string
  created_by: string | RawUUID
  updated_at: string
  updated_by: string | RawUUID
}

type RawPayment = {
  id: string | RawUUID
  expense_id: string | RawUUID
  user_id: string | RawUUID
  pending_user_id?: string | RawUUID
  amount: string
  payment_method?: string
  created_at: string
  user?: {
    email: string
    name?: string
    avatar_url?: string
  }
  pending_user?: {
    email: string
    name?: string
    avatar_url?: string
  }
}

type RawGroupExpenseResponse = {
  expense: RawExpense
  payments: Array<RawPayment>
}

type RawSplit = {
  id: string | RawUUID
  expense_id: string | RawUUID
  user_id: string | RawUUID
  pending_user_id?: string | RawUUID
  amount_owned: string
  split_type: string
  share_value?: string
  created_at: string
  user?: {
    email: string
    name?: string
    avatar_url?: string
  }
  pending_user?: {
    email: string
    name?: string
    avatar_url?: string
  }
}

type RawCreateExpenseResponse = {
  expense: RawExpense
  payments: Array<RawPayment>
  splits: Array<RawSplit>
}

const RECENT_GROUP_EXPENSES_CACHE_TTL_MS = 12_000
const GROUP_EXPENSE_DETAIL_CACHE_TTL_MS = 15_000

export type GroupExpense = {
  expense: {
    id: string
    groupId: string
    title: string
    notes: string
    amount: string
    currencyCode: string
    date: string
    categoryId: string
    tags: Array<string>
    createdAt: string
    createdBy: string
    updatedAt: string
    updatedBy: string
  }
  payments: Array<{
    id: string
    expenseId: string
    userId: string
    pendingUserId: string
    amount: string
    paymentMethod: string
    createdAt: string
    user: {
      email: string
      name: string
      avatarUrl: string
    }
    pendingUser?: {
      email: string
      name: string
      avatarUrl: string
    }
  }>
}

export type GroupExpenseSplit = {
  id: string
  expenseId: string
  userId: string
  pendingUserId: string
  amountOwned: string
  splitType: string
  shareValue: string
  createdAt: string
  user: {
    email: string
    name: string
    avatarUrl: string
  }
  pendingUser?: {
    email: string
    name: string
    avatarUrl: string
  }
}

export type GroupExpenseDetail = {
  expense: GroupExpense['expense']
  payments: GroupExpense['payments']
  splits: Array<GroupExpenseSplit>
}

export type CreatedGroupExpense = GroupExpenseDetail

export type CreateGroupExpenseInput = {
  title: string
  notes?: string
  amount: string
  currencyCode?: string
  date: string
  categoryId?: string
  tags?: Array<string>
  payments: Array<{
    userId?: string
    pendingUserId?: string
    amount: string
    paymentMethod?: string
  }>
  splits: Array<{
    userId?: string
    pendingUserId?: string
    type: 'equal' | 'percentage' | 'shares' | 'fixed' | 'custom'
    percentage?: string
    shares?: number
    amount?: string
  }>
}

const recentGroupExpensesCache = new Map<
  string,
  { value: Array<GroupExpense>; expiresAt: number }
>()
const recentGroupExpensesInFlight = new Map<
  string,
  Promise<Array<GroupExpense>>
>()

const groupExpenseDetailCache = new Map<
  string,
  { value: GroupExpenseDetail; expiresAt: number }
>()
const groupExpenseDetailInFlight = new Map<
  string,
  Promise<GroupExpenseDetail>
>()

function getRecentExpensesCacheKey(
  groupId: string,
  limit: number,
  offset: number,
) {
  return `${groupId}:${limit}:${offset}`
}

function getExpenseDetailCacheKey(groupId: string, expenseId: string) {
  return `${groupId}:${expenseId}`
}

function formatUUIDFromBytes(bytes: Array<number>) {
  const hex = bytes
    .map((value) => value.toString(16).padStart(2, '0'))
    .join('')
    .toLowerCase()

  return `${hex.slice(0, 8)}-${hex.slice(8, 12)}-${hex.slice(
    12,
    16,
  )}-${hex.slice(16, 20)}-${hex.slice(20, 32)}`
}

function parseUUID(value: string | RawUUID | undefined) {
  if (!value) return ''
  if (typeof value === 'string') return value
  if (!value.Valid) return ''
  if (!Array.isArray(value.Bytes) || value.Bytes.length !== 16) return ''
  return formatUUIDFromBytes(value.Bytes)
}

function mapGroupExpense(raw: RawGroupExpenseResponse): GroupExpense {
  return {
    expense: {
      id: parseUUID(raw.expense.id),
      groupId: parseUUID(raw.expense.group_id),
      title: raw.expense.title,
      notes: raw.expense.notes || '',
      amount: raw.expense.amount,
      currencyCode: raw.expense.currency_code,
      date: raw.expense.date,
      categoryId: parseUUID(raw.expense.category_id),
      tags: raw.expense.tags || [],
      createdAt: raw.expense.created_at,
      createdBy: parseUUID(raw.expense.created_by),
      updatedAt: raw.expense.updated_at,
      updatedBy: parseUUID(raw.expense.updated_by),
    },
    payments: raw.payments.map((payment) => ({
      id: parseUUID(payment.id),
      expenseId: parseUUID(payment.expense_id),
      userId: parseUUID(payment.user_id),
      pendingUserId: parseUUID(payment.pending_user_id),
      amount: payment.amount,
      paymentMethod: payment.payment_method || '',
      createdAt: payment.created_at,
      user: {
        email: payment.user?.email || '',
        name: payment.user?.name || '',
        avatarUrl: payment.user?.avatar_url || '',
      },
      pendingUser: payment.pending_user
        ? {
            email: payment.pending_user.email || '',
            name: payment.pending_user.name || '',
            avatarUrl: payment.pending_user.avatar_url || '',
          }
        : undefined,
    })),
  }
}

function mapSplit(raw: RawSplit): GroupExpenseSplit {
  return {
    id: parseUUID(raw.id),
    expenseId: parseUUID(raw.expense_id),
    userId: parseUUID(raw.user_id),
    pendingUserId: parseUUID(raw.pending_user_id),
    amountOwned: raw.amount_owned,
    splitType: raw.split_type,
    shareValue: raw.share_value || '',
    createdAt: raw.created_at,
    user: {
      email: raw.user?.email || '',
      name: raw.user?.name || '',
      avatarUrl: raw.user?.avatar_url || '',
    },
    pendingUser: raw.pending_user
      ? {
          email: raw.pending_user.email || '',
          name: raw.pending_user.name || '',
          avatarUrl: raw.pending_user.avatar_url || '',
        }
      : undefined,
  }
}

function mapGroupExpenseDetail(
  raw: RawCreateExpenseResponse,
): GroupExpenseDetail {
  const base = mapGroupExpense({
    expense: raw.expense,
    payments: raw.payments,
  })

  return {
    expense: base.expense,
    payments: base.payments,
    splits: raw.splits.map(mapSplit),
  }
}

export async function listRecentGroupExpenses(
  groupId: string,
  options?: { limit?: number; offset?: number; force?: boolean },
) {
  const limit = options?.limit ?? 10
  const offset = options?.offset ?? 0
  const force = options?.force === true
  const key = getRecentExpensesCacheKey(groupId, limit, offset)
  const now = Date.now()
  const cached = recentGroupExpensesCache.get(key)

  if (!force && cached && cached.expiresAt > now) return cached.value

  if (!force) {
    const inFlight = recentGroupExpensesInFlight.get(key)
    if (inFlight) return inFlight
  }

  const query = new URLSearchParams({
    limit: String(limit),
    offset: String(offset),
  })

  const request = apiRequest<Array<RawGroupExpenseResponse>>(
    `/groups/${groupId}/expenses/search?${query.toString()}`,
    {
      method: 'GET',
      auth: true,
    },
  )
    .then((rows) => {
      const mapped = rows.map(mapGroupExpense)
      recentGroupExpensesCache.set(key, {
        value: mapped,
        expiresAt: Date.now() + RECENT_GROUP_EXPENSES_CACHE_TTL_MS,
      })
      return mapped
    })
    .finally(() => {
      recentGroupExpensesInFlight.delete(key)
    })

  recentGroupExpensesInFlight.set(key, request)
  return request
}

export async function createGroupExpense(
  groupId: string,
  input: CreateGroupExpenseInput,
) {
  const created = await apiRequest<RawCreateExpenseResponse>(
    `/groups/${groupId}/expenses`,
    {
      method: 'POST',
      auth: true,
      body: JSON.stringify({
        title: input.title,
        notes: input.notes || '',
        amount: input.amount,
        currency_code: input.currencyCode || 'USD',
        date: input.date,
        category_id: input.categoryId || '',
        tags: input.tags || [],
        payments: input.payments.map((payment) => ({
          user_id: payment.userId || '',
          pending_user_id: payment.pendingUserId || '',
          amount: payment.amount,
          payment_method: payment.paymentMethod || '',
        })),
        splits: input.splits.map((split) => ({
          user_id: split.userId || '',
          pending_user_id: split.pendingUserId || '',
          type: split.type,
          percentage: split.percentage,
          shares: split.shares,
          amount: split.amount,
        })),
      }),
    },
  )

  const mapped = mapGroupExpenseDetail(created)
  invalidateGroupExpensesCache(groupId)
  groupExpenseDetailCache.set(
    getExpenseDetailCacheKey(groupId, mapped.expense.id),
    {
      value: mapped,
      expiresAt: Date.now() + GROUP_EXPENSE_DETAIL_CACHE_TTL_MS,
    },
  )
  return mapped
}

export function invalidateGroupExpensesCache(groupId?: string) {
  if (!groupId) {
    recentGroupExpensesCache.clear()
    recentGroupExpensesInFlight.clear()
    groupExpenseDetailCache.clear()
    groupExpenseDetailInFlight.clear()
    return
  }

  for (const key of recentGroupExpensesCache.keys()) {
    if (key.startsWith(`${groupId}:`)) recentGroupExpensesCache.delete(key)
  }
  for (const key of recentGroupExpensesInFlight.keys()) {
    if (key.startsWith(`${groupId}:`)) recentGroupExpensesInFlight.delete(key)
  }
  for (const key of groupExpenseDetailCache.keys()) {
    if (key.startsWith(`${groupId}:`)) groupExpenseDetailCache.delete(key)
  }
  for (const key of groupExpenseDetailInFlight.keys()) {
    if (key.startsWith(`${groupId}:`)) groupExpenseDetailInFlight.delete(key)
  }
}

export function getGroupExpenseById(
  groupId: string,
  expenseId: string,
  options?: { force?: boolean },
) {
  const force = options?.force === true
  const key = getExpenseDetailCacheKey(groupId, expenseId)
  const now = Date.now()
  const cached = groupExpenseDetailCache.get(key)

  if (!force && cached && cached.expiresAt > now) return cached.value

  if (!force) {
    const inFlight = groupExpenseDetailInFlight.get(key)
    if (inFlight) return inFlight
  }

  const request = apiRequest<RawCreateExpenseResponse>(
    `/groups/${groupId}/expenses/${expenseId}`,
    {
      method: 'GET',
      auth: true,
    },
  )
    .then((detail) => {
      const mapped = mapGroupExpenseDetail(detail)
      groupExpenseDetailCache.set(key, {
        value: mapped,
        expiresAt: Date.now() + GROUP_EXPENSE_DETAIL_CACHE_TTL_MS,
      })
      return mapped
    })
    .finally(() => {
      groupExpenseDetailInFlight.delete(key)
    })

  groupExpenseDetailInFlight.set(key, request)
  return request
}
