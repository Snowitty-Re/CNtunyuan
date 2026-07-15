'use client'

import { createContext, PropsWithChildren, useCallback, useContext, useEffect, useMemo, useState } from 'react'
import { clearAuth, getAccessToken, getCurrentUser, saveCurrentUser } from '@/lib/auth'
import { authService } from '@/services/auth'
import type { User } from '@/types/api'

type AuthContextValue = {
  ready: boolean
  user: User | null
  refreshUser: () => Promise<User | null>
  logout: () => void
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
      const local = getCurrentUser()
      setUser(local)
      return local
    }
  }, [])

  const logout = useCallback(() => {
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

  const value = useMemo(
    () => ({ ready, user, refreshUser, logout, setUser }),
    [ready, user, refreshUser, logout],
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
