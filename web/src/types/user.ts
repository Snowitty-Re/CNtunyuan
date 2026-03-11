import type { UserRole, UserStatus } from './enums';
import type { PaginationParams } from './api';

export interface UserResponse {
  id: string;
  nickname: string;
  phone: string;
  email: string;
  role: UserRole;
  status: UserStatus;
  org_id: string;
  org_name?: string;
  avatar: string;
  last_login_at?: string;
  created_at: string;
}

export interface CreateUserRequest {
  nickname: string;
  phone: string;
  email?: string;
  password: string;
  role: UserRole;
  org_id: string;
}

export interface UpdateUserRequest {
  nickname?: string;
  email?: string;
  role?: UserRole;
  org_id?: string;
  status?: UserStatus;
}

export interface UserListRequest extends PaginationParams {
  keyword?: string;
  role?: string;
  status?: string;
  org_id?: string;
}

export interface ChangePasswordRequest {
  old_password: string;
  new_password: string;
}

export interface UpdateProfileRequest {
  nickname?: string;
  email?: string;
  real_name?: string;
  id_card?: string;
  gender?: string;
  address?: string;
  emergency?: string;
  emergency_tel?: string;
  introduction?: string;
}

export interface UserProfileResponse extends UserResponse {
  real_name: string;
  id_card: string;
  gender: string;
  address: string;
  emergency: string;
  emergency_tel: string;
  introduction: string;
}

export interface UserStatsResponse {
  total_cases: number;
  active_cases: number;
  completed_cases: number;
  total_tasks: number;
  pending_tasks: number;
}
