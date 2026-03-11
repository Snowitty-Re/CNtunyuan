import type { UserRole } from '@/types';

const ROLE_HIERARCHY: Record<UserRole, number> = {
  super_admin: 4,
  admin: 3,
  manager: 2,
  volunteer: 1,
};

export function hasRole(userRole: UserRole, requiredRole: UserRole): boolean {
  return (ROLE_HIERARCHY[userRole] || 0) >= (ROLE_HIERARCHY[requiredRole] || 0);
}

export function isAdmin(role: UserRole): boolean {
  return hasRole(role, 'admin');
}

export function isManager(role: UserRole): boolean {
  return hasRole(role, 'manager');
}

export function isSuperAdmin(role: UserRole): boolean {
  return role === 'super_admin';
}
