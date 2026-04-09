import { http } from '@/lib/request'
import type { Paginated } from '@/types/api'

export type AuditLog = {
  id: string
  user_id?: string
  username?: string
  module?: string
  action?: string
  log_type?: string
  type?: string
  method?: string
  request_method?: string
  path?: string
  request_url?: string
  ip?: string
  request_ip?: string
  status_code?: number
  response_code?: number
  status?: string
  created_at?: string
}

export const auditService = {
  list(params: Record<string, string | number | undefined> = {}) {
    return http<Paginated<AuditLog>>('/audit/logs', { query: params })
  },
  stats(params: Record<string, string | number | undefined> = {}) {
    return http<Record<string, any>>('/audit/stats', { query: params })
  },
  moduleStats(params: Record<string, string | number | undefined> = {}) {
    return http<any>('/audit/module-stats', { query: params })
  },
  userActivity(userId: string, params: Record<string, string | number | undefined> = {}) {
    return http<any>(`/audit/user-activity/${userId}`, { query: params })
  },
}
