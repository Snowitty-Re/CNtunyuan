'use client'

import { FormEvent, useEffect, useState } from 'react'
import { AppShell } from '@/components/layout/AppShell'
import { ModuleHeader } from '@/components/shared/ModuleHeader'
import { PageState } from '@/components/shared/PageState'
import { Pagination } from '@/components/shared/Pagination'
import { StatusTag } from '@/components/shared/StatusTag'
import { useAuthGuard } from '@/hooks/useAuthGuard'
import { fmtTime, listFrom } from '@/lib/data'
import { auditService, type AuditLog } from '@/services/audit'

export default function AuditPage() {
  const { ready } = useAuthGuard()
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [items, setItems] = useState<AuditLog[]>([])
  const [stats, setStats] = useState<Record<string, number>>({})
  const [moduleStats, setModuleStats] = useState<Array<{ name: string; count: number }>>([])
  const [total, setTotal] = useState(0)
  const [page, setPage] = useState(1)
  const [logType, setLogType] = useState('')
  const [module, setModule] = useState('')
  const [pathKeyword, setPathKeyword] = useState('')
  const [statusCode, setStatusCode] = useState('')
  const [dateFrom, setDateFrom] = useState('')
  const [dateTo, setDateTo] = useState('')
  const [activityUserId, setActivityUserId] = useState('')
  const [activityLoading, setActivityLoading] = useState(false)
  const [activityError, setActivityError] = useState('')
  const [activityItems, setActivityItems] = useState<any[]>([])

  async function load(nextPage = page) {
    setLoading(true)
    setError('')
    try {
      const data = await auditService.list({
        page: nextPage,
        page_size: 30,
        log_type: logType || undefined,
        module: module || undefined,
        path: pathKeyword || undefined,
        status_code: statusCode || undefined,
        start_time: dateFrom ? new Date(dateFrom).toISOString() : undefined,
        end_time: dateTo ? new Date(dateTo).toISOString() : undefined,
      })
      const normalized = listFrom<AuditLog>(data)
      setItems(normalized.list)
      setTotal(normalized.total)
      setPage(nextPage)
    } catch (err) {
      setError(err instanceof Error ? err.message : '加载失败')
    } finally {
      setLoading(false)
    }
  }

  function exportCsv() {
    if (items.length === 0) return
    const headers = ['time', 'module', 'type', 'user', 'method', 'path', 'status_code', 'ip']
    const lines = items.map((row) =>
      [
        row.created_at || '',
        row.module || '',
        row.log_type || '',
        row.username || row.user_id || '',
        row.method || '',
        row.path || '',
        row.status_code || '',
        row.ip || '',
      ]
        .map((v) => `"${String(v).replace(/"/g, '""')}"`)
        .join(','),
    )
    const content = `\ufeff${[headers.join(','), ...lines].join('\n')}`
    const blob = new Blob([content], { type: 'text/csv;charset=utf-8;' })
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = `audit-logs-${Date.now()}.csv`
    a.click()
    URL.revokeObjectURL(url)
  }

  async function loadStats() {
    try {
      const [data, moduleData] = await Promise.all([auditService.stats(), auditService.moduleStats({ days: 7 })])
      setStats({
        total: Number(data.total || data.total_logs || 0),
        today: Number(data.today || data.today_logs || 0),
        error: Number(data.error || data.error_count || 0),
        login: Number(data.login || data.login_count || 0),
      })
      const rawList = listFrom<any>(moduleData).list
      const normalized = rawList.map((m: any) => ({
        name: m.module || m.name || 'unknown',
        count: Number(m.count || m.total || m.value || 0),
      }))
      setModuleStats(normalized.sort((a: any, b: any) => b.count - a.count).slice(0, 8))
    } catch {
      setStats({})
      setModuleStats([])
    }
  }

  async function loadUserActivity(targetUserId?: string) {
    const uid = (targetUserId ?? activityUserId).trim()
    if (!uid) {
      setActivityItems([])
      setActivityError('')
      return
    }
    setActivityLoading(true)
    setActivityError('')
    try {
      const data = await auditService.userActivity(uid, { page: 1, page_size: 50, days: 7 })
      setActivityItems(listFrom<any>(data).list)
    } catch (err) {
      setActivityError(err instanceof Error ? err.message : '查询失败')
      setActivityItems([])
    } finally {
      setActivityLoading(false)
    }
  }

  useEffect(() => {
    if (ready) {
      load(1)
      loadStats()
      const uid = new URLSearchParams(window.location.search).get('user_id') || ''
      if (uid) {
        setActivityUserId(uid)
        loadUserActivity(uid)
      }
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [ready])

  if (!ready) return null

  return (
    <AppShell>
      <ModuleHeader title="审计中心" desc="追踪关键操作、请求状态与风险行为" />
      <div className="kpi-grid" style={{ marginBottom: 12 }}>
        <div className="kpi">
          <div className="label">日志总数</div>
          <div className="value">{stats.total || 0}</div>
        </div>
        <div className="kpi">
          <div className="label">今日新增</div>
          <div className="value">{stats.today || 0}</div>
        </div>
        <div className="kpi">
          <div className="label">异常日志</div>
          <div className="value">{stats.error || 0}</div>
        </div>
        <div className="kpi">
          <div className="label">登录事件</div>
          <div className="value">{stats.login || 0}</div>
        </div>
      </div>
      <form
        className="panel row wrap"
        onSubmit={(e: FormEvent) => {
          e.preventDefault()
          load(1)
        }}
      >
        <input className="input" placeholder="模块名（如 task/user）" value={module} onChange={(e) => setModule(e.target.value)} />
        <input className="input" placeholder="请求路径（如 /tasks）" value={pathKeyword} onChange={(e) => setPathKeyword(e.target.value)} />
        <input
          className="input"
          placeholder="状态码（如 500）"
          value={statusCode}
          onChange={(e) => setStatusCode(e.target.value.replace(/[^\d]/g, ''))}
        />
        <input className="input" type="datetime-local" value={dateFrom} onChange={(e) => setDateFrom(e.target.value)} />
        <input className="input" type="datetime-local" value={dateTo} onChange={(e) => setDateTo(e.target.value)} />
        <select className="select" value={logType} onChange={(e) => setLogType(e.target.value)}>
          <option value="">全部类型</option>
          <option value="login">login</option>
          <option value="create">create</option>
          <option value="update">update</option>
          <option value="delete">delete</option>
          <option value="query">query</option>
          <option value="upload">upload</option>
        </select>
        <button className="btn primary" type="submit">
          筛选
        </button>
        <button className="btn" type="button" onClick={exportCsv}>
          导出CSV
        </button>
      </form>
      <div className="panel">
        <b>近 7 天模块操作分布</b>
        {moduleStats.length === 0 ? (
          <div style={{ marginTop: 10, color: '#6b7280' }}>暂无模块统计数据</div>
        ) : (
          <div className="bar-list" style={{ marginTop: 10 }}>
            {moduleStats.map((m) => {
              const max = moduleStats[0]?.count || 1
              const pct = Math.max(4, Math.round((m.count / max) * 100))
              return (
                <div className="bar-row" key={m.name}>
                  <span>{m.name}</span>
                  <div className="bar-track">
                    <div className="bar-fill" style={{ width: `${pct}%` }} />
                  </div>
                  <b style={{ textAlign: 'right' }}>{m.count}</b>
                </div>
              )
            })}
          </div>
        )}
      </div>
      <div className="panel">
        <b>用户行为追踪（近 7 天）</b>
        <form
          className="row wrap"
          style={{ marginTop: 10 }}
          onSubmit={(e: FormEvent) => {
            e.preventDefault()
            loadUserActivity()
          }}
        >
          <input
            className="input"
            placeholder="输入用户 ID"
            style={{ minWidth: 260 }}
            value={activityUserId}
            onChange={(e) => setActivityUserId(e.target.value)}
          />
          <button className="btn" type="submit" disabled={activityLoading}>
            {activityLoading ? '查询中...' : '查询'}
          </button>
        </form>
        {activityError ? <div style={{ marginTop: 10, color: '#dc2626' }}>{activityError}</div> : null}
        {activityItems.length > 0 ? (
          <div className="table-wrap" style={{ marginTop: 10 }}>
            <table className="table">
              <thead>
                <tr>
                  <th>时间</th>
                  <th>模块</th>
                  <th>类型</th>
                  <th>请求</th>
                  <th>状态码</th>
                </tr>
              </thead>
              <tbody>
                {activityItems.map((row) => (
                  <tr key={row.id || `${row.created_at}-${row.path}`}>
                    <td>{fmtTime(row.created_at)}</td>
                    <td>{row.module || '-'}</td>
                    <td>
                      <StatusTag status={row.log_type || row.type || '-'} />
                    </td>
                    <td>
                      {(row.method || '').toUpperCase()} {row.path || '-'}
                    </td>
                    <td>{row.status_code || '-'}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        ) : activityUserId && !activityLoading && !activityError ? (
          <div style={{ marginTop: 10, color: '#6b7280' }}>未查询到该用户近期行为</div>
        ) : null}
      </div>
      <PageState loading={loading} error={error} empty={!loading && !error && items.length === 0} onRetry={() => load(page)} />
      {!loading && !error && items.length > 0 ? (
        <>
          <div className="table-wrap">
            <table className="table">
              <thead>
                <tr>
                  <th>时间</th>
                  <th>模块</th>
                  <th>类型</th>
                  <th>用户</th>
                  <th>请求</th>
                  <th>状态码</th>
                  <th>IP</th>
                </tr>
              </thead>
              <tbody>
                {items.map((row) => (
                  <tr key={row.id}>
                    <td>{fmtTime(row.created_at)}</td>
                    <td>{row.module || '-'}</td>
                    <td>
                      <StatusTag status={row.log_type || '-'} />
                    </td>
                    <td>{row.username || row.user_id || '-'}</td>
                    <td>
                      {(row.method || '').toUpperCase()} {row.path || '-'}
                    </td>
                    <td>{row.status_code || '-'}</td>
                    <td>{row.ip || '-'}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
          <Pagination page={page} pageSize={30} total={total} onChange={load} />
        </>
      ) : null}
    </AppShell>
  )
}
