'use client'

import Link from 'next/link'
import { FormEvent, useEffect, useState } from 'react'
import { AppShell } from '@/components/layout/AppShell'
import { ModuleHeader } from '@/components/shared/ModuleHeader'
import { NoticeBar, type Notice } from '@/components/shared/NoticeBar'
import { PageState } from '@/components/shared/PageState'
import { Pagination } from '@/components/shared/Pagination'
import { StatusTag } from '@/components/shared/StatusTag'
import { ConfirmButton } from '@/components/shared/ConfirmButton'
import { Dialog } from '@/components/ui/Dialog'
import { useAuthGuard } from '@/hooks/useAuthGuard'
import { joinLocation, listFrom, fmtTime } from '@/lib/data'
import { missingPersonService } from '@/services/missingPersons'
import type { MissingPerson } from '@/types/api'

export default function CasesPage() {
  const { ready } = useAuthGuard()
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [items, setItems] = useState<MissingPerson[]>([])
  const [total, setTotal] = useState(0)
  const [page, setPage] = useState(1)
  const [keyword, setKeyword] = useState('')
  const [status, setStatus] = useState('')
  const [caseType, setCaseType] = useState('')
  const [city, setCity] = useState('')
  const [startTime, setStartTime] = useState('')
  const [endTime, setEndTime] = useState('')
  const [name, setName] = useState('')
  const [gender, setGender] = useState('male')
  const [contactName, setContactName] = useState('')
  const [contactPhone, setContactPhone] = useState('')
  const [selectedIds, setSelectedIds] = useState<string[]>([])
  const [createOpen, setCreateOpen] = useState(false)
  const [batchDeleteOpen, setBatchDeleteOpen] = useState(false)
  const [notice, setNotice] = useState<Notice | null>(null)
  const allSelected = items.length > 0 && selectedIds.length === items.length

  async function load(
    nextPage = page,
    filters?: {
      keyword?: string
      status?: string
      caseType?: string
      city?: string
      startTime?: string
      endTime?: string
    },
  ) {
    setLoading(true)
    setError('')
    try {
      const qKeyword = filters?.keyword ?? keyword
      const qStatus = filters?.status ?? status
      const qCaseType = filters?.caseType ?? caseType
      const qCity = filters?.city ?? city
      const qStart = filters?.startTime ?? startTime
      const qEnd = filters?.endTime ?? endTime
      const data = await missingPersonService.list({
        page: nextPage,
        page_size: 20,
        keyword: qKeyword || undefined,
        status: qStatus || undefined,
        case_type: qCaseType || undefined,
        city: qCity || undefined,
        start_time: qStart ? new Date(qStart).toISOString() : undefined,
        end_time: qEnd ? new Date(qEnd).toISOString() : undefined,
      })
      const normalized = listFrom<MissingPerson>(data)
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
    if (!name.trim() || !contactName.trim() || !contactPhone.trim()) return
    try {
      await missingPersonService.create({
        name: name.trim(),
        gender,
        missing_time: new Date().toISOString(),
        contact_name: contactName.trim(),
        contact_phone: contactPhone.trim(),
        status: 'missing',
      })
      setName('')
      setGender('male')
      setContactName('')
      setContactPhone('')
      load(1)
      setCreateOpen(false)
      setNotice({ type: 'success', text: '案件创建成功' })
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

  async function batchUpdateStatus(nextStatus: string) {
    if (selectedIds.length === 0) return
    try {
      await Promise.all(selectedIds.map((id) => missingPersonService.updateStatus(id, nextStatus)))
      load(page)
      setNotice({ type: 'success', text: `已更新 ${selectedIds.length} 条案件状态` })
    } catch (err) {
      setNotice({ type: 'error', text: err instanceof Error ? err.message : '批量更新失败' })
    }
  }

  async function batchDelete() {
    if (selectedIds.length === 0) return
    try {
      await Promise.all(selectedIds.map((id) => missingPersonService.remove(id)))
      load(page)
      setNotice({ type: 'success', text: `已删除 ${selectedIds.length} 条案件` })
    } catch (err) {
      setNotice({ type: 'error', text: err instanceof Error ? err.message : '批量删除失败' })
    }
  }

  function exportCsv() {
    const rows = selectedIds.length > 0 ? items.filter((x) => selectedIds.includes(x.id)) : items
    if (rows.length === 0) return
    const headers = ['id', 'name', 'gender', 'age', 'status', 'case_type', 'missing_time', 'location', 'contact_name', 'contact_phone']
    const lines = rows.map((row) =>
      [
        row.id,
        row.name,
        row.gender || '',
        row.age || '',
        row.status || '',
        row.case_type || '',
        row.missing_time || '',
        joinLocation(row),
        row.contact_name || '',
        row.contact_phone || '',
      ]
        .map((v) => `"${String(v).replace(/"/g, '""')}"`)
        .join(','),
    )
    const content = `\ufeff${[headers.join(','), ...lines].join('\n')}`
    const blob = new Blob([content], { type: 'text/csv;charset=utf-8;' })
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = `cases-${Date.now()}.csv`
    a.click()
    URL.revokeObjectURL(url)
  }

  function resetFilters() {
    setKeyword('')
    setStatus('')
    setCaseType('')
    setCity('')
    setStartTime('')
    setEndTime('')
    load(1, { keyword: '', status: '', caseType: '', city: '', startTime: '', endTime: '' })
  }

  useEffect(() => {
    if (ready) load(1)
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [ready])

  if (!ready) return null

  return (
    <AppShell>
      <ModuleHeader
        title="案件中心"
        desc="管理走失人员案件，维护状态与线索轨迹"
        right={
          <button className="btn primary" type="button" onClick={() => setCreateOpen(true)}>
            新建案件
          </button>
        }
      />
      <NoticeBar notice={notice} onClose={() => setNotice(null)} />
      <form
        className="panel row wrap"
        onSubmit={(e: FormEvent) => {
          e.preventDefault()
          load(1)
        }}
      >
        <input className="input" style={{ maxWidth: 240 }} placeholder="姓名/关键词" value={keyword} onChange={(e) => setKeyword(e.target.value)} />
        <select className="select" style={{ maxWidth: 180 }} value={status} onChange={(e) => setStatus(e.target.value)}>
          <option value="">全部状态</option>
          <option value="missing">失踪中</option>
          <option value="searching">寻找中</option>
          <option value="found">已找到</option>
          <option value="reunited">已团圆</option>
          <option value="closed">已结案</option>
        </select>
        <select className="select" style={{ maxWidth: 180 }} value={caseType} onChange={(e) => setCaseType(e.target.value)}>
          <option value="">全部案件类型</option>
          <option value="child">child</option>
          <option value="adult">adult</option>
          <option value="elderly">elderly</option>
          <option value="other">other</option>
        </select>
        <input className="input" style={{ maxWidth: 180 }} placeholder="城市" value={city} onChange={(e) => setCity(e.target.value)} />
        <input className="input" type="datetime-local" value={startTime} onChange={(e) => setStartTime(e.target.value)} />
        <input className="input" type="datetime-local" value={endTime} onChange={(e) => setEndTime(e.target.value)} />
        <button className="btn" type="submit">
          查询
        </button>
        <button className="btn ghost" type="button" onClick={resetFilters}>
          重置
        </button>
      </form>
      <div className="panel row wrap">
        <b>批量操作</b>
        <span>已选 {selectedIds.length} 条</span>
        <button className="btn ghost" type="button" onClick={() => toggleSelectAll(true)}>
          全选本页
        </button>
        <button className="btn ghost" type="button" onClick={() => toggleSelectAll(false)}>
          清空选择
        </button>
        <button className="btn ghost" type="button" onClick={invertSelection}>
          反选
        </button>
        <button className="btn" type="button" onClick={() => batchUpdateStatus('searching')}>
          批量设为寻找中
        </button>
        <button className="btn" type="button" onClick={() => batchUpdateStatus('found')}>
          批量设为已找到
        </button>
        <button className="btn" type="button" onClick={() => batchUpdateStatus('reunited')}>
          批量设为已团圆
        </button>
        <button className="btn danger" type="button" onClick={() => setBatchDeleteOpen(true)} disabled={selectedIds.length === 0}>
          批量删除
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
                  <th>姓名</th>
                  <th>性别/年龄</th>
                  <th>状态</th>
                  <th>走失时间</th>
                  <th>走失地点</th>
                  <th>操作</th>
                </tr>
              </thead>
              <tbody>
                {items.map((row) => (
                  <tr key={row.id}>
                    <td>
                      <input type="checkbox" checked={selectedIds.includes(row.id)} onChange={() => toggleSelect(row.id)} />
                    </td>
                    <td>{row.name}</td>
                    <td>
                      {row.gender} {row.age ? `${row.age}岁` : '-'}
                    </td>
                    <td>
                      <StatusTag status={row.status || '-'} />
                    </td>
                    <td>{fmtTime(row.missing_time)}</td>
                    <td>{joinLocation(row)}</td>
                    <td>
                      <div className="row wrap">
                        <Link className="btn ghost" href={`/cases/${row.id}`}>
                          查看详情
                        </Link>
                        <ConfirmButton
                          text="删除"
                          message={`确认删除案件「${row.name}」？`}
                          onConfirm={() => {
                            missingPersonService.remove(row.id).then(() => load(page))
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
      <Dialog open={createOpen} title="新建案件" onClose={() => setCreateOpen(false)}>
        <form className="grid cols-2" onSubmit={quickCreate}>
          <input className="input" placeholder="姓名（必填）" value={name} onChange={(e) => setName(e.target.value)} />
          <select className="select" value={gender} onChange={(e) => setGender(e.target.value)}>
            <option value="male">male</option>
            <option value="female">female</option>
            <option value="other">other</option>
          </select>
          <input className="input" placeholder="联系人（必填）" value={contactName} onChange={(e) => setContactName(e.target.value)} />
          <input className="input" placeholder="联系电话（必填）" value={contactPhone} onChange={(e) => setContactPhone(e.target.value)} />
          <div className="row" style={{ gridColumn: '1 / -1' }}>
            <button className="btn primary" type="submit">
              创建案件
            </button>
          </div>
        </form>
      </Dialog>
      <Dialog open={batchDeleteOpen} title="确认批量删除" onClose={() => setBatchDeleteOpen(false)}>
        <div className="grid">
          <div>确认删除选中的 {selectedIds.length} 条案件？该操作不可恢复。</div>
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
