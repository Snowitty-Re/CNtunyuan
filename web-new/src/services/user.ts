import { http } from '@/utils/request';
import type { User, UserProfile } from '@/types';

export interface UpdateUserRequest {
  nickname?: string;
  real_name?: string;
  avatar?: string;
  email?: string;
  gender?: string;
  address?: string;
}

export interface ChangePasswordRequest {
  old_password: string;
  new_password: string;
}

export const userApi = {
  // 获取用户列表
  getUsers: (params?: { page?: number; page_size?: number; keyword?: string; org_id?: string }) => 
    http.get<{ list: User[]; total: number }>('/users', { params }),
  
  // 获取用户详情
  getUserById: (id: string) => http.get<User>(`/users/${id}`),
  
  // 创建用户
  createUser: (data: Partial<User>) => http.post<User>('/users', data),
  
  // 更新用户
  updateUser: (id: string, data: UpdateUserRequest) => http.put<User>(`/users/${id}`, data),
  
  // 删除用户
  deleteUser: (id: string) => http.delete(`/users/${id}`),
  
  // 更新用户状态
  updateUserStatus: (id: string, status: string) => http.put(`/users/${id}/status`, { status }),
  
  // 更新用户角色
  updateUserRole: (id: string, role: string) => http.put(`/users/${id}/role`, { role }),
  
  // 获取当前用户资料
  getProfile: () => http.get<UserProfile>('/profile'),
  
  // 更新当前用户资料
  updateProfile: (data: UpdateUserRequest) => http.put<UserProfile>('/profile', data),
  
  // 修改密码
  changePassword: (data: ChangePasswordRequest) => http.put('/profile/password', data),
};
