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
import { Dialog } from '@/components/ui/Dialog'
import { useAuthGuard } from '@/hooks/useAuthGuard'
import { fmtTime, listFrom } from '@/lib/data'
import { organizationService } from '@/services/organizations'
import { userService } from '@/services/users'
import type { Organization, User } from '@/types/api'

export default function UsersPage() {
  const { ready } = useAuthGuard()
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [items, setItems] = useState<User[]>([])
  const [total, setTotal] = useState(0)
  const [page, setPage] = useState(1)
  const [orgs, setOrgs] = useState<Organization[]>([])
  const [roleMap, setRoleMap] = useState<Record<string, string>>({})
  const [selectedIds, setSelectedIds] = useState<string[]>([])
  const [batchRole, setBatchRole] = useState('volunteer')
  const [keyword, setKeyword] = useState('')
  const [statusFilter, setStatusFilter] = useState('')
  const [roleFilter, setRoleFilter] = useState('')

  const [nickname, setNickname] = useState('')
  const [phone, setPhone] = useState('')
  const [password, setPassword] = useState('')
  const [role, setRole] = useState('volunteer')
  const [orgId, setOrgId] = useState('')
  const [editingUser, setEditingUser] = useState<User | null>(null)
  const [detailOpen, setDetailOpen] = useState(false)
  const [savingDetail, setSavingDetail] = useState(false)
  const [editNickname, setEditNickname] = useState('')
  const [editPhone, setEditPhone] = useState('')
  const [editEmail, setEditEmail] = useState('')
  const [editRole, setEditRole] = useState('volunteer')
  const [editStatus, setEditStatus] = useState('active')
  const [editOrgId, setEditOrgId] = useState('')
  const allSelected = items.length > 0 && selectedIds.length === items.length
  const [notice, setNotice] = useState<Notice | null>(null)

  async function load(nextPage = page) {
    setLoading(true)
    setError('')
    try {
      const data = await userService.list({
        page: nextPage,
        page_size: 20,
        keyword: keyword || undefined,
        status: statusFilter || undefined,
        role: roleFilter || undefined,
      })
      const normalized = listFrom<User>(data)
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

  async function loadOrgs() {
    try {
      const data = await organizationService.list({ page: 1, page_size: 200 })
      setOrgs(listFrom<Organization>(data).list)
    } catch {
      setOrgs([])
    }
  }

  async function quickCreate(e: FormEvent) {
    e.preventDefault()
    if (!nickname.trim() || !phone.trim() || !password.trim()) return
    try {
      await userService.create({
        nickname: nickname.trim(),
        phone: phone.trim(),
        password: password.trim(),
        role,
        org_id: orgId || null,
      })
      setNickname('')
      setPhone('')
      setPassword('')
      setRole('volunteer')
      setOrgId('')
      load(1)
      setNotice({ type: 'success', text: '用户创建成功' })
    } catch (err) {
      setNotice({ type: 'error', text: err instanceof Error ? err.message : '创建失败' })
    }
  }

  function toggleSelect(id: string) {
    setSelectedIds((prev) => (prev.includes(id) ? prev.filter((x) => x !== id) : [...prev, id]))
  }

  function toggleSelectAll(checked: boolean) {
    setSelectedIds(checked ? items.map((x) => x.id) : [])
  }

  function invertSelection() {
    const visibleIds = items.map((x) => x.id)
    setSelectedIds((prev) => visibleIds.filter((id) => !prev.includes(id)))
  }

  async function batchUpdateStatus(status: string) {
    if (selectedIds.length === 0) return
    try {
      await Promise.all(selectedIds.map((id) => userService.updateStatus(id, status)))
      load(page)
      setNotice({ type: 'success', text: `已更新 ${selectedIds.length} 位用户状态` })
    } catch (err) {
      setNotice({ type: 'error', text: err instanceof Error ? err.message : '批量操作失败' })
    }
  }

  async function batchUpdateRole() {
    if (selectedIds.length === 0) return
    try {
      await Promise.all(selectedIds.map((id) => userService.updateRole(id, batchRole)))
      load(page)
      setNotice({ type: 'success', text: `已更新 ${selectedIds.length} 位用户角色` })
    } catch (err) {
      setNotice({ type: 'error', text: err instanceof Error ? err.message : '批量角色更新失败' })
    }
  }

  function openDetail(user: User) {
    setEditingUser(user)
    setEditNickname(user.nickname || '')
    setEditPhone(user.phone || '')
    setEditEmail(user.email || '')
    setEditRole(user.role || 'volunteer')
    setEditStatus(user.status || 'active')
    setEditOrgId(user.org_id || '')
    setDetailOpen(true)
  }

  async function saveDetail(e: FormEvent) {
    e.preventDefault()
    if (!editingUser) return
    setSavingDetail(true)
    try {
      await userService.update(editingUser.id, {
        nickname: editNickname.trim(),
        phone: editPhone.trim(),
        email: editEmail.trim(),
        org_id: editOrgId || null,
      })
      if ((editingUser.role || '') !== editRole) {
        await userService.updateRole(editingUser.id, editRole)
      }
      if ((editingUser.status || '') !== editStatus) {
        await userService.updateStatus(editingUser.id, editStatus)
      }
      setDetailOpen(false)
      setEditingUser(null)
      load(page)
      setNotice({ type: 'success', text: '用户信息已更新' })
    } catch (err) {
      setNotice({ type: 'error', text: err instanceof Error ? err.message : '保存失败' })
    } finally {
      setSavingDetail(false)
    }
  }

  function exportCsv() {
    const rows = selectedIds.length > 0 ? items.filter((x) => selectedIds.includes(x.id)) : items
    if (rows.length === 0) return
    const headers = ['id', 'nickname', 'phone', 'email', 'role', 'status', 'organization', 'created_at']
    const lines = rows.map((row) =>
      [
        row.id,
        row.nickname || '',
        row.phone || '',
        row.email || '',
        row.role || '',
        row.status || '',
        row.organization?.name || '',
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
    a.download = `users-${Date.now()}.csv`
    a.click()
    URL.revokeObjectURL(url)
    setNotice({ type: 'success', text: `已导出 ${rows.length} 位用户` })
  }

  useEffect(() => {
    if (ready) {
      load(1)
      loadOrgs()
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [ready])

  if (!ready) return null

  return (
    <AppShell>
      <ModuleHeader title="人员管理" desc="用户账号、角色、组织归属与状态管理" />
      <NoticeBar notice={notice} onClose={() => setNotice(null)} />
      <form
        className="panel row wrap"
        onSubmit={(e) => {
          e.preventDefault()
          load(1)
        }}
      >
        <input className="input" placeholder="关键词（昵称/手机号）" value={keyword} onChange={(e) => setKeyword(e.target.value)} />
        <select className="select" value={roleFilter} onChange={(e) => setRoleFilter(e.target.value)}>
          <option value="">全部角色</option>
          <option value="volunteer">volunteer</option>
          <option value="manager">manager</option>
          <option value="admin">admin</option>
          <option value="super_admin">super_admin</option>
        </select>
        <select className="select" value={statusFilter} onChange={(e) => setStatusFilter(e.target.value)}>
          <option value="">全部状态</option>
          <option value="active">active</option>
          <option value="inactive">inactive</option>
          <option value="pending">pending</option>
        </select>
        <button className="btn" type="submit">
          筛选
        </button>
      </form>
      <form className="panel grid cols-3" onSubmit={quickCreate}>
        <input className="input" placeholder="昵称" value={nickname} onChange={(e) => setNickname(e.target.value)} />
        <input className="input" placeholder="手机号" value={phone} onChange={(e) => setPhone(e.target.value)} />
        <input className="input" placeholder="初始密码" value={password} onChange={(e) => setPassword(e.target.value)} />
        <select className="select" value={role} onChange={(e) => setRole(e.target.value)}>
          <option value="volunteer">volunteer</option>
          <option value="manager">manager</option>
          <option value="admin">admin</option>
          <option value="super_admin">super_admin</option>
        </select>
        <select className="select" value={orgId} onChange={(e) => setOrgId(e.target.value)}>
          <option value="">不绑定组织</option>
          {orgs.map((o) => (
            <option key={o.id} value={o.id}>
              {o.name}
            </option>
          ))}
        </select>
        <button className="btn primary" type="submit">
          创建用户
        </button>
      </form>
      <div className="panel row wrap">
        <b>批量操作</b>
        <span>已选 {selectedIds.length} 人</span>
        <button className="btn ghost" type="button" onClick={() => toggleSelectAll(true)}>
          全选本页
        </button>
        <button className="btn ghost" type="button" onClick={() => toggleSelectAll(false)}>
          清空选择
        </button>
        <button className="btn ghost" type="button" onClick={invertSelection}>
          反选
        </button>
        <button className="btn" type="button" onClick={() => batchUpdateStatus('active')}>
          批量激活
        </button>
        <button className="btn danger" type="button" onClick={() => batchUpdateStatus('inactive')}>
          批量禁用
        </button>
        <select className="select" value={batchRole} onChange={(e) => setBatchRole(e.target.value)}>
          <option value="volunteer">volunteer</option>
          <option value="manager">manager</option>
          <option value="admin">admin</option>
          <option value="super_admin">super_admin</option>
        </select>
        <button className="btn" type="button" onClick={batchUpdateRole}>
          批量更新角色
        </button>
        <button className="btn" type="button" onClick={exportCsv}>
          导出CSV
        </button>
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
                  <th>昵称</th>
                  <th>手机号</th>
                  <th>角色</th>
                  <th>状态</th>
                  <th>组织</th>
                  <th>操作</th>
                </tr>
              </thead>
              <tbody>
                {items.map((row) => (
                  <tr key={row.id}>
                    <td>
                      <input type="checkbox" checked={selectedIds.includes(row.id)} onChange={() => toggleSelect(row.id)} />
                    </td>
                    <td>{row.nickname || '-'}</td>
                    <td>{row.phone || '-'}</td>
                    <td>
                      <div className="row wrap">
                        <select
                          className="select"
                          style={{ minWidth: 130 }}
                          value={roleMap[row.id] || row.role}
                          onChange={(e) =>
                            setRoleMap((prev) => ({
                              ...prev,
                              [row.id]: e.target.value,
                            }))
                          }
                        >
                          <option value="volunteer">volunteer</option>
                          <option value="manager">manager</option>
                          <option value="admin">admin</option>
                          <option value="super_admin">super_admin</option>
                        </select>
                        <button
                          className="btn"
                          type="button"
                          onClick={() => userService.updateRole(row.id, roleMap[row.id] || row.role).then(() => load(page))}
                        >
                          保存角色
                        </button>
                      </div>
                    </td>
                    <td>
                      <StatusTag status={row.status || '-'} />
                    </td>
                    <td>{row.organization?.name || '-'}</td>
                    <td>
                      <div className="row wrap">
                        <button className="btn" type="button" onClick={() => userService.updateStatus(row.id, 'active').then(() => load(page))}>
                          激活
                        </button>
                        <button className="btn danger" type="button" onClick={() => userService.updateStatus(row.id, 'inactive').then(() => load(page))}>
                          禁用
                        </button>
                        <ConfirmButton
                          text="删除"
                          message={`确认删除用户「${row.nickname || row.phone}」？`}
                          onConfirm={() => {
                            userService.remove(row.id).then(() => load(page))
                          }}
                          className="btn danger"
                        />
                        <button className="btn ghost" type="button" onClick={() => openDetail(row)}>
                          详情/编辑
                        </button>
                        <Link className="btn ghost" href={`/audit?user_id=${row.id}`}>
                          审计记录
                        </Link>
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
      <Dialog
        open={detailOpen}
        title={editingUser ? `用户详情：${editingUser.nickname || editingUser.phone || editingUser.id}` : '用户详情'}
        onClose={() => {
          if (!savingDetail) setDetailOpen(false)
        }}
      >
        {editingUser ? (
          <form className="grid cols-2" onSubmit={saveDetail}>
            <div>用户ID：{editingUser.id}</div>
            <div>创建时间：{fmtTime(editingUser.created_at)}</div>
            <label>
              <div>昵称</div>
              <input className="input" value={editNickname} onChange={(e) => setEditNickname(e.target.value)} />
            </label>
            <label>
              <div>手机号</div>
              <input className="input" value={editPhone} onChange={(e) => setEditPhone(e.target.value)} />
            </label>
            <label>
              <div>邮箱</div>
              <input className="input" value={editEmail} onChange={(e) => setEditEmail(e.target.value)} />
            </label>
            <label>
              <div>角色</div>
              <select className="select" value={editRole} onChange={(e) => setEditRole(e.target.value)}>
                <option value="volunteer">volunteer</option>
                <option value="manager">manager</option>
                <option value="admin">admin</option>
                <option value="super_admin">super_admin</option>
              </select>
            </label>
            <label>
              <div>状态</div>
              <select className="select" value={editStatus} onChange={(e) => setEditStatus(e.target.value)}>
                <option value="active">active</option>
                <option value="inactive">inactive</option>
                <option value="pending">pending</option>
              </select>
            </label>
            <label>
              <div>组织</div>
              <select className="select" value={editOrgId} onChange={(e) => setEditOrgId(e.target.value)}>
                <option value="">不绑定组织</option>
                {orgs.map((o) => (
                  <option key={o.id} value={o.id}>
                    {o.name}
                  </option>
                ))}
              </select>
            </label>
            <div className="row" style={{ gridColumn: '1 / -1' }}>
              <button className="btn primary" type="submit" disabled={savingDetail}>
                {savingDetail ? '保存中...' : '保存变更'}
              </button>
            </div>
          </form>
        ) : null}
      </Dialog>
    </AppShell>
  )
}
