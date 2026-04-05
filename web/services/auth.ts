import { http } from '@/lib/request'
import type { AuthLoginResponse, User, WechatLoginResponse } from '@/types/api'

export const authService = {
  login(username: string, password: string) {
    return http<AuthLoginResponse>('/auth/login', {
      method: 'POST',
      body: { username, password },
      auth: false,
    })
  },
  wechatLogin(code: string, nickname = '', avatar = '') {
    return http<AuthLoginResponse | WechatLoginResponse>('/auth/wechat-login', {
      method: 'POST',
      body: { code, nickname, avatar },
      auth: false,
    })
  },
  bindPhone(data: { phone?: string; code?: string; wechat_code?: string }) {
    return http<AuthLoginResponse>('/auth/bind-phone', {
      method: 'POST',
      body: data,
    })
  },
  sendCode(phone: string) {
    return http<{ message: string; expire?: number }>('/auth/send-code', {
      method: 'POST',
      body: { phone },
      auth: false,
    })
  },
  bindWechat(code: string) {
    return http<{ message?: string }>('/auth/bind-wechat', {
      method: 'POST',
      body: { code },
    })
  },
  unbindWechat() {
    return http<{ message?: string }>('/auth/unbind-wechat', {
      method: 'POST',
    })
  },
  wechatWebLogin(code: string) {
    return http<AuthLoginResponse>('/auth/wechat-web-login', {
      method: 'POST',
      body: { code },
      auth: false,
    })
  },
  me() {
    return http<User>('/auth/me')
  },
  updateProfile(data: {
    avatar?: string
    nickname?: string
    email?: string
    real_name?: string
    id_card?: string
    gender?: string
    address?: string
    emergency?: string
    emergency_tel?: string
    introduction?: string
  }) {
    return http<User>('/profile', {
      method: 'PUT',
      body: data,
    })
  },
  changePassword(oldPassword: string, newPassword: string) {
    return http<null>('/profile/password', {
      method: 'PUT',
      body: {
        old_password: oldPassword,
        new_password: newPassword,
      },
    })
  },
  logout() {
    return http<null>('/auth/logout', { method: 'POST' })
  },
}
