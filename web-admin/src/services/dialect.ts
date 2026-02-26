import request from '../utils/request'
import type { Dialect, PageData } from '../types'

export const dialectApi = {
  // 获取列表
  getList: (params: { page?: number; page_size?: number; province?: string; city?: string; keyword?: string }) =>
    request.get<PageData<Dialect>>('/dialects', { params }),

  // 获取详情
  getById: (id: string) => request.get<Dialect>(`/dialects/${id}`),

  // 创建
  create: (data: Partial<Dialect>) => request.post('/dialects', data),

  // 更新
  update: (id: string, data: Partial<Dialect>) => request.put(`/dialects/${id}`, data),

  // 删除
  delete: (id: string) => request.delete(`/dialects/${id}`),

  // 播放
  play: (id: string, duration?: number) => request.post(`/dialects/${id}/play`, null, { params: { duration } }),

  // 点赞
  like: (id: string) => request.post(`/dialects/${id}/like`),

  // 取消点赞
  unlike: (id: string) => request.post(`/dialects/${id}/unlike`),

  // 获取附近方言
  getNearby: (lat: number, lng: number, radius?: number) =>
    request.get<Dialect[]>('/dialects/nearby', { params: { lat, lng, radius } }),

  // 获取统计
  getStatistics: () => request.get('/dialects/statistics'),
}
