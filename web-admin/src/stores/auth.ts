import { create } from 'zustand'
import { persist } from 'zustand/middleware'

interface AuthState {
  token: string | null
  refreshToken: string | null
  user: any | null
  setToken: (token: string, refreshToken: string) => void
  setUser: (user: any) => void
  logout: () => void
}

export const useAuthStore = create<AuthState>()(
  persist(
    (set) => ({
      token: null,
      refreshToken: null,
      user: null,
      setToken: (token, refreshToken) => set({ token, refreshToken }),
      setUser: (user) => set({ user }),
      logout: () => set({ token: null, refreshToken: null, user: null }),
    }),
    {
      name: 'auth-storage',
    }
  )
)
