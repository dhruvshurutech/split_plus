import { apiRequest } from '@/lib/api/client'

type RawUUID = {
  Bytes?: Array<number>
  Valid?: boolean
}

type RawGroupBalance = {
  user_id: string | RawUUID
  user_email: string
  user_name?: string
  user_avatar_url?: string
  total_paid: string
  total_owed: string
  balance: string
}

type RawDebt = {
  debtor_id: string | RawUUID
  debtor_pending_user_id?: string | RawUUID
  debtor_email: string
  debtor_name?: string
  creditor_id: string | RawUUID
  creditor_pending_user_id?: string | RawUUID
  creditor_email: string
  creditor_name?: string
  amount: string
}

type RawOverallUserBalance = {
  group_id: string | RawUUID
  group_name: string
  currency_code: string
  total_paid: string
  total_owed: string
  balance: string
}

export type GroupBalance = {
  userId: string
  userEmail: string
  userName: string
  userAvatarUrl: string
  totalPaid: string
  totalOwed: string
  balance: string
}

export type GroupDebt = {
  debtorId: string
  debtorPendingUserId: string
  debtorEmail: string
  debtorName: string
  creditorId: string
  creditorPendingUserId: string
  creditorEmail: string
  creditorName: string
  amount: string
}

export type OverallUserBalance = {
  groupId: string
  groupName: string
  currencyCode: string
  totalPaid: string
  totalOwed: string
  balance: string
}

const GROUP_BALANCES_CACHE_TTL_MS = 12_000
const GROUP_DEBTS_CACHE_TTL_MS = 12_000
const OVERALL_BALANCES_CACHE_TTL_MS = 12_000

const groupBalancesCache = new Map<
  string,
  { value: Array<GroupBalance>; expiresAt: number }
>()
const groupBalancesInFlight = new Map<string, Promise<Array<GroupBalance>>>()

const groupDebtsCache = new Map<
  string,
  { value: Array<GroupDebt>; expiresAt: number }
>()
const groupDebtsInFlight = new Map<string, Promise<Array<GroupDebt>>>()
let overallBalancesCache: {
  value: Array<OverallUserBalance>
  expiresAt: number
} | null = null
let overallBalancesInFlight: Promise<Array<OverallUserBalance>> | null = null

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

function mapGroupBalance(raw: RawGroupBalance): GroupBalance {
  return {
    userId: parseUUID(raw.user_id),
    userEmail: raw.user_email,
    userName: raw.user_name || '',
    userAvatarUrl: raw.user_avatar_url || '',
    totalPaid: raw.total_paid,
    totalOwed: raw.total_owed,
    balance: raw.balance,
  }
}

function mapGroupDebt(raw: RawDebt): GroupDebt {
  return {
    debtorId: parseUUID(raw.debtor_id),
    debtorPendingUserId: parseUUID(raw.debtor_pending_user_id),
    debtorEmail: raw.debtor_email,
    debtorName: raw.debtor_name || '',
    creditorId: parseUUID(raw.creditor_id),
    creditorPendingUserId: parseUUID(raw.creditor_pending_user_id),
    creditorEmail: raw.creditor_email,
    creditorName: raw.creditor_name || '',
    amount: raw.amount,
  }
}

function mapOverallUserBalance(raw: RawOverallUserBalance): OverallUserBalance {
  return {
    groupId: parseUUID(raw.group_id),
    groupName: raw.group_name,
    currencyCode: raw.currency_code,
    totalPaid: raw.total_paid,
    totalOwed: raw.total_owed,
    balance: raw.balance,
  }
}

export function invalidateGroupBalancesCache(groupId?: string) {
  if (!groupId) {
    groupBalancesCache.clear()
    groupBalancesInFlight.clear()
    overallBalancesCache = null
    overallBalancesInFlight = null
    return
  }

  groupBalancesCache.delete(groupId)
  groupBalancesInFlight.delete(groupId)
}

export function invalidateGroupDebtsCache(groupId?: string) {
  if (!groupId) {
    groupDebtsCache.clear()
    groupDebtsInFlight.clear()
    return
  }

  groupDebtsCache.delete(groupId)
  groupDebtsInFlight.delete(groupId)
}

export function listGroupBalances(
  groupId: string,
  options?: { force?: boolean },
) {
  const force = options?.force === true
  const now = Date.now()
  const cached = groupBalancesCache.get(groupId)

  if (!force && cached && cached.expiresAt > now) return cached.value

  if (!force) {
    const inFlight = groupBalancesInFlight.get(groupId)
    if (inFlight) return inFlight
  }

  const request = apiRequest<Array<RawGroupBalance>>(
    `/groups/${groupId}/balances`,
    {
      method: 'GET',
      auth: true,
    },
  )
    .then((rows) => {
      const mapped = rows.map(mapGroupBalance)
      groupBalancesCache.set(groupId, {
        value: mapped,
        expiresAt: Date.now() + GROUP_BALANCES_CACHE_TTL_MS,
      })
      return mapped
    })
    .finally(() => {
      groupBalancesInFlight.delete(groupId)
    })

  groupBalancesInFlight.set(groupId, request)
  return request
}

export function listGroupDebts(groupId: string, options?: { force?: boolean }) {
  const force = options?.force === true
  const now = Date.now()
  const cached = groupDebtsCache.get(groupId)

  if (!force && cached && cached.expiresAt > now) return cached.value

  if (!force) {
    const inFlight = groupDebtsInFlight.get(groupId)
    if (inFlight) return inFlight
  }

  const request = apiRequest<Array<RawDebt>>(`/groups/${groupId}/debts`, {
    method: 'GET',
    auth: true,
  })
    .then((rows) => {
      const mapped = rows.map(mapGroupDebt)
      groupDebtsCache.set(groupId, {
        value: mapped,
        expiresAt: Date.now() + GROUP_DEBTS_CACHE_TTL_MS,
      })
      return mapped
    })
    .finally(() => {
      groupDebtsInFlight.delete(groupId)
    })

  groupDebtsInFlight.set(groupId, request)
  return request
}

export function listOverallUserBalances(options?: { force?: boolean }) {
  const force = options?.force === true
  const now = Date.now()

  if (!force && overallBalancesCache && overallBalancesCache.expiresAt > now) {
    return overallBalancesCache.value
  }

  if (!force && overallBalancesInFlight) return overallBalancesInFlight

  overallBalancesInFlight = apiRequest<Array<RawOverallUserBalance>>(
    '/users/me/balances',
    {
      method: 'GET',
      auth: true,
    },
  )
    .then((rows) => {
      const mapped = rows.map(mapOverallUserBalance)
      overallBalancesCache = {
        value: mapped,
        expiresAt: Date.now() + OVERALL_BALANCES_CACHE_TTL_MS,
      }
      return mapped
    })
    .finally(() => {
      overallBalancesInFlight = null
    })

  return overallBalancesInFlight
}
