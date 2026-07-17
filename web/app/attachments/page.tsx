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
import { uploadService } from '@/services/upload'
import { systemService } from '@/services/system'

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
  const [previewUrl, setPreviewUrl] = useState('')
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
      // errors handled in children
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
      const blob = await uploadService.downloadBlob(file.id)
      const url = URL.createObjectURL(blob)
      const a = document.createElement('a')
      a.href = url
      a.download = file.original_name || file.file_name || file.id
      document.body.appendChild(a)
      a.click()
      a.remove()
      URL.revokeObjectURL(url)
      setNotice({ type: 'success', text: '已开始下载' })
    } catch (err) {
      setNotice({ type: 'error', text: err instanceof Error ? err.message : '下载失败' })
    }
  }

  useEffect(() => {
    if (ready) loadAll()
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [ready])

  useEffect(() => {
    let revoked = false
    let objectUrl = ''
    async function loadPreview() {
      if (!previewFile?.id) {
        setPreviewUrl('')
        return
      }
      try {
        const blob = await uploadService.downloadBlob(previewFile.id)
        if (revoked) return
        objectUrl = URL.createObjectURL(blob)
        setPreviewUrl(objectUrl)
      } catch {
        if (!revoked) setPreviewUrl('')
      }
    }
    loadPreview()
    return () => {
      revoked = true
      if (objectUrl) URL.revokeObjectURL(objectUrl)
    }
  }, [previewFile?.id])

  const currentFile = previewFile
  const previewKind = useMemo(() => detectPreviewKind(currentFile), [currentFile])
  const totalPages = Math.max(1, Math.ceil(total / pageSize))

  if (!ready) {
    return (
      <AppShell>
        <PageState loading />
      </AppShell>
    )
  }
  if (!hasPermission(user, ACTIONS.USER_MODIFY)) {
    return (
      <AppShell>
        <ModuleHeader title="附件中心" desc="无权限访问" />
        <div className="empty-preview">需要管理员权限</div>
      </AppShell>
    )
  }

  return (
    <AppShell>
      <ModuleHeader title="附件中心" desc="文件统计、筛选与鉴权下载" />
      <NoticeBar notice={notice} onClose={() => setNotice(null)} />
      <PageState loading={loading} error={error} onRetry={loadAll} />
      {!loading && !error ? (
        <div className="grid">
          <div className="insight-grid">
            <div className="stat-card">
              <span>文件总数</span>
              <strong>{stats.total_count ?? files.length}</strong>
            </div>
            <div className="stat-card">
              <span>总大小</span>
              <strong>{stats.total_size_readable || formatSize(Number(stats.total_size || 0))}</strong>
            </div>
          </div>

          <section className="section-card">
            <div className="row wrap" style={{ gap: 8 }}>
              <input
                className="input"
                placeholder="关键词"
                value={filters.keyword}
                onChange={(e) => setFilters((s) => ({ ...s, keyword: e.target.value }))}
              />
              <select className="select" value={filters.fileType} onChange={(e) => setFilters((s) => ({ ...s, fileType: e.target.value }))}>
                {fileTypeOptions.map((o) => (
                  <option key={o.value || 'all-type'} value={o.value}>
                    {o.label}
                  </option>
                ))}
              </select>
              <select className="select" value={filters.entityType} onChange={(e) => setFilters((s) => ({ ...s, entityType: e.target.value }))}>
                {entityTypeOptions.map((o) => (
                  <option key={o.value || 'all-entity'} value={o.value}>
                    {o.label}
                  </option>
                ))}
              </select>
              <select className="select" value={filters.storageType} onChange={(e) => setFilters((s) => ({ ...s, storageType: e.target.value }))}>
                {storageTypeOptions.map((o) => (
                  <option key={o.value || 'all-storage'} value={o.value}>
                    {o.label}
                  </option>
                ))}
              </select>
              <button className="btn" type="button" onClick={() => loadFiles(1, filters)}>
                查询
              </button>
            </div>
          </section>

          <div className="grid cols-2">
            <section className="section-card">
              <b>预览</b>
              <div className="hint">通过鉴权下载接口加载，无需公开 /uploads</div>
              {currentFile && previewUrl ? (
                <>
                  <div style={{ marginTop: 10 }}>{renderPreview(previewUrl, previewKind, currentFile)}</div>
                  <div className="meta-grid" style={{ marginTop: 12 }}>
                    <div>文件名：{currentFile.original_name || currentFile.file_name || '-'}</div>
                    <div>类型：{currentFile.mime_type || currentFile.file_type || '-'}</div>
                    <div>大小：{currentFile.size_readable || formatSize(Number(currentFile.size || 0))}</div>
                    <div>业务：{currentFile.entity_type || '未绑定'}</div>
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
                <div className="empty-preview">{currentFile ? '预览加载中或无权预览' : '暂无可预览文件'}</div>
              )}
            </section>

            <section className="section-card">
              <b>文件列表</b>
              <div className="hint">当前共 {total} 个文件</div>
              <div className="table-wrap" style={{ marginTop: 10 }}>
                <table className="table">
                  <thead>
                    <tr>
                      <th>文件</th>
                      <th>类型</th>
                      <th>业务</th>
                      <th>时间</th>
                      <th>操作</th>
                    </tr>
                  </thead>
                  <tbody>
                    {files.length === 0 ? (
                      <tr>
                        <td colSpan={5}>暂无数据</td>
                      </tr>
                    ) : (
                      files.map((row) => (
                        <tr key={row.id} className={currentFile?.id === row.id ? 'table-row-active' : ''}>
                          <td>
                            <button className="link-button" type="button" onClick={() => setPreviewFile(row)}>
                              {row.original_name || row.file_name || row.id}
                            </button>
                          </td>
                          <td>{row.file_type || '-'}</td>
                          <td>{row.entity_type || '未绑定'}</td>
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
                <span className="hint">
                  第 {page}/{totalPages} 页
                </span>
                <button className="btn ghost" type="button" disabled={page >= totalPages} onClick={() => loadFiles(page + 1, filters)}>
                  下一页
                </button>
              </div>
            </section>
          </div>
        </div>
      ) : null}
    </AppShell>
  )
}

function detectPreviewKind(file: any) {
  const mime = String(file?.mime_type || '').toLowerCase()
  const type = String(file?.file_type || '').toLowerCase()
  const name = String(file?.original_name || file?.file_name || '').toLowerCase()
  if (type === 'image' || mime.startsWith('image/')) return 'image'
  if (type === 'audio' || mime.startsWith('audio/')) return 'audio'
  if (type === 'video' || mime.startsWith('video/')) return 'video'
  if (mime.includes('pdf') || name.endsWith('.pdf')) return 'pdf'
  return 'file'
}

function formatSize(n: number) {
  if (!n) return '0 B'
  if (n < 1024) return `${n} B`
  if (n < 1024 * 1024) return `${(n / 1024).toFixed(1)} KB`
  return `${(n / 1024 / 1024).toFixed(1)} MB`
}

function renderPreview(url: string, kind: string, file: any) {
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
  if (kind === 'audio') return <audio controls preload="metadata" style={{ width: '100%' }} src={url} />
  if (kind === 'video') return <video controls style={{ width: '100%', maxHeight: 420 }} src={url} />
  if (kind === 'pdf') return <iframe title="pdf" src={url} style={{ width: '100%', height: 420, border: '1px solid #eee' }} />
  return <div className="hint">该类型请使用下载查看</div>
}
