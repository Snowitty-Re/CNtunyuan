'use client'

import { useEffect, useMemo, useState } from 'react'
import { AppShell } from '@/components/layout/AppShell'
import { ModuleHeader } from '@/components/shared/ModuleHeader'
import { NoticeBar, type Notice } from '@/components/shared/NoticeBar'
import { PageState } from '@/components/shared/PageState'
import { useAuthGuard } from '@/hooks/useAuthGuard'
import { ACTIONS, hasPermission } from '@/lib/rbac'
import { fmtTime } from '@/lib/data'
import { systemService } from '@/services/system'

export default function MonitorPage() {
  const { ready, user } = useAuthGuard()
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [notice, setNotice] = useState<Notice | null>(null)
  const [health, setHealth] = useState<any>(null)
  const [detail, setDetail] = useState<any>(null)
  const [metricsPreview, setMetricsPreview] = useState('')
  const [updatedAt, setUpdatedAt] = useState('')

  async function loadAll() {
    setLoading(true)
    setError('')
    try {
      const [h, d, m] = await Promise.all([
        systemService.health(),
        systemService.detailedHealth(),
        systemService.metricsRaw(),
      ])
      setHealth(h)
      setDetail(d)
      setMetricsPreview((m || '').split('\n').slice(0, 80).join('\n'))
      setUpdatedAt(new Date().toISOString())
    } catch (err) {
      setError(err instanceof Error ? err.message : '监控数据加载失败')
      setHealth(null)
      setDetail(null)
      setMetricsPreview('')
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    if (ready) loadAll()
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [ready])

  const checks = detail?.checks || health?.checks || {}
  const runtime = detail?.runtime || {}
  const environment = String(detail?.environment || health?.environment || 'unknown')
  const version = String(detail?.version || health?.version || '-')
  const overallStatus = String(detail?.status || health?.status || 'unknown')
  const dbCheck = checks.database || {}
  const cacheCheck = checks.cache || {}
  const sysCheck = checks.system || {}

  const monitorCards = useMemo(() => ([
    { label: '服务状态', value: overallStatus, detail: `版本 ${version}` },
    { label: '数据库', value: String(dbCheck.status || 'unknown'), detail: String(dbCheck.message || '未返回数据库状态') },
    { label: '缓存', value: String(cacheCheck.status || 'unknown'), detail: String(cacheCheck.message || '未返回缓存状态') },
    { label: '运行环境', value: environment, detail: `${runtime.go_os || '-'} / ${runtime.go_arch || '-'}` },
  ]), [overallStatus, version, dbCheck.status, dbCheck.message, cacheCheck.status, cacheCheck.message, environment, runtime.go_os, runtime.go_arch])

  if (!ready) return null
  if (!hasPermission(user, ACTIONS.USER_MODIFY)) {
    return (
      <AppShell>
        <ModuleHeader title="服务监控" desc="健康检查、依赖状态与指标预览" />
        <PageState error="当前账号无权限访问该页面（需要 user:modify 权限）" />
      </AppShell>
    )
  }

  return (
    <AppShell>
      <ModuleHeader
        title="服务监控"
        desc="健康检查、依赖状态、运行环境与指标预览"
        right={
          <button className="btn" type="button" onClick={loadAll}>
            立即刷新
          </button>
        }
      />
      <NoticeBar notice={notice} onClose={() => setNotice(null)} />
      <PageState loading={loading} error={error} onRetry={loadAll} />
      {!loading && !error ? (
        <>
          <div className="panel row wrap">
            <b>最近刷新时间：</b>
            <span>{fmtTime(updatedAt)}</span>
          </div>

          <div className="insight-grid">
            {monitorCards.map((item) => (
              <div className="stat-card stat-card-blue" key={item.label}>
                <span>{item.label}</span>
                <strong>{item.value}</strong>
                <small>{item.detail}</small>
              </div>
            ))}
          </div>

          <div className="dashboard-grid">
            <section className="section-card">
              <b>依赖状态</b>
              <div className="heat-panel" style={{ marginTop: 12 }}>
                <StatusRow label="数据库响应" response={dbCheck.response_ms} status={dbCheck.status} />
                <StatusRow label="缓存响应" response={cacheCheck.response_ms} status={cacheCheck.status} />
                <StatusRow label="系统资源" response={sysCheck.response_ms} status={sysCheck.status} />
              </div>
            </section>

            <section className="section-card">
              <b>运行时信息</b>
              <div className="attachment-meta-grid" style={{ marginTop: 12 }}>
                <div>Go 版本：{runtime.go_version || '-'}</div>
                <div>CPU 数：{runtime.cpu_count || '-'}</div>
                <div>Goroutines：{runtime.goroutines || '-'}</div>
                <div>GOMAXPROCS：{runtime.gomaxprocs || '-'}</div>
                <div>系统：{runtime.go_os || '-'}</div>
                <div>架构：{runtime.go_arch || '-'}</div>
              </div>
            </section>
          </div>

          <div className="grid cols-2">
            <div className="section-card">
              <b>健康检查详情</b>
              <pre style={{ whiteSpace: 'pre-wrap', marginTop: 10 }}>{JSON.stringify(detail || health || {}, null, 2)}</pre>
            </div>
            <div className="section-card">
              <b>Prometheus 指标预览（前 80 行）</b>
              <pre style={{ whiteSpace: 'pre-wrap', marginTop: 10 }}>{metricsPreview || '暂无 metrics 内容'}</pre>
            </div>
          </div>
        </>
      ) : null}
    </AppShell>
  )
}

function StatusRow({ label, response, status }: { label: string; response?: number; status?: string }) {
  const normalized = String(status || 'unknown').toUpperCase()
  const value = normalized === 'UP' ? 100 : normalized === 'WARNING' ? 60 : normalized === 'UNKNOWN' ? 20 : 15
  return (
    <div className="heat-row">
      <span>{label}</span>
      <div className="heat-track">
        <div className="heat-fill" style={{ width: `${value}%` }} />
      </div>
      <b>{normalized} {response !== undefined ? `${response}ms` : ''}</b>
    </div>
  )
}
