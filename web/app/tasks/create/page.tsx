'use client'

import { FormEvent, useEffect, useState } from 'react'
import { useRouter } from 'next/navigation'
import { AppShell } from '@/components/layout/AppShell'
import { ModuleHeader } from '@/components/shared/ModuleHeader'
import { NoticeBar, type Notice } from '@/components/shared/NoticeBar'
import { useAuthGuard } from '@/hooks/useAuthGuard'
import { listFrom } from '@/lib/data'
import { missingPersonService } from '@/services/missingPersons'
import { taskService } from '@/services/tasks'
import { userService } from '@/services/users'
import type { MissingPerson, User } from '@/types/api'

export default function TaskCreatePage() {
  const { ready } = useAuthGuard()
  const router = useRouter()
  const [users, setUsers] = useState<User[]>([])
  const [cases, setCases] = useState<MissingPerson[]>([])
  const [submitting, setSubmitting] = useState(false)
  const [notice, setNotice] = useState<Notice | null>(null)

  const [title, setTitle] = useState('')
  const [description, setDescription] = useState('')
  const [type, setType] = useState('general')
  const [priority, setPriority] = useState('medium')
  const [deadline, setDeadline] = useState('')
  const [assigneeId, setAssigneeId] = useState('')
  const [missingPersonId, setMissingPersonId] = useState('')
  const [location, setLocation] = useState('')
  const [province, setProvince] = useState('')
  const [city, setCity] = useState('')
  const [district, setDistrict] = useState('')
  const [address, setAddress] = useState('')
  const [lat, setLat] = useState('')
  const [lng, setLng] = useState('')

  async function loadOptions() {
    try {
      const [userData, caseData] = await Promise.all([
        userService.list({ page: 1, page_size: 300, status: 'active' }),
        missingPersonService.list({ page: 1, page_size: 200 }),
      ])
      setUsers(listFrom<User>(userData).list)
      setCases(listFrom<MissingPerson>(caseData).list)
    } catch {
      setUsers([])
      setCases([])
    }
  }

  async function submit(e: FormEvent) {
    e.preventDefault()
    if (!title.trim()) {
      setNotice({ type: 'error', text: '请填写任务标题' })
      return
    }
    if (!type) {
      setNotice({ type: 'error', text: '请选择任务类型' })
      return
    }
    setSubmitting(true)
    setNotice(null)
    try {
      await taskService.create({
        title: title.trim(),
        description: description.trim(),
        type,
        priority,
        assignee_id: assigneeId || null,
        missing_person_id: missingPersonId || null,
        deadline: deadline ? new Date(deadline).toISOString() : null,
        location: location.trim(),
        province: province.trim(),
        city: city.trim(),
        district: district.trim(),
        address: address.trim(),
        lat: Number(lat) || 0,
        lng: Number(lng) || 0,
      })
      router.push('/tasks')
    } catch (err) {
      setNotice({ type: 'error', text: err instanceof Error ? err.message : '创建失败' })
    } finally {
      setSubmitting(false)
    }
  }

  useEffect(() => {
    if (ready) loadOptions()
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [ready])

  if (!ready) return null

  return (
    <AppShell>
      <ModuleHeader title="创建任务" desc="创建可执行、可分配、可追踪的任务单" />
      <NoticeBar notice={notice} onClose={() => setNotice(null)} />
      <form className="grid" onSubmit={submit}>
        <div className="form-section">
          <h3 className="form-section-title">任务基础</h3>
          <div className="grid cols-3">
            <label>
              <span className="field-label">任务标题<span className="required">*</span></span>
              <input className="input" value={title} onChange={(e) => setTitle(e.target.value)} placeholder="请输入任务标题" />
            </label>
            <label>
              <span className="field-label">任务类型<span className="required">*</span></span>
              <select className="select" value={type} onChange={(e) => setType(e.target.value)}>
                <option value="general">general</option>
                <option value="search">search</option>
                <option value="verify">verify</option>
                <option value="field">field</option>
              </select>
            </label>
            <label>
              <span className="field-label">优先级</span>
              <select className="select" value={priority} onChange={(e) => setPriority(e.target.value)}>
                <option value="low">low</option>
                <option value="normal">normal</option>
                <option value="medium">medium</option>
                <option value="high">high</option>
                <option value="urgent">urgent</option>
              </select>
            </label>
            <label style={{ gridColumn: '1 / -1' }}>
              <span className="field-label">任务描述</span>
              <textarea className="textarea" value={description} onChange={(e) => setDescription(e.target.value)} placeholder="补充执行说明、范围与要求" />
            </label>
          </div>
        </div>

        <div className="form-section">
          <h3 className="form-section-title">分配与关联</h3>
          <div className="grid cols-3">
            <label>
              <span className="field-label">截止时间</span>
              <input className="input" type="datetime-local" value={deadline} onChange={(e) => setDeadline(e.target.value)} />
            </label>
            <label>
              <span className="field-label">执行人</span>
              <select className="select" value={assigneeId} onChange={(e) => setAssigneeId(e.target.value)}>
                <option value="">暂不分配</option>
                {users.map((u) => (
                  <option key={u.id} value={u.id}>
                    {u.nickname || u.phone || u.id}
                  </option>
                ))}
              </select>
            </label>
            <label>
              <span className="field-label">关联案件</span>
              <select className="select" value={missingPersonId} onChange={(e) => setMissingPersonId(e.target.value)}>
                <option value="">不关联案件</option>
                {cases.map((c) => (
                  <option key={c.id} value={c.id}>
                    {c.name} {c.status ? `(${c.status})` : ''}
                  </option>
                ))}
              </select>
            </label>
          </div>
        </div>

        <div className="form-section">
          <h3 className="form-section-title">执行地点（可选）</h3>
          <div className="grid cols-3">
            <label>
              <span className="field-label">地点摘要</span>
              <input className="input" value={location} onChange={(e) => setLocation(e.target.value)} placeholder="例如：越秀区北京街道" />
            </label>
            <label>
              <span className="field-label">省</span>
              <input className="input" value={province} onChange={(e) => setProvince(e.target.value)} placeholder="广东省" />
            </label>
            <label>
              <span className="field-label">市</span>
              <input className="input" value={city} onChange={(e) => setCity(e.target.value)} placeholder="广州市" />
            </label>
            <label>
              <span className="field-label">区/县</span>
              <input className="input" value={district} onChange={(e) => setDistrict(e.target.value)} placeholder="越秀区" />
            </label>
            <label>
              <span className="field-label">详细地址</span>
              <input className="input" value={address} onChange={(e) => setAddress(e.target.value)} placeholder="可选" />
            </label>
            <label>
              <span className="field-label">纬度 / 经度</span>
              <div className="row">
                <input className="input" value={lat} onChange={(e) => setLat(e.target.value.replace(/[^\d.-]/g, ''))} placeholder="lat" />
                <input className="input" value={lng} onChange={(e) => setLng(e.target.value.replace(/[^\d.-]/g, ''))} placeholder="lng" />
              </div>
            </label>
          </div>
        </div>

        <div className="panel row wrap" style={{ marginTop: 0 }}>
          <button className="btn primary" type="submit" disabled={submitting}>
            {submitting ? '创建中...' : '创建任务'}
          </button>
          <button type="button" className="btn" onClick={() => router.back()}>
            返回
          </button>
          <span className="hint">先填写“任务基础”，其余可按需要补充。</span>
        </div>
      </form>
    </AppShell>
  )
}
