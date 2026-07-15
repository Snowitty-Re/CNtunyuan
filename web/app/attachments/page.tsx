'use client'

import { useEffect, useMemo, useState } from 'react'
import { AppShell } from '@/components/layout/AppShell'
import { ModuleHeader } from '@/components/shared/ModuleHeader'
import { NoticeBar, type Notice } from '@/components/shared/NoticeBar'
import { PageState } from '@/components/shared/PageState'
import { SafeImage } from '@/components/shared/SafeImage'
import { useAuthGuard } from '@/hooks/useAuthGuard'
import { ACTIONS, hasPermission } from '@/lib/rbac'
import { fmtTime, listFrom } from '@/lib/data'
import { getAccessToken } from '@/lib/auth'
import { API_BASE } from '@/lib/request'
import { systemService } from '@/services/system'

async function downloadWithAuth(fileId: string, filename?: string) {
  const token = getAccessToken()
  const res = await fetch(`${API_BASE}/upload/${fileId}/download`, {
    headers: token ? { Authorization: `Bearer ${token}` } : {},
  })
  if (!res.ok) {
    throw new Error(`下载失败 HTTP ${res.status}`)
  }
  const blob = await res.blob()
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = filename || fileId
  document.body.appendChild(a)
  a.click()
  a.remove()
  URL.revokeObjectURL(url)
}

const fileTypeOptions = [
  { value: '', label: '全部类型' },
  { value: 'image', label: '图片' },
  { value: 'audio', label: '音频' },
  { value: 'video', label: '视频' },
  { value: 'document', label: '文档' },
]

const entityTypeOptions = [
  { value: '', label: '全部业务' },
  { value: 'missing_person', label: '案件' },
  { value: 'task', label: '任务' },
  { value: 'task_follow_up', label: '任务跟进' },
  { value: 'dialect', label: '方言录音' },
  { value: 'dialect_card', label: '方言卡片' },
  { value: 'user', label: '用户资料' },
]

const storageTypeOptions = [
  { value: '', label: '全部存储' },
  { value: 'local', label: '本地存储' },
  { value: 'oss', label: '阿里云 OSS' },
  { value: 'cos', label: '腾讯云 COS' },
]

export default function AttachmentsPage() {
  const { ready, user } = useAuthGuard()
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [notice, setNotice] = useState<Notice | null>(null)
  const [stats, setStats] = useState<Record<string, any>>({})
  const [files, setFiles] = useState<any[]>([])
  const [previewFile, setPreviewFile] = useState<any | null>(null)
  const [page, setPage] = useState(1)
  const [pageSize] = useState(12)
  const [total, setTotal] = useState(0)
  const [filters, setFilters] = useState({
    keyword: '',
    fileType: '',
    entityType: '',
    storageType: '',
    uploaderID: '',
  })

  async function loadStats() {
    const data = await systemService.uploadStats()
    setStats(data || {})
  }

  async function loadFiles(nextPage = page, nextFilters = filters) {
    setLoading(true)
    setError('')
    try {
      const data = await systemService.listFiles({
        page: nextPage,
        page_size: pageSize,
        keyword: nextFilters.keyword.trim() || undefined,
        file_type: nextFilters.fileType || undefined,
        entity_type: nextFilters.entityType || undefined,
        storage_type: nextFilters.storageType || undefined,
        uploader_id: nextFilters.uploaderID.trim() || undefined,
      })
      const result = listFrom<any>(data)
      setFiles(result.list)
      setTotal(Number(result.total || 0))
      setPreviewFile((current: any | null) => {
        if (current) {
          const matched = result.list.find((item: any) => item.id === current.id)
          if (matched) return matched
        }
        return result.list[0] || null
      })
      setPage(nextPage)
    } catch (err) {
      setError(err instanceof Error ? err.message : '文件列表加载失败')
      setFiles([])
      setPreviewFile(null)
      setTotal(0)
    } finally {
      setLoading(false)
    }
  }

  async function loadAll() {
    try {
      await Promise.all([loadStats(), loadFiles(1, filters)])
    } catch {
      // errors are handled in children
    }
  }

  async function deleteFile(id: string) {
    try {
      await systemService.deleteFile(id)
      setNotice({ type: 'success', text: '文件删除成功' })
      await Promise.all([loadStats(), loadFiles(page, filters)])
    } catch (err) {
      setNotice({ type: 'error', text: err instanceof Error ? err.message : '删除失败' })
    }
  }

  async function handleDownload(file: any) {
    if (!file?.id) return
    try {
      await downloadWithAuth(file.id, file.original_name || file.file_name)
      setNotice({ type: 'success', text: '已开始下载' })
    } catch (err) {
      setNotice({ type: 'error', text: err instanceof Error ? err.message : '下载失败' })
    }
  }

  useEffect(() => {
    if (ready) {
      loadAll()
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [ready])

  const currentFile = previewFile
  const previewKind = useMemo(() => detectPreviewKind(currentFile), [currentFile])
  const totalPages = Math.max(1, Math.ceil(total / pageSize))

  if (!ready) return null
  if (!hasPermission(user, ACTIONS.USER_MODIFY)) {
    return (
      <AppShell>
        <ModuleHeader title="文件管理中心" desc="统一查看、筛选、预览、下载与删除平台文件" />
        <PageState error="当前账号无权限访问该页面（需要 user:modify 权限）" />
      </AppShell>
    )
  }

  return (
    <AppShell>
      <ModuleHeader
        title="文件管理中心"
        desc="按文件维度统一管理平台内图片、音频、视频和文档"
        right={
          <button className="btn" type="button" onClick={loadAll}>
            刷新数据
          </button>
        }
      />
      <NoticeBar notice={notice} onClose={() => setNotice(null)} />
      <PageState loading={loading && files.length === 0} error={error} onRetry={loadAll} />
      {!error ? (
        <>
          <div className="insight-grid">
            <div className="stat-card stat-card-amber">
              <span>文件总数</span>
              <strong>{Number(stats.total_count || stats.total_files || 0)}</strong>
              <small>全站当前已接管文件</small>
            </div>
            <div className="stat-card stat-card-blue">
              <span>总大小</span>
              <strong>{formatSize(Number(stats.total_size || 0))}</strong>
              <small>存储资源整体占用</small>
            </div>
            <div className="stat-card stat-card-green">
              <span>音频文件</span>
              <strong>{Number(stats.audio_count || 0)}</strong>
              <small>适用于方言录音和跟进材料</small>
            </div>
            <div className="stat-card stat-card-rose">
              <span>图片文件</span>
              <strong>{Number(stats.image_count || 0)}</strong>
              <small>适用于案件照片和卡片图片</small>
            </div>
          </div>

          <section className="section-card">
            <div className="filters-grid">
              <input
                className="input"
                placeholder="搜索文件名、原始文件名、描述"
                value={filters.keyword}
                onChange={(e) => setFilters((prev) => ({ ...prev, keyword: e.target.value }))}
              />
              <select className="select" value={filters.fileType} onChange={(e) => setFilters((prev) => ({ ...prev, fileType: e.target.value }))}>
                {fileTypeOptions.map((item) => (
                  <option key={item.value} value={item.value}>{item.label}</option>
                ))}
              </select>
              <select className="select" value={filters.entityType} onChange={(e) => setFilters((prev) => ({ ...prev, entityType: e.target.value }))}>
                {entityTypeOptions.map((item) => (
                  <option key={item.value} value={item.value}>{item.label}</option>
                ))}
              </select>
              <select className="select" value={filters.storageType} onChange={(e) => setFilters((prev) => ({ ...prev, storageType: e.target.value }))}>
                {storageTypeOptions.map((item) => (
                  <option key={item.value} value={item.value}>{item.label}</option>
                ))}
              </select>
              <input
                className="input"
                placeholder="上传人 ID"
                value={filters.uploaderID}
                onChange={(e) => setFilters((prev) => ({ ...prev, uploaderID: e.target.value }))}
              />
              <div className="row wrap">
                <button className="btn primary" type="button" onClick={() => loadFiles(1, filters)}>
                  应用筛选
                </button>
                <button
                  className="btn ghost"
                  type="button"
                  onClick={() => {
                    const next = { keyword: '', fileType: '', entityType: '', storageType: '', uploaderID: '' }
                    setFilters(next)
                    loadFiles(1, next)
                  }}
                >
                  重置
                </button>
              </div>
            </div>
          </section>

          <div className="attachment-layout">
            <section className="section-card attachment-preview-shell">
              <div className="row wrap" style={{ justifyContent: 'space-between' }}>
                <div>
                  <b>文件预览</b>
                  <div className="hint">点击右侧任一文件即可在此查看</div>
                </div>
                {currentFile ? (
                  <a className="btn ghost" href={resolveFileUrl(currentFile)} target="_blank" rel="noreferrer">
                    打开原文件
                  </a>
                ) : null}
              </div>
              {currentFile ? (
                <>
                  <div className="attachment-preview-box">{renderPreview(currentFile, previewKind)}</div>
                  <div className="attachment-meta-grid">
                    <div>文件名：{currentFile.original_name || currentFile.file_name || '-'}</div>
                    <div>类型：{currentFile.mime_type || currentFile.file_type || '-'}</div>
                    <div>大小：{currentFile.size_readable || formatSize(Number(currentFile.size || 0))}</div>
                    <div>上传时间：{fmtTime(currentFile.created_at)}</div>
                    <div>业务类型：{currentFile.entity_type || '未绑定'}</div>
                    <div>业务 ID：{currentFile.entity_id || '-'}</div>
                    <div>存储方式：{currentFile.storage_type || '-'}</div>
                    <div>上传人：{currentFile.uploader_id || '-'}</div>
                    <div>路径：{currentFile.path || '-'}</div>
                  </div>
                  <div className="row wrap" style={{ marginTop: 12 }}>
                    <button className="btn" type="button" onClick={() => handleDownload(currentFile)}>
                      下载
                    </button>
                    <button className="btn danger" type="button" onClick={() => deleteFile(currentFile.id)}>
                      删除
                    </button>
                  </div>
                </>
              ) : (
                <div className="empty-preview">当前筛选条件下暂无可预览文件</div>
              )}
            </section>

            <section className="section-card">
              <div className="row wrap" style={{ justifyContent: 'space-between' }}>
                <div>
                  <b>文件列表</b>
                  <div className="hint">当前共 {total} 个文件</div>
                </div>
                <div className="hint">第 {page} / {totalPages} 页</div>
              </div>
              <div className="table-wrap" style={{ marginTop: 12 }}>
                <table className="table">
                  <thead>
                    <tr>
                      <th>文件</th>
                      <th>类型</th>
                      <th>业务</th>
                      <th>上传人</th>
                      <th>时间</th>
                      <th>操作</th>
                    </tr>
                  </thead>
                  <tbody>
                    {files.length === 0 ? (
                      <tr>
                        <td colSpan={6}>暂无文件数据</td>
                      </tr>
                    ) : (
                      files.map((row) => (
                        <tr key={row.id} className={currentFile?.id === row.id ? 'table-row-active' : ''}>
                          <td>
                            <button className="link-button" type="button" onClick={() => setPreviewFile(row)}>
                              {row.original_name || row.file_name || row.id}
                            </button>
                            <div className="hint">{row.size_readable || formatSize(Number(row.size || 0))}</div>
                          </td>
                          <td>{row.file_type || '-'}</td>
                          <td>{row.entity_type || '未绑定'}</td>
                          <td>{row.uploader_id || '-'}</td>
                          <td>{fmtTime(row.created_at)}</td>
                          <td>
                            <div className="row wrap">
                              <button className="btn ghost" type="button" onClick={() => handleDownload(row)}>
                                下载
                              </button>
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
              <div className="row wrap" style={{ justifyContent: 'space-between', marginTop: 12 }}>
                <button className="btn ghost" type="button" disabled={page <= 1} onClick={() => loadFiles(page - 1, filters)}>
                  上一页
                </button>
                <button className="btn ghost" type="button" disabled={page >= totalPages} onClick={() => loadFiles(page + 1, filters)}>
                  下一页
                </button>
              </div>
            </section>
          </div>
        </>
      ) : null}
    </AppShell>
  )
}

function detectPreviewKind(file: any) {
  const mime = String(file?.mime_type || '').toLowerCase()
  const type = String(file?.file_type || '').toLowerCase()
  const url = resolveFileUrl(file)
  if (type === 'image' || mime.startsWith('image/')) return 'image'
  if (type === 'audio' || mime.startsWith('audio/')) return 'audio'
  if (type === 'video' || mime.startsWith('video/')) return 'video'
  if (mime.includes('pdf') || url.endsWith('.pdf')) return 'pdf'
  return 'file'
}

function resolveFileUrl(file: any) {
  const path = String(file?.path || '').trim()
  if (path) return buildUploadUrl(path)
  const raw = String(file?.url || '').trim()
  if (raw) return normalizeFileUrl(raw)
  return `${API_BASE}/upload/${file?.id}/download`
}

function buildUploadUrl(path: string) {
  try {
    const apiOrigin = new URL(API_BASE).origin
    const normalized = path.replace(/^\/+/, '').replace(/^uploads\/+/, '')
    return `${apiOrigin}/uploads/${normalized}`
  } catch {
    const normalized = path.replace(/^\/+/, '').replace(/^uploads\/+/, '')
    return `/uploads/${normalized}`
  }
}

function normalizeFileUrl(raw: string) {
  try {
    const apiOrigin = new URL(API_BASE).origin
    return new URL(raw, apiOrigin).toString()
  } catch {
    return raw
  }
}

function renderPreview(file: any, kind: string) {
  const url = resolveFileUrl(file)
  if (kind === 'image') {
    return (
      <SafeImage
        className="attachment-preview-image"
        src={url}
        alt={file?.original_name || 'preview'}
        width={1200}
        height={900}
        style={{ width: '100%', height: 'auto', maxHeight: 420, objectFit: 'contain' }}
      />
    )
  }
  if (kind === 'audio') {
    return <audio controls preload="metadata" style={{ width: '100%' }} src={url} />
  }
  if (kind === 'video') {
    return <video controls preload="metadata" className="attachment-preview-video" src={url} />
  }
  if (kind === 'pdf') {
    return <iframe className="attachment-preview-frame" src={url} title="attachment-preview" />
  }
  return (
    <div className="attachment-file-fallback">
      <b>当前文件暂不支持内嵌预览</b>
      <span>{file?.original_name || file?.file_name || '未命名文件'}</span>
      <a className="btn" href={url} target="_blank" rel="noreferrer">打开文件</a>
    </div>
  )
}

function formatSize(size: number) {
  if (!size) return '0 B'
  if (size >= 1024 * 1024 * 1024) return `${(size / 1024 / 1024 / 1024).toFixed(2)} GB`
  if (size >= 1024 * 1024) return `${(size / 1024 / 1024).toFixed(2)} MB`
  if (size >= 1024) return `${(size / 1024).toFixed(2)} KB`
  return `${size} B`
}
