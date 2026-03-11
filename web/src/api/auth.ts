import request from './request';
import type { LoginRequest, LoginResponse, SendCodeResponse, ResetPasswordRequest } from '@/types';
import type { UserResponse } from '@/types';

export function login(data: LoginRequest) {
  return request.post<never, LoginResponse>('/auth/login', data);
}

export function adminLogin(data: LoginRequest) {
  return request.post<never, LoginResponse>('/auth/admin-login', data);
}

export function refreshToken(refresh_token: string) {
  return request.post<never, LoginResponse>('/auth/refresh', { refresh_token });
}

export function logout() {
  return request.post<never, null>('/auth/logout');
}

export function getCurrentUser() {
  return request.get<never, UserResponse>('/auth/me');
}

export function sendCode(phone: string) {
  return request.post<never, SendCodeResponse>('/auth/send-code', { phone });
}

export function resetPassword(data: ResetPasswordRequest) {
  return request.post<never, { message: string }>('/auth/reset-password', data);
}
