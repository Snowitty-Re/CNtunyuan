'use client'

import { FormEvent, useEffect, useState } from 'react'
import { useParams } from 'next/navigation'
import { AppShell } from '@/components/layout/AppShell'
import { ModuleHeader } from '@/components/shared/ModuleHeader'
import { PageState } from '@/components/shared/PageState'
import { StatusTag } from '@/components/shared/StatusTag'
import { useAuthGuard } from '@/hooks/useAuthGuard'
import { fmtTime, listFrom } from '@/lib/data'
import { dialectService } from '@/services/dialects'
import type { Dialect } from '@/types/api'

export default function DialectDetailPage() {
  const { ready } = useAuthGuard()
  const params = useParams<{ id: string }>()
  const id = params.id
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [item, setItem] = useState<Dialect | null>(null)
  const [comments, setComments] = useState<any[]>([])
  const [comment, setComment] = useState('')

  async function load() {
    setLoading(true)
    setError('')
    try {
      const [data, commentData] = await Promise.all([dialectService.byId(id), dialectService.comments(id, { page: 1, page_size: 100 })])
      setItem(data)
      setComments(listFrom<any>(commentData).list)
    } catch (err) {
      setError(err instanceof Error ? err.message : '加载失败')
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    if (ready && id) load()
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [ready, id])

  if (!ready) return null

  async function submitComment(e: FormEvent) {
    e.preventDefault()
    if (!comment.trim()) return
    try {
      await dialectService.addComment(id, comment.trim())
      setComment('')
      load()
    } catch (err) {
      alert(err instanceof Error ? err.message : '评论失败')
    }
  }

  return (
    <AppShell>
      <ModuleHeader title="方言详情" desc="查看录音信息、状态与使用情况" />
      <PageState loading={loading} error={error} onRetry={load} />
      {!loading && !error && item ? (
        <div className="section-card">
          <div className="grid cols-2">
            <div>标题：{item.title}</div>
            <div>
              状态：<StatusTag status={item.status || '-'} />
            </div>
            <div>区域：{item.region || '-'}</div>
            <div>创建时间：{fmtTime(item.created_at)}</div>
            <div>播放数：{item.play_count || 0}</div>
            <div>点赞数：{item.like_count || 0}</div>
          </div>
          <div style={{ marginTop: 10, color: '#6b7280' }}>{item.content || '暂无内容'}</div>
          <div className="row wrap" style={{ marginTop: 10 }}>
            <button className="btn" type="button" onClick={() => dialectService.updateStatus(id, 'approved').then(load)}>
              审核通过
            </button>
            <button className="btn danger" type="button" onClick={() => dialectService.updateStatus(id, 'rejected').then(load)}>
              审核驳回
            </button>
            <button className="btn" type="button" onClick={() => dialectService.updateStatus(id, 'inactive').then(load)}>
              设为不可见
            </button>
            <button className="btn" type="button" onClick={() => dialectService.feature(id).then(load)}>
              设为精选
            </button>
          </div>
          {item.audio_url ? (
            <div style={{ marginTop: 12 }}>
              <audio controls src={item.audio_url} style={{ width: '100%' }} />
            </div>
          ) : null}
        </div>
      ) : null}

      {!loading && !error && item ? (
        <div className="section-card" style={{ marginTop: 12 }}>
          <b>评论管理</b>
          <form className="row" style={{ marginTop: 10 }} onSubmit={submitComment}>
            <input className="input" placeholder="输入评论内容" value={comment} onChange={(e) => setComment(e.target.value)} />
            <button className="btn primary" type="submit">
              发表评论
            </button>
          </form>
          <div className="table-wrap" style={{ marginTop: 10 }}>
            <table className="table">
              <thead>
                <tr>
                  <th>时间</th>
                  <th>用户</th>
                  <th>内容</th>
                </tr>
              </thead>
              <tbody>
                {comments.length === 0 ? (
                  <tr>
                    <td colSpan={3}>暂无评论</td>
                  </tr>
                ) : (
                  comments.map((c) => (
                    <tr key={c.id || `${c.created_at}-${c.content}`}>
                      <td>{fmtTime(c.created_at)}</td>
                      <td>{c.user?.nickname || c.user_name || c.user_id || '-'}</td>
                      <td>{c.content || '-'}</td>
                    </tr>
                  ))
                )}
              </tbody>
            </table>
          </div>
        </div>
      ) : null}
    </AppShell>
  )
}
