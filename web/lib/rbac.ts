import type { User } from '@/types/api'

const weights: Record<string, number> = {
  super_admin: 100,
  admin: 80,
  manager: 60,
  volunteer: 40,
}

export const ACTIONS = {
  USER_CREATE: 'user:create',
  USER_MODIFY: 'user:modify',
  USER_VIEW: 'user:view',
  TASK_MANAGE: 'task:manage',
  TASK_VIEW: 'task:view',
  TASK_EDIT: 'task:edit',
  TASK_EXECUTE: 'task:execute',
  MISSING_MODIFY: 'missing:modify',
  MISSING_MANAGE: 'missing:manage',
  DIALECT_MODIFY: 'dialect:modify',
  DIALECT_MANAGE: 'dialect:manage',
  ORG_MANAGE: 'org:manage',
} as const

const ROLE_PERMISSIONS: Record<string, string[]> = {
  super_admin: [
    ACTIONS.USER_CREATE,
    ACTIONS.USER_MODIFY,
    ACTIONS.USER_VIEW,
    ACTIONS.TASK_MANAGE,
    ACTIONS.TASK_VIEW,
    ACTIONS.TASK_EDIT,
    ACTIONS.TASK_EXECUTE,
    ACTIONS.MISSING_MODIFY,
    ACTIONS.MISSING_MANAGE,
    ACTIONS.DIALECT_MODIFY,
    ACTIONS.DIALECT_MANAGE,
    ACTIONS.ORG_MANAGE,
  ],
  admin: [
    ACTIONS.USER_CREATE,
    ACTIONS.USER_MODIFY,
    ACTIONS.USER_VIEW,
    ACTIONS.TASK_MANAGE,
    ACTIONS.TASK_VIEW,
    ACTIONS.TASK_EDIT,
    ACTIONS.TASK_EXECUTE,
    ACTIONS.MISSING_MODIFY,
    ACTIONS.MISSING_MANAGE,
    ACTIONS.DIALECT_MODIFY,
    ACTIONS.DIALECT_MANAGE,
    ACTIONS.ORG_MANAGE,
  ],
  manager: [
    ACTIONS.USER_MODIFY,
    ACTIONS.USER_VIEW,
    ACTIONS.TASK_MANAGE,
    ACTIONS.TASK_VIEW,
    ACTIONS.TASK_EDIT,
    ACTIONS.TASK_EXECUTE,
    ACTIONS.MISSING_MODIFY,
    ACTIONS.MISSING_MANAGE,
    ACTIONS.DIALECT_MODIFY,
    ACTIONS.DIALECT_MANAGE,
    ACTIONS.ORG_MANAGE,
  ],
  volunteer: [
    ACTIONS.USER_VIEW,
    ACTIONS.USER_MODIFY,
    ACTIONS.TASK_VIEW,
    ACTIONS.TASK_EDIT,
    ACTIONS.TASK_EXECUTE,
    ACTIONS.MISSING_MODIFY,
    ACTIONS.DIALECT_MODIFY,
  ],
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

function collectExplicitPermissions(user: User | null): Set<string> {
  const set = new Set<string>()
  if (!user) return set
  const candidates = [
    user.permissions,
    user.permission_codes,
    user.permissionCodes,
    user.effective_permissions,
    user.effectivePermissions,
    user.authz?.permissions,
    user.authz?.permission_codes,
  ]
  candidates.forEach((items) => {
    if (!Array.isArray(items)) return
    items.forEach((code) => {
      const normalized = String(code || '').trim()
      if (normalized) set.add(normalized)
    })
  })
  return set
}

export function resolvePermissionSet(user: User | null): Set<string> {
  const explicit = collectExplicitPermissions(user)
  if (explicit.size > 0) return explicit
  if (!user) return explicit
  const fallback = ROLE_PERMISSIONS[user.role || ''] || []
  return new Set(fallback)
}

export function hasPermission(user: User | null, action: string): boolean {
  const permission = String(action || '').trim()
  if (!permission) return false
  if (permission === ACTIONS.ORG_MANAGE) return isManager(user)
  const set = resolvePermissionSet(user)
  return set.has(permission)
}

export function canAny(user: User | null, actions: string[]): boolean {
  if (!Array.isArray(actions) || actions.length === 0) return false
  return actions.some((action) => hasPermission(user, action))
}

export function canAll(user: User | null, actions: string[]): boolean {
  if (!Array.isArray(actions) || actions.length === 0) return false
  return actions.every((action) => hasPermission(user, action))
}
