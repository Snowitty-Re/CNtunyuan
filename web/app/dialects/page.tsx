'use client'

import Link from 'next/link'
import { useEffect, useState } from 'react'
import { AppShell } from '@/components/layout/AppShell'
import { ModuleHeader } from '@/components/shared/ModuleHeader'
import { ConfirmButton } from '@/components/shared/ConfirmButton'
import { PageState } from '@/components/shared/PageState'
import { Pagination } from '@/components/shared/Pagination'
import { NoticeBar, type Notice } from '@/components/shared/NoticeBar'
import { StatusTag } from '@/components/shared/StatusTag'
import { useAuthGuard } from '@/hooks/useAuthGuard'
import { fmtTime, listFrom } from '@/lib/data'
import { ACTIONS, hasPermission } from '@/lib/rbac'
import { dialectService } from '@/services/dialects'
import type { Dialect } from '@/types/api'

export default function DialectsPage() {
  const { ready, user } = useAuthGuard()
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [items, setItems] = useState<Dialect[]>([])
  const [total, setTotal] = useState(0)
  const [page, setPage] = useState(1)
  const [status, setStatus] = useState('')
  const [notice, setNotice] = useState<Notice | null>(null)

  const canModify = hasPermission(user, ACTIONS.DIALECT_MODIFY)
  const canManage = hasPermission(user, ACTIONS.DIALECT_MANAGE)

  async function load(nextPage = page) {
    setLoading(true)
    setError('')
    try {
      const data = await dialectService.list({ page: nextPage, page_size: 20, status: status || undefined })
      const normalized = listFrom<Dialect>(data)
      setItems(normalized.list)
      setTotal(normalized.total)
      setPage(nextPage)
    } catch (err) {
      setError(err instanceof Error ? err.message : '加载失败')
    } finally {
      setLoading(false)
    }
  }

  async function runManagedAction(runner: () => Promise<unknown>, successText: string) {
    try {
      await runner()
      await load(page)
      setNotice({ type: 'success', text: successText })
    } catch (err) {
      setNotice({ type: 'error', text: err instanceof Error ? err.message : '操作失败' })
    }
  }

  useEffect(() => {
    if (ready) load(1)
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [ready])

  if (!ready) return null

  return (
    <AppShell>
      <ModuleHeader
        title="方言中心"
        desc="按卡片模板采集、审核和管理方言录音"
        right={
          <div className="row wrap">
            {canManage ? (
              <Link className="btn" href="/dialects/cards">
                卡片管理
              </Link>
            ) : null}
            {canModify ? (
              <Link className="btn primary" href="/dialects/create">
                新建方言批次
              </Link>
            ) : null}
          </div>
        }
      />
      <NoticeBar notice={notice} onClose={() => setNotice(null)} />

      <form
        className="panel row wrap"
        onSubmit={(e) => {
          e.preventDefault()
          load(1)
        }}
      >
        <select className="select" value={status} onChange={(e) => setStatus(e.target.value)}>
          <option value="">全部状态</option>
          <option value="pending">待审核</option>
          <option value="active">已通过</option>
          <option value="inactive">不可见</option>
        </select>
        <button className="btn" type="submit">
          筛选
        </button>
      </form>

      <PageState loading={loading} error={error} empty={!loading && !error && items.length === 0} onRetry={() => load(page)} />
      {!loading && !error && items.length > 0 ? (
        <>
          <div className="table-wrap">
            <table className="table">
              <thead>
                <tr>
                  <th>标题</th>
                  <th>卡片词汇</th>
                  <th>采集地区</th>
                  <th>状态</th>
                  <th>批次</th>
                  <th>创建时间</th>
                  <th>操作</th>
                </tr>
              </thead>
              <tbody>
                {items.map((row) => (
                  <tr key={row.id}>
                    <td>{row.title || '-'}</td>
                    <td>{row.card_content || '-'}</td>
                    <td>{row.region || [row.province, row.city, row.district].filter(Boolean).join(' ') || '-'}</td>
                    <td>
                      <StatusTag status={row.status || '-'} />
                    </td>
                    <td>{row.batch_id || '-'}</td>
                    <td>{fmtTime(row.created_at)}</td>
                    <td>
                      <div className="row wrap">
                        <Link className="btn ghost" href={`/dialects/${row.id}`}>
                          详情
                        </Link>
                        {canManage ? (
                          <>
                            <button
                              className="btn"
                              type="button"
                              onClick={() => runManagedAction(() => dialectService.feature(row.id), '已设为精选')}
                            >
                              设为精选
                            </button>
                            <button
                              className="btn"
                              type="button"
                              onClick={() => runManagedAction(() => dialectService.updateStatus(row.id, 'active'), '审核通过')}
                            >
                              通过
                            </button>
                            <button
                              className="btn"
                              type="button"
                              onClick={() => runManagedAction(() => dialectService.updateStatus(row.id, 'inactive'), '已设为不可见')}
                            >
                              设为不可见
                            </button>
                            <ConfirmButton
                              text="删除"
                              message="确认删除该方言记录？"
                              onConfirm={() => runManagedAction(() => dialectService.remove(row.id), '方言记录已删除')}
                              className="btn danger"
                            />
                          </>
                        ) : null}
                      </div>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
          <Pagination page={page} pageSize={20} total={total} onChange={load} />
        </>
      ) : null}
    </AppShell>
  )
}
