import { create } from 'zustand';
import type { UserResponse } from '@/types';
import { adminLogin, getCurrentUser, logout as logoutApi } from '@/api/auth';
import { setTokens, clearTokens, hasToken } from '@/utils/token';

interface AuthState {
  user: UserResponse | null;
  loading: boolean;
  initialized: boolean;
  login: (username: string, password: string) => Promise<void>;
  logout: () => Promise<void>;
  fetchUser: () => Promise<void>;
  init: () => Promise<void>;
}

export const useAuthStore = create<AuthState>((set, get) => ({
  user: null,
  loading: false,
  initialized: false,

  login: async (username, password) => {
    set({ loading: true });
    try {
      const res = await adminLogin({ username, password });
      setTokens(res.access_token, res.refresh_token);
      set({ user: res.user, loading: false });
    } catch {
      set({ loading: false });
      throw new Error('登录失败');
    }
  },

  logout: async () => {
    try {
      await logoutApi();
    } catch {
      // ignore
    }
    clearTokens();
    set({ user: null });
  },

  fetchUser: async () => {
    try {
      const user = await getCurrentUser();
      set({ user });
    } catch {
      clearTokens();
      set({ user: null });
    }
  },

  init: async () => {
    if (get().initialized) return;
    if (hasToken()) {
      await get().fetchUser();
    }
    set({ initialized: true });
  },
}));
