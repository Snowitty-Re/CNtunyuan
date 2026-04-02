'use client'

import { FormEvent, useEffect, useState } from 'react'
import { useRouter } from 'next/navigation'
import { AppShell } from '@/components/layout/AppShell'
import { ModuleHeader } from '@/components/shared/ModuleHeader'
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
  const [error, setError] = useState('')

  const [title, setTitle] = useState('')
  const [description, setDescription] = useState('')
  const [type, setType] = useState('general')
  const [priority, setPriority] = useState('medium')
  const [deadline, setDeadline] = useState('')
  const [assigneeId, setAssigneeId] = useState('')
  const [missingPersonId, setMissingPersonId] = useState('')

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
    if (!title.trim()) return
    setSubmitting(true)
    setError('')
    try {
      await taskService.create({
        title: title.trim(),
        description: description.trim(),
        type,
        priority,
        assignee_id: assigneeId || null,
        missing_person_id: missingPersonId || null,
        deadline: deadline ? new Date(deadline).toISOString() : null,
      })
      router.push('/tasks')
    } catch (err) {
      setError(err instanceof Error ? err.message : '创建失败')
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
      <form className="section-card grid cols-2" onSubmit={submit}>
        <label>
          <div>任务标题 *</div>
          <input className="input" value={title} onChange={(e) => setTitle(e.target.value)} placeholder="请输入任务标题" />
        </label>
        <label>
          <div>任务类型</div>
          <select className="select" value={type} onChange={(e) => setType(e.target.value)}>
            <option value="general">general</option>
            <option value="search">search</option>
            <option value="verify">verify</option>
            <option value="field">field</option>
          </select>
        </label>
        <label>
          <div>优先级</div>
          <select className="select" value={priority} onChange={(e) => setPriority(e.target.value)}>
            <option value="low">low</option>
            <option value="normal">normal</option>
            <option value="medium">medium</option>
            <option value="high">high</option>
            <option value="urgent">urgent</option>
          </select>
        </label>
        <label>
          <div>截止时间</div>
          <input className="input" type="datetime-local" value={deadline} onChange={(e) => setDeadline(e.target.value)} />
        </label>
        <label>
          <div>执行人</div>
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
          <div>关联案件</div>
          <select className="select" value={missingPersonId} onChange={(e) => setMissingPersonId(e.target.value)}>
            <option value="">不关联案件</option>
            {cases.map((c) => (
              <option key={c.id} value={c.id}>
                {c.name} {c.status ? `(${c.status})` : ''}
              </option>
            ))}
          </select>
        </label>
        <label style={{ gridColumn: '1 / -1' }}>
          <div>任务描述</div>
          <textarea className="textarea" value={description} onChange={(e) => setDescription(e.target.value)} placeholder="补充执行说明、范围与要求" />
        </label>
        <div className="row" style={{ gridColumn: '1 / -1' }}>
          <button className="btn primary" type="submit" disabled={submitting}>
            {submitting ? '创建中...' : '创建任务'}
          </button>
          <button type="button" className="btn" onClick={() => router.back()}>
            返回
          </button>
          {error ? <span className="alert">{error}</span> : null}
        </div>
      </form>
    </AppShell>
  )
}
