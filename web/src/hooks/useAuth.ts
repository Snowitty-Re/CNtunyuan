import { useAuthStore } from '@/stores/authStore';

export function useAuth() {
  const { user, loading, login, logout } = useAuthStore();
  return { user, loading, login, logout };
}
