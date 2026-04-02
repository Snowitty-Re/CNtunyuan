'use client'

import { FormEvent, useEffect, useState } from 'react'
import { AppShell } from '@/components/layout/AppShell'
import { ModuleHeader } from '@/components/shared/ModuleHeader'
import { ConfirmButton } from '@/components/shared/ConfirmButton'
import { NoticeBar, type Notice } from '@/components/shared/NoticeBar'
import { PageState } from '@/components/shared/PageState'
import { Pagination } from '@/components/shared/Pagination'
import { Dialog } from '@/components/ui/Dialog'
import { useAuthGuard } from '@/hooks/useAuthGuard'
import { fmtTime, listFrom } from '@/lib/data'
import { organizationService } from '@/services/organizations'
import type { Organization } from '@/types/api'

export default function OrganizationsPage() {
  const { ready } = useAuthGuard()
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [items, setItems] = useState<Organization[]>([])
  const [tree, setTree] = useState<Organization[]>([])
  const [total, setTotal] = useState(0)
  const [page, setPage] = useState(1)
  const [keyword, setKeyword] = useState('')
  const [typeFilter, setTypeFilter] = useState('')
  const [moveTarget, setMoveTarget] = useState<Record<string, string>>({})

  const [name, setName] = useState('')
  const [code, setCode] = useState('')
  const [type, setType] = useState('team')
  const [editingOrg, setEditingOrg] = useState<Organization | null>(null)
  const [detailOpen, setDetailOpen] = useState(false)
  const [savingDetail, setSavingDetail] = useState(false)
  const [editName, setEditName] = useState('')
  const [editCode, setEditCode] = useState('')
  const [editType, setEditType] = useState('team')
  const [editParentId, setEditParentId] = useState('')
  const [editDesc, setEditDesc] = useState('')
  const [editAddress, setEditAddress] = useState('')
  const [editContactName, setEditContactName] = useState('')
  const [editContactPhone, setEditContactPhone] = useState('')
  const [editSortOrder, setEditSortOrder] = useState('0')
  const [notice, setNotice] = useState<Notice | null>(null)

  async function load(
    nextPage = page,
    filters?: {
      keyword?: string
      type?: string
    },
  ) {
    setLoading(true)
    setError('')
    try {
      const qKeyword = filters?.keyword ?? keyword
      const qType = filters?.type ?? typeFilter
      const [data, treeData] = await Promise.all([
        organizationService.list({ page: nextPage, page_size: 20, keyword: qKeyword || undefined, type: qType || undefined }),
        organizationService.tree(),
      ])
      const normalized = listFrom<Organization>(data)
      setItems(normalized.list)
      setTotal(normalized.total)
      setTree(Array.isArray(treeData) ? treeData : [])
      setPage(nextPage)
    } catch (err) {
      setError(err instanceof Error ? err.message : '加载失败')
    } finally {
      setLoading(false)
    }
  }

  function flattenTree(nodes: Organization[], level = 0): Array<{ id: string; name: string; level: number }> {
    const out: Array<{ id: string; name: string; level: number }> = []
    nodes.forEach((n) => {
      out.push({ id: n.id, name: n.name, level })
      if (Array.isArray(n.children) && n.children.length > 0) {
        out.push(...flattenTree(n.children, level + 1))
      }
    })
    return out
  }

  async function moveOrg(id: string) {
    const parentId = moveTarget[id] || null
    try {
      await organizationService.move(id, parentId || null)
      load(page)
      setNotice({ type: 'success', text: '组织移动成功' })
    } catch (err) {
      setNotice({ type: 'error', text: err instanceof Error ? err.message : '移动失败' })
    }
  }

  function resetFilters() {
    setKeyword('')
    setTypeFilter('')
    load(1, { keyword: '', type: '' })
  }

  async function quickCreate(e: FormEvent) {
    e.preventDefault()
    if (!name.trim() || !code.trim()) return
    try {
      await organizationService.create({
        name: name.trim(),
        code: code.trim(),
        type: type || 'team',
      })
      setName('')
      setCode('')
      setType('team')
      load(1)
      setNotice({ type: 'success', text: '组织创建成功' })
    } catch (err) {
      setNotice({ type: 'error', text: err instanceof Error ? err.message : '创建失败' })
    }
  }

  function openDetail(org: Organization) {
    setEditingOrg(org)
    setEditName(org.name || '')
    setEditCode(org.code || '')
    setEditType(org.type || 'team')
    setEditParentId(org.parent_id || '')
    setEditDesc(org.description || '')
    setEditAddress(org.address || '')
    setEditContactName(org.contact_name || '')
    setEditContactPhone(org.contact_phone || '')
    setEditSortOrder(String(org.sort_order || 0))
    setDetailOpen(true)
  }

  async function saveDetail(e: FormEvent) {
    e.preventDefault()
    if (!editingOrg) return
    setSavingDetail(true)
    try {
      await organizationService.update(editingOrg.id, {
        name: editName.trim(),
        code: editCode.trim(),
        type: editType || 'team',
        parent_id: editParentId || null,
        description: editDesc.trim(),
        address: editAddress.trim(),
        contact_name: editContactName.trim(),
        contact_phone: editContactPhone.trim(),
        sort_order: Number(editSortOrder) || 0,
      })
      setDetailOpen(false)
      setEditingOrg(null)
      load(page)
      setNotice({ type: 'success', text: '组织信息已更新' })
    } catch (err) {
      setNotice({ type: 'error', text: err instanceof Error ? err.message : '保存失败' })
    } finally {
      setSavingDetail(false)
    }
  }

  function exportCsv() {
    const rows = items
    if (rows.length === 0) return
    const headers = ['id', 'name', 'code', 'type', 'parent_id', 'contact_name', 'contact_phone', 'address', 'sort_order', 'created_at']
    const lines = rows.map((row) =>
      [
        row.id,
        row.name || '',
        row.code || '',
        row.type || '',
        row.parent_id || '',
        row.contact_name || '',
        row.contact_phone || '',
        row.address || '',
        row.sort_order ?? '',
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
    a.download = `organizations-${Date.now()}.csv`
    a.click()
    URL.revokeObjectURL(url)
    setNotice({ type: 'success', text: `已导出 ${rows.length} 条组织数据` })
  }

  useEffect(() => {
    if (ready) load(1)
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [ready])

  if (!ready) return null

  return (
    <AppShell>
      <ModuleHeader title="组织管理" desc="组织结构维护、编码治理与组织信息管理" />
      <NoticeBar notice={notice} onClose={() => setNotice(null)} />
      <form
        className="panel row wrap"
        onSubmit={(e) => {
          e.preventDefault()
          load(1)
        }}
      >
        <input className="input" placeholder="组织名称/编码关键词" value={keyword} onChange={(e) => setKeyword(e.target.value)} />
        <select className="select" value={typeFilter} onChange={(e) => setTypeFilter(e.target.value)}>
          <option value="">全部类型</option>
          <option value="team">team</option>
          <option value="group">group</option>
          <option value="branch">branch</option>
        </select>
        <button className="btn" type="submit">
          筛选
        </button>
        <button className="btn ghost" type="button" onClick={resetFilters}>
          重置
        </button>
      </form>
      <form className="panel row wrap" onSubmit={quickCreate}>
        <input className="input" placeholder="组织名称" value={name} onChange={(e) => setName(e.target.value)} />
        <input className="input" placeholder="组织编码（唯一）" value={code} onChange={(e) => setCode(e.target.value)} />
        <select className="select" value={type} onChange={(e) => setType(e.target.value)}>
          <option value="team">team</option>
          <option value="group">group</option>
          <option value="branch">branch</option>
        </select>
        <button className="btn primary" type="submit">
          创建组织
        </button>
        <button className="btn" type="button" onClick={exportCsv}>
          导出CSV
        </button>
      </form>
      <PageState loading={loading} error={error} empty={!loading && !error && items.length === 0} onRetry={() => load(page)} />
      {!loading && !error && items.length > 0 ? (
        <>
          <div className="section-card" style={{ marginBottom: 12 }}>
            <b>组织树视图</b>
            <div className="table-wrap" style={{ marginTop: 10 }}>
              <table className="table">
                <thead>
                  <tr>
                    <th>组织</th>
                    <th>层级</th>
                  </tr>
                </thead>
                <tbody>
                  {flattenTree(tree).map((node) => (
                    <tr key={node.id}>
                      <td>{`${'　'.repeat(node.level)}${node.name}`}</td>
                      <td>L{node.level}</td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </div>
          <div className="table-wrap">
            <table className="table">
              <thead>
                <tr>
                  <th>名称</th>
                  <th>编码</th>
                  <th>类型</th>
                  <th>父组织</th>
                  <th>操作</th>
                </tr>
              </thead>
              <tbody>
                {items.map((row) => (
                  <tr key={row.id}>
                    <td>{row.name}</td>
                    <td>{row.code}</td>
                    <td>{row.type || '-'}</td>
                    <td>
                      <select
                        className="select"
                        value={moveTarget[row.id] ?? row.parent_id ?? ''}
                        onChange={(e) =>
                          setMoveTarget((prev) => ({
                            ...prev,
                            [row.id]: e.target.value,
                          }))
                        }
                      >
                        <option value="">顶级组织</option>
                        {flattenTree(tree)
                          .filter((n) => n.id !== row.id)
                          .map((n) => (
                            <option key={n.id} value={n.id}>
                              {`${'　'.repeat(n.level)}${n.name}`}
                            </option>
                          ))}
                      </select>
                    </td>
                    <td>
                      <div className="row wrap">
                        <button className="btn" type="button" onClick={() => moveOrg(row.id)}>
                          移动
                        </button>
                        <ConfirmButton
                          text="删除"
                          message={`确认删除组织「${row.name}」？`}
                          onConfirm={() => {
                            organizationService.remove(row.id).then(() => load(page))
                          }}
                          className="btn danger"
                        />
                        <button className="btn ghost" type="button" onClick={() => openDetail(row)}>
                          详情/编辑
                        </button>
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
        title={editingOrg ? `组织详情：${editingOrg.name}` : '组织详情'}
        onClose={() => {
          if (!savingDetail) setDetailOpen(false)
        }}
      >
        {editingOrg ? (
          <form className="grid cols-2" onSubmit={saveDetail}>
            <div>组织ID：{editingOrg.id}</div>
            <div>创建时间：{fmtTime(editingOrg.created_at)}</div>
            <label>
              <div>名称</div>
              <input className="input" value={editName} onChange={(e) => setEditName(e.target.value)} />
            </label>
            <label>
              <div>编码</div>
              <input className="input" value={editCode} onChange={(e) => setEditCode(e.target.value)} />
            </label>
            <label>
              <div>类型</div>
              <select className="select" value={editType} onChange={(e) => setEditType(e.target.value)}>
                <option value="team">team</option>
                <option value="group">group</option>
                <option value="branch">branch</option>
              </select>
            </label>
            <label>
              <div>父组织</div>
              <select className="select" value={editParentId} onChange={(e) => setEditParentId(e.target.value)}>
                <option value="">顶级组织</option>
                {flattenTree(tree)
                  .filter((n) => n.id !== editingOrg.id)
                  .map((n) => (
                    <option key={n.id} value={n.id}>
                      {`${'　'.repeat(n.level)}${n.name}`}
                    </option>
                  ))}
              </select>
            </label>
            <label>
              <div>联系人</div>
              <input className="input" value={editContactName} onChange={(e) => setEditContactName(e.target.value)} />
            </label>
            <label>
              <div>联系电话</div>
              <input className="input" value={editContactPhone} onChange={(e) => setEditContactPhone(e.target.value)} />
            </label>
            <label>
              <div>地址</div>
              <input className="input" value={editAddress} onChange={(e) => setEditAddress(e.target.value)} />
            </label>
            <label>
              <div>排序</div>
              <input className="input" value={editSortOrder} onChange={(e) => setEditSortOrder(e.target.value.replace(/[^\d-]/g, ''))} />
            </label>
            <label style={{ gridColumn: '1 / -1' }}>
              <div>描述</div>
              <textarea className="textarea" value={editDesc} onChange={(e) => setEditDesc(e.target.value)} />
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
