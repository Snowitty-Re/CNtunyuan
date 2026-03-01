import { useAuthStore } from '@/stores/auth';

/**
 * 角色级别
 */
export enum RoleLevel {
  Volunteer = 0,   // 志愿者
  Manager = 1,     // 管理者
  Admin = 2,       // 管理员
  SuperAdmin = 3,  // 超级管理员
}

/**
 * 角色映射
 */
export const roleLevelMap: Record<string, RoleLevel> = {
  volunteer: RoleLevel.Volunteer,
  manager: RoleLevel.Manager,
  admin: RoleLevel.Admin,
  super_admin: RoleLevel.SuperAdmin,
};

/**
 * 权限资源类型
 */
export enum PermissionResource {
  User = 'user',
  Org = 'org',
  Task = 'task',
  Case = 'case',
  Dialect = 'dialect',
  Workflow = 'workflow',
  Log = 'log',
}

/**
 * 权限操作类型
 */
export enum PermissionAction {
  Create = 'create',
  Read = 'read',
  Update = 'update',
  Delete = 'delete',
  Approve = 'approve',
  Assign = 'assign',
}

/**
 * 权限矩阵
 */
export const permissionMatrix: Record<string, Array<{ resource: PermissionResource; action: PermissionAction }>> = {
  super_admin: [
    { resource: PermissionResource.User, action: PermissionAction.Create },
    { resource: PermissionResource.User, action: PermissionAction.Read },
    { resource: PermissionResource.User, action: PermissionAction.Update },
    { resource: PermissionResource.User, action: PermissionAction.Delete },
    { resource: PermissionResource.Org, action: PermissionAction.Create },
    { resource: PermissionResource.Org, action: PermissionAction.Read },
    { resource: PermissionResource.Org, action: PermissionAction.Update },
    { resource: PermissionResource.Org, action: PermissionAction.Delete },
    { resource: PermissionResource.Task, action: PermissionAction.Create },
    { resource: PermissionResource.Task, action: PermissionAction.Read },
    { resource: PermissionResource.Task, action: PermissionAction.Update },
    { resource: PermissionResource.Task, action: PermissionAction.Delete },
    { resource: PermissionResource.Task, action: PermissionAction.Assign },
    { resource: PermissionResource.Case, action: PermissionAction.Create },
    { resource: PermissionResource.Case, action: PermissionAction.Read },
    { resource: PermissionResource.Case, action: PermissionAction.Update },
    { resource: PermissionResource.Case, action: PermissionAction.Delete },
    { resource: PermissionResource.Dialect, action: PermissionAction.Create },
    { resource: PermissionResource.Dialect, action: PermissionAction.Read },
    { resource: PermissionResource.Dialect, action: PermissionAction.Update },
    { resource: PermissionResource.Dialect, action: PermissionAction.Delete },
    { resource: PermissionResource.Workflow, action: PermissionAction.Create },
    { resource: PermissionResource.Workflow, action: PermissionAction.Read },
    { resource: PermissionResource.Workflow, action: PermissionAction.Update },
    { resource: PermissionResource.Workflow, action: PermissionAction.Delete },
    { resource: PermissionResource.Workflow, action: PermissionAction.Approve },
    { resource: PermissionResource.Log, action: PermissionAction.Read },
    { resource: PermissionResource.Log, action: PermissionAction.Delete },
  ],
  admin: [
    { resource: PermissionResource.User, action: PermissionAction.Create },
    { resource: PermissionResource.User, action: PermissionAction.Read },
    { resource: PermissionResource.User, action: PermissionAction.Update },
    { resource: PermissionResource.User, action: PermissionAction.Delete },
    { resource: PermissionResource.Org, action: PermissionAction.Read },
    { resource: PermissionResource.Org, action: PermissionAction.Update },
    { resource: PermissionResource.Task, action: PermissionAction.Create },
    { resource: PermissionResource.Task, action: PermissionAction.Read },
    { resource: PermissionResource.Task, action: PermissionAction.Update },
    { resource: PermissionResource.Task, action: PermissionAction.Delete },
    { resource: PermissionResource.Task, action: PermissionAction.Assign },
    { resource: PermissionResource.Case, action: PermissionAction.Create },
    { resource: PermissionResource.Case, action: PermissionAction.Read },
    { resource: PermissionResource.Case, action: PermissionAction.Update },
    { resource: PermissionResource.Case, action: PermissionAction.Delete },
    { resource: PermissionResource.Dialect, action: PermissionAction.Create },
    { resource: PermissionResource.Dialect, action: PermissionAction.Read },
    { resource: PermissionResource.Dialect, action: PermissionAction.Update },
    { resource: PermissionResource.Dialect, action: PermissionAction.Delete },
    { resource: PermissionResource.Workflow, action: PermissionAction.Create },
    { resource: PermissionResource.Workflow, action: PermissionAction.Read },
    { resource: PermissionResource.Workflow, action: PermissionAction.Update },
    { resource: PermissionResource.Workflow, action: PermissionAction.Delete },
    { resource: PermissionResource.Workflow, action: PermissionAction.Approve },
  ],
  manager: [
    { resource: PermissionResource.User, action: PermissionAction.Read },
    { resource: PermissionResource.User, action: PermissionAction.Update },
    { resource: PermissionResource.Org, action: PermissionAction.Read },
    { resource: PermissionResource.Org, action: PermissionAction.Update },
    { resource: PermissionResource.Task, action: PermissionAction.Create },
    { resource: PermissionResource.Task, action: PermissionAction.Read },
    { resource: PermissionResource.Task, action: PermissionAction.Update },
    { resource: PermissionResource.Task, action: PermissionAction.Assign },
    { resource: PermissionResource.Case, action: PermissionAction.Create },
    { resource: PermissionResource.Case, action: PermissionAction.Read },
    { resource: PermissionResource.Case, action: PermissionAction.Update },
    { resource: PermissionResource.Dialect, action: PermissionAction.Create },
    { resource: PermissionResource.Dialect, action: PermissionAction.Read },
    { resource: PermissionResource.Dialect, action: PermissionAction.Update },
    { resource: PermissionResource.Workflow, action: PermissionAction.Read },
    { resource: PermissionResource.Workflow, action: PermissionAction.Approve },
  ],
  volunteer: [
    { resource: PermissionResource.User, action: PermissionAction.Read },
    { resource: PermissionResource.User, action: PermissionAction.Update },
    { resource: PermissionResource.Org, action: PermissionAction.Read },
    { resource: PermissionResource.Task, action: PermissionAction.Read },
    { resource: PermissionResource.Task, action: PermissionAction.Update },
    { resource: PermissionResource.Case, action: PermissionAction.Create },
    { resource: PermissionResource.Case, action: PermissionAction.Read },
    { resource: PermissionResource.Dialect, action: PermissionAction.Create },
    { resource: PermissionResource.Dialect, action: PermissionAction.Read },
    { resource: PermissionResource.Dialect, action: PermissionAction.Update },
  ],
};

/**
 * 获取当前用户角色级别
 */
export function getCurrentRoleLevel(): RoleLevel {
  const authStore = useAuthStore.getState();
  const role = authStore.user?.role;
  if (!role) return RoleLevel.Volunteer;
  return roleLevelMap[role] ?? RoleLevel.Volunteer;
}

/**
 * 检查是否拥有指定角色级别
 */
export function hasRoleLevel(minLevel: RoleLevel): boolean {
  return getCurrentRoleLevel() >= minLevel;
}

/**
 * 检查是否管理员
 */
export function isAdmin(): boolean {
  const level = getCurrentRoleLevel();
  return level >= RoleLevel.Admin;
}

/**
 * 检查是否管理者（管理员或管理者）
 */
export function isManager(): boolean {
  const level = getCurrentRoleLevel();
  return level >= RoleLevel.Manager;
}

/**
 * 检查是否超级管理员
 */
export function isSuperAdmin(): boolean {
  return getCurrentRoleLevel() === RoleLevel.SuperAdmin;
}

/**
 * 检查是否拥有指定权限
 */
export function hasPermission(resource: PermissionResource, action: PermissionAction): boolean {
  const authStore = useAuthStore.getState();
  const role = authStore.user?.role;
  if (!role) return false;

  // 超级管理员拥有所有权限
  if (role === 'super_admin') return true;

  const permissions = permissionMatrix[role] || [];
  return permissions.some(p => p.resource === resource && p.action === action);
}

/**
 * 权限检查Hook（用于React组件）
 */
export function usePermission() {
  return {
    roleLevel: getCurrentRoleLevel(),
    isAdmin: isAdmin(),
    isManager: isManager(),
    isSuperAdmin: isSuperAdmin(),
    hasRoleLevel,
    hasPermission,
  };
}

/**
 * 路由权限配置
 */
export const routePermissions: Record<string, { minRole?: RoleLevel; permission?: { resource: PermissionResource; action: PermissionAction } }> = {
  '/users': { minRole: RoleLevel.Admin },
  '/organizations': { minRole: RoleLevel.Volunteer },
  '/cases': { minRole: RoleLevel.Volunteer },
  '/tasks': { minRole: RoleLevel.Volunteer },
  '/tasks/create': { minRole: RoleLevel.Volunteer },
  '/workflows': { minRole: RoleLevel.Volunteer },
  '/workflows/create': { minRole: RoleLevel.Admin },
  '/dialects': { minRole: RoleLevel.Volunteer },
  '/settings': { minRole: RoleLevel.Admin },
  '/logs': { minRole: RoleLevel.SuperAdmin },
};

/**
 * 检查路由权限
 */
export function checkRoutePermission(path: string): boolean {
  const config = routePermissions[path];
  if (!config) return true;

  if (config.minRole !== undefined) {
    return hasRoleLevel(config.minRole);
  }

  if (config.permission) {
    return hasPermission(config.permission.resource, config.permission.action);
  }

  return true;
}
