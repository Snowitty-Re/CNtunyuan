import request from './request'
import type { ApiResponse, PageData, User, UserProfile, UserQuery, CreateUserRequest, UpdateUserRequest, UpdateProfileRequest, ChangePasswordRequest } from '@/types'

export const userApi = {
  list: (params: UserQuery) =>
    request.get<ApiResponse<PageData<User>>>('/users', { params }),

  getById: (id: string) =>
    request.get<ApiResponse<User>>(`/users/${id}`),

  create: (data: CreateUserRequest) =>
    request.post<ApiResponse<User>>('/users', data),

  update: (id: string, data: UpdateUserRequest) =>
    request.put<ApiResponse<User>>(`/users/${id}`, data),

  delete: (id: string) =>
    request.delete<ApiResponse<null>>(`/users/${id}`),

  updateStatus: (id: string, status: string) =>
    request.put<ApiResponse<null>>(`/users/${id}/status`, { status }),

  updateRole: (id: string, role: string) =>
    request.put<ApiResponse<null>>(`/users/${id}/role`, { role }),

  getProfile: () =>
    request.get<ApiResponse<UserProfile>>('/profile'),

  updateProfile: (data: UpdateProfileRequest) =>
    request.put<ApiResponse<UserProfile>>('/profile', data),

  changePassword: (data: ChangePasswordRequest) =>
    request.put<ApiResponse<null>>('/profile/password', data),
}
