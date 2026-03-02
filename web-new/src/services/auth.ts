import { http } from '@/utils/request';
import type { User } from '@/types';

export interface LoginRequest {
  phone: string;
  password: string;
}

export interface LoginResponse {
  token: string;
  refresh_token: string;
  user: User;
}

export const authApi = {
  // 用户登录
  login: (data: LoginRequest) => http.post<LoginResponse>('/auth/login', data),
  
  // 管理员登录
  adminLogin: (data: LoginRequest) => http.post<LoginResponse>('/auth/admin-login', data),
  
  // 获取当前用户信息
  getCurrentUser: () => http.get<User>('/auth/me'),
  
  // 刷新token
  refreshToken: (refreshToken: string) => http.post<LoginResponse>('/auth/refresh', { refresh_token: refreshToken }),
  
  // 退出登录
  logout: () => http.post('/auth/logout'),
};
