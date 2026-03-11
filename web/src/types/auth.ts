import type { UserResponse } from './user';

export interface LoginRequest {
  username: string;
  password: string;
}

export interface LoginResponse {
  access_token: string;
  refresh_token: string;
  expires_in: number;
  token_type: string;
  user: UserResponse;
}

export interface RefreshTokenRequest {
  refresh_token: string;
}

export interface SendCodeRequest {
  phone: string;
}

export interface SendCodeResponse {
  message: string;
  expire: number;
}

export interface ResetPasswordRequest {
  phone: string;
  code: string;
  new_password: string;
}
