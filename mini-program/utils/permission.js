/**
 * Client RBAC aligned with backend route middleware:
 * - org write / user create-update: admin+
 * - user list/status, task manage, case manage: manager+
 * - case/dialect modify: volunteer+
 */
const ACTIONS = {
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
  ORG_MANAGE: 'org:manage'
}

const ROLE_LEVEL = {
  volunteer: 40,
  manager: 60,
  admin: 80,
  super_admin: 100
}

const ROLE_PERMISSIONS = {
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
    ACTIONS.ORG_MANAGE
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
    ACTIONS.ORG_MANAGE
  ],
  manager: [
    ACTIONS.USER_VIEW,
    ACTIONS.TASK_MANAGE,
    ACTIONS.TASK_VIEW,
    ACTIONS.TASK_EDIT,
    ACTIONS.TASK_EXECUTE,
    ACTIONS.MISSING_MODIFY,
    ACTIONS.MISSING_MANAGE,
    ACTIONS.DIALECT_MODIFY,
    ACTIONS.DIALECT_MANAGE
  ],
  volunteer: [
    ACTIONS.TASK_VIEW,
    ACTIONS.TASK_EDIT,
    ACTIONS.TASK_EXECUTE,
    ACTIONS.MISSING_MODIFY,
    ACTIONS.DIALECT_MODIFY
  ]
}

function getUserRole(userInfo = {}) {
  return String((userInfo && userInfo.role) || '').trim()
}

function getRoleLevel(role = '') {
  return ROLE_LEVEL[String(role).trim()] || 0
}

function isManagerRole(userInfo = {}) {
  return getRoleLevel(getUserRole(userInfo)) >= ROLE_LEVEL.manager
}

function isAdminRole(userInfo = {}) {
  return getRoleLevel(getUserRole(userInfo)) >= ROLE_LEVEL.admin
}

function isSuperAdminRole(userInfo = {}) {
  return getUserRole(userInfo) === 'super_admin'
}

/** Align with backend canAssignRole: strict higher level; only super_admin may assign super_admin */
function canAssignRole(operator = {}, targetRole = '') {
  const role = String(targetRole || '').trim()
  if (!role) return false
  if (role === 'super_admin') return isSuperAdminRole(operator)
  return getRoleLevel(getUserRole(operator)) > getRoleLevel(role)
}

const ALL_ASSIGNABLE_ROLES = [
  { key: 'volunteer', label: '志愿者' },
  { key: 'manager', label: '管理者' },
  { key: 'admin', label: '管理员' },
  { key: 'super_admin', label: '超级管理员' }
]

function assignableRoles(operator = {}) {
  return ALL_ASSIGNABLE_ROLES.filter((item) => canAssignRole(operator, item.key))
}

function collectExplicitPermissions(userInfo = {}) {
  const candidates = [
    userInfo.permissions,
    userInfo.permission_codes,
    userInfo.permissionCodes,
    userInfo.effective_permissions,
    userInfo.effectivePermissions,
    userInfo.authz && userInfo.authz.permissions,
    userInfo.authz && userInfo.authz.permission_codes
  ]
  const set = new Set()
  candidates.forEach((items) => {
    if (!Array.isArray(items)) return
    items.forEach((code) => {
      const normalized = String(code || '').trim()
      if (normalized) set.add(normalized)
    })
  })
  return set
}

function resolvePermissionSet(userInfo = {}) {
  const explicit = collectExplicitPermissions(userInfo)
  if (explicit.size > 0) return explicit

  const role = getUserRole(userInfo)
  const fallback = ROLE_PERMISSIONS[role] || []
  return new Set(fallback)
}

function hasPermission(userInfo = {}, action = '') {
  const permission = String(action || '').trim()
  if (!permission) return false
  // Backend org write + user create/update require admin
  if (permission === ACTIONS.ORG_MANAGE) return isAdminRole(userInfo)
  if (permission === ACTIONS.USER_CREATE) return isAdminRole(userInfo)
  if (permission === ACTIONS.USER_MODIFY) return isAdminRole(userInfo)

  const set = resolvePermissionSet(userInfo)
  return set.has(permission)
}

function canAny(userInfo = {}, actions = []) {
  if (!Array.isArray(actions) || actions.length === 0) return false
  return actions.some((action) => hasPermission(userInfo, action))
}

function canAll(userInfo = {}, actions = []) {
  if (!Array.isArray(actions) || actions.length === 0) return false
  return actions.every((action) => hasPermission(userInfo, action))
}

module.exports = {
  ACTIONS,
  ROLE_LEVEL,
  ROLE_PERMISSIONS,
  getUserRole,
  getRoleLevel,
  isManagerRole,
  isAdminRole,
  isSuperAdminRole,
  canAssignRole,
  assignableRoles,
  hasPermission,
  canAny,
  canAll,
  resolvePermissionSet
}
