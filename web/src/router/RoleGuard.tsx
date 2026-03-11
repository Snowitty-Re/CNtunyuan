import { Navigate } from 'react-router-dom';
import { useAuthStore } from '@/stores/authStore';
import { hasRole } from '@/utils/permission';
import type { UserRole } from '@/types';

interface Props {
  minRole: UserRole;
  children: React.ReactNode;
}

export default function RoleGuard({ minRole, children }: Props) {
  const user = useAuthStore((s) => s.user);
  const role = (user?.role || 'volunteer') as UserRole;

  if (!hasRole(role, minRole)) {
    return <Navigate to="/403" replace />;
  }

  return <>{children}</>;
}
