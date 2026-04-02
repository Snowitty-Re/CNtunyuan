import type { User } from '@/types/api'

const weights: Record<string, number> = {
  super_admin: 100,
  admin: 80,
  manager: 60,
  volunteer: 40,
}

export function roleWeight(role?: string): number {
  return weights[role || ''] || 0
}

export function hasMinRole(user: User | null, minRole: string): boolean {
  if (!user) return false
  return roleWeight(user.role) >= roleWeight(minRole)
}

export function isManager(user: User | null): boolean {
  return hasMinRole(user, 'manager')
}

export function isAdmin(user: User | null): boolean {
  return hasMinRole(user, 'admin')
}
