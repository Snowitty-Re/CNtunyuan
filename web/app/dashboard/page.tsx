'use client'

import Link from 'next/link'
import { useEffect, useState } from 'react'
import { AppShell } from '@/components/layout/AppShell'
import { useAuthGuard } from '@/hooks/useAuthGuard'
import { dashboardService } from '@/services/dashboard'
import { taskService } from '@/services/tasks'
import { missingPersonService } from '@/services/missingPersons'
import { PageState } from '@/components/shared/PageState'
import { ModuleHeader } from '@/components/shared/ModuleHeader'
import { listFrom, fmtTime } from '@/lib/data'

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
  const [pendingTasks, setPendingTasks] = useState<Array<{ id: string; title: string; deadline?: string }>>([])
  const [activeCases, setActiveCases] = useState<Array<{ id: string; name: string; missing_time?: string }>>([])
  const [trendDays, setTrendDays] = useState(7)
  const [trendRows, setTrendRows] = useState<Array<{ date: string; tasks: number; cases: number; resolved: number }>>([])

  async function load() {
    setLoading(true)
    setError('')
    try {
      const [stats, overview, tasksData, casesData, trendData] = await Promise.all([
        dashboardService.stats(),
        dashboardService.overview(),
        taskService.list({ page: 1, page_size: 5, status: 'pending' }),
        missingPersonService.list({ page: 1, page_size: 5, status: 'searching' }),
        dashboardService.trend(trendDays),
      ])
      setKpi({
        totalCases: Number(stats?.missing_persons?.total || overview?.total_cases || 0),
        resolvedCases: Number(
          (stats?.missing_persons?.found || 0) + (stats?.missing_persons?.reunited || 0) || overview?.resolved_cases || 0,
        ),
        totalUsers: Number(stats?.users?.total || overview?.total_users || 0),
        totalTasks: Number(stats?.tasks?.total || overview?.total_tasks || 0),
      })
      const t = listFrom<any>(tasksData).list
      const c = listFrom<any>(casesData).list
      const trendList = listFrom<any>(trendData).list.length > 0 ? listFrom<any>(trendData).list : (Array.isArray((trendData as any)?.data) ? (trendData as any).data : [])
      setPendingTasks(t.map((x) => ({ id: x.id, title: x.title || '未命名任务', deadline: x.deadline || x.created_at })))
      setActiveCases(c.map((x) => ({ id: x.id, name: x.name || '未命名案件', missing_time: x.missing_time || x.created_at })))
      setTrendRows(
        trendList.map((r: any) => ({
          date: String(r.date || r.day || r.time || '-'),
          tasks: Number(r.tasks || r.task_count || 0),
          cases: Number(r.cases || r.case_count || r.missing_count || 0),
          resolved: Number(r.resolved || r.resolved_count || r.found_count || 0),
        })),
      )
    } catch (err) {
      setError(err instanceof Error ? err.message : '加载失败')
      setPendingTasks([])
      setActiveCases([])
      setTrendRows([])
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    if (ready) load()
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [ready, trendDays])

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
          <div className="section-card">
            <div className="row wrap" style={{ justifyContent: 'space-between' }}>
              <b>趋势分析</b>
              <div className="row wrap">
                <button className={`btn ${trendDays === 7 ? 'primary' : ''}`} type="button" onClick={() => setTrendDays(7)}>
                  近7天
                </button>
                <button className={`btn ${trendDays === 30 ? 'primary' : ''}`} type="button" onClick={() => setTrendDays(30)}>
                  近30天
                </button>
              </div>
            </div>
            {trendRows.length === 0 ? (
              <div style={{ marginTop: 10, color: '#6b7280' }}>暂无趋势数据</div>
            ) : (
              <div className="bar-list" style={{ marginTop: 12 }}>
                {trendRows.map((row) => {
                  const max = Math.max(...trendRows.map((x) => x.tasks + x.cases + x.resolved), 1)
                  const total = row.tasks + row.cases + row.resolved
                  const pct = Math.max(4, Math.round((total / max) * 100))
                  return (
                    <div className="bar-row" key={row.date}>
                      <span>{row.date}</span>
                      <div className="bar-track">
                        <div className="bar-fill" style={{ width: `${pct}%` }} />
                      </div>
                      <b style={{ textAlign: 'right' }}>
                        任务{row.tasks} / 案件{row.cases} / 闭环{row.resolved}
                      </b>
                    </div>
                  )
                })}
              </div>
            )}
          </div>
          <div className="grid cols-2">
            <div className="section-card">
              <div className="row wrap" style={{ justifyContent: 'space-between' }}>
                <b>待分配任务提醒</b>
                <Link className="btn ghost" href="/tasks">
                  前往任务中心
                </Link>
              </div>
              {pendingTasks.length === 0 ? (
                <div style={{ marginTop: 10, color: '#6b7280' }}>当前无待分配任务</div>
              ) : (
                <ul className="soft-list">
                  {pendingTasks.map((t) => (
                    <li key={t.id}>
                      <div className="row wrap" style={{ justifyContent: 'space-between' }}>
                        <Link href={`/tasks/${t.id}`}>{t.title}</Link>
                        <span className="hint">{fmtTime(t.deadline)}</span>
                      </div>
                    </li>
                  ))}
                </ul>
              )}
            </div>
            <div className="section-card">
              <div className="row wrap" style={{ justifyContent: 'space-between' }}>
                <b>进行中案件提醒</b>
                <Link className="btn ghost" href="/cases">
                  前往案件中心
                </Link>
              </div>
              {activeCases.length === 0 ? (
                <div style={{ marginTop: 10, color: '#6b7280' }}>当前无进行中案件</div>
              ) : (
                <ul className="soft-list">
                  {activeCases.map((c) => (
                    <li key={c.id}>
                      <div className="row wrap" style={{ justifyContent: 'space-between' }}>
                        <Link href={`/cases/${c.id}`}>{c.name}</Link>
                        <span className="hint">{fmtTime(c.missing_time)}</span>
                      </div>
                    </li>
                  ))}
                </ul>
              )}
            </div>
          </div>
        </>
      ) : null}
    </AppShell>
  )
}
