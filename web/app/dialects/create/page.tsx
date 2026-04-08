'use client'

import { useEffect, useMemo, useState } from 'react'
import { useRouter } from 'next/navigation'
import { AppShell } from '@/components/layout/AppShell'
import { ModuleHeader } from '@/components/shared/ModuleHeader'
import { NoticeBar, type Notice } from '@/components/shared/NoticeBar'
import { PageState } from '@/components/shared/PageState'
import { useAuthGuard } from '@/hooks/useAuthGuard'
import { ACTIONS, hasPermission } from '@/lib/rbac'
import { dialectService } from '@/services/dialects'
import { missingPersonService } from '@/services/missingPersons'
import { uploadService } from '@/services/upload'
import type { DialectCard, MissingPerson } from '@/types/api'

type RecordingState = {
  card_id: string
  audio_url: string
  duration: number
  file_size: number
  format: string
  uploading?: boolean
  file_name?: string
}

function flattenCards(groups: any[]): DialectCard[] {
  const cards: DialectCard[] = []
  groups.forEach((group) => {
    const groupCards = Array.isArray(group?.cards) ? group.cards : []
    groupCards.forEach((card: DialectCard) => {
      if (!card?.id) return
      cards.push({
        ...card,
        required: card.required !== false,
      })
    })
  })
  return cards
}

function getFileFormat(file: File): string {
  const ext = file.name.split('.').pop()?.toLowerCase() || ''
  return ext || 'mp3'
}

function readAudioDuration(file: File): Promise<number> {
  return new Promise((resolve) => {
    const objectUrl = URL.createObjectURL(file)
    const audio = new Audio()
    audio.preload = 'metadata'
    audio.src = objectUrl
    audio.onloadedmetadata = () => {
      const duration = Math.max(1, Math.round(audio.duration || 0))
      URL.revokeObjectURL(objectUrl)
      resolve(duration)
    }
    audio.onerror = () => {
      URL.revokeObjectURL(objectUrl)
      resolve(3)
    }
  })
}

export default function DialectCreatePage() {
  const { ready, user } = useAuthGuard()
  const router = useRouter()

  const [loading, setLoading] = useState(true)
  const [saving, setSaving] = useState(false)
  const [notice, setNotice] = useState<Notice | null>(null)

  const [groups, setGroups] = useState<any[]>([])
  const [cards, setCards] = useState<DialectCard[]>([])
  const [activeCardIndex, setActiveCardIndex] = useState(0)
  const [recordings, setRecordings] = useState<Record<string, RecordingState>>({})

  const [cases, setCases] = useState<MissingPerson[]>([])
  const [form, setForm] = useState({
    description: '',
    tags: '',
    province: '',
    city: '',
    district: '',
    collect_address: '',
    collect_latitude: '',
    collect_longitude: '',
    missing_person_id: '',
  })

  const currentCard = cards[activeCardIndex] || null
  const completedCount = useMemo(() => {
    return cards.filter((card) => {
      const rec = recordings[card.id]
      return !!(rec && rec.audio_url && rec.duration > 0)
    }).length
  }, [cards, recordings])
  const allCardsRecorded = cards.length > 0 && completedCount === cards.length

  async function loadBaseData() {
    setLoading(true)
    try {
      const [template, casesData] = await Promise.all([
        dialectService.cardTemplate(false),
        missingPersonService.list({ page: 1, page_size: 100, status: 'missing' }).catch(() => null),
      ])

      const nextGroups = template?.groups || []
      const nextCards = flattenCards(nextGroups)
      setGroups(nextGroups)
      setCards(nextCards)
      setActiveCardIndex(0)

      const caseList = Array.isArray(casesData?.list) ? casesData.list : []
      setCases(caseList)
      if (!nextCards.length) {
        setNotice({ type: 'error', text: '暂无可录入方言卡片，请先在方言卡片管理中配置模板' })
      }
    } catch (err) {
      setNotice({ type: 'error', text: err instanceof Error ? err.message : '加载录入模板失败' })
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    if (ready) loadBaseData()
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [ready])

  async function uploadCardAudio(card: DialectCard, file: File | null) {
    if (!file || !card?.id) return
    const cardID = card.id
    setRecordings((prev) => ({
      ...prev,
      [cardID]: {
        card_id: cardID,
        audio_url: prev[cardID]?.audio_url || '',
        duration: prev[cardID]?.duration || 0,
        file_size: file.size || 0,
        format: getFileFormat(file),
        uploading: true,
        file_name: file.name,
      },
    }))

    try {
      const [uploaded, duration] = await Promise.all([
        uploadService.uploadSingle(file, { type: 'audio', entity_type: 'dialect' }),
        readAudioDuration(file),
      ])
      const url = uploaded.url || uploaded.path || ''
      if (!url) throw new Error('上传成功但未返回可用地址')

      setRecordings((prev) => ({
        ...prev,
        [cardID]: {
          card_id: cardID,
          audio_url: url,
          duration,
          file_size: file.size || 0,
          format: getFileFormat(file),
          uploading: false,
          file_name: file.name,
        },
      }))
      setNotice({ type: 'success', text: `卡片「${card.content || '未命名'}」录音已上传` })
    } catch (err) {
      setRecordings((prev) => {
        const next = { ...prev }
        delete next[cardID]
        return next
      })
      setNotice({ type: 'error', text: err instanceof Error ? err.message : '卡片录音上传失败' })
    }
  }

  function removeCardRecording(cardID: string) {
    setRecordings((prev) => {
      const next = { ...prev }
      delete next[cardID]
      return next
    })
  }

  async function submitBatch() {
    if (!cards.length) {
      setNotice({ type: 'error', text: '暂无录入卡片模板' })
      return
    }
    if (!allCardsRecorded) {
      setNotice({ type: 'error', text: '请先完成全部卡片录音' })
      return
    }
    if (!form.province.trim() || !form.city.trim()) {
      setNotice({ type: 'error', text: '请完善采集地区（省/市）' })
      return
    }
    if (!form.collect_address.trim()) {
      setNotice({ type: 'error', text: '请填写采集地址' })
      return
    }

    const lat = Number(form.collect_latitude || 0)
    const lng = Number(form.collect_longitude || 0)
    if ((lat && !lng) || (!lat && lng)) {
      setNotice({ type: 'error', text: '采集经纬度需同时填写或同时留空' })
      return
    }

    setSaving(true)
    try {
      const payload = {
        region: [form.province, form.city, form.district].filter(Boolean).join(' '),
        province: form.province.trim(),
        city: form.city.trim(),
        district: form.district.trim(),
        description: form.description.trim(),
        tags: form.tags.trim(),
        collect_address: form.collect_address.trim(),
        collect_latitude: lat || 0,
        collect_longitude: lng || 0,
        missing_person_id: form.missing_person_id || undefined,
        recordings: cards.map((card) => {
          const rec = recordings[card.id]
          return {
            card_id: card.id,
            audio_url: rec.audio_url,
            duration: rec.duration,
            file_size: rec.file_size || 0,
            format: rec.format || 'mp3',
          }
        }),
      }

      await dialectService.createBatch(payload)
      setNotice({ type: 'success', text: '方言批次提交成功' })
      setTimeout(() => router.push('/dialects'), 700)
    } catch (err) {
      setNotice({ type: 'error', text: err instanceof Error ? err.message : '提交失败' })
    } finally {
      setSaving(false)
    }
  }

  if (!ready) return null
  if (!hasPermission(user, ACTIONS.DIALECT_MODIFY)) {
    return (
      <AppShell>
        <ModuleHeader title="方言批次录入" desc="按卡片模板完成分段录音并提交" />
        <PageState error="当前账号无权限访问该页面（需要 dialect:modify 权限）" />
      </AppShell>
    )
  }

  return (
    <AppShell>
      <ModuleHeader title="方言批次录入" desc="请先完成全部卡片录音，再填写采集信息并提交" />
      <NoticeBar notice={notice} onClose={() => setNotice(null)} />
      <PageState loading={loading} empty={!loading && cards.length === 0} />

      {!loading && cards.length > 0 ? (
        <>
          <div className="section-card" style={{ marginBottom: 12 }}>
            <div className="row wrap" style={{ justifyContent: 'space-between' }}>
              <b>步骤 1/2：卡片录音（必填）</b>
              <span className="state-badge">
                已完成 {completedCount} / {cards.length}
              </span>
            </div>

            <div
              style={{
                marginTop: 10,
                height: 8,
                borderRadius: 999,
                background: '#f5e7d7',
                overflow: 'hidden',
              }}
            >
              <div
                style={{
                  height: '100%',
                  width: `${cards.length ? Math.floor((completedCount / cards.length) * 100) : 0}%`,
                  background: 'linear-gradient(90deg, #f59e0b, #d97706)',
                }}
              />
            </div>

            <div className="row wrap" style={{ marginTop: 12 }}>
              {cards.map((card, index) => {
                const done = !!recordings[card.id]?.audio_url
                return (
                  <button
                    key={card.id}
                    className={`btn ${index === activeCardIndex ? 'primary' : ''}`}
                    type="button"
                    onClick={() => setActiveCardIndex(index)}
                  >
                    {done ? '✓ ' : ''}
                    {card.content || `卡片${index + 1}`}
                  </button>
                )
              })}
            </div>
          </div>

          {currentCard ? (
            <div className="section-card" style={{ marginBottom: 12 }}>
              <b>当前卡片：{currentCard.content || '未命名卡片'}</b>
              <div className="hint" style={{ marginTop: 6 }}>
                分组：{groups.find((g) => g.id === currentCard.group_id)?.name || '未分组'}
              </div>
              <div style={{ marginTop: 12 }}>
                {currentCard.image_url ? (
                  <img
                    src={currentCard.image_url}
                    alt={currentCard.content || 'dialect-card'}
                    style={{
                      width: '100%',
                      maxHeight: 360,
                      objectFit: 'contain',
                      borderRadius: 12,
                      border: '1px solid var(--border)',
                      background: '#fff',
                    }}
                  />
                ) : (
                  <div
                    style={{
                      width: '100%',
                      minHeight: 180,
                      display: 'grid',
                      placeItems: 'center',
                      border: '1px dashed var(--border)',
                      borderRadius: 12,
                      color: 'var(--subtext)',
                    }}
                  >
                    当前卡片未配置图片
                  </div>
                )}
              </div>

              <div className="row wrap" style={{ marginTop: 12 }}>
                <input
                  className="input"
                  style={{ maxWidth: 340 }}
                  type="file"
                  accept="audio/*"
                  onChange={(event) => {
                    const file = event.target.files?.[0] || null
                    uploadCardAudio(currentCard, file)
                    event.currentTarget.value = ''
                  }}
                />
                <button
                  className="btn"
                  type="button"
                  onClick={() => setActiveCardIndex((prev) => Math.max(0, prev - 1))}
                  disabled={activeCardIndex <= 0}
                >
                  上一张
                </button>
                <button
                  className="btn"
                  type="button"
                  onClick={() => setActiveCardIndex((prev) => Math.min(cards.length - 1, prev + 1))}
                  disabled={activeCardIndex >= cards.length - 1}
                >
                  下一张
                </button>
              </div>

              {recordings[currentCard.id] ? (
                <div className="panel" style={{ marginTop: 12 }}>
                  <div className="row wrap" style={{ justifyContent: 'space-between' }}>
                    <span className="hint">
                      已上传：{recordings[currentCard.id].file_name || '-'} · {recordings[currentCard.id].duration}s
                    </span>
                    <button className="btn danger" type="button" onClick={() => removeCardRecording(currentCard.id)}>
                      清除本卡片录音
                    </button>
                  </div>
                  {recordings[currentCard.id].audio_url ? (
                    <audio controls src={recordings[currentCard.id].audio_url} style={{ width: '100%', marginTop: 10 }} />
                  ) : null}
                </div>
              ) : (
                <div className="hint" style={{ marginTop: 10 }}>
                  当前卡片尚未上传录音
                </div>
              )}
            </div>
          ) : null}

          <div className="section-card">
            <div className="row wrap" style={{ justifyContent: 'space-between' }}>
              <b>步骤 2/2：采集信息</b>
              {!allCardsRecorded ? <span className="state-badge">请先完成全部卡片录音</span> : null}
            </div>

            {!allCardsRecorded ? (
              <div className="hint" style={{ marginTop: 10 }}>
                当前仅可进行卡片录音，全部完成后可填写采集信息并提交。
              </div>
            ) : (
              <div className="grid" style={{ marginTop: 12 }}>
                <div className="grid cols-3">
                  <label>
                    <span className="field-label">
                      省<span className="required">*</span>
                    </span>
                    <input
                      className="input"
                      value={form.province}
                      onChange={(e) => setForm((prev) => ({ ...prev, province: e.target.value }))}
                      placeholder="例如：广东省"
                    />
                  </label>
                  <label>
                    <span className="field-label">
                      市<span className="required">*</span>
                    </span>
                    <input
                      className="input"
                      value={form.city}
                      onChange={(e) => setForm((prev) => ({ ...prev, city: e.target.value }))}
                      placeholder="例如：广州市"
                    />
                  </label>
                  <label>
                    <span className="field-label">区/县</span>
                    <input
                      className="input"
                      value={form.district}
                      onChange={(e) => setForm((prev) => ({ ...prev, district: e.target.value }))}
                      placeholder="例如：天河区"
                    />
                  </label>
                </div>

                <div className="grid cols-2">
                  <label>
                    <span className="field-label">
                      采集地址<span className="required">*</span>
                    </span>
                    <input
                      className="input"
                      value={form.collect_address}
                      onChange={(e) => setForm((prev) => ({ ...prev, collect_address: e.target.value }))}
                      placeholder="建议填写地图定位地址"
                    />
                  </label>
                  <label>
                    <span className="field-label">关联走失人员</span>
                    <select
                      className="select"
                      value={form.missing_person_id}
                      onChange={(e) => setForm((prev) => ({ ...prev, missing_person_id: e.target.value }))}
                    >
                      <option value="">不关联</option>
                      {cases.map((caseItem) => (
                        <option key={caseItem.id} value={caseItem.id}>
                          {caseItem.name}
                        </option>
                      ))}
                    </select>
                  </label>
                </div>

                <div className="grid cols-3">
                  <label>
                    <span className="field-label">纬度（可选）</span>
                    <input
                      className="input"
                      value={form.collect_latitude}
                      onChange={(e) => setForm((prev) => ({ ...prev, collect_latitude: e.target.value }))}
                      placeholder="31.2304"
                    />
                  </label>
                  <label>
                    <span className="field-label">经度（可选）</span>
                    <input
                      className="input"
                      value={form.collect_longitude}
                      onChange={(e) => setForm((prev) => ({ ...prev, collect_longitude: e.target.value }))}
                      placeholder="121.4737"
                    />
                  </label>
                  <label>
                    <span className="field-label">标签</span>
                    <input
                      className="input"
                      value={form.tags}
                      onChange={(e) => setForm((prev) => ({ ...prev, tags: e.target.value }))}
                      placeholder="例如：家禽,日常词汇"
                    />
                  </label>
                </div>

                <label>
                  <span className="field-label">补充描述</span>
                  <textarea
                    className="textarea"
                    value={form.description}
                    onChange={(e) => setForm((prev) => ({ ...prev, description: e.target.value }))}
                    placeholder="可填写采集说明、口音特点等"
                  />
                </label>

                <div className="row wrap">
                  <button className="btn primary" type="button" disabled={saving} onClick={submitBatch}>
                    {saving ? '提交中...' : '提交方言批次'}
                  </button>
                  <button className="btn" type="button" onClick={() => router.push('/dialects')}>
                    返回方言列表
                  </button>
                </div>
              </div>
            )}
          </div>
        </>
      ) : null}
    </AppShell>
  )
}
