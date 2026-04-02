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
import { taskService } from '@/services/tasks'
import type { TaskFollowUp } from '@/types/api'

export default function TaskFollowUpDetailPage() {
  const { ready } = useAuthGuard()
  const params = useParams<{ id: string; followUpId: string }>()
  const taskId = params.id
  const followUpId = params.followUpId

  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [followUp, setFollowUp] = useState<TaskFollowUp | null>(null)
  const [comments, setComments] = useState<any[]>([])
  const [comment, setComment] = useState('')
  const [notice, setNotice] = useState<Notice | null>(null)

  async function load() {
    setLoading(true)
    setError('')
    try {
      const [detail, commentData] = await Promise.all([
        taskService.followUpById(taskId, followUpId),
        taskService.followUpComments(taskId, followUpId, { page: 1, page_size: 100 }),
      ])
      setFollowUp(detail)
      setComments(listFrom<any>(commentData).list)
    } catch (err) {
      setError(err instanceof Error ? err.message : '加载失败')
    } finally {
      setLoading(false)
    }
  }

  async function submitComment(e: FormEvent) {
    e.preventDefault()
    if (!comment.trim()) return
    try {
      await taskService.addFollowUpComment(taskId, followUpId, comment.trim())
      setComment('')
      load()
      setNotice({ type: 'success', text: '评论发送成功' })
    } catch (err) {
      setNotice({ type: 'error', text: err instanceof Error ? err.message : '评论失败' })
    }
  }

  useEffect(() => {
    if (ready && taskId && followUpId) load()
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [ready, taskId, followUpId])

  if (!ready) return null

  return (
    <AppShell>
      <ModuleHeader title="跟进记录详情" desc="查看跟进内容、审批结果与评论讨论" />
      <NoticeBar notice={notice} onClose={() => setNotice(null)} />
      <PageState loading={loading} error={error} onRetry={load} />
      {!loading && !error && followUp ? (
        <div className="grid">
          <div className="section-card">
            <h3 className="card-title">记录概览</h3>
            <div className="meta-grid">
              <div className="meta-item">
                <div className="k">创建时间</div>
                <div className="v">{fmtTime(followUp.created_at)}</div>
              </div>
              <div className="meta-item">
                <div className="k">进度</div>
                <div className="v">{followUp.progress ?? 0}%</div>
              </div>
              <div className="meta-item">
                <div className="k">审批状态</div>
                <div className="v">
                  <StatusTag status={followUp.review_status || 'pending'} />
                </div>
              </div>
            </div>
            <div className="panel" style={{ marginTop: 10 }}>
              <div className="hint">跟进内容</div>
              <div style={{ marginTop: 6 }}>{followUp.content}</div>
            </div>
            <div className="panel" style={{ marginTop: 10 }}>
              <div className="hint">审批意见</div>
              <div style={{ marginTop: 6 }}>{followUp.review_remark || '暂无'}</div>
            </div>
            {followUp.attachments && followUp.attachments.length > 0 ? (
              <div style={{ marginTop: 12 }}>
                <h3 className="card-title">附件</h3>
                <ul className="soft-list">
                  {followUp.attachments.map((a) => (
                    <li key={a}>
                      <a href={a} target="_blank" rel="noreferrer">
                        {a}
                      </a>
                    </li>
                  ))}
                </ul>
              </div>
            ) : null}
          </div>

          <details className="section-toggle" open>
            <summary>评论讨论</summary>
            <div className="section-toggle-body">
            <form className="row" style={{ marginTop: 10 }} onSubmit={submitComment}>
              <input className="input" placeholder="输入评论内容" value={comment} onChange={(e) => setComment(e.target.value)} />
              <button className="btn primary" type="submit">
                发送
              </button>
            </form>
            <div className="table-wrap" style={{ marginTop: 10 }}>
              <table className="table">
                <thead>
                  <tr>
                    <th>时间</th>
                    <th>用户</th>
                    <th>评论</th>
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
          </details>
        </div>
      ) : null}
    </AppShell>
  )
}
