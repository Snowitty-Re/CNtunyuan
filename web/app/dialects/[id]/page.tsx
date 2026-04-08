'use client'

import { FormEvent, useEffect, useState } from 'react'
import { useParams } from 'next/navigation'
import { AppShell } from '@/components/layout/AppShell'
import { ModuleHeader } from '@/components/shared/ModuleHeader'
import { NoticeBar, type Notice } from '@/components/shared/NoticeBar'
import { PageState } from '@/components/shared/PageState'
import { StatusTag } from '@/components/shared/StatusTag'
import { useAuthGuard } from '@/hooks/useAuthGuard'
import { fmtTime, listFrom } from '@/lib/data'
import { ACTIONS, hasPermission } from '@/lib/rbac'
import { dialectService } from '@/services/dialects'
import type { Dialect } from '@/types/api'

export default function DialectDetailPage() {
  const { ready, user } = useAuthGuard()
  const params = useParams<{ id: string }>()
  const id = params.id
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [item, setItem] = useState<Dialect | null>(null)
  const [comments, setComments] = useState<any[]>([])
  const [comment, setComment] = useState('')
  const [notice, setNotice] = useState<Notice | null>(null)

  const canManage = hasPermission(user, ACTIONS.DIALECT_MANAGE)

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
      await load()
      setNotice({ type: 'success', text: '评论发布成功' })
    } catch (err) {
      setNotice({ type: 'error', text: err instanceof Error ? err.message : '评论失败' })
    }
  }

  async function runManageAction(action: () => Promise<unknown>, successText: string) {
    try {
      await action()
      await load()
      setNotice({ type: 'success', text: successText })
    } catch (err) {
      setNotice({ type: 'error', text: err instanceof Error ? err.message : '操作失败' })
    }
  }

  return (
    <AppShell>
      <ModuleHeader title="方言详情" desc="查看卡片录音、采集信息、审核状态和评论" />
      <NoticeBar notice={notice} onClose={() => setNotice(null)} />
      <PageState loading={loading} error={error} onRetry={load} />
      {!loading && !error && item ? (
        <div className="section-card">
          <div className="grid cols-2">
            <div>标题：{item.title || '-'}</div>
            <div>
              状态：<StatusTag status={item.status || '-'} />
            </div>
            <div>卡片词汇：{item.card_content || '-'}</div>
            <div>批次ID：{item.batch_id || '-'}</div>
            <div>采集地区：{item.region || [item.province, item.city, item.district].filter(Boolean).join(' ') || '-'}</div>
            <div>采集地址：{item.collect_address || '-'}</div>
            <div>录音时长：{item.duration || 0}s</div>
            <div>格式：{item.format || '-'}</div>
            <div>创建时间：{fmtTime(item.created_at)}</div>
            <div>播放/点赞：{item.play_count || 0}/{item.like_count || 0}</div>
          </div>

          {item.card_image_url ? (
            <div style={{ marginTop: 12 }}>
              <b>卡片图片</b>
              <img
                src={item.card_image_url}
                alt={item.card_content || 'dialect-card'}
                style={{ width: '100%', maxHeight: 320, objectFit: 'contain', marginTop: 8, borderRadius: 10, border: '1px solid var(--border)' }}
              />
            </div>
          ) : null}

          <div style={{ marginTop: 10, color: '#6b7280' }}>{item.description || item.content || '暂无内容'}</div>
          {canManage ? (
            <div className="row wrap" style={{ marginTop: 10 }}>
              <button className="btn" type="button" onClick={() => runManageAction(() => dialectService.updateStatus(id, 'active'), '审核通过')}>
                审核通过
              </button>
              <button className="btn" type="button" onClick={() => runManageAction(() => dialectService.updateStatus(id, 'inactive'), '已设为不可见')}>
                设为不可见
              </button>
              <button className="btn" type="button" onClick={() => runManageAction(() => dialectService.feature(id), '已设为精选')}>
                设为精选
              </button>
            </div>
          ) : null}
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
                  comments.map((entry) => (
                    <tr key={entry.id || `${entry.created_at}-${entry.content}`}>
                      <td>{fmtTime(entry.created_at)}</td>
                      <td>{entry.user?.nickname || entry.user_name || entry.user_id || '-'}</td>
                      <td>{entry.content || '-'}</td>
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
