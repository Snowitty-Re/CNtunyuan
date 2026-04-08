'use client'

import { useEffect, useState } from 'react'
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
        desc="健康检查、依赖状态与指标预览"
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
          <div className="kpi-grid">
            <div className="kpi">
              <div className="label">服务状态</div>
              <div className="value">{String(health?.status || detail?.status || 'unknown')}</div>
            </div>
            <div className="kpi">
              <div className="label">数据库</div>
              <div className="value">{String(detail?.database?.status || detail?.components?.database?.status || 'unknown')}</div>
            </div>
            <div className="kpi">
              <div className="label">缓存</div>
              <div className="value">{String(detail?.redis?.status || detail?.components?.redis?.status || 'unknown')}</div>
            </div>
            <div className="kpi">
              <div className="label">运行环境</div>
              <div className="value">{String(detail?.env || detail?.environment || '-')}</div>
            </div>
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
          <div className="panel row wrap">
            <button
              className="btn ghost"
              type="button"
              onClick={() => {
                if (!metricsPreview) {
                  setNotice({ type: 'info', text: '暂无可复制的指标文本' })
                  return
                }
                navigator.clipboard.writeText(metricsPreview)
                  .then(() => setNotice({ type: 'success', text: '指标预览已复制到剪贴板' }))
                  .catch(() => setNotice({ type: 'error', text: '复制失败，请手动复制' }))
              }}
            >
              复制指标预览
            </button>
          </div>
        </>
      ) : null}
    </AppShell>
  )
}
