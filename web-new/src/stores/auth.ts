import { create } from 'zustand';
import { persist, createJSONStorage } from 'zustand/middleware';
import type { User } from '@/types';

interface AuthState {
  token: string | null;
  refreshToken: string | null;
  user: User | null;
  isAuthenticated: boolean;
  hasHydrated: boolean;
  
  // Actions
  setToken: (token: string, refreshToken: string) => void;
  setUser: (user: User) => void;
  logout: () => void;
  updateUser: (updates: Partial<User>) => void;
  setHasHydrated: (hasHydrated: boolean) => void;
  // 登录（一次性设置 token 和 user）
  login: (token: string, refreshToken: string, user: User) => void;
}

export const useAuthStore = create<AuthState>()(
  persist(
    (set) => ({
      token: null,
      refreshToken: null,
      user: null,
      isAuthenticated: false,
      hasHydrated: false,

      setToken: (token, refreshToken) =>
        set({ token, refreshToken, isAuthenticated: true }),

      setUser: (user) => set({ user }),

      logout: () =>
        set({
          token: null,
          refreshToken: null,
          user: null,
          isAuthenticated: false,
        }),

      updateUser: (updates) =>
        set((state) => ({
          user: state.user ? { ...state.user, ...updates } : null,
        })),

      setHasHydrated: (hasHydrated) => set({ hasHydrated }),

      login: (token, refreshToken, user) =>
        set({ token, refreshToken, user, isAuthenticated: true }),
    }),
    {
      name: 'auth-storage',
      storage: createJSONStorage(() => localStorage),
      partialize: (state) => ({
        token: state.token,
        refreshToken: state.refreshToken,
        user: state.user,
        isAuthenticated: state.isAuthenticated,
        // hasHydrated 不持久化，每次页面加载重新从 false 开始
      }),
      onRehydrateStorage: () => (_state, error) => {
        // 使用 setTimeout 确保在 store 创建后再调用 setHasHydrated
        if (!error) {
          setTimeout(() => {
            useAuthStore.setState({ hasHydrated: true });
          }, 0);
        }
      },
    }
  )
);
