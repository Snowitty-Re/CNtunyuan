import request from './request';
import type { PaginatedData, UserResponse, UserListRequest, CreateUserRequest, UpdateUserRequest, UserProfileResponse, UpdateProfileRequest, ChangePasswordRequest, UserStatsResponse } from '@/types';

export function getUsers(params: UserListRequest) {
  return request.get<never, PaginatedData<UserResponse>>('/users', { params });
}

export function getUser(id: string) {
  return request.get<never, UserResponse>(`/users/${id}`);
}

export function createUser(data: CreateUserRequest) {
  return request.post<never, UserResponse>('/users', data);
}

export function updateUser(id: string, data: UpdateUserRequest) {
  return request.put<never, UserResponse>(`/users/${id}`, data);
}

export function deleteUser(id: string) {
  return request.delete<never, null>(`/users/${id}`);
}

export function updateUserStatus(id: string, status: string) {
  return request.put<never, null>(`/users/${id}/status`, { status });
}

export function updateUserRole(id: string, role: string) {
  return request.put<never, null>(`/users/${id}/role`, { role });
}

export function getProfile() {
  return request.get<never, UserProfileResponse>('/profile');
}

export function updateProfile(data: UpdateProfileRequest) {
  return request.put<never, UserProfileResponse>('/profile', data);
}

export function changePassword(data: ChangePasswordRequest) {
  return request.put<never, null>('/profile/password', data);
}

export function getProfileStats() {
  return request.get<never, UserStatsResponse>('/profile/stats');
}
