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
import { SafeImage } from '@/components/shared/SafeImage'
import { Dialog } from '@/components/ui/Dialog'
import { useAuthGuard } from '@/hooks/useAuthGuard'
import { ACTIONS, hasPermission } from '@/lib/rbac'
import { fmtTime, listFrom } from '@/lib/data'
import { organizationService } from '@/services/organizations'
import { userService } from '@/services/users'
import type { Organization, User } from '@/types/api'

const USER_FILTER_KEY = 'web_users_filters_v1'
const USER_COL_KEY = 'web_users_columns_v1'

export default function UsersPage() {
  const { ready, user } = useAuthGuard()
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
  const [booted, setBooted] = useState(false)
  const [columnVisible, setColumnVisible] = useState<Record<string, boolean>>({
    avatar: true,
    nickname: true,
    phone: true,
    role: true,
    status: true,
    wxBound: true,
    organization: true,
    actions: true,
  })
  const allSelected = items.length > 0 && selectedIds.length === items.length
  const [notice, setNotice] = useState<Notice | null>(null)

  function normalizeUser(row: User): User {
    const anyRow = row as any
    const orgId = row.org_id || anyRow.orgId || row.organization?.id || ''
    const orgName = row.org_name || anyRow.orgName || row.organization?.name || ''
    const organization = row.organization || (orgId ? { id: orgId, name: orgName } : null)
    return {
      ...row,
      org_id: orgId,
      org_name: orgName,
      organization,
    }
  }

  function resolveOrgIdForEdit(row: User): string {
    const orgId = row.org_id || row.organization?.id || ''
    if (orgId) return orgId
    const orgName = row.org_name || row.organization?.name || ''
    if (!orgName) return ''
    const found = orgs.find((o) => o.name === orgName)
    return found?.id || ''
  }

  async function load(
    nextPage = page,
    filters?: {
      keyword?: string
      status?: string
      role?: string
    },
  ) {
    setLoading(true)
    setError('')
    try {
      const qKeyword = filters?.keyword ?? keyword
      const qStatus = filters?.status ?? statusFilter
      const qRole = filters?.role ?? roleFilter
      const data = await userService.list({
        page: nextPage,
        page_size: 20,
        keyword: qKeyword || undefined,
        status: qStatus || undefined,
        role: qRole || undefined,
      })
      const normalized = listFrom<User>(data)
      setItems(normalized.list.map(normalizeUser))
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
      const pageSize = 100
      let current = 1
      let keepGoing = true
      const merged: Organization[] = []

      while (keepGoing && current <= 20) {
        const data = await organizationService.list({ page: current, page_size: pageSize })
        const chunk = listFrom<Organization>(data).list
        merged.push(...chunk)
        keepGoing = chunk.length === pageSize
        current += 1
      }

      setOrgs(merged)
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
    setEditOrgId(resolveOrgIdForEdit(user))
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

  function resetFilters() {
    setKeyword('')
    setStatusFilter('')
    setRoleFilter('')
    load(1, { keyword: '', status: '', role: '' })
  }

  useEffect(() => {
    if (ready && !booted) {
      if (typeof window !== 'undefined') {
        try {
          const savedFilter = JSON.parse(localStorage.getItem(USER_FILTER_KEY) || '{}')
          const savedCols = JSON.parse(localStorage.getItem(USER_COL_KEY) || '{}')
          if (savedFilter && typeof savedFilter === 'object') {
            setKeyword(savedFilter.keyword || '')
            setStatusFilter(savedFilter.statusFilter || '')
            setRoleFilter(savedFilter.roleFilter || '')
            load(1, { keyword: savedFilter.keyword || '', status: savedFilter.statusFilter || '', role: savedFilter.roleFilter || '' })
          } else {
            load(1)
          }
          if (savedCols && typeof savedCols === 'object') {
            setColumnVisible((prev) => ({ ...prev, ...savedCols }))
          }
        } catch {
          load(1)
        }
      } else {
        load(1)
      }
      loadOrgs()
      setBooted(true)
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [ready, booted])

  useEffect(() => {
    if (!ready || !booted || typeof window === 'undefined') return
    localStorage.setItem(USER_FILTER_KEY, JSON.stringify({ keyword, statusFilter, roleFilter }))
  }, [ready, booted, keyword, statusFilter, roleFilter])

  useEffect(() => {
    if (!ready || !booted || typeof window === 'undefined') return
    localStorage.setItem(USER_COL_KEY, JSON.stringify(columnVisible))
  }, [ready, booted, columnVisible])

  useEffect(() => {
    if (!detailOpen || !editingUser || editOrgId || orgs.length === 0) return
    const resolved = resolveOrgIdForEdit(editingUser)
    if (resolved) setEditOrgId(resolved)
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [detailOpen, editingUser, editOrgId, orgs])

  if (!ready) return null
  if (!hasPermission(user, ACTIONS.USER_VIEW)) {
    return (
      <AppShell>
        <ModuleHeader title="人员管理" desc="用户账号、角色、组织归属与状态管理" />
        <PageState error="当前账号无权限访问该页面（需要 user:view 权限）" />
      </AppShell>
    )
  }

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
          <option value="active">正常 (active)</option>
          <option value="inactive">待审批/禁用 (inactive)</option>
          <option value="banned">封禁 (banned)</option>
        </select>
        <button className="btn" type="submit">
          筛选
        </button>
        <button className="btn" type="button" onClick={() => { setStatusFilter('inactive'); load(1, { status: 'inactive' }) }}>
          待审批
        </button>
        <button className="btn ghost" type="button" onClick={resetFilters}>
          重置
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
      <div className="panel row wrap">
        <b>列显示</b>
        <label><input type="checkbox" checked={columnVisible.avatar} onChange={(e) => setColumnVisible((v) => ({ ...v, avatar: e.target.checked }))} /> 头像</label>
        <label><input type="checkbox" checked={columnVisible.nickname} onChange={(e) => setColumnVisible((v) => ({ ...v, nickname: e.target.checked }))} /> 昵称</label>
        <label><input type="checkbox" checked={columnVisible.phone} onChange={(e) => setColumnVisible((v) => ({ ...v, phone: e.target.checked }))} /> 手机号</label>
        <label><input type="checkbox" checked={columnVisible.role} onChange={(e) => setColumnVisible((v) => ({ ...v, role: e.target.checked }))} /> 角色</label>
        <label><input type="checkbox" checked={columnVisible.status} onChange={(e) => setColumnVisible((v) => ({ ...v, status: e.target.checked }))} /> 状态</label>
        <label><input type="checkbox" checked={columnVisible.wxBound} onChange={(e) => setColumnVisible((v) => ({ ...v, wxBound: e.target.checked }))} /> 微信绑定</label>
        <label><input type="checkbox" checked={columnVisible.organization} onChange={(e) => setColumnVisible((v) => ({ ...v, organization: e.target.checked }))} /> 组织</label>
        <label><input type="checkbox" checked={columnVisible.actions} onChange={(e) => setColumnVisible((v) => ({ ...v, actions: e.target.checked }))} /> 操作</label>
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
                  {columnVisible.avatar ? <th>头像</th> : null}
                  {columnVisible.nickname ? <th>昵称</th> : null}
                  {columnVisible.phone ? <th>手机号</th> : null}
                  {columnVisible.role ? <th>角色</th> : null}
                  {columnVisible.status ? <th>状态</th> : null}
                  {columnVisible.wxBound ? <th>微信绑定</th> : null}
                  {columnVisible.organization ? <th>组织</th> : null}
                  {columnVisible.actions ? <th>操作</th> : null}
                </tr>
              </thead>
              <tbody>
                {items.map((row) => (
                  <tr key={row.id}>
                    <td>
                      <input type="checkbox" checked={selectedIds.includes(row.id)} onChange={() => toggleSelect(row.id)} />
                    </td>
                    {columnVisible.avatar ? <td>
                      <SafeImage
                        className="cell-avatar"
                        src={row.avatar || '/default-avatar.svg'}
                        alt={row.nickname || row.phone || 'avatar'}
                        width={34}
                        height={34}
                      />
                    </td> : null}
                    {columnVisible.nickname ? <td>{row.nickname || '-'}</td> : null}
                    {columnVisible.phone ? <td>{row.phone || '-'}</td> : null}
                    {columnVisible.role ? <td>
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
                    </td> : null}
                    {columnVisible.status ? <td>
                      <StatusTag status={row.status || '-'} />
                    </td> : null}
                    {columnVisible.wxBound ? <td>{row.wx_bound ? '已绑定' : '未绑定'}</td> : null}
                    {columnVisible.organization ? <td>{row.organization?.name || row.org_name || '-'}</td> : null}
                    {columnVisible.actions ? <td>
                      <div className="row wrap">
                        <button className="btn" type="button" onClick={() => userService.updateStatus(row.id, 'active').then(() => load(page))}>
                          {row.status === 'inactive' ? '审批通过' : '激活'}
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
                    </td> : null}
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
                <option value="active">正常 (active)</option>
                <option value="inactive">待审批/禁用 (inactive)</option>
                <option value="banned">封禁 (banned)</option>
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
