import { apiRequest } from '@/lib/api/client'

type RawUUID = {
  Bytes?: Array<number>
  Valid?: boolean
}

type RawUserGroup = {
  id: string | RawUUID
  name: string
  description?: string
  currency_code: string
  created_at: string
  membership_id: string | RawUUID
  member_role: string
  member_status: string
  member_joined_at?: string
}

type RawGroupMember = {
  id: string | RawUUID
  group_id: string | RawUUID
  user_id: string | RawUUID
  invitation_token?: string
  role: string
  status: string
  invited_at?: string
  joined_at?: string
  user: {
    email: string
    name?: string
    avatar_url?: string
  }
}

type RawCreateGroupResponse = {
  id: string | RawUUID
  name: string
  description?: string
  currency_code: string
  created_at: string
  role: string
}

type RawCreateInvitationResponse = {
  message: string
  token: string
}

export type UserGroup = {
  id: string
  name: string
  description: string
  currencyCode: string
  createdAt: string
  membershipId: string
  memberRole: string
  memberStatus: string
  memberJoinedAt: string
}

export type GroupMember = {
  id: string
  groupId: string
  userId: string
  invitationToken: string
  role: string
  status: string
  invitedAt: string
  joinedAt: string
  user: {
    email: string
    name: string
    avatarUrl: string
  }
}

export type CreatedGroup = {
  id: string
  name: string
  description: string
  currencyCode: string
  createdAt: string
  role: string
}

export type CreateInvitationResult = {
  message: string
  token: string
}

const USER_GROUPS_CACHE_TTL_MS = 20_000
const GROUP_MEMBERS_CACHE_TTL_MS = 15_000

let userGroupsCache: { value: Array<UserGroup>; expiresAt: number } | null =
  null
let userGroupsInFlight: Promise<Array<UserGroup>> | null = null
const groupMembersCache = new Map<
  string,
  { value: Array<GroupMember>; expiresAt: number }
>()
const groupMembersInFlight = new Map<string, Promise<Array<GroupMember>>>()

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

function mapUserGroup(raw: RawUserGroup): UserGroup {
  return {
    id: parseUUID(raw.id),
    name: raw.name,
    description: raw.description || '',
    currencyCode: raw.currency_code,
    createdAt: raw.created_at,
    membershipId: parseUUID(raw.membership_id),
    memberRole: raw.member_role,
    memberStatus: raw.member_status,
    memberJoinedAt: raw.member_joined_at || '',
  }
}

function mapGroupMember(raw: RawGroupMember): GroupMember {
  return {
    id: parseUUID(raw.id),
    groupId: parseUUID(raw.group_id),
    userId: parseUUID(raw.user_id),
    invitationToken: raw.invitation_token || '',
    role: raw.role,
    status: raw.status,
    invitedAt: raw.invited_at || '',
    joinedAt: raw.joined_at || '',
    user: {
      email: raw.user.email,
      name: raw.user.name || '',
      avatarUrl: raw.user.avatar_url || '',
    },
  }
}

function mapCreatedGroup(raw: RawCreateGroupResponse): CreatedGroup {
  return {
    id: parseUUID(raw.id),
    name: raw.name,
    description: raw.description || '',
    currencyCode: raw.currency_code,
    createdAt: raw.created_at,
    role: raw.role,
  }
}

export function listUserGroups() {
  const now = Date.now()

  if (userGroupsCache && userGroupsCache.expiresAt > now) {
    return userGroupsCache.value
  }

  if (userGroupsInFlight) return userGroupsInFlight

  userGroupsInFlight = apiRequest<Array<RawUserGroup>>('/groups', {
    method: 'GET',
    auth: true,
  })
    .then((groups) => {
      const mapped = groups.map(mapUserGroup)
      userGroupsCache = {
        value: mapped,
        expiresAt: Date.now() + USER_GROUPS_CACHE_TTL_MS,
      }
      return mapped
    })
    .finally(() => {
      userGroupsInFlight = null
    })

  return userGroupsInFlight
}

export function invalidateUserGroupsCache() {
  userGroupsCache = null
  userGroupsInFlight = null
}

export async function createGroup(input: {
  name: string
  description?: string
  currencyCode?: string
}) {
  const group = await apiRequest<RawCreateGroupResponse>('/groups', {
    method: 'POST',
    auth: true,
    body: JSON.stringify({
      name: input.name,
      description: input.description || '',
      currency_code: input.currencyCode || 'USD',
    }),
  })
  const mapped = mapCreatedGroup(group)
  invalidateUserGroupsCache()
  return mapped
}

export function invalidateGroupMembersCache(groupId?: string) {
  if (!groupId) {
    groupMembersCache.clear()
    groupMembersInFlight.clear()
    return
  }

  groupMembersCache.delete(groupId)
  groupMembersInFlight.delete(groupId)
}

export function listGroupMembers(
  groupId: string,
  options?: { force?: boolean },
) {
  const force = options?.force === true
  const now = Date.now()
  const cached = groupMembersCache.get(groupId)

  if (!force && cached && cached.expiresAt > now) {
    return cached.value
  }

  if (!force) {
    const inFlight = groupMembersInFlight.get(groupId)
    if (inFlight) return inFlight
  }

  const request = apiRequest<Array<RawGroupMember>>(
    `/groups/${groupId}/members`,
    {
      method: 'GET',
      auth: true,
    },
  )
    .then((members) => {
      const mapped = members.map(mapGroupMember)
      groupMembersCache.set(groupId, {
        value: mapped,
        expiresAt: Date.now() + GROUP_MEMBERS_CACHE_TTL_MS,
      })
      return mapped
    })
    .finally(() => {
      groupMembersInFlight.delete(groupId)
    })

  groupMembersInFlight.set(groupId, request)
  return request
}

export async function createGroupInvitation(
  groupId: string,
  input: { email: string; name?: string; role?: 'member' | 'admin' },
) {
  const result = (await apiRequest<RawCreateInvitationResponse>(
    `/groups/${groupId}/invitations`,
    {
      method: 'POST',
      auth: true,
      body: JSON.stringify({
        email: input.email,
        name: input.name || '',
        role: input.role || 'member',
      }),
    },
  )) as CreateInvitationResult

  invalidateGroupMembersCache(groupId)
  return result
}

export function buildInvitationLink(token: string) {
  if (typeof window === 'undefined') return `/invite/${token}`
  return `${window.location.origin}/invite/${token}`
}
