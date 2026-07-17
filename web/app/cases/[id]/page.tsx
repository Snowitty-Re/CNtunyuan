'use client'

import Image from 'next/image'
import Link from 'next/link'
import { FormEvent, useEffect, useMemo, useState } from 'react'
import { useParams, useRouter } from 'next/navigation'
import { AppShell } from '@/components/layout/AppShell'
import { ConfirmButton } from '@/components/shared/ConfirmButton'
import { ModuleHeader } from '@/components/shared/ModuleHeader'
import { NoticeBar, type Notice } from '@/components/shared/NoticeBar'
import { PageState } from '@/components/shared/PageState'
import { StatusTag } from '@/components/shared/StatusTag'
import { useAuthGuard } from '@/hooks/useAuthGuard'
import { fmtTime, joinLocation, listFrom } from '@/lib/data'
import { missingPersonService } from '@/services/missingPersons'
import { taskService } from '@/services/tasks'
import { uploadService } from '@/services/upload'
import type { MissingPerson, MissingTrack, Task } from '@/types/api'

const CASE_LIST_IDS_KEY = 'web_cases_list_ids_v1'

export default function CaseDetailPage() {
  const { ready } = useAuthGuard()
  const params = useParams<{ id: string }>()
  const router = useRouter()
  const id = params.id

  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [item, setItem] = useState<MissingPerson | null>(null)
  const [tracks, setTracks] = useState<MissingTrack[]>([])
  const [tasks, setTasks] = useState<Task[]>([])
  const [trackContent, setTrackContent] = useState('')
  const [trackLocation, setTrackLocation] = useState('')
  const [foundLocation, setFoundLocation] = useState('')
  const [foundDesc, setFoundDesc] = useState('')
  const [editing, setEditing] = useState(false)
  const [saving, setSaving] = useState(false)
  const [taskCreating, setTaskCreating] = useState(false)
  const [taskTitle, setTaskTitle] = useState('')
  const [taskDesc, setTaskDesc] = useState('')
  const [taskPriority, setTaskPriority] = useState('medium')
  const [photoUploading, setPhotoUploading] = useState(false)
  const [notice, setNotice] = useState<Notice | null>(null)
  const [listIds, setListIds] = useState<string[]>([])
  const nav = useMemo(() => {
    const idx = listIds.indexOf(id)
    return {
      prevId: idx > 0 ? listIds[idx - 1] : '',
      nextId: idx >= 0 && idx < listIds.length - 1 ? listIds[idx + 1] : '',
    }
  }, [id, listIds])

  const [editForm, setEditForm] = useState({
    name: '',
    age: '',
    gender: 'male',
    province: '',
    city: '',
    district: '',
    address: '',
    description: '',
    contact_name: '',
    contact_phone: '',
    case_type: 'other',
    photo_url: '',
  })

  async function load() {
    setLoading(true)
    setError('')
    try {
      const [detail, trackData, taskData] = await Promise.all([
        missingPersonService.byId(id),
        missingPersonService.tracks(id),
        taskService.list({ page: 1, page_size: 50, missing_person_id: id }),
      ])
      const normalized = listFrom<MissingTrack>(trackData)
      const taskNormalized = listFrom<Task>(taskData)
      setItem(detail)
      setTracks(normalized.list)
      setTasks(taskNormalized.list.filter((t) => t.missing_person_id === id || t.missing_person?.id === id))
      setEditForm({
        name: detail.name || '',
        age: detail.age ? String(detail.age) : '',
        gender: detail.gender || 'male',
        province: detail.province || '',
        city: detail.city || '',
        district: detail.district || '',
        address: detail.address || '',
        description: detail.description || '',
        contact_name: detail.contact_name || '',
        contact_phone: detail.contact_phone || '',
        case_type: detail.case_type || 'other',
        photo_url: detail.photo_url || '',
      })
      setTaskTitle(`案件跟进：${detail.name || '未命名'}`)
    } catch (err) {
      setError(err instanceof Error ? err.message : '加载失败')
    } finally {
      setLoading(false)
    }
  }

  async function submitTrack(e: FormEvent) {
    e.preventDefault()
    if (!trackContent.trim()) return
    try {
      await missingPersonService.addTrack(id, {
        description: trackContent.trim(),
        location: trackLocation.trim(),
        time: new Date().toISOString(),
      })
      setTrackContent('')
      setTrackLocation('')
      load()
      setNotice({ type: 'success', text: '线索添加成功' })
    } catch (err) {
      setNotice({ type: 'error', text: err instanceof Error ? err.message : '新增线索失败' })
    }
  }

  async function saveCase(e: FormEvent) {
    e.preventDefault()
    if (!editForm.name.trim()) return
    setSaving(true)
    try {
      await missingPersonService.update(id, {
        name: editForm.name.trim(),
        age: Number(editForm.age) || 0,
        gender: editForm.gender,
        province: editForm.province.trim(),
        city: editForm.city.trim(),
        district: editForm.district.trim(),
        address: editForm.address.trim(),
        description: editForm.description.trim(),
        contact_name: editForm.contact_name.trim(),
        contact_phone: editForm.contact_phone.trim(),
        case_type: editForm.case_type,
        photo_url: editForm.photo_url.trim(),
      })
      setEditing(false)
      load()
      setNotice({ type: 'success', text: '案件信息已保存' })
    } catch (err) {
      setNotice({ type: 'error', text: err instanceof Error ? err.message : '保存失败' })
    } finally {
      setSaving(false)
    }
  }

  async function createRelatedTask(e: FormEvent) {
    e.preventDefault()
    if (!taskTitle.trim()) return
    setTaskCreating(true)
    try {
      await taskService.create({
        title: taskTitle.trim(),
        description: taskDesc.trim(),
        missing_person_id: id,
        priority: taskPriority,
        type: 'search',
      })
      setTaskDesc('')
      load()
      setNotice({ type: 'success', text: '关联任务创建成功' })
    } catch (err) {
      setNotice({ type: 'error', text: err instanceof Error ? err.message : '创建关联任务失败' })
    } finally {
      setTaskCreating(false)
    }
  }

  async function uploadCasePhoto(files: FileList | null) {
    if (!files || files.length === 0) return
    setPhotoUploading(true)
    try {
      const file = files[0]
      const uploaded = await uploadService.uploadSingle(file, {
        entity_type: 'missing_person',
        entity_id: id,
      })
      const url = uploaded.url || uploaded.path || ''
      if (uploaded.id) {
        await uploadService.bind(uploaded.id, 'missing_person', id)
      }
      if (url) {
        setEditForm((s) => ({ ...s, photo_url: url }))
      }
      setNotice({ type: 'success', text: '案件照片上传成功' })
    } catch (err) {
      setNotice({ type: 'error', text: err instanceof Error ? err.message : '上传照片失败' })
    } finally {
      setPhotoUploading(false)
    }
  }

  useEffect(() => {
    if (ready && id) load()
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [ready, id])

  useEffect(() => {
    if (typeof window === 'undefined') return
    try {
      const ids = JSON.parse(localStorage.getItem(CASE_LIST_IDS_KEY) || '[]')
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

  async function runAction(action: () => Promise<unknown>, okText: string) {
    try {
      await action()
      await load()
      setNotice({ type: 'success', text: okText })
    } catch (err) {
      setNotice({ type: 'error', text: err instanceof Error ? err.message : '操作失败' })
    }
  }

  return (
    <AppShell>
      <ModuleHeader title="案件详情" desc="维护案件状态、线索与闭环信息" />
      <NoticeBar notice={notice} onClose={() => setNotice(null)} />
      <div className="panel row wrap" style={{ marginTop: 0 }}>
        <button className="btn ghost" type="button" disabled={!nav.prevId} onClick={() => nav.prevId && router.push(`/cases/${nav.prevId}`)}>
          上一个案件
        </button>
        <button className="btn ghost" type="button" disabled={!nav.nextId} onClick={() => nav.nextId && router.push(`/cases/${nav.nextId}`)}>
          下一个案件
        </button>
      </div>
      <PageState loading={loading} error={error} onRetry={load} />
      {!loading && !error && item ? (
        <div className="grid">
          <div className="section-card">
            <div className="row wrap" style={{ justifyContent: 'space-between' }}>
              <h3 className="card-title">{item.name}</h3>
              <div className="row wrap">
                <button className="btn" type="button" onClick={() => setEditing((v) => !v)}>
                  {editing ? '取消编辑' : '编辑案件'}
                </button>
                <button
                  className="btn"
                  type="button"
                  onClick={() => runAction(() => missingPersonService.updateStatus(id, 'searching'), '已设为寻找中')}
                >
                  设为寻找中
                </button>
                <button
                  className="btn"
                  type="button"
                  onClick={() =>
                    runAction(
                      () =>
                        missingPersonService.markFound(id, {
                          found_location: foundLocation.trim(),
                          found_time: new Date().toISOString(),
                          description: foundDesc.trim(),
                        }),
                      '已标记找到',
                    )
                  }
                >
                  标记已找到
                </button>
                <button
                  className="btn"
                  type="button"
                  onClick={() => runAction(() => missingPersonService.markReunited(id), '已标记团圆')}
                >
                  标记已团圆
                </button>
                <ConfirmButton
                  text="删除案件"
                  message="确认删除当前案件？"
                  onConfirm={() => {
                    missingPersonService
                      .remove(id)
                      .then(() => router.push('/cases'))
                      .catch((err) => setNotice({ type: 'error', text: err instanceof Error ? err.message : '删除失败' }))
                  }}
                />
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
                <div className="k">走失时间</div>
                <div className="v">{fmtTime(item.missing_time)}</div>
              </div>
              <div className="meta-item">
                <div className="k">走失地点</div>
                <div className="v">{joinLocation(item)}</div>
              </div>
              <div className="meta-item">
                <div className="k">联系人</div>
                <div className="v">{item.contact_name || '-'}</div>
              </div>
              <div className="meta-item">
                <div className="k">联系电话</div>
                <div className="v">{item.contact_phone || '—'}</div>
              </div>
              <div className="meta-item">
                <div className="k">关系</div>
                <div className="v">{item.contact_rel || '-'}</div>
              </div>
            </div>
            <div className="panel" style={{ marginTop: 10 }}>
              <div className="hint">案件描述</div>
              <div style={{ marginTop: 6 }}>{item.description || '暂无描述'}</div>
            </div>
            {item.photo_url ? (
              <div style={{ marginTop: 10 }}>
                <Image
                  src={item.photo_url}
                  alt={item.name}
                  width={220}
                  height={160}
                  unoptimized
                  style={{ borderRadius: 8, border: '1px solid #e5e7eb', objectFit: 'cover', width: 220, height: 160 }}
                />
              </div>
            ) : null}
            <div className="grid cols-2" style={{ marginTop: 12 }}>
              <input
                className="input"
                placeholder="找到地点（可选）"
                value={foundLocation}
                onChange={(e) => setFoundLocation(e.target.value)}
              />
              <input
                className="input"
                placeholder="找到说明（可选）"
                value={foundDesc}
                onChange={(e) => setFoundDesc(e.target.value)}
              />
            </div>
            {editing ? (
              <form className="grid cols-2" style={{ marginTop: 12 }} onSubmit={saveCase}>
                <input className="input" value={editForm.name} onChange={(e) => setEditForm((s) => ({ ...s, name: e.target.value }))} placeholder="姓名" />
                <input className="input" value={editForm.age} onChange={(e) => setEditForm((s) => ({ ...s, age: e.target.value.replace(/[^\d]/g, '') }))} placeholder="年龄" />
                <select className="select" value={editForm.gender} onChange={(e) => setEditForm((s) => ({ ...s, gender: e.target.value }))}>
                  <option value="male">male</option>
                  <option value="female">female</option>
                  <option value="other">other</option>
                </select>
                <select className="select" value={editForm.case_type} onChange={(e) => setEditForm((s) => ({ ...s, case_type: e.target.value }))}>
                  <option value="other">other</option>
                  <option value="child">child</option>
                  <option value="elderly">elderly</option>
                  <option value="adult">adult</option>
                </select>
                <input className="input" value={editForm.province} onChange={(e) => setEditForm((s) => ({ ...s, province: e.target.value }))} placeholder="省" />
                <input className="input" value={editForm.city} onChange={(e) => setEditForm((s) => ({ ...s, city: e.target.value }))} placeholder="市" />
                <input className="input" value={editForm.district} onChange={(e) => setEditForm((s) => ({ ...s, district: e.target.value }))} placeholder="区/县" />
                <input className="input" value={editForm.address} onChange={(e) => setEditForm((s) => ({ ...s, address: e.target.value }))} placeholder="详细地址" />
                <input className="input" value={editForm.contact_name} onChange={(e) => setEditForm((s) => ({ ...s, contact_name: e.target.value }))} placeholder="联系人" />
                <input className="input" value={editForm.contact_phone} onChange={(e) => setEditForm((s) => ({ ...s, contact_phone: e.target.value }))} placeholder="联系电话" />
                <textarea className="textarea" style={{ gridColumn: '1 / -1' }} value={editForm.description} onChange={(e) => setEditForm((s) => ({ ...s, description: e.target.value }))} placeholder="案件描述" />
                <label style={{ gridColumn: '1 / -1' }}>
                  <div style={{ marginBottom: 6 }}>案件照片上传</div>
                  <input className="input" type="file" accept="image/*" onChange={(e) => uploadCasePhoto(e.target.files)} />
                </label>
                <input
                  className="input"
                  style={{ gridColumn: '1 / -1' }}
                  value={editForm.photo_url}
                  onChange={(e) => setEditForm((s) => ({ ...s, photo_url: e.target.value }))}
                  placeholder="照片 URL"
                />
                {editForm.photo_url ? (
                  <div style={{ gridColumn: '1 / -1' }}>
                    <Image
                      src={editForm.photo_url}
                      alt="案件照片"
                      width={220}
                      height={160}
                      unoptimized
                      style={{ borderRadius: 8, border: '1px solid #e5e7eb', objectFit: 'cover', width: 220, height: 160 }}
                    />
                  </div>
                ) : null}
                <div className="row" style={{ gridColumn: '1 / -1' }}>
                  <button className="btn primary" type="submit" disabled={saving || photoUploading}>
                    {photoUploading ? '照片上传中...' : saving ? '保存中...' : '保存案件'}
                  </button>
                </div>
              </form>
            ) : null}
          </div>

          <div className="section-card">
            <h3 className="card-title">新增线索</h3>
            <div className="hint" style={{ marginTop: 4 }}>
              补充新的目击信息或排查结果，支持地点摘要与详细描述。
            </div>
            <form className="grid" style={{ marginTop: 10 }} onSubmit={submitTrack}>
              <input className="input" placeholder="地点摘要（可选）" value={trackLocation} onChange={(e) => setTrackLocation(e.target.value)} />
              <textarea className="textarea" placeholder="线索描述（必填）" value={trackContent} onChange={(e) => setTrackContent(e.target.value)} />
              <div className="row">
                <button className="btn primary" type="submit">
                  提交线索
                </button>
              </div>
            </form>
          </div>

          <details className="section-toggle" open>
            <summary>关联任务</summary>
            <div className="section-toggle-body">
            <div className="hint" style={{ marginTop: 2 }}>
              可从案件直接创建任务并分配执行，形成案件处理闭环。
            </div>
            <form className="grid cols-3" style={{ margin: '10px 0' }} onSubmit={createRelatedTask}>
              <input className="input" value={taskTitle} onChange={(e) => setTaskTitle(e.target.value)} placeholder="任务标题" />
              <select className="select" value={taskPriority} onChange={(e) => setTaskPriority(e.target.value)}>
                <option value="low">low</option>
                <option value="normal">normal</option>
                <option value="medium">medium</option>
                <option value="high">high</option>
                <option value="urgent">urgent</option>
              </select>
              <button className="btn primary" type="submit" disabled={taskCreating}>
                {taskCreating ? '创建中...' : '创建关联任务'}
              </button>
              <textarea
                className="textarea"
                style={{ gridColumn: '1 / -1' }}
                value={taskDesc}
                onChange={(e) => setTaskDesc(e.target.value)}
                placeholder="任务说明（可选）"
              />
            </form>
            {tasks.length === 0 ? (
              <div style={{ marginTop: 10, color: '#6b7280' }}>暂无关联任务</div>
            ) : (
              <div className="table-wrap" style={{ marginTop: 10 }}>
                <table className="table">
                  <thead>
                    <tr>
                      <th>标题</th>
                      <th>状态</th>
                      <th>优先级</th>
                      <th>执行人</th>
                      <th>操作</th>
                    </tr>
                  </thead>
                  <tbody>
                    {tasks.map((t) => (
                      <tr key={t.id}>
                        <td>{t.title}</td>
                        <td>
                          <StatusTag status={t.status || '-'} />
                        </td>
                        <td>{t.priority || '-'}</td>
                        <td>{t.assignee?.nickname || '-'}</td>
                        <td>
                          <Link className="btn ghost" href={`/tasks/${t.id}`}>
                            查看
                          </Link>
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            )}
            </div>
          </details>

          <details className="section-toggle" open>
            <summary>线索记录</summary>
            <div className="section-toggle-body">
            {tracks.length === 0 ? (
              <div style={{ marginTop: 10, color: '#6b7280' }}>暂无线索</div>
            ) : (
              <div className="timeline" style={{ marginTop: 10 }}>
                {tracks.map((t) => (
                  <div className="timeline-item" key={t.id}>
                    <div className="timeline-head">
                      <b>{joinLocation(t) || t.location || '地点待补充'}</b>
                      <span className="timeline-time">{fmtTime(t.time || t.created_at)}</span>
                    </div>
                    <div className="timeline-body">{t.description || '-'}</div>
                  </div>
                ))}
              </div>
            )}
            </div>
          </details>
        </div>
      ) : null}
    </AppShell>
  )
}
