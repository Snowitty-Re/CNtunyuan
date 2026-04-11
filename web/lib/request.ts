import { clearAuth, getAccessToken, getRefreshToken, saveAuth } from '@/lib/auth'
import type { ApiEnvelope, AuthLoginResponse } from '@/types/api'

const API_BASE = process.env.NEXT_PUBLIC_API_BASE || 'http://localhost:8080/api/v1'

type HttpMethod = 'GET' | 'POST' | 'PUT' | 'DELETE'

type RequestOptions = {
  method?: HttpMethod
  body?: unknown
  auth?: boolean
  query?: Record<string, string | number | boolean | undefined | null>
}

function buildUrl(path: string, query?: RequestOptions['query']): string {
  const url = new URL(`${API_BASE}${path}`)
  if (query) {
    Object.entries(query).forEach(([k, v]) => {
      if (v !== undefined && v !== null && `${v}` !== '') url.searchParams.set(k, String(v))
    })
  }
  return url.toString()
}

async function refreshToken(): Promise<boolean> {
  const refresh = getRefreshToken()
  if (!refresh) return false

  const res = await fetch(`${API_BASE}/auth/refresh`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ refresh_token: refresh }),
  })
  if (!res.ok) return false

  const payload = (await res.json()) as ApiEnvelope<AuthLoginResponse>
  if (!payload || (payload.code !== 0 && payload.code !== 200)) return false
  saveAuth(payload.data)
  return true
}

export async function http<T>(path: string, options: RequestOptions = {}): Promise<T> {
  const { method = 'GET', body, auth = true, query } = options

  const headers: Record<string, string> = {
    'Content-Type': 'application/json',
  }

  if (auth) {
    const token = getAccessToken()
    if (token) headers.Authorization = `Bearer ${token}`
  }

  const run = async () => {
    const res = await fetch(buildUrl(path, query), {
      method,
      headers,
      body: body === undefined ? undefined : JSON.stringify(body),
      cache: 'no-store',
    })

    if (res.status === 204) return null as T

    let payload: ApiEnvelope<T> | null = null
    try {
      payload = (await res.json()) as ApiEnvelope<T>
    } catch {
      throw new Error(`HTTP ${res.status}`)
    }

    if (res.status === 401 || payload.code === 401) {
      return undefined as T
    }

    if (!res.ok || (payload.code !== 0 && payload.code !== 200)) {
      throw new Error(payload.message || `HTTP ${res.status}`)
    }

    return payload.data
  }

  let data = await run()
  if (data !== undefined) return data

  if (!auth) {
    throw new Error('未授权访问')
  }

  const refreshed = auth ? await refreshToken() : false
  if (!refreshed) {
    clearAuth()
    throw new Error('登录已过期，请重新登录')
  }

  const token = getAccessToken()
  if (token) headers.Authorization = `Bearer ${token}`
  data = await run()
  if (data === undefined) {
    clearAuth()
    throw new Error('登录已过期，请重新登录')
  }
  return data
}

export { API_BASE }
