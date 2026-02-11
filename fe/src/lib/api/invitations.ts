import { apiRequest } from '@/lib/api/client'

type RawUUID = {
  Bytes?: Array<number>
  Valid?: boolean
}

type RawInvitation = {
  id: string | RawUUID
  group_id: string | RawUUID
  group_name?: string
  email: string
  role: string
  status: string
  expires_at: string
  invited_by: string | RawUUID
  inviter_name?: string
  inviter_email?: string
}

export type Invitation = {
  id: string
  groupId: string
  groupName: string
  email: string
  role: string
  status: string
  expiresAt: string
  invitedBy: string
  inviterName?: string
  inviterEmail?: string
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

function mapInvitation(raw: RawInvitation): Invitation {
  return {
    id: parseUUID(raw.id),
    groupId: parseUUID(raw.group_id),
    groupName: raw.group_name || '',
    email: raw.email,
    role: raw.role,
    status: raw.status,
    expiresAt: raw.expires_at,
    invitedBy: parseUUID(raw.invited_by),
    inviterName: raw.inviter_name || undefined,
    inviterEmail: raw.inviter_email || undefined,
  }
}

export async function getInvitation(token: string) {
  const invitation = await apiRequest<RawInvitation>(`/invitations/${token}`, {
    method: 'GET',
    retryOnAuthFail: false,
  })
  return mapInvitation(invitation)
}

export async function acceptInvitation(token: string) {
  return apiRequest<{ message: string }>(`/invitations/${token}/accept`, {
    method: 'POST',
    auth: true,
  })
}

export async function joinGroupViaInvitation(
  token: string,
  input: { password: string; name?: string },
) {
  return apiRequest<{ message: string }>(`/invitations/${token}/join`, {
    method: 'POST',
    retryOnAuthFail: false,
    body: JSON.stringify({
      password: input.password,
      name: input.name || '',
    }),
  })
}
