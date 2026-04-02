import { API_BASE, http } from '@/lib/request'
import { getAccessToken } from '@/lib/auth'

export type UploadResult = {
  id?: string
  url?: string
  path?: string
  filename?: string
}

export const uploadService = {
  async uploadSingle(file: File, formData: Record<string, string> = {}): Promise<UploadResult> {
    const token = getAccessToken()
    const url = `${API_BASE}/upload`
    const body = new FormData()
    body.append('file', file)
    Object.entries(formData).forEach(([k, v]) => {
      body.append(k, v)
    })

    const res = await fetch(url, {
      method: 'POST',
      headers: token ? { Authorization: `Bearer ${token}` } : undefined,
      body,
    })

    if (res.status === 204) return {}
    const payload = await res.json()
    if (!res.ok || (payload.code !== 0 && payload.code !== 200)) {
      throw new Error(payload.message || `上传失败: ${res.status}`)
    }
    return payload.data || {}
  },
  bind(fileId: string, entityType: string, entityId: string) {
    return http(`/upload/${fileId}/bind`, {
      method: 'PUT',
      body: {
        entity_type: entityType,
        entity_id: entityId,
      },
    })
  },
}

