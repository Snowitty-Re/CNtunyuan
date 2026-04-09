'use client'

import Link from 'next/link'
import { useEffect, useMemo, useState } from 'react'
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

type TrendRow = { date: string; tasks: number; cases: number; resolved: number }

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
  const [trendRows, setTrendRows] = useState<TrendRow[]>([])

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
      const totalCases = Number(stats?.missing_persons?.total || overview?.total_cases || 0)
      const resolvedCases = Number(
        (stats?.missing_persons?.found || 0) + (stats?.missing_persons?.reunited || 0) || overview?.resolved_cases || 0,
      )
      const totalUsers = Number(stats?.users?.total || overview?.total_users || 0)
      const totalTasks = Number(stats?.tasks?.total || overview?.total_tasks || 0)
      setKpi({ totalCases, resolvedCases, totalUsers, totalTasks })

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

  const closureRate = useMemo(() => {
    if (!kpi.totalCases) return 0
    return Math.min(100, Math.round((kpi.resolvedCases / kpi.totalCases) * 100))
  }, [kpi.resolvedCases, kpi.totalCases])

  const collaborationScore = useMemo(() => {
    const numerator = kpi.totalTasks + kpi.totalUsers * 2 + kpi.resolvedCases * 3
    const denominator = Math.max(1, kpi.totalCases + kpi.totalTasks + kpi.totalUsers)
    return Math.min(100, Math.round((numerator / denominator) * 10))
  }, [kpi])

  const maxTrend = useMemo(() => Math.max(...trendRows.map((row) => row.tasks + row.cases + row.resolved), 1), [trendRows])

  if (!ready) return null

  return (
    <AppShell>
      <ModuleHeader title="工作台总览" desc="实时查看案件、任务、志愿者协作与闭环趋势" />
      <PageState loading={loading} error={error} onRetry={load} />
      {!loading && !error ? (
        <>
          <div className="overview-hero">
            <div className="overview-copy">
              <span className="overview-badge">今日工作建议</span>
              <h3>优先跟进待分配任务，持续补全线索闭环。</h3>
              <p>通过案件、任务、方言三条链路统一推进，缩短走失人员与亲属重聚路径。</p>
            </div>
            <div className="overview-rings">
              <MetricRing value={closureRate} label="案件闭环率" accent="var(--primary)" />
              <MetricRing value={collaborationScore} label="协作活跃度" accent="var(--accent)" />
            </div>
          </div>

          <div className="insight-grid">
            <StatCard label="走失人员总数" value={kpi.totalCases} detail="全平台累计案件" tone="amber" />
            <StatCard label="已找到 / 团圆" value={kpi.resolvedCases} detail={`闭环率 ${closureRate}%`} tone="green" />
            <StatCard label="平台人员数" value={kpi.totalUsers} detail="含志愿者与管理人员" tone="blue" />
            <StatCard label="任务总数" value={kpi.totalTasks} detail="可分配、可审批、可追踪" tone="rose" />
          </div>

          <div className="dashboard-grid">
            <section className="section-card chart-panel">
              <div className="row wrap" style={{ justifyContent: 'space-between' }}>
                <div>
                  <b>趋势分析</b>
                  <div className="hint">案件、任务与闭环数据图形化展示</div>
                </div>
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
                <div className="trend-chart">
                  {trendRows.map((row) => {
                    const total = row.tasks + row.cases + row.resolved
                    const height = `${Math.max(12, Math.round((total / maxTrend) * 100))}%`
                    return (
                      <div className="trend-column" key={row.date}>
                        <div className="trend-stack">
                          <div className="trend-bar trend-bar-resolved" style={{ height: `${Math.max(4, Math.round((row.resolved / Math.max(total, 1)) * 100))}%` }} />
                          <div className="trend-bar trend-bar-cases" style={{ height: `${Math.max(4, Math.round((row.cases / Math.max(total, 1)) * 100))}%` }} />
                          <div className="trend-bar trend-bar-tasks" style={{ height: `${Math.max(4, Math.round((row.tasks / Math.max(total, 1)) * 100))}%` }} />
                        </div>
                        <div className="trend-column-shell" style={{ height }}>
                          <span className="trend-total">{total}</span>
                        </div>
                        <span className="trend-label">{row.date.slice(5)}</span>
                      </div>
                    )
                  })}
                </div>
              )}
              <div className="chart-legend">
                <span><i className="legend-dot legend-tasks" />任务</span>
                <span><i className="legend-dot legend-cases" />案件</span>
                <span><i className="legend-dot legend-resolved" />闭环</span>
              </div>
            </section>

            <section className="section-card">
              <b>协作热度</b>
              <div className="hint">根据人员、任务与案件闭环形成的综合活跃度</div>
              <div className="heat-panel">
                <HeatRow label="志愿者参与" value={Math.min(100, kpi.totalUsers * 8)} />
                <HeatRow label="任务推进" value={Math.min(100, kpi.totalTasks * 6)} />
                <HeatRow label="线索闭环" value={closureRate} />
              </div>
            </section>
          </div>

          <div className="grid cols-2">
            <div className="section-card">
              <div className="row wrap" style={{ justifyContent: 'space-between' }}>
                <div>
                  <b>待分配任务提醒</b>
                  <div className="hint">优先处理待指派或待启动的任务</div>
                </div>
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
                <div>
                  <b>进行中案件提醒</b>
                  <div className="hint">持续更新轨迹、附件与重点线索</div>
                </div>
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

function StatCard({ label, value, detail, tone }: { label: string; value: number; detail: string; tone: string }) {
  return (
    <div className={`stat-card stat-card-${tone}`}>
      <span>{label}</span>
      <strong>{value}</strong>
      <small>{detail}</small>
    </div>
  )
}

function MetricRing({ value, label, accent }: { value: number; label: string; accent: string }) {
  const radius = 48
  const circumference = 2 * Math.PI * radius
  const offset = circumference - (Math.max(0, Math.min(value, 100)) / 100) * circumference
  return (
    <div className="metric-ring">
      <svg viewBox="0 0 120 120">
        <circle cx="60" cy="60" r={radius} className="metric-ring-track" />
        <circle cx="60" cy="60" r={radius} className="metric-ring-progress" style={{ stroke: accent, strokeDasharray: circumference, strokeDashoffset: offset }} />
      </svg>
      <div className="metric-ring-content">
        <strong>{value}%</strong>
        <span>{label}</span>
      </div>
    </div>
  )
}

function HeatRow({ label, value }: { label: string; value: number }) {
  return (
    <div className="heat-row">
      <span>{label}</span>
      <div className="heat-track">
        <div className="heat-fill" style={{ width: `${Math.max(6, Math.min(value, 100))}%` }} />
      </div>
      <b>{Math.min(value, 100)}%</b>
    </div>
  )
}
