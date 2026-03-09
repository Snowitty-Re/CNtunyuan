import request from './request'
import type { ApiResponse, PageData, MissingPerson, MissingPersonTrack, MissingPersonQuery, CreateMissingPersonRequest, UpdateMissingPersonRequest, CreateTrackRequest, MarkFoundRequest, MissingPersonStats } from '@/types'

export const missingPersonApi = {
  list: (params: MissingPersonQuery) =>
    request.get<ApiResponse<PageData<MissingPerson>>>('/missing-persons', { params }),

  search: (keyword: string, page = 1, page_size = 10) =>
    request.get<ApiResponse<PageData<MissingPerson>>>('/missing-persons/search', { params: { keyword, page, page_size } }),

  getStats: () =>
    request.get<ApiResponse<MissingPersonStats>>('/missing-persons/stats'),

  getById: (id: string) =>
    request.get<ApiResponse<MissingPerson>>(`/missing-persons/${id}`),

  create: (data: CreateMissingPersonRequest) =>
    request.post<ApiResponse<MissingPerson>>('/missing-persons', data),

  update: (id: string, data: UpdateMissingPersonRequest) =>
    request.put<ApiResponse<MissingPerson>>(`/missing-persons/${id}`, data),

  delete: (id: string) =>
    request.delete<ApiResponse<null>>(`/missing-persons/${id}`),

  updateStatus: (id: string, status: string) =>
    request.put<ApiResponse<null>>(`/missing-persons/${id}/status`, { status }),

  markFound: (id: string, data: MarkFoundRequest) =>
    request.post<ApiResponse<null>>(`/missing-persons/${id}/found`, data),

  markReunited: (id: string) =>
    request.post<ApiResponse<null>>(`/missing-persons/${id}/reunited`),

  getTracks: (id: string) =>
    request.get<ApiResponse<MissingPersonTrack[]>>(`/missing-persons/${id}/tracks`),

  addTrack: (id: string, data: CreateTrackRequest) =>
    request.post<ApiResponse<MissingPersonTrack>>(`/missing-persons/${id}/tracks`, data),
}
