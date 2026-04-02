'use client'

import { FormEvent, useEffect, useState } from 'react'
import { useRouter } from 'next/navigation'
import { AppShell } from '@/components/layout/AppShell'
import { ModuleHeader } from '@/components/shared/ModuleHeader'
import { NoticeBar, type Notice } from '@/components/shared/NoticeBar'
import { useAuthGuard } from '@/hooks/useAuthGuard'
import { listFrom } from '@/lib/data'
import { dialectService } from '@/services/dialects'
import { missingPersonService } from '@/services/missingPersons'
import { uploadService } from '@/services/upload'
import type { MissingPerson } from '@/types/api'

export default function DialectCreatePage() {
  const { ready } = useAuthGuard()
  const router = useRouter()

  const [submitting, setSubmitting] = useState(false)
  const [uploading, setUploading] = useState(false)
  const [notice, setNotice] = useState<Notice | null>(null)
  const [cases, setCases] = useState<MissingPerson[]>([])

  const [title, setTitle] = useState('')
  const [region, setRegion] = useState('')
  const [province, setProvince] = useState('')
  const [city, setCity] = useState('')
  const [dialectType, setDialectType] = useState('other')
  const [content, setContent] = useState('')
  const [description, setDescription] = useState('')
  const [tags, setTags] = useState('')

  const [audioUrl, setAudioUrl] = useState('')
  const [duration, setDuration] = useState('')
  const [format, setFormat] = useState('')
  const [fileSize, setFileSize] = useState('')
  const [missingPersonId, setMissingPersonId] = useState('')

  const [uploadedFileId, setUploadedFileId] = useState('')

  async function loadCases() {
    try {
      const data = await missingPersonService.list({ page: 1, page_size: 200 })
      setCases(listFrom<MissingPerson>(data).list)
    } catch {
      setCases([])
    }
  }

  async function uploadAudio(files: FileList | null) {
    if (!files || files.length === 0) return
    setUploading(true)
    try {
      const uploaded = await uploadService.uploadSingle(files[0], { entity_type: 'dialect' })
      const url = uploaded.url || uploaded.path || ''
      if (!url) throw new Error('上传成功但未返回可用地址')
      setAudioUrl(url)
      setUploadedFileId(uploaded.id || '')
      setNotice({ type: 'success', text: '音频上传成功，可继续提交方言记录' })
    } catch (err) {
      setNotice({ type: 'error', text: err instanceof Error ? err.message : '上传音频失败' })
    } finally {
      setUploading(false)
    }
  }

  async function submit(e: FormEvent) {
    e.preventDefault()
    if (!title.trim()) {
      setNotice({ type: 'error', text: '请填写方言标题' })
      return
    }
    if (!region.trim()) {
      setNotice({ type: 'error', text: '请填写区域' })
      return
    }
    if (!audioUrl.trim()) {
      setNotice({ type: 'error', text: '请上传或填写音频地址' })
      return
    }

    setSubmitting(true)
    setNotice(null)
    try {
      const created = await dialectService.create({
        title: title.trim(),
        region: region.trim(),
        province: province.trim(),
        city: city.trim(),
        dialect_type: dialectType,
        content: content.trim(),
        audio_url: audioUrl.trim(),
        duration: Number(duration) || 0,
        file_size: Number(fileSize) || 0,
        format: format.trim(),
        tags: tags.trim(),
        description: description.trim(),
        missing_person_id: missingPersonId || '',
      })
      if (uploadedFileId && created?.id) {
        try {
          await uploadService.bind(uploadedFileId, 'dialect', created.id)
        } catch {
          // ignore bind failure, create already success
        }
      }
      router.push(created?.id ? `/dialects/${created.id}` : '/dialects')
    } catch (err) {
      setNotice({ type: 'error', text: err instanceof Error ? err.message : '创建方言记录失败' })
    } finally {
      setSubmitting(false)
    }
  }

  useEffect(() => {
    if (ready) loadCases()
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [ready])

  if (!ready) return null

  return (
    <AppShell>
      <ModuleHeader title="新建方言记录" desc="补充语音与方言元信息，支持后续审核、评论和精选" />
      <NoticeBar notice={notice} onClose={() => setNotice(null)} />

      <form className="grid" onSubmit={submit}>
        <div className="form-section">
          <h3 className="form-section-title">基础信息</h3>
          <div className="grid cols-3">
            <label>
              <span className="field-label">标题<span className="required">*</span></span>
              <input className="input" value={title} onChange={(e) => setTitle(e.target.value)} placeholder="例如：粤语问路片段" />
            </label>
            <label>
              <span className="field-label">区域<span className="required">*</span></span>
              <input className="input" value={region} onChange={(e) => setRegion(e.target.value)} placeholder="例如：广东-广州" />
            </label>
            <label>
              <span className="field-label">方言类型</span>
              <input className="input" value={dialectType} onChange={(e) => setDialectType(e.target.value)} placeholder="例如：cantonese" />
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
          <h3 className="form-section-title">音频信息</h3>
          <div className="grid cols-3">
            <label>
              <span className="field-label">上传音频</span>
              <input className="input" type="file" accept="audio/*" onChange={(e) => uploadAudio(e.target.files)} disabled={uploading} />
            </label>
            <label>
              <span className="field-label">音频 URL<span className="required">*</span></span>
              <input className="input" value={audioUrl} onChange={(e) => setAudioUrl(e.target.value)} placeholder="可粘贴外部 URL" />
            </label>
            <label>
              <span className="field-label">时长(秒)</span>
              <input className="input" value={duration} onChange={(e) => setDuration(e.target.value.replace(/[^\d]/g, ''))} placeholder="例如：18" />
            </label>
            <label>
              <span className="field-label">格式</span>
              <input className="input" value={format} onChange={(e) => setFormat(e.target.value)} placeholder="mp3 / wav / m4a" />
            </label>
            <label>
              <span className="field-label">文件大小(bytes)</span>
              <input className="input" value={fileSize} onChange={(e) => setFileSize(e.target.value.replace(/[^\d]/g, ''))} placeholder="可选" />
            </label>
          </div>
        </div>

        <div className="form-section">
          <h3 className="form-section-title">内容说明</h3>
          <div className="grid cols-2">
            <label style={{ gridColumn: '1 / -1' }}>
              <span className="field-label">语音文本内容</span>
              <textarea className="textarea" value={content} onChange={(e) => setContent(e.target.value)} placeholder="可填写转写文本" />
            </label>
            <label>
              <span className="field-label">标签</span>
              <input className="input" value={tags} onChange={(e) => setTags(e.target.value)} placeholder="以逗号分隔，例如：问路,广州" />
            </label>
            <label>
              <span className="field-label">描述</span>
              <input className="input" value={description} onChange={(e) => setDescription(e.target.value)} placeholder="补充录制场景、用途" />
            </label>
          </div>
        </div>

        <div className="panel row wrap" style={{ marginTop: 0 }}>
          <button className="btn primary" type="submit" disabled={submitting || uploading}>
            {uploading ? '音频上传中...' : submitting ? '提交中...' : '提交方言记录'}
          </button>
          <button className="btn" type="button" onClick={() => router.push('/dialects')}>
            返回方言列表
          </button>
          <span className="hint">带 <span className="required">*</span> 为必填</span>
        </div>
      </form>
    </AppShell>
  )
}
