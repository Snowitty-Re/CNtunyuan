import { http } from '@/lib/request'
import type { AuthLoginResponse, User } from '@/types/api'

export const authService = {
  login(username: string, password: string) {
    return http<AuthLoginResponse>('/auth/login', {
      method: 'POST',
      body: { username, password },
      auth: false,
    })
  },
  me() {
    return http<User>('/auth/me')
  },
  logout() {
    return http<null>('/auth/logout', { method: 'POST' })
  },
}
