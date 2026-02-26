import request from '../utils/request'
import type { User, PageData } from '../types'

export const userApi = {
  // 获取当前用户
  getCurrentUser: () => request.get('/auth/me'),

  // 用户登录
  login: (code: string) => request.post('/auth/wechat-login', { code }),

  // 管理后台登录
  adminLogin: (phone: string, password: string) => request.post('/auth/admin-login', { phone, password }),

  // 刷新token
  refreshToken: (refreshToken: string) => request.post('/auth/refresh', { refresh_token: refreshToken }),

  // 获取用户列表
  getUsers: (params: { page?: number; page_size?: number; role?: string; status?: string }) =>
    request.get<PageData<User>>('/users', { params }),

  // 获取用户详情
  getUser: (id: string) => request.get<User>(`/users/${id}`),

  // 更新用户
  updateUser: (id: string, data: Partial<User>) => request.put(`/users/${id}`, data),

  // 删除用户
  deleteUser: (id: string) => request.delete(`/users/${id}`),

  // 获取用户统计
  getStatistics: () => request.get('/users/statistics'),

  // 分配用户到组织
  assignToOrg: (userId: string, orgId: string) =>
    request.post('/users/assign-to-org', { user_id: userId, org_id: orgId }),
}
