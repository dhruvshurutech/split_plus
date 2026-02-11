import { apiRequest } from '@/lib/api/client'

type RawUUID = {
  Bytes?: Array<number>
  Valid?: boolean
}

type RawGroupActivity = {
  id: string | RawUUID
  group_id: string | RawUUID
  user_id: string | RawUUID
  action: string
  entity_type: string
  entity_id: string | RawUUID
  metadata?: Record<string, unknown>
  created_at: string
  user?: {
    email: string
    name?: string
    avatar_url?: string
  }
}

export type GroupActivity = {
  id: string
  groupId: string
  userId: string
  action: string
  entityType: string
  entityId: string
  metadata: Record<string, unknown>
  createdAt: string
  user: {
    email: string
    name: string
    avatarUrl: string
  }
}

const GROUP_ACTIVITY_CACHE_TTL_MS = 10_000
const groupActivityCache = new Map<
  string,
  { value: Array<GroupActivity>; expiresAt: number }
>()
const groupActivityInFlight = new Map<string, Promise<Array<GroupActivity>>>()

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

function mapGroupActivity(raw: RawGroupActivity): GroupActivity {
  return {
    id: parseUUID(raw.id),
    groupId: parseUUID(raw.group_id),
    userId: parseUUID(raw.user_id),
    action: raw.action,
    entityType: raw.entity_type,
    entityId: parseUUID(raw.entity_id),
    metadata: raw.metadata || {},
    createdAt: raw.created_at,
    user: {
      email: raw.user?.email || '',
      name: raw.user?.name || '',
      avatarUrl: raw.user?.avatar_url || '',
    },
  }
}

function getCacheKey(groupId: string, limit: number, offset: number) {
  return `${groupId}:${limit}:${offset}`
}

export function invalidateGroupActivityCache(groupId?: string) {
  if (!groupId) {
    groupActivityCache.clear()
    groupActivityInFlight.clear()
    return
  }

  for (const key of groupActivityCache.keys()) {
    if (key.startsWith(`${groupId}:`)) groupActivityCache.delete(key)
  }
  for (const key of groupActivityInFlight.keys()) {
    if (key.startsWith(`${groupId}:`)) groupActivityInFlight.delete(key)
  }
}

export function listGroupActivities(
  groupId: string,
  options?: { limit?: number; offset?: number; force?: boolean },
) {
  const limit = options?.limit ?? 8
  const offset = options?.offset ?? 0
  const force = options?.force === true
  const key = getCacheKey(groupId, limit, offset)
  const now = Date.now()
  const cached = groupActivityCache.get(key)

  if (!force && cached && cached.expiresAt > now) return cached.value

  if (!force) {
    const inFlight = groupActivityInFlight.get(key)
    if (inFlight) return inFlight
  }

  const query = new URLSearchParams({
    limit: String(limit),
    offset: String(offset),
  })

  const request = apiRequest<Array<RawGroupActivity>>(
    `/groups/${groupId}/activity?${query.toString()}`,
    {
      method: 'GET',
      auth: true,
    },
  )
    .then((rows) => {
      const mapped = rows.map(mapGroupActivity)
      groupActivityCache.set(key, {
        value: mapped,
        expiresAt: Date.now() + GROUP_ACTIVITY_CACHE_TTL_MS,
      })
      return mapped
    })
    .finally(() => {
      groupActivityInFlight.delete(key)
    })

  groupActivityInFlight.set(key, request)
  return request
}
