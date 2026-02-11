const ACCESS_TOKEN_KEY = 'splitplus.accessToken'
const REFRESH_TOKEN_KEY = 'splitplus.refreshToken'

export function canUseStorage() {
  return typeof window !== 'undefined' && typeof localStorage !== 'undefined'
}

export function getAccessToken() {
  if (!canUseStorage()) return null
  return localStorage.getItem(ACCESS_TOKEN_KEY)
}

export function getRefreshToken() {
  if (!canUseStorage()) return null
  return localStorage.getItem(REFRESH_TOKEN_KEY)
}

export function setSessionTokens(accessToken: string, refreshToken: string) {
  if (!canUseStorage()) return
  localStorage.setItem(ACCESS_TOKEN_KEY, accessToken)
  localStorage.setItem(REFRESH_TOKEN_KEY, refreshToken)
}

export function clearSessionTokens() {
  if (!canUseStorage()) return
  localStorage.removeItem(ACCESS_TOKEN_KEY)
  localStorage.removeItem(REFRESH_TOKEN_KEY)
}

export function isAuthenticated() {
  return Boolean(getAccessToken())
}

export function hasSession() {
  return Boolean(getAccessToken() && getRefreshToken())
}
