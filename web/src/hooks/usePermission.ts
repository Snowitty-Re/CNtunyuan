import { useAuthStore } from '@/store/auth'
import type { UserRole } from '@/types'

const roleWeight: Record<UserRole, number> = {
  super_admin: 100,
  admin: 80,
  manager: 60,
  volunteer: 40,
}

export function usePermission() {
  const user = useAuthStore((s) => s.user)
  const currentRole = user?.role || 'volunteer'

  const hasRole = (required: UserRole): boolean => {
    return roleWeight[currentRole] >= roleWeight[required]
  }

  const isAdmin = hasRole('admin')
  const isManager = hasRole('manager')

  return { user, currentRole, hasRole, isAdmin, isManager }
}
