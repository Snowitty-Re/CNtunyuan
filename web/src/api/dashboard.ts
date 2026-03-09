import request from './request'
import type { ApiResponse, DashboardStats, TrendData } from '@/types'

export const dashboardApi = {
  getStats: () =>
    request.get<ApiResponse<DashboardStats>>('/dashboard/stats'),

  getOverview: () =>
    request.get<ApiResponse<Record<string, number>>>('/dashboard/overview'),

  getTrend: (days = 7) =>
    request.get<ApiResponse<TrendData[]>>('/dashboard/trend', { params: { days } }),
}
