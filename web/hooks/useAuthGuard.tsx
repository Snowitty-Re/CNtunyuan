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
  const { ready, user, logout, blockReason, blockMessage } = useAuth()

  useEffect(() => {
    if (!ready) return
    const token = getAccessToken()
    if (!token) {
      router.replace('/login')
      return
    }
    if (blockReason) {
      // 阻断页由各页面/壳层处理；此处仅保证不进入无权管理路由
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
  }, [ready, user, router, pathname, options?.requireAdmin, options?.requireManager, blockReason])

  return {
    ready: ready && !!getAccessToken(),
    user: user as User | null,
    logout,
    blocked: !!blockReason,
    blockReason,
    blockMessage,
  }
}
