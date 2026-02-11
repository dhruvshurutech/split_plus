import { apiRequest } from '@/lib/api/client'

type RawUUID = {
  Bytes?: Array<number>
  Valid?: boolean
}

type RawSettlement = {
  id: string | RawUUID
  group_id: string | RawUUID
  payer_id?: string | RawUUID
  payer_pending_user_id?: string | RawUUID
  payee_id?: string | RawUUID
  payee_pending_user_id?: string | RawUUID
  amount: string
  currency_code: string
  status: 'pending' | 'completed' | 'cancelled'
  payment_method?: string
  transaction_reference?: string
  paid_at?: string
  notes?: string
  created_at: string
  created_by: string | RawUUID
  updated_at: string
  updated_by: string | RawUUID
}

type RawSettlementWithUsers = RawSettlement & {
  payer?: {
    user_id?: string | RawUUID
    pending_user_id?: string | RawUUID
    email: string
    name?: string
    avatar_url?: string
    is_pending?: boolean
  }
  payee?: {
    user_id?: string | RawUUID
    pending_user_id?: string | RawUUID
    email: string
    name?: string
    avatar_url?: string
    is_pending?: boolean
  }
}

export type GroupSettlement = {
  id: string
  groupId: string
  payerId: string
  payerPendingUserId: string
  payeeId: string
  payeePendingUserId: string
  amount: string
  currencyCode: string
  status: 'pending' | 'completed' | 'cancelled'
  paymentMethod: string
  transactionReference: string
  paidAt: string
  notes: string
  createdAt: string
  createdBy: string
  updatedAt: string
  updatedBy: string
}

export type GroupSettlementWithUsers = GroupSettlement & {
  payer: {
    userId: string
    pendingUserId: string
    email: string
    name: string
    avatarUrl: string
    isPending: boolean
  }
  payee: {
    userId: string
    pendingUserId: string
    email: string
    name: string
    avatarUrl: string
    isPending: boolean
  }
}

export type CreateGroupSettlementInput = {
  payerId?: string
  payerPendingUserId?: string
  payeeId?: string
  payeePendingUserId?: string
  amount: string
  currencyCode?: string
  status?: 'pending' | 'completed' | 'cancelled'
  paymentMethod?: string
  transactionReference?: string
  notes?: string
}

const GROUP_SETTLEMENTS_CACHE_TTL_MS = 12_000
const groupSettlementsCache = new Map<
  string,
  { value: Array<GroupSettlementWithUsers>; expiresAt: number }
>()
const groupSettlementsInFlight = new Map<
  string,
  Promise<Array<GroupSettlementWithUsers>>
>()

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

function mapSettlement(raw: RawSettlement): GroupSettlement {
  return {
    id: parseUUID(raw.id),
    groupId: parseUUID(raw.group_id),
    payerId: parseUUID(raw.payer_id),
    payerPendingUserId: parseUUID(raw.payer_pending_user_id),
    payeeId: parseUUID(raw.payee_id),
    payeePendingUserId: parseUUID(raw.payee_pending_user_id),
    amount: raw.amount,
    currencyCode: raw.currency_code,
    status: raw.status,
    paymentMethod: raw.payment_method || '',
    transactionReference: raw.transaction_reference || '',
    paidAt: raw.paid_at || '',
    notes: raw.notes || '',
    createdAt: raw.created_at,
    createdBy: parseUUID(raw.created_by),
    updatedAt: raw.updated_at,
    updatedBy: parseUUID(raw.updated_by),
  }
}

function mapSettlementWithUsers(
  raw: RawSettlementWithUsers,
): GroupSettlementWithUsers {
  return {
    ...mapSettlement(raw),
    payer: {
      userId: parseUUID(raw.payer?.user_id),
      pendingUserId: parseUUID(raw.payer?.pending_user_id),
      email: raw.payer?.email || '',
      name: raw.payer?.name || '',
      avatarUrl: raw.payer?.avatar_url || '',
      isPending: raw.payer?.is_pending === true,
    },
    payee: {
      userId: parseUUID(raw.payee?.user_id),
      pendingUserId: parseUUID(raw.payee?.pending_user_id),
      email: raw.payee?.email || '',
      name: raw.payee?.name || '',
      avatarUrl: raw.payee?.avatar_url || '',
      isPending: raw.payee?.is_pending === true,
    },
  }
}

export function invalidateGroupSettlementsCache(groupId?: string) {
  if (!groupId) {
    groupSettlementsCache.clear()
    groupSettlementsInFlight.clear()
    return
  }

  groupSettlementsCache.delete(groupId)
  groupSettlementsInFlight.delete(groupId)
}

export function listGroupSettlements(
  groupId: string,
  options?: { force?: boolean },
) {
  const force = options?.force === true
  const now = Date.now()
  const cached = groupSettlementsCache.get(groupId)

  if (!force && cached && cached.expiresAt > now) return cached.value

  if (!force) {
    const inFlight = groupSettlementsInFlight.get(groupId)
    if (inFlight) return inFlight
  }

  const request = apiRequest<Array<RawSettlementWithUsers>>(
    `/groups/${groupId}/settlements`,
    {
      method: 'GET',
      auth: true,
    },
  )
    .then((rows) => {
      const mapped = rows.map(mapSettlementWithUsers)
      groupSettlementsCache.set(groupId, {
        value: mapped,
        expiresAt: Date.now() + GROUP_SETTLEMENTS_CACHE_TTL_MS,
      })
      return mapped
    })
    .finally(() => {
      groupSettlementsInFlight.delete(groupId)
    })

  groupSettlementsInFlight.set(groupId, request)
  return request
}

export async function createGroupSettlement(
  groupId: string,
  input: CreateGroupSettlementInput,
) {
  const settlement = await apiRequest<RawSettlement>(
    `/groups/${groupId}/settlements`,
    {
      method: 'POST',
      auth: true,
      body: JSON.stringify({
        payer_id: input.payerId || '',
        payer_pending_user_id: input.payerPendingUserId || '',
        payee_id: input.payeeId || '',
        payee_pending_user_id: input.payeePendingUserId || '',
        amount: input.amount,
        currency_code: input.currencyCode || 'USD',
        status: input.status || 'pending',
        payment_method: input.paymentMethod || '',
        transaction_reference: input.transactionReference || '',
        notes: input.notes || '',
      }),
    },
  )

  return mapSettlement(settlement)
}
