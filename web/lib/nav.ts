import type { User } from '@/types/api'
import { ACTIONS, hasPermission, isAdmin, isManager } from '@/lib/rbac'

export type WorkbenchKind = 'admin' | 'ops' | 'field'

export type NavItem = {
  href: string
  label: string
  /** backend-aligned gate; omit = any authenticated */
  action?: string
  /** extra role gate when action alone is insufficient */
  minRole?: 'admin' | 'manager'
}

const ADMIN_NAV: NavItem[] = [
  { href: '/dashboard', label: '系统概览' },
  { href: '/organizations', label: '组织管理', action: ACTIONS.ORG_MANAGE, minRole: 'admin' },
  { href: '/users', label: '人员管理', action: ACTIONS.USER_VIEW, minRole: 'admin' },
  { href: '/cases', label: '案件中心', action: ACTIONS.MISSING_MODIFY },
  { href: '/tasks', label: '任务中心', action: ACTIONS.TASK_VIEW },
  { href: '/dialects', label: '方言中心', action: ACTIONS.DIALECT_MODIFY },
  { href: '/dialects/cards', label: '方言卡片', action: ACTIONS.DIALECT_MANAGE, minRole: 'manager' },
  { href: '/attachments', label: '附件管理', minRole: 'admin' },
  { href: '/site-settings', label: '网站设置', minRole: 'admin' },
  { href: '/feature-settings', label: '功能设置', minRole: 'admin' },
  { href: '/monitor', label: '服务监控', minRole: 'admin' },
  { href: '/audit', label: '审计中心', minRole: 'admin' },
  { href: '/settings', label: '个人设置' },
]

const OPS_NAV: NavItem[] = [
  { href: '/dashboard', label: '运营工作台' },
  { href: '/cases', label: '案件中心', action: ACTIONS.MISSING_MODIFY },
  { href: '/tasks', label: '任务中心', action: ACTIONS.TASK_VIEW },
  { href: '/dialects', label: '方言中心', action: ACTIONS.DIALECT_MODIFY },
  { href: '/dialects/cards', label: '方言卡片', action: ACTIONS.DIALECT_MANAGE },
  { href: '/users', label: '人员查看', action: ACTIONS.USER_VIEW },
  { href: '/settings', label: '个人设置' },
]

const FIELD_NAV: NavItem[] = [
  { href: '/dashboard', label: '我的工作台' },
  { href: '/tasks', label: '我的任务', action: ACTIONS.TASK_VIEW },
  { href: '/cases', label: '案件查看', action: ACTIONS.MISSING_MODIFY },
  { href: '/dialects', label: '方言', action: ACTIONS.DIALECT_MODIFY },
  { href: '/settings', label: '个人设置' },
]

export function resolveWorkbench(user: User | null): WorkbenchKind {
  if (isAdmin(user)) return 'admin'
  if (isManager(user)) return 'ops'
  return 'field'
}

export function workbenchLabel(kind: WorkbenchKind): string {
  if (kind === 'admin') return '系统管理台'
  if (kind === 'ops') return '运营工作台'
  return '现场工作台'
}

export function defaultHomePath(user: User | null): string {
  return '/dashboard'
}

function passMinRole(user: User | null, minRole?: NavItem['minRole']): boolean {
  if (!minRole) return true
  if (minRole === 'admin') return isAdmin(user)
  if (minRole === 'manager') return isManager(user)
  return true
}

export function navItemsForUser(user: User | null): NavItem[] {
  const kind = resolveWorkbench(user)
  const source = kind === 'admin' ? ADMIN_NAV : kind === 'ops' ? OPS_NAV : FIELD_NAV
  return source.filter((item) => {
    if (!passMinRole(user, item.minRole)) return false
    if (item.action && !hasPermission(user, item.action)) return false
    return true
  })
}

/** Prefer longest matching path so /dialects/cards does not dual-activate /dialects */
export function isNavActive(pathname: string, href: string, allHrefs: string[]): boolean {
  if (pathname === href) return true
  if (!pathname.startsWith(`${href}/`)) return false
  const longerMatch = allHrefs.some(
    (other) => other !== href && other.startsWith(`${href}/`) && (pathname === other || pathname.startsWith(`${other}/`)),
  )
  return !longerMatch
}

/** Routes restricted to admin workbench (direct URL guard) */
export function isAdminOnlyPath(pathname: string): boolean {
  const prefixes = [
    '/organizations',
    '/attachments',
    '/site-settings',
    '/feature-settings',
    '/monitor',
    '/audit',
  ]
  return prefixes.some((p) => pathname === p || pathname.startsWith(`${p}/`))
}
