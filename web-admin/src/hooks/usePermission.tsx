import { useMemo } from 'react';
import { useAuthStore } from '@/stores/auth';
import {
  RoleLevel,
  roleLevelMap,
  PermissionResource,
  PermissionAction,
  permissionMatrix,
} from '@/utils/permission';

// Re-export RoleLevel for convenience
export { RoleLevel, PermissionResource, PermissionAction } from '@/utils/permission';

/**
 * 权限检查Hook
 */
export function usePermission() {
  const user = useAuthStore((state) => state.user);

  const roleLevel = useMemo(() => {
    const role = user?.role;
    if (!role) return RoleLevel.Volunteer;
    return roleLevelMap[role] ?? RoleLevel.Volunteer;
  }, [user?.role]);

  const role = useMemo(() => user?.role || 'volunteer', [user?.role]);

  /**
   * 检查是否拥有指定角色级别
   */
  const hasRoleLevel = (minLevel: RoleLevel): boolean => {
    return roleLevel >= minLevel;
  };

  /**
   * 检查是否管理员
   */
  const isAdmin = useMemo(() => {
    return roleLevel >= RoleLevel.Admin;
  }, [roleLevel]);

  /**
   * 检查是否管理者
   */
  const isManager = useMemo(() => {
    return roleLevel >= RoleLevel.Manager;
  }, [roleLevel]);

  /**
   * 检查是否超级管理员
   */
  const isSuperAdmin = useMemo(() => {
    return roleLevel === RoleLevel.SuperAdmin;
  }, [roleLevel]);

  /**
   * 检查是否拥有指定权限
   */
  const hasPermission = (resource: PermissionResource, action: PermissionAction): boolean => {
    if (!role) return false;

    // 超级管理员拥有所有权限
    if (role === 'super_admin') return true;

    const permissions = permissionMatrix[role] || [];
    return permissions.some((p) => p.resource === resource && p.action === action);
  };

  /**
   * 检查是否拥有任一权限
   */
  const hasAnyPermission = (
    permissions: Array<{ resource: PermissionResource; action: PermissionAction }>
  ): boolean => {
    return permissions.some((p) => hasPermission(p.resource, p.action));
  };

  /**
   * 检查是否拥有所有权限
   */
  const hasAllPermissions = (
    permissions: Array<{ resource: PermissionResource; action: PermissionAction }>
  ): boolean => {
    return permissions.every((p) => hasPermission(p.resource, p.action));
  };

  return {
    role,
    roleLevel,
    isAdmin,
    isManager,
    isSuperAdmin,
    hasRoleLevel,
    hasPermission,
    hasAnyPermission,
    hasAllPermissions,
  };
}

/**
 * 权限组件Props
 */
interface PermissionProps {
  resource?: PermissionResource;
  action?: PermissionAction;
  minRole?: RoleLevel;
  any?: Array<{ resource: PermissionResource; action: PermissionAction }>;
  all?: Array<{ resource: PermissionResource; action: PermissionAction }>;
  children: React.ReactNode;
  fallback?: React.ReactNode;
}

/**
 * 权限包装组件
 * 根据权限控制子组件的显示
 */
export function PermissionGuard({
  resource,
  action,
  minRole,
  any: anyPermissions,
  all: allPermissions,
  children,
  fallback = null,
}: PermissionProps) {
  const { hasPermission, hasRoleLevel, hasAnyPermission, hasAllPermissions } = usePermission();

  const hasAccess = useMemo(() => {
    if (minRole !== undefined) {
      return hasRoleLevel(minRole);
    }

    if (resource && action) {
      return hasPermission(resource, action);
    }

    if (anyPermissions && anyPermissions.length > 0) {
      return hasAnyPermission(anyPermissions);
    }

    if (allPermissions && allPermissions.length > 0) {
      return hasAllPermissions(allPermissions);
    }

    return true;
  }, [
    minRole,
    resource,
    action,
    anyPermissions,
    allPermissions,
    hasRoleLevel,
    hasPermission,
    hasAnyPermission,
    hasAllPermissions,
  ]);

  if (!hasAccess) {
    return <>{fallback}</>;
  }

  return <>{children}</>;
}

/**
 * 管理员权限组件
 */
export function AdminOnly({
  children,
  fallback = null,
}: {
  children: React.ReactNode;
  fallback?: React.ReactNode;
}) {
  return (
    <PermissionGuard minRole={RoleLevel.Admin} fallback={fallback}>
      {children}
    </PermissionGuard>
  );
}

/**
 * 管理者权限组件
 */
export function ManagerOnly({
  children,
  fallback = null,
}: {
  children: React.ReactNode;
  fallback?: React.ReactNode;
}) {
  return (
    <PermissionGuard minRole={RoleLevel.Manager} fallback={fallback}>
      {children}
    </PermissionGuard>
  );
}

/**
 * 超级管理员权限组件
 */
export function SuperAdminOnly({
  children,
  fallback = null,
}: {
  children: React.ReactNode;
  fallback?: React.ReactNode;
}) {
  return (
    <PermissionGuard minRole={RoleLevel.SuperAdmin} fallback={fallback}>
      {children}
    </PermissionGuard>
  );
}

export default usePermission;
