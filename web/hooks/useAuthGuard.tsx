'use client'

import { useEffect, useState } from 'react'
import { useRouter } from 'next/navigation'
import { getCurrentUser, getAccessToken, saveCurrentUser } from '@/lib/auth'
import { authService } from '@/services/auth'
import type { User } from '@/types/api'

export function useAuthGuard() {
  const router = useRouter()
  const [ready, setReady] = useState(false)
  const [user, setUser] = useState<User | null>(null)

  useEffect(() => {
    const token = getAccessToken()
    if (!token) {
      router.replace('/login')
      return
    }
    const local = getCurrentUser()
    setUser(local)
    authService
      .me()
      .then((me) => {
        setUser(me)
        saveCurrentUser(me)
      })
      .catch(() => {
        if (!local) {
          router.replace('/login')
          return
        }
      })
      .finally(() => {
        setReady(true)
      })
  }, [router])

  return { ready, user }
}
