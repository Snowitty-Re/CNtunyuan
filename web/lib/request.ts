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

export class ApiError extends Error {
  status: number
  code?: number

  constructor(message: string, status: number, code?: number) {
    super(message)
    this.name = 'ApiError'
    this.status = status
    this.code = code
  }
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

let refreshPromise: Promise<boolean> | null = null

async function refreshTokenOnce(): Promise<boolean> {
  if (refreshPromise) return refreshPromise
  refreshPromise = (async () => {
    const refresh = getRefreshToken()
    if (!refresh) return false
    try {
      const res = await fetch(`${API_BASE}/auth/refresh`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ refresh_token: refresh }),
      })
      if (!res.ok) return false
      const payload = (await res.json()) as ApiEnvelope<AuthLoginResponse>
      if (!payload || (payload.code !== 0 && payload.code !== 200) || !payload.data) return false
      saveAuth(payload.data)
      return true
    } catch {
      return false
    } finally {
      refreshPromise = null
    }
  })()
  return refreshPromise
}

function mapErrorMessage(message: string, status: number): string {
  const msg = String(message || '')
  if (status === 403) {
    if (msg.includes('手机号') || msg.includes('绑定')) return msg || '请先绑定真实手机号'
    if (msg.includes('激活') || msg.includes('审批') || msg.includes('禁用')) return msg || '账号未激活或已禁用'
    return msg || '权限不足'
  }
  if (status === 401) return msg || '登录已过期，请重新登录'
  return msg || `请求失败 (${status})`
}

/** 带鉴权与单飞 refresh 的通用 fetch（支持 JSON / FormData / blob） */
export async function authedFetch(path: string, init: RequestInit = {}, auth = true): Promise<Response> {
  const headers = new Headers(init.headers || {})
  if (auth) {
    const token = getAccessToken()
    if (token) headers.set('Authorization', `Bearer ${token}`)
  }

  const url = path.startsWith('http') ? path : `${API_BASE}${path.startsWith('/') ? path : `/${path}`}`
  let res = await fetch(url, { ...init, headers, cache: 'no-store' })

  if (auth && res.status === 401) {
    const ok = await refreshTokenOnce()
    if (!ok) {
      clearAuth()
      throw new ApiError('登录已过期，请重新登录', 401)
    }
    const token = getAccessToken()
    if (token) headers.set('Authorization', `Bearer ${token}`)
    res = await fetch(url, { ...init, headers, cache: 'no-store' })
  }

  return res
}

export async function http<T>(path: string, options: RequestOptions = {}): Promise<T> {
  const { method = 'GET', body, auth = true, query } = options

  const headers: Record<string, string> = {
    'Content-Type': 'application/json',
  }

  const run = async (token?: string) => {
    if (auth && token) headers.Authorization = `Bearer ${token}`
    else if (auth) {
      const t = getAccessToken()
      if (t) headers.Authorization = `Bearer ${t}`
    }

    const res = await fetch(buildUrl(path, query), {
      method,
      headers,
      body: body === undefined ? undefined : JSON.stringify(body),
      cache: 'no-store',
    })

    if (res.status === 204) return { kind: 'ok' as const, data: null as T }

    let payload: ApiEnvelope<T> | null = null
    try {
      payload = (await res.json()) as ApiEnvelope<T>
    } catch {
      throw new ApiError(`HTTP ${res.status}`, res.status)
    }

    if (res.status === 401 || payload.code === 401) {
      return { kind: 'unauthorized' as const }
    }

    if (!res.ok || (payload.code !== 0 && payload.code !== 200)) {
      throw new ApiError(mapErrorMessage(payload.message, res.status), res.status, payload.code)
    }

    return { kind: 'ok' as const, data: payload.data }
  }

  let result = await run()
  if (result.kind === 'ok') return result.data

  if (!auth) throw new ApiError('未授权访问', 401)

  const refreshed = await refreshTokenOnce()
  if (!refreshed) {
    clearAuth()
    throw new ApiError('登录已过期，请重新登录', 401)
  }

  result = await run(getAccessToken())
  if (result.kind === 'unauthorized') {
    clearAuth()
    throw new ApiError('登录已过期，请重新登录', 401)
  }
  return result.data
}

export { API_BASE }
