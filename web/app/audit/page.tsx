'use client'

import { FormEvent, useEffect, useMemo, useState } from 'react'
import { AppShell } from '@/components/layout/AppShell'
import { ModuleHeader } from '@/components/shared/ModuleHeader'
import { NoticeBar, type Notice } from '@/components/shared/NoticeBar'
import { PageState } from '@/components/shared/PageState'
import { Pagination } from '@/components/shared/Pagination'
import { StatusTag } from '@/components/shared/StatusTag'
import { useAuthGuard } from '@/hooks/useAuthGuard'
import { ACTIONS, hasPermission } from '@/lib/rbac'
import { fmtTime, listFrom } from '@/lib/data'
import { auditService, type AuditLog } from '@/services/audit'

const rangeOptions = [
  { value: 1, label: '近1天' },
  { value: 7, label: '近7天' },
  { value: 30, label: '近30天' },
]

const typeOptions = [
  { value: '', label: '全部类型' },
  { value: 'login', label: '登录' },
  { value: 'logout', label: '登出' },
  { value: 'create', label: '创建' },
  { value: 'update', label: '更新' },
  { value: 'delete', label: '删除' },
  { value: 'query', label: '查询' },
  { value: 'upload', label: '上传' },
  { value: 'download', label: '下载' },
  { value: 'other', label: '其他' },
]

export default function AuditPage() {
  const { ready, user } = useAuthGuard()
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [notice, setNotice] = useState<Notice | null>(null)
  const [items, setItems] = useState<AuditLog[]>([])
  const [stats, setStats] = useState<Record<string, number>>({})
  const [total, setTotal] = useState(0)
  const [page, setPage] = useState(1)
  const [pageSize] = useState(20)
  const [filters, setFilters] = useState({
    rangeDays: 7,
    type: '',
    module: '',
    username: '',
    path: '',
    statusCode: '',
    keyword: '',
  })

  async function loadStats() {
    try {
      const start = calcStartDate(filters.rangeDays)
      const end = new Date().toISOString()
      const data = await auditService.stats({
        start_time: start.slice(0, 10),
        end_time: end.slice(0, 10),
      })
      setStats({
        total: Number(data.total_count || 0),
        today: Number(data.today_count || 0),
        operation: Number(data.operation_count || 0),
        error: Number(data.error_count || 0),
      })
    } catch {
      setStats({})
    }
  }

  async function load(nextPage = page, nextFilters = filters) {
    setLoading(true)
    setError('')
    try {
      const data = await auditService.list({
        page: nextPage,
        page_size: pageSize,
        type: nextFilters.type || undefined,
        module: nextFilters.module.trim() || undefined,
        username: nextFilters.username.trim() || undefined,
        path: nextFilters.path.trim() || undefined,
        status_code: nextFilters.statusCode.trim() || undefined,
        keyword: nextFilters.keyword.trim() || undefined,
        start_time: calcStartDate(nextFilters.rangeDays),
        end_time: new Date().toISOString(),
      })
      const normalized = listFrom<AuditLog>(data)
      setItems(normalized.list)
      setTotal(normalized.total)
      setPage(nextPage)
    } catch (err) {
      setError(err instanceof Error ? err.message : '审计记录加载失败')
      setItems([])
      setTotal(0)
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    if (ready) {
      load(1, filters)
      loadStats()
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [ready])

  const totalPages = useMemo(() => Math.max(1, Math.ceil(total / pageSize)), [total, pageSize])

  if (!ready) return null
  if (!hasPermission(user, ACTIONS.USER_MODIFY)) {
    return (
      <AppShell>
        <ModuleHeader title="审计记录" desc="查看近期关键操作与异常请求" />
        <PageState error="当前账号无权限访问该页面（需要 user:modify 权限）" />
      </AppShell>
    )
  }

  return (
    <AppShell>
      <ModuleHeader
        title="审计记录"
        desc="默认聚焦近期记录，支持按类型、模块、人员、路径与状态码快速定位"
        right={
          <button className="btn" type="button" onClick={() => { load(1, filters); loadStats() }}>
            刷新
          </button>
        }
      />
      <NoticeBar notice={notice} onClose={() => setNotice(null)} />

      <div className="insight-grid" style={{ marginTop: 0 }}>
        <div className="stat-card stat-card-amber">
          <span>当前范围日志数</span>
          <strong>{stats.total || 0}</strong>
          <small>默认只展示近期审计记录</small>
        </div>
        <div className="stat-card stat-card-blue">
          <span>今日新增</span>
          <strong>{stats.today || 0}</strong>
          <small>今日产生的审计日志</small>
        </div>
        <div className="stat-card stat-card-green">
          <span>关键操作</span>
          <strong>{stats.operation || 0}</strong>
          <small>排除登录/登出/查询后的操作</small>
        </div>
        <div className="stat-card stat-card-rose">
          <span>异常记录</span>
          <strong>{stats.error || 0}</strong>
          <small>失败请求或错误事件</small>
        </div>
      </div>

      <section className="section-card">
        <form
          className="filters-grid audit-filters"
          onSubmit={(e: FormEvent) => {
            e.preventDefault()
            load(1, filters)
            loadStats()
          }}
        >
          <select className="select" value={String(filters.rangeDays)} onChange={(e) => setFilters((prev) => ({ ...prev, rangeDays: Number(e.target.value) }))}>
            {rangeOptions.map((item) => (
              <option key={item.value} value={item.value}>{item.label}</option>
            ))}
          </select>
          <select className="select" value={filters.type} onChange={(e) => setFilters((prev) => ({ ...prev, type: e.target.value }))}>
            {typeOptions.map((item) => (
              <option key={item.value} value={item.value}>{item.label}</option>
            ))}
          </select>
          <input className="input" placeholder="模块，如 用户管理" value={filters.module} onChange={(e) => setFilters((prev) => ({ ...prev, module: e.target.value }))} />
          <input className="input" placeholder="人员名" value={filters.username} onChange={(e) => setFilters((prev) => ({ ...prev, username: e.target.value }))} />
          <input className="input" placeholder="请求路径关键字" value={filters.path} onChange={(e) => setFilters((prev) => ({ ...prev, path: e.target.value }))} />
          <input className="input" placeholder="状态码" value={filters.statusCode} onChange={(e) => setFilters((prev) => ({ ...prev, statusCode: e.target.value.replace(/[^\d]/g, '') }))} />
          <input className="input audit-filters-keyword" placeholder="关键词（描述 / 用户 / trace_id / 请求）" value={filters.keyword} onChange={(e) => setFilters((prev) => ({ ...prev, keyword: e.target.value }))} />
          <div className="row wrap">
            <button className="btn primary" type="submit">应用筛选</button>
            <button
              className="btn ghost"
              type="button"
              onClick={() => {
                const next = { rangeDays: 7, type: '', module: '', username: '', path: '', statusCode: '', keyword: '' }
                setFilters(next)
                load(1, next)
              }}
            >
              重置
            </button>
          </div>
        </form>
      </section>

      <PageState loading={loading && items.length === 0} error={error} empty={!loading && !error && items.length === 0} onRetry={() => load(page, filters)} />

      {!error && items.length > 0 ? (
        <>
          <div className="section-card">
            <div className="row wrap" style={{ justifyContent: 'space-between', marginBottom: 12 }}>
              <div>
                <b>审计记录列表</b>
                <div className="hint">第 {page} / {totalPages} 页，共 {total} 条</div>
              </div>
            </div>
            <div className="table-wrap">
              <table className="table">
                <thead>
                  <tr>
                    <th>时间</th>
                    <th>模块</th>
                    <th>类型</th>
                    <th>人员</th>
                    <th>请求</th>
                    <th>状态码</th>
                    <th>状态</th>
                    <th>IP</th>
                  </tr>
                </thead>
                <tbody>
                  {items.map((row) => (
                    <tr key={row.id}>
                      <td>{fmtTime(row.created_at)}</td>
                      <td>{row.module || '-'}</td>
                      <td><StatusTag status={row.log_type || row.type || '-'} /></td>
                      <td>{row.username || row.user_id || '-'}</td>
                      <td className="audit-request-cell">
                        <div>{(row.method || row.request_method || '').toUpperCase()} {row.path || row.request_url || '-'}</div>
                        {row.action ? <div className="hint">{row.action}</div> : null}
                      </td>
                      <td>{row.status_code || row.response_code || '-'}</td>
                      <td><StatusTag status={row.status || '-'} /></td>
                      <td>{row.ip || row.request_ip || '-'}</td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </div>
          <Pagination page={page} pageSize={pageSize} total={total} onChange={(nextPage) => load(nextPage, filters)} />
        </>
      ) : null}
    </AppShell>
  )
}

function calcStartDate(days: number) {
  const date = new Date()
  date.setDate(date.getDate() - Math.max(0, days))
  return date.toISOString()
}
