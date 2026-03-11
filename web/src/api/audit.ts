import request from './request';
import type { PaginatedData, AuditLogResponse, AuditLogListRequest, AuditLogStats, AuditUserActivity, AuditModuleStat } from '@/types';

export function getAuditLogs(params: AuditLogListRequest) {
  return request.get<never, PaginatedData<AuditLogResponse>>('/audit/logs', { params });
}

export function getAuditLog(id: string) {
  return request.get<never, AuditLogResponse>(`/audit/logs/${id}`);
}

export function getAuditStats(params?: { start_time?: string; end_time?: string }) {
  return request.get<never, AuditLogStats>('/audit/stats', { params });
}

export function getUserActivity(userId: string, days = 7) {
  return request.get<never, AuditUserActivity[]>(`/audit/user-activity/${userId}`, { params: { days } });
}

export function getModuleStats(params?: { start_time?: string; end_time?: string }) {
  return request.get<never, AuditModuleStat[]>('/audit/module-stats', { params });
}

export function cleanupAuditLogs(days: number) {
  return request.post<never, { deleted_count: number; before: string }>('/audit/cleanup', { days });
}
