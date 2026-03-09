import request from './request'
import type { ApiResponse, PageData, AuditLog, AuditQuery } from '@/types'

export const auditApi = {
  list: (params: AuditQuery) =>
    request.get<ApiResponse<PageData<AuditLog>>>('/audit/logs', { params }),

  getById: (id: string) =>
    request.get<ApiResponse<AuditLog>>(`/audit/logs/${id}`),

  getStats: () =>
    request.get<ApiResponse<Record<string, unknown>>>('/audit/stats'),

  getUserActivity: (userId: string) =>
    request.get<ApiResponse<AuditLog[]>>(`/audit/user-activity/${userId}`),

  getModuleStats: () =>
    request.get<ApiResponse<Record<string, unknown>>>('/audit/module-stats'),

  cleanup: (days: number) =>
    request.post<ApiResponse<{ deleted_count: number; before: string }>>('/audit/cleanup', { days }),
}
