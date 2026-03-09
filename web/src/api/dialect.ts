import request from './request'
import type { ApiResponse, PageData, Dialect, DialectComment, DialectQuery, DialectStats, CreateDialectRequest, UpdateDialectRequest } from '@/types'

export const dialectApi = {
  list: (params: DialectQuery) =>
    request.get<ApiResponse<PageData<Dialect>>>('/dialects', { params }),

  featured: (params: { page?: number; page_size?: number }) =>
    request.get<ApiResponse<PageData<Dialect>>>('/dialects/featured', { params }),

  getStats: () =>
    request.get<ApiResponse<DialectStats>>('/dialects/stats'),

  getById: (id: string) =>
    request.get<ApiResponse<Dialect>>(`/dialects/${id}`),

  create: (data: CreateDialectRequest) =>
    request.post<ApiResponse<Dialect>>('/dialects', data),

  update: (id: string, data: UpdateDialectRequest) =>
    request.put<ApiResponse<Dialect>>(`/dialects/${id}`, data),

  delete: (id: string) =>
    request.delete<ApiResponse<null>>(`/dialects/${id}`),

  updateStatus: (id: string, status: string) =>
    request.put<ApiResponse<null>>(`/dialects/${id}/status`, { status }),

  feature: (id: string) =>
    request.post<ApiResponse<null>>(`/dialects/${id}/feature`),

  unfeature: (id: string) =>
    request.delete<ApiResponse<null>>(`/dialects/${id}/feature`),

  play: (id: string) =>
    request.post<ApiResponse<null>>(`/dialects/${id}/play`),

  like: (id: string) =>
    request.post<ApiResponse<null>>(`/dialects/${id}/like`),

  unlike: (id: string) =>
    request.delete<ApiResponse<null>>(`/dialects/${id}/like`),

  getComments: (id: string, params: { page?: number; page_size?: number }) =>
    request.get<ApiResponse<PageData<DialectComment>>>(`/dialects/${id}/comments`, { params }),

  addComment: (id: string, data: { content: string; parent_id?: string }) =>
    request.post<ApiResponse<DialectComment>>(`/dialects/${id}/comments`, data),
}
