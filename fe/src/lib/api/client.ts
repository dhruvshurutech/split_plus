import {
  clearSessionTokens,
  getAccessToken,
  getRefreshToken,
  setSessionTokens,
} from '@/lib/session'
import { getUserFacingErrorMessage } from '@/lib/api/error-messages'

const API_BASE_URL =
  import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080'

type ApiEnvelope<T> = {
  status: boolean
  data?: T
  error?: {
    code?: string
    message?: string | Array<string>
    details?: unknown
  }
}

type RequestOptions = Omit<RequestInit, 'headers'> & {
  auth?: boolean
  headers?: Record<string, string>
  retryOnAuthFail?: boolean
}

export class ApiError extends Error {
  statusCode: number
  code?: string
  details?: unknown

  constructor(
    message: string,
    statusCode: number,
    details?: unknown,
    code?: string,
  ) {
    super(message)
    this.name = 'ApiError'
    this.statusCode = statusCode
    this.details = details
    this.code = code
  }
}

function isFormData(body: BodyInit | null | undefined) {
  return typeof FormData !== 'undefined' && body instanceof FormData
}

function normalizeErrorMessage(message: string | Array<string> | undefined) {
  if (Array.isArray(message)) return message.join(', ')
  return message || 'Something went wrong'
}

async function parseEnvelope<T>(
  response: Response,
): Promise<ApiEnvelope<T> | null> {
  const contentType = response.headers.get('content-type')
  if (!contentType?.includes('application/json')) return null

  try {
    return (await response.json()) as ApiEnvelope<T>
  } catch {
    return null
  }
}

let refreshPromise: Promise<boolean> | null = null

async function tryRefreshAccessToken() {
  if (refreshPromise) return refreshPromise

  refreshPromise = (async () => {
    const refreshToken = getRefreshToken()
    if (!refreshToken) {
      clearSessionTokens()
      return false
    }

    try {
      const response = await fetch(`${API_BASE_URL}/auth/refresh`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          refresh_token: refreshToken,
        }),
      })

      const envelope = await parseEnvelope<{ access_token: string }>(response)

      if (!response.ok || !envelope?.status || !envelope.data?.access_token) {
        clearSessionTokens()
        return false
      }

      setSessionTokens(envelope.data.access_token, refreshToken)
      return true
    } catch {
      clearSessionTokens()
      return false
    }
  })()

  try {
    return await refreshPromise
  } finally {
    refreshPromise = null
  }
}

async function getValidAccessToken() {
  const existingAccessToken = getAccessToken()
  if (existingAccessToken) return existingAccessToken

  const refreshed = await tryRefreshAccessToken()
  if (!refreshed) return null

  return getAccessToken()
}

export async function apiRequest<T>(
  path: string,
  options: RequestOptions = {},
): Promise<T> {
  const {
    auth = false,
    retryOnAuthFail = true,
    headers,
    body,
    ...rest
  } = options

  const finalHeaders: Record<string, string> = {
    ...(headers || {}),
  }

  if (body && !isFormData(body) && !finalHeaders['Content-Type']) {
    finalHeaders['Content-Type'] = 'application/json'
  }

  if (auth) {
    const accessToken = await getValidAccessToken()
    if (!accessToken) {
      throw new ApiError(
        'Session expired. Please sign in again.',
        401,
        undefined,
        'auth.refresh_token.invalid',
      )
    }
    finalHeaders.Authorization = `Bearer ${accessToken}`
  }

  const response = await fetch(`${API_BASE_URL}${path}`, {
    ...rest,
    body,
    headers: finalHeaders,
  })

  if (response.status === 401 && auth && retryOnAuthFail) {
    const refreshed = await tryRefreshAccessToken()
    if (refreshed) {
      return apiRequest<T>(path, {
        ...options,
        retryOnAuthFail: false,
      })
    }

    throw new ApiError(
      'Session expired. Please sign in again.',
      401,
      undefined,
      'auth.refresh_token.invalid',
    )
  }

  if (response.status === 204) {
    return undefined as T
  }

  const envelope = await parseEnvelope<T>(response)

  if (!response.ok || !envelope?.status) {
    const code = envelope?.error?.code
    const rawMessage = normalizeErrorMessage(envelope?.error?.message)
    const message = getUserFacingErrorMessage(code, rawMessage)
    throw new ApiError(message, response.status, envelope?.error?.details, code)
  }

  return envelope.data as T
}
