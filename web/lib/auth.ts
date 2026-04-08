import type { AuthLoginResponse, User } from '@/types/api'

const ACCESS_TOKEN_KEY = 'web_access_token'
const REFRESH_TOKEN_KEY = 'web_refresh_token'
const USER_KEY = 'web_user'

export function getAccessToken(): string {
  if (typeof window === 'undefined') return ''
  return localStorage.getItem(ACCESS_TOKEN_KEY) || ''
}

export function getRefreshToken(): string {
  if (typeof window === 'undefined') return ''
  return localStorage.getItem(REFRESH_TOKEN_KEY) || ''
}

export function getCurrentUser(): User | null {
  if (typeof window === 'undefined') return null
  const raw = localStorage.getItem(USER_KEY)
  if (!raw) return null
  try {
    return JSON.parse(raw) as User
  } catch {
    return null
  }
}

export function saveAuth(payload: AuthLoginResponse): void {
  if (typeof window === 'undefined') return
  localStorage.setItem(ACCESS_TOKEN_KEY, payload.access_token)
  localStorage.setItem(REFRESH_TOKEN_KEY, payload.refresh_token)
  localStorage.setItem(USER_KEY, JSON.stringify(payload.user))
}

export function saveCurrentUser(user: User): void {
  if (typeof window === 'undefined') return
  localStorage.setItem(USER_KEY, JSON.stringify(user))
}

export function clearAuth(): void {
  if (typeof window === 'undefined') return
  localStorage.removeItem(ACCESS_TOKEN_KEY)
  localStorage.removeItem(REFRESH_TOKEN_KEY)
  localStorage.removeItem(USER_KEY)
}
