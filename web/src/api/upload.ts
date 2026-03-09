import request from './request'
import type { ApiResponse, FileInfo } from '@/types'

export const uploadApi = {
  upload: (file: File, onProgress?: (percent: number) => void) => {
    const formData = new FormData()
    formData.append('file', file)
    return request.post<ApiResponse<FileInfo>>('/upload', formData, {
      headers: { 'Content-Type': 'multipart/form-data' },
      onUploadProgress: (e) => {
        if (onProgress && e.total) {
          onProgress(Math.round((e.loaded * 100) / e.total))
        }
      },
    })
  },

  batchUpload: (files: File[]) => {
    const formData = new FormData()
    files.forEach((file) => formData.append('files', file))
    return request.post<ApiResponse<FileInfo[]>>('/upload/batch', formData, {
      headers: { 'Content-Type': 'multipart/form-data' },
    })
  },

  getById: (id: string) =>
    request.get<ApiResponse<FileInfo>>(`/upload/${id}`),

  download: (id: string) =>
    request.get(`/upload/${id}/download`, { responseType: 'blob' }),

  getByEntity: (entityType: string, entityId: string) =>
    request.get<ApiResponse<FileInfo[]>>(`/upload/entity/${entityType}/${entityId}`),

  delete: (id: string) =>
    request.delete<ApiResponse<null>>(`/upload/${id}`),

  bind: (id: string, entity_type: string, entity_id: string) =>
    request.put<ApiResponse<null>>(`/upload/${id}/bind`, { entity_type, entity_id }),
}
