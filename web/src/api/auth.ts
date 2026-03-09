import request from './request'
import type { ApiResponse, LoginRequest, LoginResponse, WechatLoginRequest, WechatLoginResponse, SendCodeRequest, BindPhoneRequest, User } from '@/types'

export const authApi = {
  login: (data: LoginRequest) =>
    request.post<ApiResponse<LoginResponse>>('/auth/admin-login', data),

  wechatLogin: (data: WechatLoginRequest) =>
    request.post<ApiResponse<WechatLoginResponse>>('/auth/wechat-login', data),

  refresh: (refresh_token: string) =>
    request.post<ApiResponse<LoginResponse>>('/auth/refresh', { refresh_token }),

  logout: () =>
    request.post<ApiResponse<null>>('/auth/logout'),

  sendCode: (data: SendCodeRequest) =>
    request.post<ApiResponse<null>>('/auth/send-code', data),

  bindPhone: (data: BindPhoneRequest) =>
    request.post<ApiResponse<LoginResponse>>('/auth/bind-phone', data),

  getMe: () =>
    request.get<ApiResponse<User>>('/auth/me'),
}
