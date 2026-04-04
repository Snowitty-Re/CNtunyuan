'use client'

import Link from 'next/link'
import { FormEvent, useEffect, useState } from 'react'
import { AppShell } from '@/components/layout/AppShell'
import { ModuleHeader } from '@/components/shared/ModuleHeader'
import { ConfirmButton } from '@/components/shared/ConfirmButton'
import { PageState } from '@/components/shared/PageState'
import { Pagination } from '@/components/shared/Pagination'
import { NoticeBar, type Notice } from '@/components/shared/NoticeBar'
import { StatusTag } from '@/components/shared/StatusTag'
import { useAuthGuard } from '@/hooks/useAuthGuard'
import { fmtTime, listFrom } from '@/lib/data'
import { dialectService } from '@/services/dialects'
import type { Dialect } from '@/types/api'

export default function DialectsPage() {
  const { ready } = useAuthGuard()
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [items, setItems] = useState<Dialect[]>([])
  const [total, setTotal] = useState(0)
  const [page, setPage] = useState(1)
  const [status, setStatus] = useState('')
  const [title, setTitle] = useState('')
  const [audioUrl, setAudioUrl] = useState('')
  const [region, setRegion] = useState('')
  const [quickOpen, setQuickOpen] = useState(false)
  const [notice, setNotice] = useState<Notice | null>(null)

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

  async function quickCreate(e: FormEvent) {
    e.preventDefault()
    if (!title.trim() || !audioUrl.trim() || !region.trim()) return
    try {
      await dialectService.create({
        title: title.trim(),
        region: region.trim(),
        audio_url: audioUrl.trim(),
      })
      setTitle('')
      setRegion('')
      setAudioUrl('')
      load(1)
      setNotice({ type: 'success', text: '方言记录创建成功' })
    } catch (err) {
      setNotice({ type: 'error', text: err instanceof Error ? err.message : '创建失败' })
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
        desc="管理方言录音、审核状态与精选内容"
        right={
          <div className="row wrap">
            <button className="btn" type="button" onClick={() => setQuickOpen((v) => !v)}>
              {quickOpen ? '收起快速录入' : '快速录入'}
            </button>
            <Link className="btn primary" href="/dialects/create">
              完整新建
            </Link>
          </div>
        }
      />
      <NoticeBar notice={notice} onClose={() => setNotice(null)} />
      <div className="grid cols-2">
        {quickOpen ? (
          <form className="panel grid" onSubmit={quickCreate}>
            <b>快速新建方言记录</b>
            <input className="input" placeholder="标题（必填）" value={title} onChange={(e) => setTitle(e.target.value)} />
            <input className="input" placeholder="区域（必填）" value={region} onChange={(e) => setRegion(e.target.value)} />
            <input className="input" placeholder="音频URL（必填）" value={audioUrl} onChange={(e) => setAudioUrl(e.target.value)} />
            <div className="row wrap">
              <button className="btn primary" type="submit">
                创建
              </button>
              <span className="hint">复杂字段请使用“完整新建”页面</span>
            </div>
          </form>
        ) : (
          <div className="panel">
            <b>快速录入已收起</b>
            <div className="hint" style={{ marginTop: 8 }}>
              推荐使用“完整新建”录入标题、区域、语音元数据、标签和关联案件。
            </div>
          </div>
        )}
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
      </div>
      <PageState loading={loading} error={error} empty={!loading && !error && items.length === 0} onRetry={() => load(page)} />
      {!loading && !error && items.length > 0 ? (
        <>
          <div className="table-wrap">
            <table className="table">
              <thead>
                <tr>
                  <th>标题</th>
                  <th>区域</th>
                  <th>状态</th>
                  <th>播放/点赞</th>
                  <th>创建时间</th>
                  <th>操作</th>
                </tr>
              </thead>
              <tbody>
                {items.map((row) => (
                  <tr key={row.id}>
                    <td>{row.title}</td>
                    <td>{row.region || `${row.province || ''}${row.city || ''}` || '-'}</td>
                    <td>
                      <StatusTag status={row.status || '-'} />
                    </td>
                    <td>
                      {row.play_count || 0}/{row.like_count || 0}
                    </td>
                    <td>{fmtTime(row.created_at)}</td>
                    <td>
                      <div className="row wrap">
                        <Link className="btn ghost" href={`/dialects/${row.id}`}>
                          详情
                        </Link>
                        <button className="btn" type="button" onClick={() => dialectService.feature(row.id).then(() => load(page))}>
                          设为精选
                        </button>
                        <button className="btn" type="button" onClick={() => dialectService.updateStatus(row.id, 'active').then(() => load(page))}>
                          通过
                        </button>
                        <button className="btn" type="button" onClick={() => dialectService.updateStatus(row.id, 'inactive').then(() => load(page))}>
                          设为不可见
                        </button>
                        <ConfirmButton
                          text="删除"
                          message="确认删除该方言记录？"
                          onConfirm={() => {
                            dialectService.remove(row.id).then(() => load(page))
                          }}
                          className="btn danger"
                        />
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
