'use client'

import Link from 'next/link'
import { FormEvent, useEffect, useMemo, useState } from 'react'
import { useParams, useRouter } from 'next/navigation'
import { AppShell } from '@/components/layout/AppShell'
import { ModuleHeader } from '@/components/shared/ModuleHeader'
import { NoticeBar, type Notice } from '@/components/shared/NoticeBar'
import { PageState } from '@/components/shared/PageState'
import { StatusTag } from '@/components/shared/StatusTag'
import { useAuthGuard } from '@/hooks/useAuthGuard'
import { fmtTime, listFrom } from '@/lib/data'
import { taskService } from '@/services/tasks'
import { uploadService } from '@/services/upload'
import type { Task, TaskFollowUp } from '@/types/api'

const TASK_LIST_IDS_KEY = 'web_tasks_list_ids_v1'

type UploadedAttachment = {
  id: string
  url: string
}

export default function TaskDetailPage() {
  const { ready } = useAuthGuard()
  const params = useParams<{ id: string }>()
  const router = useRouter()
  const id = params.id

  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [item, setItem] = useState<Task | null>(null)
  const [followUps, setFollowUps] = useState<TaskFollowUp[]>([])
  const [logs, setLogs] = useState<any[]>([])
  const [content, setContent] = useState('')
  const [progress, setProgress] = useState('50')
  const [uploading, setUploading] = useState(false)
  const [attachments, setAttachments] = useState<UploadedAttachment[]>([])
  const [notice, setNotice] = useState<Notice | null>(null)
  const [listIds, setListIds] = useState<string[]>([])
  const nav = useMemo(() => {
    const idx = listIds.indexOf(id)
    return {
      prevId: idx > 0 ? listIds[idx - 1] : '',
      nextId: idx >= 0 && idx < listIds.length - 1 ? listIds[idx + 1] : '',
    }
  }, [id, listIds])

  async function load() {
    setLoading(true)
    setError('')
    try {
      const [detail, ups, logsData] = await Promise.all([
        taskService.byId(id),
        taskService.followUps(id, { page: 1, page_size: 50 }),
        taskService.logs(id, { page: 1, page_size: 50 }),
      ])
      setItem(detail)
      setFollowUps(listFrom<TaskFollowUp>(ups).list)
      setLogs(listFrom<any>(logsData).list)
    } catch (err) {
      setError(err instanceof Error ? err.message : '加载失败')
    } finally {
      setLoading(false)
    }
  }

  async function onCreateFollowUp(e: FormEvent) {
    e.preventDefault()
    if (!content.trim()) return
    try {
      const created = await taskService.createFollowUp(id, {
        content: content.trim(),
        progress: Number(progress) || 0,
        attachments: attachments.map((a) => a.url),
      })
      if (created?.id) {
        await Promise.all(
          attachments
            .filter((a) => a.id)
            .map((a) => uploadService.bind(a.id, 'task_follow_up', created.id)),
        )
      }
      setContent('')
      setProgress('50')
      setAttachments([])
      load()
      setNotice({ type: 'success', text: '跟进记录提交成功' })
    } catch (err) {
      setNotice({ type: 'error', text: err instanceof Error ? err.message : '新增失败' })
    }
  }

  async function onUploadFiles(files: FileList | null) {
    if (!files || files.length === 0) return
    setUploading(true)
    try {
      for (let i = 0; i < files.length; i += 1) {
        const file = files[i]
        const uploaded = await uploadService.uploadSingle(file, {
          entity_type: 'task_follow_up',
        })
        const url = uploaded.url || uploaded.path || ''
        if (uploaded.id && url) {
          setAttachments((prev) => [...prev, { id: uploaded.id as string, url }])
        } else if (url) {
          setAttachments((prev) => [...prev, { id: '', url }])
        }
      }
      setNotice({ type: 'success', text: '附件上传成功' })
    } catch (err) {
      setNotice({ type: 'error', text: err instanceof Error ? err.message : '上传失败' })
    } finally {
      setUploading(false)
    }
  }

  useEffect(() => {
    if (ready && id) load()
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [ready, id])

  useEffect(() => {
    if (typeof window === 'undefined') return
    try {
      const ids = JSON.parse(localStorage.getItem(TASK_LIST_IDS_KEY) || '[]')
      setListIds(Array.isArray(ids) ? ids : [])
    } catch {
      setListIds([])
    }
  }, [id])

  if (!ready) {
    return (
      <AppShell>
        <PageState loading />
      </AppShell>
    )
  }

  return (
    <AppShell>
      <ModuleHeader title="任务详情" desc="执行进度、跟进记录与审批协作" />
      <NoticeBar notice={notice} onClose={() => setNotice(null)} />
      <div className="panel row wrap" style={{ marginTop: 0 }}>
        <button className="btn ghost" type="button" disabled={!nav.prevId} onClick={() => nav.prevId && router.push(`/tasks/${nav.prevId}`)}>
          上一个任务
        </button>
        <button className="btn ghost" type="button" disabled={!nav.nextId} onClick={() => nav.nextId && router.push(`/tasks/${nav.nextId}`)}>
          下一个任务
        </button>
      </div>
      <PageState loading={loading} error={error} onRetry={load} />
      {!loading && !error && item ? (
        <div className="grid">
          <div className="section-card">
            <div className="row wrap" style={{ justifyContent: 'space-between' }}>
              <h3 className="card-title">{item.title}</h3>
              <div className="row wrap">
                {item.status === 'assigned' ? (
                  <button
                    className="btn"
                    type="button"
                    onClick={() =>
                      taskService
                        .start(id)
                        .then(() => {
                          setNotice({ type: 'success', text: '任务已开始' })
                          return load()
                        })
                        .catch((err) => setNotice({ type: 'error', text: err instanceof Error ? err.message : '开始失败' }))
                    }
                  >
                    开始任务
                  </button>
                ) : null}
                {item.status === 'pending' ? (
                  <span className="hint">待分配后执行人可开始任务</span>
                ) : null}
                {item.status === 'processing' || item.status === 'assigned' ? (
                  <button
                    className="btn"
                    type="button"
                    onClick={() =>
                      taskService
                        .complete(id, { result: 'web详情页完成' })
                        .then(() => {
                          setNotice({ type: 'success', text: '任务已完成' })
                          return load()
                        })
                        .catch((err) => setNotice({ type: 'error', text: err instanceof Error ? err.message : '完成失败' }))
                    }
                  >
                    完成任务
                  </button>
                ) : null}
                {item.status !== 'cancelled' && item.status !== 'completed' ? (
                  <button
                    className="btn danger"
                    type="button"
                    onClick={() =>
                      taskService
                        .cancel(id, 'web端取消')
                        .then(() => {
                          setNotice({ type: 'success', text: '任务已取消' })
                          return load()
                        })
                        .catch((err) => setNotice({ type: 'error', text: err instanceof Error ? err.message : '取消失败' }))
                    }
                  >
                    取消任务
                  </button>
                ) : null}
              </div>
            </div>
            <div className="meta-grid">
              <div className="meta-item">
                <div className="k">状态</div>
                <div className="v">
                  <StatusTag status={item.status || '-'} />
                </div>
              </div>
              <div className="meta-item">
                <div className="k">优先级</div>
                <div className="v">{item.priority || '-'}</div>
              </div>
              <div className="meta-item">
                <div className="k">进度</div>
                <div className="v">{item.progress ?? 0}%</div>
              </div>
              <div className="meta-item">
                <div className="k">执行人</div>
                <div className="v">{item.assignee?.nickname || '-'}</div>
              </div>
              <div className="meta-item">
                <div className="k">关联案件</div>
                <div className="v">{item.missing_person?.name || '-'}</div>
              </div>
              <div className="meta-item">
                <div className="k">更新时间</div>
                <div className="v">{fmtTime(item.created_at)}</div>
              </div>
            </div>
            <div className="panel" style={{ marginTop: 10 }}>
              <div className="hint">任务描述</div>
              <div style={{ marginTop: 6 }}>{item.description || '暂无描述'}</div>
            </div>
          </div>

          <div className="section-card">
            <h3 className="card-title">新增跟进记录</h3>
            <div className="hint" style={{ marginTop: 4 }}>
              记录任务阶段性动作，支持附件上传，后续可评论与审批。
            </div>
            <form className="grid" style={{ marginTop: 10 }} onSubmit={onCreateFollowUp}>
              <textarea className="textarea" value={content} onChange={(e) => setContent(e.target.value)} placeholder="跟进内容" />
              <div className="progress-input-wrap">
                <input className="input" value={progress} onChange={(e) => setProgress(e.target.value.replace(/[^\d]/g, ''))} placeholder="进度 0-100" />
                <span className="progress-badge">{progress || 0}%</span>
              </div>
              <label>
                <div style={{ marginBottom: 6 }}>附件上传</div>
                <input className="input" type="file" multiple onChange={(e) => onUploadFiles(e.target.files)} />
              </label>
              {attachments.length > 0 ? (
                <div className="table-wrap">
                  <table className="table">
                    <thead>
                      <tr>
                        <th>已上传附件</th>
                        <th>操作</th>
                      </tr>
                    </thead>
                    <tbody>
                      {attachments.map((a) => (
                        <tr key={`${a.id}-${a.url}`}>
                          <td>{a.url}</td>
                          <td>
                            <button
                              className="btn danger"
                              type="button"
                              onClick={() => setAttachments((prev) => prev.filter((x) => x.url !== a.url))}
                            >
                              移除
                            </button>
                          </td>
                        </tr>
                      ))}
                    </tbody>
                  </table>
                </div>
              ) : null}
              <div className="row">
                <button className="btn primary" type="submit" disabled={uploading}>
                  {uploading ? '上传中...' : '提交跟进'}
                </button>
              </div>
            </form>
          </div>

          <details className="section-toggle" open>
            <summary>跟进记录</summary>
            <div className="section-toggle-body">
            {followUps.length === 0 ? (
              <div style={{ marginTop: 10, color: '#6b7280' }}>暂无记录</div>
            ) : (
              <div className="table-wrap" style={{ marginTop: 10 }}>
                <table className="table">
                  <thead>
                    <tr>
                      <th>时间</th>
                      <th>内容</th>
                      <th>进度</th>
                      <th>审批状态</th>
                      <th>操作</th>
                    </tr>
                  </thead>
                  <tbody>
                    {followUps.map((f) => (
                      <tr key={f.id}>
                        <td>{fmtTime(f.created_at)}</td>
                        <td>{f.content}</td>
                        <td>{f.progress ?? 0}%</td>
                        <td>
                          <StatusTag status={f.review_status || 'pending'} />
                        </td>
                        <td>
                          <div className="row wrap">
                            <Link className="btn ghost" href={`/tasks/${id}/follow-ups/${f.id}`}>
                              详情
                            </Link>
                            <button
                              className="btn"
                              type="button"
                              onClick={() => taskService.reviewFollowUp(id, f.id, { approve: true, remark: '通过' }).then(load)}
                            >
                              通过
                            </button>
                            <button
                              className="btn danger"
                              type="button"
                              onClick={() => taskService.reviewFollowUp(id, f.id, { approve: false, remark: '驳回' }).then(load)}
                            >
                              驳回
                            </button>
                          </div>
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            )}
            </div>
          </details>

          <details className="section-toggle">
            <summary>任务日志</summary>
            <div className="section-toggle-body">
            {logs.length === 0 ? (
              <div style={{ marginTop: 10, color: '#6b7280' }}>暂无日志</div>
            ) : (
              <div className="table-wrap" style={{ marginTop: 10 }}>
                <table className="table">
                  <thead>
                    <tr>
                      <th>时间</th>
                      <th>操作</th>
                      <th>操作者</th>
                      <th>备注</th>
                    </tr>
                  </thead>
                  <tbody>
                    {logs.map((l) => (
                      <tr key={l.id || `${l.created_at}-${l.action}`}>
                        <td>{fmtTime(l.created_at || l.time)}</td>
                        <td>{l.action || l.type || '-'}</td>
                        <td>{l.operator_name || l.operator || l.user_name || '-'}</td>
                        <td>{l.remark || l.content || '-'}</td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            )}
            </div>
          </details>
        </div>
      ) : null}
    </AppShell>
  )
}
