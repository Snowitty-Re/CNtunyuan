import { API_BASE, authedFetch, http, ApiError } from '@/lib/request'

export type UploadResult = {
  id?: string
  url?: string
  path?: string
  filename?: string
}

export const uploadService = {
  async uploadSingle(file: File, formData: Record<string, string> = {}): Promise<UploadResult> {
    const body = new FormData()
    body.append('file', file)
    Object.entries(formData).forEach(([k, v]) => {
      body.append(k, v)
    })

    const res = await authedFetch('/upload', {
      method: 'POST',
      body,
    })

    if (res.status === 204) return {}
    const payload = await res.json()
    if (!res.ok || (payload.code !== 0 && payload.code !== 200)) {
      throw new ApiError(payload.message || `上传失败: ${res.status}`, res.status, payload.code)
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
  async downloadBlob(fileId: string): Promise<Blob> {
    const res = await authedFetch(`/upload/${fileId}/download`)
    if (!res.ok) {
      throw new ApiError(`下载失败 HTTP ${res.status}`, res.status)
    }
    return res.blob()
  },
}

export { API_BASE }
