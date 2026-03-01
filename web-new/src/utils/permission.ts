import { useAuthStore } from '@/stores/auth';

// 角色权重
const roleWeights: Record<string, number> = {
  super_admin: 100,
  admin: 80,
  manager: 60,
  volunteer: 40,
};

// 角色标签
const roleLabels: Record<string, string> = {
  super_admin: '超级管理员',
  admin: '管理员',
  manager: '组织者',
  volunteer: '志愿者',
};

// 角色颜色
const roleColors: Record<string, string> = {
  super_admin: '#f5222d',
  admin: '#fa8c16',
  manager: '#1890ff',
  volunteer: '#52c41a',
};

export function getRoleLabel(role: string): string {
  return roleLabels[role] || role;
}

export function getRoleColor(role: string): string {
  return roleColors[role] || '#999';
}

export function usePermission() {
  const { user } = useAuthStore();
  
  const userRole = user?.role || 'volunteer';
  const userWeight = roleWeights[userRole] || 0;
  
  const hasRole = (roles: string | string[]) => {
    if (!user) return false;
    const roleList = Array.isArray(roles) ? roles : [roles];
    return roleList.includes(userRole);
  };
  
  const hasMinRole = (minRole: string) => {
    const minWeight = roleWeights[minRole] || 0;
    return userWeight >= minWeight;
  };
  
  const isSuperAdmin = userRole === 'super_admin';
  const isAdmin = hasMinRole('admin');
  const isManager = hasMinRole('manager');
  
  return {
    hasRole,
    hasMinRole,
    isSuperAdmin,
    isAdmin,
    isManager,
    userRole,
    getRoleLabel: () => getRoleLabel(userRole),
    getRoleColor: () => getRoleColor(userRole),
  };
}

export function checkPermission(userRole: string, requiredRole: string): boolean {
  const userWeight = roleWeights[userRole] || 0;
  const requiredWeight = roleWeights[requiredRole] || 0;
  return userWeight >= requiredWeight;
}
