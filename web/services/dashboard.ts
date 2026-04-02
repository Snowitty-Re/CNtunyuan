import { http } from '@/lib/request'

export const dashboardService = {
  stats() {
    return http<Record<string, any>>('/dashboard/stats')
  },
  overview() {
    return http<Record<string, any>>('/dashboard/overview')
  },
  trend(days = 7) {
    return http<Record<string, any>>('/dashboard/trend', { query: { days } })
  },
}
