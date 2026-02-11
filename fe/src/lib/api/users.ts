import { apiRequest } from '@/lib/api/client'

export type MeResponse = {
  id: string
  name: string
  email: string
  created_at: string
}

const ME_CACHE_TTL_MS = 30_000

let meCache: { value: MeResponse; expiresAt: number } | null = null
let meInFlight: Promise<MeResponse> | null = null

export function invalidateMeCache() {
  meCache = null
}

export function getMe(options?: { force?: boolean }) {
  const force = options?.force === true
  const now = Date.now()

  if (!force && meCache && meCache.expiresAt > now) {
    return Promise.resolve(meCache.value)
  }

  if (!force && meInFlight) return meInFlight

  meInFlight = apiRequest<MeResponse>('/users/me', {
    method: 'GET',
    auth: true,
  })
    .then((result) => {
      meCache = { value: result, expiresAt: Date.now() + ME_CACHE_TTL_MS }
      return result
    })
    .finally(() => {
      meInFlight = null
    })

  return meInFlight
}
