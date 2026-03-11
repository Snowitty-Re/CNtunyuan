import { useAuthStore } from '@/stores/authStore';
import { hasRole as checkRole, isAdmin as checkAdmin, isManager as checkManager, isSuperAdmin as checkSuperAdmin } from '@/utils/permission';
import type { UserRole } from '@/types';

export function usePermission() {
  const user = useAuthStore((s) => s.user);
  const role = user?.role || 'volunteer';

  return {
    role,
    hasRole: (r: UserRole) => checkRole(role, r),
    isAdmin: checkAdmin(role),
    isManager: checkManager(role),
    isSuperAdmin: checkSuperAdmin(role),
  };
}
