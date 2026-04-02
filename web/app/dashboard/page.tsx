'use client'

import { useEffect, useState } from 'react'
import { AppShell } from '@/components/layout/AppShell'
import { useAuthGuard } from '@/hooks/useAuthGuard'
import { dashboardService } from '@/services/dashboard'
import { PageState } from '@/components/shared/PageState'
import { ModuleHeader } from '@/components/shared/ModuleHeader'

type KPIs = {
  totalCases: number
  resolvedCases: number
  totalUsers: number
  totalTasks: number
}

export default function DashboardPage() {
  const { ready } = useAuthGuard()
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [kpi, setKpi] = useState<KPIs>({
    totalCases: 0,
    resolvedCases: 0,
    totalUsers: 0,
    totalTasks: 0,
  })

  async function load() {
    setLoading(true)
    setError('')
    try {
      const [stats, overview] = await Promise.all([dashboardService.stats(), dashboardService.overview()])
      setKpi({
        totalCases: Number(stats?.missing_persons?.total || overview?.total_cases || 0),
        resolvedCases: Number(
          (stats?.missing_persons?.found || 0) + (stats?.missing_persons?.reunited || 0) || overview?.resolved_cases || 0,
        ),
        totalUsers: Number(stats?.users?.total || overview?.total_users || 0),
        totalTasks: Number(stats?.tasks?.total || overview?.total_tasks || 0),
      })
    } catch (err) {
      setError(err instanceof Error ? err.message : '加载失败')
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    if (ready) load()
  }, [ready])

  if (!ready) return null

  return (
    <AppShell>
      <ModuleHeader title="工作台总览" desc="实时查看案件、任务与志愿者协作数据" />
      <PageState loading={loading} error={error} onRetry={load} />
      {!loading && !error ? (
        <>
          <div className="kpi-grid">
            <div className="kpi">
              <div className="label">走失人员总数</div>
              <div className="value">{kpi.totalCases}</div>
            </div>
            <div className="kpi">
              <div className="label">已找到/团圆</div>
              <div className="value">{kpi.resolvedCases}</div>
            </div>
            <div className="kpi">
              <div className="label">平台人员数</div>
              <div className="value">{kpi.totalUsers}</div>
            </div>
            <div className="kpi">
              <div className="label">任务总数</div>
              <div className="value">{kpi.totalTasks}</div>
            </div>
          </div>
          <div className="panel">
            <b>建议操作：</b> 优先处理待分配任务，持续补全案件轨迹与方言线索，缩短寻亲闭环时间。
          </div>
        </>
      ) : null}
    </AppShell>
  )
}
