import request from '../utils/request'
import type { MissingPerson, PageData } from '../types'

export const missingPersonApi = {
  // 获取列表
  getList: (params: { page?: number; page_size?: number; status?: string; case_type?: string; keyword?: string }) =>
    request.get<PageData<MissingPerson>>('/missing-persons', { params }),

  // 获取详情
  getById: (id: string) => request.get<MissingPerson>(`/missing-persons/${id}`),

  // 创建
  create: (data: Partial<MissingPerson>) => request.post('/missing-persons', data),

  // 更新状态
  updateStatus: (id: string, status: string, foundDetail?: string) =>
    request.put(`/missing-persons/${id}/status`, null, { params: { status, found_detail: foundDetail } }),

  // 获取附近案件
  getNearby: (lat: number, lng: number, radius?: number) =>
    request.get<MissingPerson[]>('/missing-persons/nearby', { params: { lat, lng, radius } }),

  // 获取统计
  getStatistics: () => request.get('/missing-persons/statistics'),

  // 添加轨迹
  addTrack: (id: string, data: { track_time: string; location: string; longitude: number; latitude: number; description?: string; photos?: string[] }) =>
    request.post(`/missing-persons/${id}/tracks`, data),

  // 获取轨迹
  getTracks: (id: string) => request.get(`/missing-persons/${id}/tracks`),
}
