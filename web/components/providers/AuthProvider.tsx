'use client'

import { createContext, PropsWithChildren, useCallback, useContext, useEffect, useMemo, useState } from 'react'
import { clearAuth, getAccessToken, getCurrentUser, saveCurrentUser } from '@/lib/auth'
import { getSessionBlockReason, sessionBlockMessage, type SessionBlockReason } from '@/lib/session'
import { authService } from '@/services/auth'
import type { User } from '@/types/api'

type AuthContextValue = {
  ready: boolean
  user: User | null
  blockReason: SessionBlockReason
  blockMessage: string
  refreshUser: () => Promise<User | null>
  logout: () => Promise<void>
  setUser: (user: User | null) => void
}

const AuthContext = createContext<AuthContextValue | null>(null)

export function AuthProvider({ children }: PropsWithChildren) {
  const [ready, setReady] = useState(false)
  const [user, setUser] = useState<User | null>(null)

  const refreshUser = useCallback(async () => {
    const token = getAccessToken()
    if (!token) {
      setUser(null)
      return null
    }
    try {
      const me = await authService.me()
      setUser(me)
      saveCurrentUser(me)
      return me
    } catch {
      // me 失败时不信任陈旧 localStorage 作为有效会话
      const local = getCurrentUser()
      if (local && getAccessToken()) {
        setUser(local)
        return local
      }
      setUser(null)
      return null
    }
  }, [])

  const logout = useCallback(async () => {
    try {
      if (getAccessToken()) await authService.logout()
    } catch {
      // ignore
    }
    clearAuth()
    setUser(null)
  }, [])

  useEffect(() => {
    const token = getAccessToken()
    if (!token) {
      setUser(null)
      setReady(true)
      return
    }
    setUser(getCurrentUser())
    refreshUser().finally(() => setReady(true))
  }, [refreshUser])

  const blockReason = getSessionBlockReason(user)
  const blockMessage = sessionBlockMessage(blockReason)

  const value = useMemo(
    () => ({ ready, user, blockReason, blockMessage, refreshUser, logout, setUser }),
    [ready, user, blockReason, blockMessage, refreshUser, logout],
  )

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>
}

export function useAuth(): AuthContextValue {
  const ctx = useContext(AuthContext)
  if (!ctx) {
    throw new Error('useAuth must be used within AuthProvider')
  }
  return ctx
}
