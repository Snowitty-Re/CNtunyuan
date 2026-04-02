'use client'

import Link from 'next/link'
import { FormEvent, useEffect, useMemo, useState } from 'react'
import { AppShell } from '@/components/layout/AppShell'
import { ModuleHeader } from '@/components/shared/ModuleHeader'
import { ConfirmButton } from '@/components/shared/ConfirmButton'
import { PageState } from '@/components/shared/PageState'
import { Pagination } from '@/components/shared/Pagination'
import { NoticeBar, type Notice } from '@/components/shared/NoticeBar'
import { StatusTag } from '@/components/shared/StatusTag'
import { Dialog } from '@/components/ui/Dialog'
import { useAuthGuard } from '@/hooks/useAuthGuard'
import { fmtTime, listFrom } from '@/lib/data'
import { taskService } from '@/services/tasks'
import { userService } from '@/services/users'
import type { Task, User } from '@/types/api'

export default function TasksPage() {
  const { ready } = useAuthGuard()
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [items, setItems] = useState<Task[]>([])
  const [total, setTotal] = useState(0)
  const [page, setPage] = useState(1)
  const [status, setStatus] = useState('')
  const [keyword, setKeyword] = useState('')
  const [priority, setPriority] = useState('')
  const [typeFilter, setTypeFilter] = useState('')
  const [assigneeId, setAssigneeId] = useState('')
  const [startTime, setStartTime] = useState('')
  const [endTime, setEndTime] = useState('')
  const [sortBy, setSortBy] = useState('deadline')
  const [sortOrder, setSortOrder] = useState<'asc' | 'desc'>('asc')
  const [newTitle, setNewTitle] = useState('')
  const [users, setUsers] = useState<User[]>([])
  const [assignMap, setAssignMap] = useState<Record<string, string>>({})
  const [batchAssigneeId, setBatchAssigneeId] = useState('')
  const [stats, setStats] = useState<Record<string, number>>({})
  const [selectedIds, setSelectedIds] = useState<string[]>([])
  const [createOpen, setCreateOpen] = useState(false)
  const [batchCancelOpen, setBatchCancelOpen] = useState(false)
  const [batchDeleteOpen, setBatchDeleteOpen] = useState(false)
  const [notice, setNotice] = useState<Notice | null>(null)
  const sortedItems = useMemo(() => {
    const list = [...items]
    const factor = sortOrder === 'asc' ? 1 : -1
    const rank: Record<string, number> = { low: 1, normal: 2, medium: 3, high: 4, urgent: 5 }
    list.sort((a, b) => {
      if (sortBy === 'progress') {
        return (((a.progress ?? 0) - (b.progress ?? 0)) * factor)
      }
      if (sortBy === 'priority') {
        return ((rank[a.priority || 'normal'] - rank[b.priority || 'normal']) * factor)
      }
      const aTime = new Date(a.deadline || a.created_at || 0).getTime()
      const bTime = new Date(b.deadline || b.created_at || 0).getTime()
      return (aTime - bTime) * factor
    })
    return list
  }, [items, sortBy, sortOrder])
  const allSelected = sortedItems.length > 0 && selectedIds.length === sortedItems.length

  async function load(
    nextPage = page,
    nextStatus = status,
    filters?: {
      keyword?: string
      priority?: string
      type?: string
      assigneeId?: string
      startTime?: string
      endTime?: string
    },
  ) {
    setLoading(true)
    setError('')
    try {
      const qKeyword = filters?.keyword ?? keyword
      const qPriority = filters?.priority ?? priority
      const qType = filters?.type ?? typeFilter
      const qAssignee = filters?.assigneeId ?? assigneeId
      const qStart = filters?.startTime ?? startTime
      const qEnd = filters?.endTime ?? endTime
      const data = await taskService.list({
        page: nextPage,
        page_size: 20,
        status: nextStatus || undefined,
        keyword: qKeyword || undefined,
        priority: qPriority || undefined,
        type: qType || undefined,
        assignee_id: qAssignee || undefined,
        start_time: qStart ? new Date(qStart).toISOString() : undefined,
        end_time: qEnd ? new Date(qEnd).toISOString() : undefined,
      })
      const normalized = listFrom<Task>(data)
      setItems(normalized.list)
      setTotal(normalized.total)
      setPage(nextPage)
      setSelectedIds([])
    } catch (err) {
      setError(err instanceof Error ? err.message : '加载失败')
    } finally {
      setLoading(false)
    }
  }

  async function quickCreate(e: FormEvent) {
    e.preventDefault()
    if (!newTitle.trim()) return
    try {
      await taskService.create({
        title: newTitle.trim(),
        type: 'general',
        priority: 'medium',
        status: 'pending',
      })
      setNewTitle('')
      load(1)
      loadStats()
      setCreateOpen(false)
      setNotice({ type: 'success', text: '任务创建成功' })
    } catch (err) {
      setNotice({ type: 'error', text: err instanceof Error ? err.message : '创建失败' })
    }
  }

  async function loadUsers() {
    try {
      const data = await userService.list({ page: 1, page_size: 200, status: 'active' })
      setUsers(listFrom<User>(data).list)
    } catch {
      setUsers([])
    }
  }

  function resetFilters() {
    setStatus('')
    setKeyword('')
    setPriority('')
    setTypeFilter('')
    setAssigneeId('')
    setStartTime('')
    setEndTime('')
    setSortBy('deadline')
    setSortOrder('asc')
    load(1, '', { keyword: '', priority: '', type: '', assigneeId: '', startTime: '', endTime: '' })
  }

  function toggleSelect(id: string) {
    setSelectedIds((prev) => (prev.includes(id) ? prev.filter((x) => x !== id) : [...prev, id]))
  }

  function toggleSelectAll(checked: boolean) {
    setSelectedIds(checked ? sortedItems.map((x) => x.id) : [])
  }

  function invertSelection() {
    const visibleIds = sortedItems.map((x) => x.id)
    setSelectedIds((prev) => visibleIds.filter((id) => !prev.includes(id)))
  }

  async function batchAssign() {
    if (selectedIds.length === 0) {
      setNotice({ type: 'info', text: '请先勾选任务' })
      return
    }
    const fallbackIds = Array.from(new Set(selectedIds.map((id) => assignMap[id]).filter(Boolean))) as string[]
    const assigneeId = batchAssigneeId || (fallbackIds.length === 1 ? fallbackIds[0] : '')
    if (!assigneeId) {
      setNotice({ type: 'info', text: '请选择批量执行人，或为所选任务设置同一个执行人' })
      return
    }
    try {
      await Promise.all(selectedIds.map((id) => taskService.assign(id, assigneeId)))
      load(page)
      loadStats()
      setBatchAssigneeId('')
      setNotice({ type: 'success', text: `已批量分配 ${selectedIds.length} 个任务` })
    } catch (err) {
      setNotice({ type: 'error', text: err instanceof Error ? err.message : '批量分配失败' })
    }
  }

  async function batchStart() {
    if (selectedIds.length === 0) return
    try {
      await Promise.all(selectedIds.map((id) => taskService.start(id)))
      load(page)
      loadStats()
      setNotice({ type: 'success', text: `已批量开始 ${selectedIds.length} 个任务` })
    } catch (err) {
      setNotice({ type: 'error', text: err instanceof Error ? err.message : '批量开始失败' })
    }
  }

  async function batchCancel() {
    if (selectedIds.length === 0) return
    try {
      await Promise.all(selectedIds.map((id) => taskService.cancel(id, '批量取消')))
      load(page)
      loadStats()
      setNotice({ type: 'success', text: `已批量取消 ${selectedIds.length} 个任务` })
    } catch (err) {
      setNotice({ type: 'error', text: err instanceof Error ? err.message : '批量取消失败' })
    }
  }

  async function batchComplete() {
    if (selectedIds.length === 0) return
    try {
      await Promise.all(selectedIds.map((id) => taskService.complete(id, { result: '批量完成', feedback: 'web批量操作' })))
      load(page)
      loadStats()
      setNotice({ type: 'success', text: `已批量完成 ${selectedIds.length} 个任务` })
    } catch (err) {
      setNotice({ type: 'error', text: err instanceof Error ? err.message : '批量完成失败' })
    }
  }

  async function batchDelete() {
    if (selectedIds.length === 0) return
    try {
      await Promise.all(selectedIds.map((id) => taskService.remove(id)))
      load(page)
      loadStats()
      setNotice({ type: 'success', text: `已批量删除 ${selectedIds.length} 个任务` })
    } catch (err) {
      setNotice({ type: 'error', text: err instanceof Error ? err.message : '批量删除失败' })
    }
  }

  function exportCsv() {
    const rows = selectedIds.length > 0 ? sortedItems.filter((x) => selectedIds.includes(x.id)) : sortedItems
    if (rows.length === 0) return
    const headers = ['id', 'title', 'type', 'priority', 'status', 'progress', 'assignee', 'missing_person', 'deadline', 'created_at']
    const lines = rows.map((row) =>
      [
        row.id,
        row.title || '',
        row.type || '',
        row.priority || '',
        row.status || '',
        row.progress ?? 0,
        row.assignee?.nickname || row.assignee?.phone || row.assignee_id || '',
        row.missing_person?.name || row.missing_person_id || '',
        row.deadline || '',
        row.created_at || '',
      ]
        .map((v) => `"${String(v).replace(/"/g, '""')}"`)
        .join(','),
    )
    const content = `\ufeff${[headers.join(','), ...lines].join('\n')}`
    const blob = new Blob([content], { type: 'text/csv;charset=utf-8;' })
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = `tasks-${Date.now()}.csv`
    a.click()
    URL.revokeObjectURL(url)
    setNotice({ type: 'success', text: `已导出 ${rows.length} 条任务` })
  }

  async function loadStats() {
    try {
      const data = await taskService.stats()
      setStats({
        total: Number(data.total || 0),
        pending: Number(data.pending || 0),
        assigned: Number(data.assigned || 0),
        processing: Number(data.processing || 0),
        completed: Number(data.completed || 0),
        cancelled: Number(data.cancelled || 0),
      })
    } catch {
      setStats({})
    }
  }

  useEffect(() => {
    if (ready) {
      load(1)
      loadUsers()
      loadStats()
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [ready])

  if (!ready) return null

  return (
    <AppShell>
      <ModuleHeader
        title="任务中心"
        desc="任务创建、分配、执行、跟进和闭环审批"
        right={
          <div className="row wrap">
            <button className="btn" type="button" onClick={() => setCreateOpen(true)}>
              快速新建
            </button>
            <Link className="btn primary" href="/tasks/create">
              新建任务
            </Link>
          </div>
        }
      />
      <NoticeBar notice={notice} onClose={() => setNotice(null)} />
      <div className="kpi-grid" style={{ marginBottom: 12 }}>
        <div className="kpi">
          <div className="label">总任务</div>
          <div className="value">{stats.total || 0}</div>
        </div>
        <div className="kpi">
          <div className="label">待分配</div>
          <div className="value">{stats.pending || 0}</div>
        </div>
        <div className="kpi">
          <div className="label">进行中</div>
          <div className="value">{stats.processing || 0}</div>
        </div>
        <div className="kpi">
          <div className="label">已完成</div>
          <div className="value">{stats.completed || 0}</div>
        </div>
      </div>
      <div className="grid cols-2">
        <form
          className="panel row wrap"
          onSubmit={(e) => {
            e.preventDefault()
            load(1)
          }}
        >
          <input
            className="input"
            style={{ minWidth: 220 }}
            placeholder="关键词（标题/描述）"
            value={keyword}
            onChange={(e) => setKeyword(e.target.value)}
          />
          <button
            className={`btn ${status === '' ? 'primary' : ''}`}
            type="button"
            onClick={() => {
              const next = ''
              setStatus(next)
              load(1, next)
            }}
          >
            全部
          </button>
          <button
            className={`btn ${status === 'pending' ? 'primary' : ''}`}
            type="button"
            onClick={() => {
              const next = 'pending'
              setStatus(next)
              load(1, next)
            }}
          >
            待分配
          </button>
          <button
            className={`btn ${status === 'processing' ? 'primary' : ''}`}
            type="button"
            onClick={() => {
              const next = 'processing'
              setStatus(next)
              load(1, next)
            }}
          >
            进行中
          </button>
          <button
            className={`btn ${status === 'completed' ? 'primary' : ''}`}
            type="button"
            onClick={() => {
              const next = 'completed'
              setStatus(next)
              load(1, next)
            }}
          >
            已完成
          </button>
          <select className="select" value={status} onChange={(e) => setStatus(e.target.value)}>
            <option value="">全部状态</option>
            <option value="pending">待分配</option>
            <option value="assigned">已分配</option>
            <option value="processing">进行中</option>
            <option value="completed">已完成</option>
            <option value="cancelled">已取消</option>
          </select>
          <select className="select" value={typeFilter} onChange={(e) => setTypeFilter(e.target.value)}>
            <option value="">全部类型</option>
            <option value="general">general</option>
            <option value="search">search</option>
            <option value="verify">verify</option>
            <option value="field">field</option>
          </select>
          <select className="select" value={priority} onChange={(e) => setPriority(e.target.value)}>
            <option value="">全部优先级</option>
            <option value="low">low</option>
            <option value="normal">normal</option>
            <option value="medium">medium</option>
            <option value="high">high</option>
            <option value="urgent">urgent</option>
          </select>
          <select className="select" value={assigneeId} onChange={(e) => setAssigneeId(e.target.value)}>
            <option value="">全部执行人</option>
            {users.map((u) => (
              <option key={u.id} value={u.id}>
                {u.nickname || u.phone || u.id}
              </option>
            ))}
          </select>
          <input className="input" type="datetime-local" value={startTime} onChange={(e) => setStartTime(e.target.value)} />
          <input className="input" type="datetime-local" value={endTime} onChange={(e) => setEndTime(e.target.value)} />
          <select className="select" value={sortBy} onChange={(e) => setSortBy(e.target.value)}>
            <option value="deadline">按截止时间</option>
            <option value="priority">按优先级</option>
            <option value="progress">按进度</option>
          </select>
          <select className="select" value={sortOrder} onChange={(e) => setSortOrder(e.target.value as 'asc' | 'desc')}>
            <option value="asc">升序</option>
            <option value="desc">降序</option>
          </select>
          <button className="btn" type="submit">
            筛选
          </button>
          <button className="btn ghost" type="button" onClick={resetFilters}>
            重置
          </button>
        </form>
        <div className="panel row wrap">
          <b>批量操作</b>
          <span>已选 {selectedIds.length} 项</span>
          <button className="btn ghost" type="button" onClick={() => toggleSelectAll(true)}>
            全选本页
          </button>
          <button className="btn ghost" type="button" onClick={() => toggleSelectAll(false)}>
            清空选择
          </button>
          <button className="btn ghost" type="button" onClick={invertSelection}>
            反选
          </button>
          <select className="select" style={{ minWidth: 180 }} value={batchAssigneeId} onChange={(e) => setBatchAssigneeId(e.target.value)}>
            <option value="">选择批量执行人</option>
            {users.map((u) => (
              <option key={u.id} value={u.id}>
                {u.nickname || u.phone || u.id}
              </option>
            ))}
          </select>
          <button className="btn" type="button" onClick={batchAssign}>
            批量分配
          </button>
          <button className="btn" type="button" onClick={batchStart}>
            批量开始
          </button>
          <button className="btn" type="button" onClick={batchComplete}>
            批量完成
          </button>
          <button className="btn danger" type="button" onClick={() => setBatchCancelOpen(true)} disabled={selectedIds.length === 0}>
            批量取消
          </button>
          <button className="btn danger" type="button" onClick={() => setBatchDeleteOpen(true)} disabled={selectedIds.length === 0}>
            批量删除
          </button>
          <button className="btn" type="button" onClick={exportCsv}>
            导出CSV
          </button>
        </div>
      </div>

      <PageState loading={loading} error={error} empty={!loading && !error && items.length === 0} onRetry={() => load(page)} />
      {!loading && !error && items.length > 0 ? (
        <>
          <div className="table-wrap">
            <table className="table">
              <thead>
                <tr>
                  <th>
                    <input type="checkbox" checked={allSelected} onChange={(e) => toggleSelectAll(e.target.checked)} />
                  </th>
                  <th>标题</th>
                  <th>状态</th>
                  <th>优先级</th>
                  <th>进度</th>
                  <th>截止时间</th>
                  <th>操作</th>
                </tr>
              </thead>
              <tbody>
                {sortedItems.map((row) => (
                  <tr key={row.id}>
                    <td>
                      <input type="checkbox" checked={selectedIds.includes(row.id)} onChange={() => toggleSelect(row.id)} />
                    </td>
                    <td>{row.title}</td>
                    <td>
                      <StatusTag status={row.status || '-'} />
                    </td>
                    <td>{row.priority || '-'}</td>
                    <td>{row.progress ?? 0}%</td>
                    <td>{fmtTime(row.deadline || row.created_at)}</td>
                    <td>
                      <div className="row wrap">
                        <Link className="btn ghost" href={`/tasks/${row.id}`}>
                          详情
                        </Link>
                        {row.status === 'assigned' ? (
                          <button className="btn" type="button" onClick={() => taskService.start(row.id).then(() => load(page))}>
                            开始
                          </button>
                        ) : null}
                        {(row.status === 'pending' || row.status === 'assigned') && users.length > 0 ? (
                          <>
                            <select
                              className="select"
                              style={{ minWidth: 150 }}
                              value={assignMap[row.id] || ''}
                              onChange={(e) =>
                                setAssignMap((prev) => ({
                                  ...prev,
                                  [row.id]: e.target.value,
                                }))
                              }
                            >
                              <option value="">选择执行人</option>
                              {users.map((u) => (
                                <option key={u.id} value={u.id}>
                                  {u.nickname || u.phone || u.id}
                                </option>
                              ))}
                            </select>
                            <button
                              className="btn"
                              type="button"
                              onClick={() => {
                                const assigneeId = assignMap[row.id]
                                if (!assigneeId) return
                                taskService.assign(row.id, assigneeId).then(() => {
                                  load(page)
                                  loadStats()
                                })
                              }}
                            >
                              分配
                            </button>
                          </>
                        ) : null}
                        {row.status === 'processing' ? (
                          <button
                            className="btn"
                            type="button"
                            onClick={() =>
                              taskService.complete(row.id, { result: 'web端提交完成', feedback: '已处理' }).then(() => {
                                load(page)
                                loadStats()
                              })
                            }
                          >
                            完成
                          </button>
                        ) : null}
                        <ConfirmButton
                          text="删除"
                          message="确认删除该任务？"
                          onConfirm={() => {
                            taskService.remove(row.id).then(() => load(page))
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
      <Dialog open={createOpen} title="快速创建任务" onClose={() => setCreateOpen(false)}>
        <form className="row wrap" onSubmit={quickCreate}>
          <input className="input" placeholder="任务标题（必填）" value={newTitle} onChange={(e) => setNewTitle(e.target.value)} />
          <button className="btn primary" type="submit">
            创建任务
          </button>
        </form>
      </Dialog>
      <Dialog open={batchCancelOpen} title="确认批量取消" onClose={() => setBatchCancelOpen(false)}>
        <div className="grid">
          <div>确认批量取消 {selectedIds.length} 个任务？</div>
          <div className="row">
            <button className="btn ghost" type="button" onClick={() => setBatchCancelOpen(false)}>
              取消
            </button>
            <button
              className="btn danger"
              type="button"
              onClick={() => {
                setBatchCancelOpen(false)
                batchCancel()
              }}
            >
              确认取消
            </button>
          </div>
        </div>
      </Dialog>
      <Dialog open={batchDeleteOpen} title="确认批量删除" onClose={() => setBatchDeleteOpen(false)}>
        <div className="grid">
          <div>确认批量删除 {selectedIds.length} 个任务？该操作不可恢复。</div>
          <div className="row">
            <button className="btn ghost" type="button" onClick={() => setBatchDeleteOpen(false)}>
              取消
            </button>
            <button
              className="btn danger"
              type="button"
              onClick={() => {
                setBatchDeleteOpen(false)
                batchDelete()
              }}
            >
              确认删除
            </button>
          </div>
        </div>
      </Dialog>
    </AppShell>
  )
}
