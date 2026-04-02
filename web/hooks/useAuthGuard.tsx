'use client'

import { useEffect, useState } from 'react'
import { useRouter } from 'next/navigation'
import { getCurrentUser, getAccessToken } from '@/lib/auth'
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
    setUser(getCurrentUser())
    setReady(true)
  }, [router])

  return { ready, user }
}
