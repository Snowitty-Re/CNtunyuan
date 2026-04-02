'use client'

import { FormEvent, useEffect, useState } from 'react'
import { AppShell } from '@/components/layout/AppShell'
import { ModuleHeader } from '@/components/shared/ModuleHeader'
import { NoticeBar, type Notice } from '@/components/shared/NoticeBar'
import { PageState } from '@/components/shared/PageState'
import { useAuthGuard } from '@/hooks/useAuthGuard'
import { hasMinRole } from '@/lib/rbac'
import { fmtTime, listFrom } from '@/lib/data'
import { API_BASE } from '@/lib/request'
import { systemService } from '@/services/system'

export default function AttachmentsPage() {
  const { ready, user } = useAuthGuard()
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [notice, setNotice] = useState<Notice | null>(null)
  const [stats, setStats] = useState<Record<string, any>>({})
  const [entityType, setEntityType] = useState('missing_person')
  const [entityId, setEntityId] = useState('')
  const [fileId, setFileId] = useState('')
  const [files, setFiles] = useState<any[]>([])
  const [singleFile, setSingleFile] = useState<any | null>(null)

  async function loadStats() {
    setLoading(true)
    setError('')
    try {
      const data = await systemService.uploadStats()
      setStats(data || {})
    } catch (err) {
      setError(err instanceof Error ? err.message : '加载失败')
      setStats({})
    } finally {
      setLoading(false)
    }
  }

  async function queryByEntity(e: FormEvent) {
    e.preventDefault()
    if (!entityType || !entityId.trim()) return
    try {
      const data = await systemService.filesByEntity(entityType.trim(), entityId.trim())
      setFiles(listFrom<any>(data).list)
      setSingleFile(null)
      setNotice({ type: 'success', text: `已查询到 ${listFrom<any>(data).list.length} 条附件` })
    } catch (err) {
      setNotice({ type: 'error', text: err instanceof Error ? err.message : '查询失败' })
      setFiles([])
    }
  }

  async function queryById(e: FormEvent) {
    e.preventDefault()
    if (!fileId.trim()) return
    try {
      const data = await systemService.fileById(fileId.trim())
      setSingleFile(data)
      setNotice({ type: 'success', text: '附件详情查询成功' })
    } catch (err) {
      setNotice({ type: 'error', text: err instanceof Error ? err.message : '查询失败' })
      setSingleFile(null)
    }
  }

  async function deleteFile(id: string) {
    try {
      await systemService.deleteFile(id)
      setFiles((prev) => prev.filter((x) => x.id !== id))
      if (singleFile?.id === id) setSingleFile(null)
      setNotice({ type: 'success', text: '附件删除成功' })
    } catch (err) {
      setNotice({ type: 'error', text: err instanceof Error ? err.message : '删除失败' })
    }
  }

  useEffect(() => {
    if (ready) loadStats()
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [ready])

  if (!ready) return null
  if (!hasMinRole(user, 'admin')) {
    return (
      <AppShell>
        <ModuleHeader title="附件管理" desc="文件查询、下载、删除与统计" />
        <PageState error="当前账号无权限访问该页面（需要 admin 及以上）" />
      </AppShell>
    )
  }

  return (
    <AppShell>
      <ModuleHeader title="附件管理" desc="文件查询、下载、删除与统计" />
      <NoticeBar notice={notice} onClose={() => setNotice(null)} />
      <PageState loading={loading} error={error} onRetry={loadStats} />
      {!loading && !error ? (
        <>
          <div className="kpi-grid">
            <div className="kpi">
              <div className="label">附件总数</div>
              <div className="value">{Number(stats.total_files || stats.total || 0)}</div>
            </div>
            <div className="kpi">
              <div className="label">总大小(MB)</div>
              <div className="value">{Math.round(Number(stats.total_size_mb || stats.total_size || 0))}</div>
            </div>
            <div className="kpi">
              <div className="label">今日上传</div>
              <div className="value">{Number(stats.today_uploads || stats.today || 0)}</div>
            </div>
            <div className="kpi">
              <div className="label">受管对象</div>
              <div className="value">{Number(stats.entity_count || 0)}</div>
            </div>
          </div>

          <div className="grid cols-2">
            <form className="panel row wrap" onSubmit={queryByEntity}>
              <b>按实体查询</b>
              <select className="select" value={entityType} onChange={(e) => setEntityType(e.target.value)}>
                <option value="missing_person">missing_person</option>
                <option value="task_follow_up">task_follow_up</option>
                <option value="dialect">dialect</option>
                <option value="task">task</option>
              </select>
              <input className="input" placeholder="实体 ID" value={entityId} onChange={(e) => setEntityId(e.target.value)} />
              <button className="btn primary" type="submit">
                查询
              </button>
            </form>
            <form className="panel row wrap" onSubmit={queryById}>
              <b>按附件ID查询</b>
              <input className="input" placeholder="附件 ID" value={fileId} onChange={(e) => setFileId(e.target.value)} />
              <button className="btn primary" type="submit">
                查询
              </button>
            </form>
          </div>

          {singleFile ? (
            <div className="section-card" style={{ marginTop: 12 }}>
              <b>附件详情</b>
              <div className="grid cols-2" style={{ marginTop: 10 }}>
                <div>ID：{singleFile.id}</div>
                <div>文件名：{singleFile.original_name || singleFile.filename || '-'}</div>
                <div>类型：{singleFile.mime_type || singleFile.file_type || '-'}</div>
                <div>大小：{singleFile.size || '-'}</div>
                <div>创建时间：{fmtTime(singleFile.created_at)}</div>
                <div>实体：{singleFile.entity_type || '-'} / {singleFile.entity_id || '-'}</div>
              </div>
              <div className="row wrap" style={{ marginTop: 10 }}>
                <a className="btn" href={`${API_BASE}/upload/${singleFile.id}/download`} target="_blank" rel="noreferrer">
                  下载
                </a>
                <button className="btn danger" type="button" onClick={() => deleteFile(singleFile.id)}>
                  删除
                </button>
              </div>
            </div>
          ) : null}

          <div className="table-wrap" style={{ marginTop: 12 }}>
            <table className="table">
              <thead>
                <tr>
                  <th>ID</th>
                  <th>文件名</th>
                  <th>类型</th>
                  <th>大小</th>
                  <th>创建时间</th>
                  <th>操作</th>
                </tr>
              </thead>
              <tbody>
                {files.length === 0 ? (
                  <tr>
                    <td colSpan={6}>暂无附件数据</td>
                  </tr>
                ) : (
                  files.map((row) => (
                    <tr key={row.id}>
                      <td>{row.id}</td>
                      <td>{row.original_name || row.filename || '-'}</td>
                      <td>{row.mime_type || row.file_type || '-'}</td>
                      <td>{row.size || '-'}</td>
                      <td>{fmtTime(row.created_at)}</td>
                      <td>
                        <div className="row wrap">
                          <a className="btn ghost" href={`${API_BASE}/upload/${row.id}/download`} target="_blank" rel="noreferrer">
                            下载
                          </a>
                          <button className="btn danger" type="button" onClick={() => deleteFile(row.id)}>
                            删除
                          </button>
                        </div>
                      </td>
                    </tr>
                  ))
                )}
              </tbody>
            </table>
          </div>
        </>
      ) : null}
    </AppShell>
  )
}
