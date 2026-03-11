import request from './request';
import type { DashboardStats, TrendData, OverviewData } from '@/types';

export function getDashboardStats() {
  return request.get<never, DashboardStats>('/dashboard/stats');
}

export function getOverview() {
  return request.get<never, OverviewData>('/dashboard/overview');
}

export function getTrend(days = 7) {
  return request.get<never, TrendData[]>('/dashboard/trend', { params: { days } });
}
