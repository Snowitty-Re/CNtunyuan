import { API_BASE, http } from '@/lib/request'
import { getAccessToken } from '@/lib/auth'

export const systemService = {
  health() {
    return http<any>('/health', { auth: false })
  },
  detailedHealth() {
    return http<any>('/health/detailed', { auth: false })
  },
  uploadStats() {
    return http<any>('/upload/stats')
  },
  filesByEntity(entityType: string, entityId: string) {
    return http<any[]>(`/upload/entity/${entityType}/${entityId}`)
  },
  fileById(id: string) {
    return http<any>(`/upload/${id}`)
  },
  deleteFile(id: string) {
    return http<null>(`/upload/${id}`, { method: 'DELETE' })
  },
  getSiteConfig() {
    return http<{ config: Record<string, unknown>; sensitive_fields: string[] }>('/system/config')
  },
  updateSiteConfig(config: Record<string, unknown>) {
    return http<{ config: Record<string, unknown>; sensitive_fields: string[] }>('/system/config', {
      method: 'PUT',
      body: { config },
    })
  },
  async metricsRaw(): Promise<string> {
    const token = getAccessToken()
    const res = await fetch(`${API_BASE}/metrics`, {
      headers: token ? { Authorization: `Bearer ${token}` } : undefined,
      cache: 'no-store',
    })
    if (!res.ok) throw new Error(`获取 metrics 失败: HTTP ${res.status}`)
    return res.text()
  },
}
