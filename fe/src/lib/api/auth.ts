import { apiRequest } from '@/lib/api/client'
import {
  clearSessionTokens,
  getRefreshToken,
  setSessionTokens,
} from '@/lib/session'

type LoginResponse = {
  access_token: string
  refresh_token: string
  expires_in: number
}

type CreateUserResponse = {
  id: string
  name: string
  email: string
  created_at: string
}

export async function login(email: string, password: string) {
  const response = await apiRequest<LoginResponse>('/auth/login', {
    method: 'POST',
    body: JSON.stringify({ email, password }),
    retryOnAuthFail: false,
  })

  setSessionTokens(response.access_token, response.refresh_token)
  return response
}

export async function signup(name: string, email: string, password: string) {
  await apiRequest<CreateUserResponse>('/users/', {
    method: 'POST',
    body: JSON.stringify({ name, email, password }),
    retryOnAuthFail: false,
  })

  return login(email, password)
}

export async function logout() {
  const refreshToken = getRefreshToken()

  try {
    if (refreshToken) {
      await apiRequest('/auth/logout', {
        method: 'POST',
        auth: true,
        body: JSON.stringify({ refresh_token: refreshToken }),
        retryOnAuthFail: false,
      })
    }
  } finally {
    clearSessionTokens()
  }
}

export async function logoutAll() {
  try {
    await apiRequest('/auth/logout-all', {
      method: 'POST',
      auth: true,
      retryOnAuthFail: false,
    })
  } finally {
    clearSessionTokens()
  }
}
