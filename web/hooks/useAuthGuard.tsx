'use client'

import { useEffect } from 'react'
import { usePathname, useRouter } from 'next/navigation'
import { useAuth } from '@/components/providers/AuthProvider'
import { getAccessToken } from '@/lib/auth'
import { isAdmin, isManager } from '@/lib/rbac'
import { isAdminOnlyPath } from '@/lib/nav'
import type { User } from '@/types/api'

export function useAuthGuard(options?: { requireAdmin?: boolean; requireManager?: boolean }) {
  const router = useRouter()
  const pathname = usePathname()
  const { ready, user, logout } = useAuth()

  useEffect(() => {
    if (!ready) return
    const token = getAccessToken()
    if (!token) {
      router.replace('/login')
      return
    }
    if (options?.requireAdmin && !isAdmin(user)) {
      router.replace('/dashboard')
      return
    }
    if (options?.requireManager && !isManager(user)) {
      router.replace('/dashboard')
      return
    }
    if (isAdminOnlyPath(pathname) && !isAdmin(user)) {
      router.replace('/dashboard')
    }
  }, [ready, user, router, pathname, options?.requireAdmin, options?.requireManager])

  return { ready: ready && !!getAccessToken(), user: user as User | null, logout }
}
