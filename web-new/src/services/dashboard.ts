import { http } from '@/utils/request';
import type { DashboardStats } from '@/types';

export interface TrendData {
  date: string;
  new_cases: number;
  resolved_cases: number;
  new_tasks: number;
  completed_tasks: number;
  new_users: number;
}

export const dashboardApi = {
  // 获取仪表盘统计数据
  getStats: () => http.get<DashboardStats>('/dashboard/stats'),
  
  // 获取概览数据
  getOverview: () => http.get<{
    total_users: number;
    total_cases: number;
    active_cases: number;
    resolved_cases: number;
    pending_tasks: number;
    processing_tasks: number;
  }>('/dashboard/overview'),
  
  // 获取趋势数据
  getTrend: (days: number = 7) => http.get<TrendData[]>('/dashboard/trend', { params: { days } }),
};
