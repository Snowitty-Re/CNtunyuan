import axios, { AxiosError, InternalAxiosRequestConfig } from 'axios'
import { message } from 'antd'
import type { ApiResponse } from '@/types'

const request = axios.create({
  baseURL: import.meta.env.VITE_API_BASE_URL || '/api/v1',
  timeout: 30000,
  headers: { 'Content-Type': 'application/json' },
})

// Token refresh state
let isRefreshing = false
let pendingRequests: Array<(token: string) => void> = []

function getToken(): string | null {
  try {
    const raw = localStorage.getItem('auth-storage')
    if (!raw) return null
    const parsed = JSON.parse(raw)
    return parsed?.state?.token || null
  } catch {
    return null
  }
}

function getRefreshToken(): string | null {
  try {
    const raw = localStorage.getItem('auth-storage')
    if (!raw) return null
    const parsed = JSON.parse(raw)
    return parsed?.state?.refreshToken || null
  } catch {
    return null
  }
}

function setTokens(token: string, refreshToken: string) {
  try {
    const raw = localStorage.getItem('auth-storage')
    if (!raw) return
    const parsed = JSON.parse(raw)
    parsed.state.token = token
    parsed.state.refreshToken = refreshToken
    localStorage.setItem('auth-storage', JSON.stringify(parsed))
  } catch { /* ignore */ }
}

function clearAuth() {
  localStorage.removeItem('auth-storage')
  window.location.href = '/login'
}

// Request interceptor
request.interceptors.request.use((config: InternalAxiosRequestConfig) => {
  const token = getToken()
  if (token && config.headers) {
    config.headers.Authorization = `Bearer ${token}`
  }
  return config
})

// Response interceptor
request.interceptors.response.use(
  (response) => {
    const data = response.data as ApiResponse
    // Backend returns code 200 or 0 for success
    if (data.code !== 200 && data.code !== 0) {
      message.error(data.message || '请求失败')
      return Promise.reject(new Error(data.message))
    }
    return response
  },
  async (error: AxiosError<ApiResponse>) => {
    const originalRequest = error.config as InternalAxiosRequestConfig & { _retry?: boolean }

    // Handle 401 - try token refresh
    if (error.response?.status === 401 && !originalRequest._retry) {
      originalRequest._retry = true
      const refreshToken = getRefreshToken()

      if (!refreshToken) {
        clearAuth()
        return Promise.reject(error)
      }

      if (isRefreshing) {
        return new Promise((resolve) => {
          pendingRequests.push((newToken: string) => {
            originalRequest.headers.Authorization = `Bearer ${newToken}`
            resolve(request(originalRequest))
          })
        })
      }

      isRefreshing = true
      try {
        const res = await axios.post(
          `${import.meta.env.VITE_API_BASE_URL || '/api/v1'}/auth/refresh`,
          { refresh_token: refreshToken }
        )
        const { access_token, refresh_token } = res.data.data
        setTokens(access_token, refresh_token)
        pendingRequests.forEach((cb) => cb(access_token))
        pendingRequests = []
        originalRequest.headers.Authorization = `Bearer ${access_token}`
        return request(originalRequest)
      } catch {
        clearAuth()
        return Promise.reject(error)
      } finally {
        isRefreshing = false
      }
    }

    // Other errors
    const msg = error.response?.data?.message || error.message || '网络错误'
    if (error.response?.status !== 401) {
      message.error(msg)
    }
    return Promise.reject(error)
  }
)

export default request
