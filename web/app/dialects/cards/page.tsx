'use client'

import { FormEvent, useEffect, useMemo, useState } from 'react'
import { AppShell } from '@/components/layout/AppShell'
import { ModuleHeader } from '@/components/shared/ModuleHeader'
import { NoticeBar, type Notice } from '@/components/shared/NoticeBar'
import { PageState } from '@/components/shared/PageState'
import { ConfirmButton } from '@/components/shared/ConfirmButton'
import { useAuthGuard } from '@/hooks/useAuthGuard'
import { ACTIONS, hasPermission } from '@/lib/rbac'
import { dialectService } from '@/services/dialects'
import { uploadService } from '@/services/upload'
import type { DialectCard, DialectCardGroup } from '@/types/api'

export default function DialectCardsPage() {
  const { ready, user } = useAuthGuard()
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [notice, setNotice] = useState<Notice | null>(null)

  const [groups, setGroups] = useState<DialectCardGroup[]>([])
  const [activeGroupId, setActiveGroupId] = useState('')

  const [groupName, setGroupName] = useState('')
  const [cardName, setCardName] = useState('')
  const [cardImageFile, setCardImageFile] = useState<File | null>(null)
  const [cardCreating, setCardCreating] = useState(false)

  const cards = useMemo(() => {
    return groups.find((g) => g.id === activeGroupId)?.cards || []
  }, [groups, activeGroupId])

  async function loadAll() {
    setLoading(true)
    setError('')
    try {
      const res = await dialectService.cardGroups()
      const nextGroups = (res.groups || []).map((g) => ({
        ...g,
        cards: Array.isArray(g.cards) ? g.cards : [],
      }))
      setGroups(nextGroups)
      setActiveGroupId((prev) => {
        if (prev && nextGroups.some((g) => g.id === prev)) return prev
        return nextGroups[0]?.id || ''
      })
    } catch (err) {
      setError(err instanceof Error ? err.message : '加载卡片模板失败')
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    if (ready) loadAll()
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [ready])

  async function createGroup(e: FormEvent) {
    e.preventDefault()
    if (!groupName.trim()) return
    try {
      await dialectService.createCardGroup({ name: groupName.trim(), status: 'active' })
      setGroupName('')
      await loadAll()
      setNotice({ type: 'success', text: '分组创建成功' })
    } catch (err) {
      setNotice({ type: 'error', text: err instanceof Error ? err.message : '分组创建失败' })
    }
  }

  async function toggleGroup(group: DialectCardGroup) {
    try {
      await dialectService.updateCardGroup(group.id, { status: group.status === 'active' ? 'inactive' : 'active' })
      await loadAll()
      setNotice({ type: 'success', text: '分组状态已更新' })
    } catch (err) {
      setNotice({ type: 'error', text: err instanceof Error ? err.message : '更新失败' })
    }
  }

  async function removeGroup(group: DialectCardGroup) {
    try {
      await dialectService.removeCardGroup(group.id)
      await loadAll()
      setNotice({ type: 'success', text: '分组已删除' })
    } catch (err) {
      setNotice({ type: 'error', text: err instanceof Error ? err.message : '删除失败，请先删除分组内卡片' })
    }
  }

  async function uploadImage(file: File): Promise<string> {
    const res = await uploadService.uploadSingle(file, {
      type: 'image',
      entity_type: 'dialect_card',
    })
    const url = res.url || res.path || ''
    if (!url) throw new Error('图片上传成功但未返回URL')
    return url
  }

  async function createCard(e: FormEvent) {
    e.preventDefault()
    if (!activeGroupId) {
      setNotice({ type: 'error', text: '请先选择分组' })
      return
    }
    if (!cardName.trim()) {
      setNotice({ type: 'error', text: '请填写卡片名称' })
      return
    }
    if (!cardImageFile) {
      setNotice({ type: 'error', text: '请上传卡片图片' })
      return
    }

    setCardCreating(true)
    try {
      const imageUrl = await uploadImage(cardImageFile)
      await dialectService.createCard({
        group_id: activeGroupId,
        content: cardName.trim(),
        image_url: imageUrl,
        status: 'active',
        required: true,
      })
      setCardName('')
      setCardImageFile(null)
      const fileInput = document.getElementById('card-image-input') as HTMLInputElement | null
      if (fileInput) fileInput.value = ''
      await loadAll()
      setNotice({ type: 'success', text: '卡片创建成功' })
    } catch (err) {
      setNotice({ type: 'error', text: err instanceof Error ? err.message : '卡片创建失败' })
    } finally {
      setCardCreating(false)
    }
  }

  async function renameCard(card: DialectCard) {
    const nextName = window.prompt('请输入卡片名称', card.content || '')?.trim() || ''
    if (!nextName) return
    try {
      await dialectService.updateCard(card.id, { content: nextName })
      await loadAll()
      setNotice({ type: 'success', text: '卡片名称已更新' })
    } catch (err) {
      setNotice({ type: 'error', text: err instanceof Error ? err.message : '更新失败' })
    }
  }

  async function replaceCardImage(card: DialectCard, file: File | null) {
    if (!file) return
    try {
      const imageUrl = await uploadImage(file)
      await dialectService.updateCard(card.id, { image_url: imageUrl })
      await loadAll()
      setNotice({ type: 'success', text: '卡片图片已更新' })
    } catch (err) {
      setNotice({ type: 'error', text: err instanceof Error ? err.message : '图片更新失败' })
    }
  }

  async function toggleCard(card: DialectCard) {
    try {
      await dialectService.updateCard(card.id, { status: card.status === 'active' ? 'inactive' : 'active' })
      await loadAll()
      setNotice({ type: 'success', text: '卡片状态已更新' })
    } catch (err) {
      setNotice({ type: 'error', text: err instanceof Error ? err.message : '更新失败' })
    }
  }

  async function removeCard(card: DialectCard) {
    try {
      await dialectService.removeCard(card.id)
      await loadAll()
      setNotice({ type: 'success', text: '卡片已删除' })
    } catch (err) {
      setNotice({ type: 'error', text: err instanceof Error ? err.message : '删除失败' })
    }
  }

  if (!ready) return null
  if (!hasPermission(user, ACTIONS.DIALECT_MANAGE)) {
    return (
      <AppShell>
        <ModuleHeader title="方言卡片管理" desc="管理方言录入卡片模板（分组与图片）" />
        <PageState error="当前账号无权限访问该页面（需要 dialect:manage 权限）" />
      </AppShell>
    )
  }

  return (
    <AppShell>
      <ModuleHeader title="方言卡片管理" desc="管理方言录入卡片模板（分组与图片）" />
      <NoticeBar notice={notice} onClose={() => setNotice(null)} />

      <div className="grid cols-2" style={{ marginBottom: 12 }}>
        <form className="section-card row wrap" onSubmit={createGroup}>
          <b>新建分组</b>
          <input className="input" placeholder="例如：家禽类" value={groupName} onChange={(e) => setGroupName(e.target.value)} />
          <button className="btn primary" type="submit">
            创建分组
          </button>
        </form>

        <form className="section-card row wrap" onSubmit={createCard}>
          <b>新建卡片</b>
          <select className="select" value={activeGroupId} onChange={(e) => setActiveGroupId(e.target.value)}>
            {groups.length === 0 ? <option value="">暂无分组</option> : null}
            {groups.map((g) => (
              <option key={g.id} value={g.id}>
                {g.name} ({(g.cards || []).length})
              </option>
            ))}
          </select>
          <input className="input" placeholder="卡片名称（如：鸡）" value={cardName} onChange={(e) => setCardName(e.target.value)} />
          <input
            id="card-image-input"
            className="input"
            type="file"
            accept="image/*"
            onChange={(e) => setCardImageFile(e.target.files?.[0] || null)}
          />
          <button className="btn primary" type="submit" disabled={cardCreating}>
            {cardCreating ? '创建中...' : '创建卡片'}
          </button>
        </form>
      </div>

      <PageState loading={loading} error={error} onRetry={loadAll} empty={!loading && !error && groups.length === 0} />

      {!loading && !error && groups.length > 0 ? (
        <>
          <div className="section-card" style={{ marginBottom: 12 }}>
            <b>分组列表</b>
            <div className="row wrap" style={{ marginTop: 10 }}>
              {groups.map((group) => (
                <div
                  key={group.id}
                  style={{
                    border: '1px solid var(--border)',
                    borderRadius: 12,
                    padding: '8px 10px',
                    background: activeGroupId === group.id ? 'var(--bg-accent)' : '#fff',
                    cursor: 'pointer',
                    minWidth: 220,
                  }}
                  onClick={() => setActiveGroupId(group.id)}
                >
                  <div className="row wrap" style={{ justifyContent: 'space-between' }}>
                    <b>{group.name}</b>
                    <span className="hint">{group.status || 'active'}</span>
                  </div>
                  <div className="row wrap" style={{ marginTop: 8 }}>
                    <button className="btn" type="button" onClick={(e) => { e.stopPropagation(); toggleGroup(group) }}>
                      {group.status === 'active' ? '停用' : '启用'}
                    </button>
                    <ConfirmButton
                      text="删除"
                      message={`确认删除分组「${group.name}」？`}
                      onConfirm={() => removeGroup(group)}
                      className="btn danger"
                    />
                  </div>
                </div>
              ))}
            </div>
          </div>

          <div className="table-wrap">
            <table className="table">
              <thead>
                <tr>
                  <th>图片</th>
                  <th>名称</th>
                  <th>状态</th>
                  <th>排序/必录</th>
                  <th>操作</th>
                </tr>
              </thead>
              <tbody>
                {cards.length === 0 ? (
                  <tr>
                    <td colSpan={5}>当前分组暂无卡片</td>
                  </tr>
                ) : (
                  cards.map((card) => (
                    <tr key={card.id}>
                      <td>
                        {card.image_url ? (
                          <img src={card.image_url} alt={card.content || 'card'} style={{ width: 88, height: 88, objectFit: 'cover', borderRadius: 10 }} />
                        ) : (
                          <div style={{ width: 88, height: 88, borderRadius: 10, background: 'var(--bg-soft)', display: 'flex', alignItems: 'center', justifyContent: 'center' }}>无图</div>
                        )}
                      </td>
                      <td>{card.content || '-'}</td>
                      <td>{card.status || 'active'}</td>
                      <td>
                        {(card.sort_order ?? 0)} / {card.required === false ? '否' : '是'}
                      </td>
                      <td>
                        <div className="row wrap">
                          <button className="btn" type="button" onClick={() => renameCard(card)}>
                            改名
                          </button>
                          <label className="btn" style={{ cursor: 'pointer' }}>
                            换图
                            <input
                              style={{ display: 'none' }}
                              type="file"
                              accept="image/*"
                              onChange={(e) => {
                                const file = e.target.files?.[0] || null
                                replaceCardImage(card, file)
                                e.currentTarget.value = ''
                              }}
                            />
                          </label>
                          <button className="btn" type="button" onClick={() => toggleCard(card)}>
                            {card.status === 'active' ? '停用' : '启用'}
                          </button>
                          <ConfirmButton
                            text="删除"
                            message={`确认删除卡片「${card.content || card.id}」？`}
                            onConfirm={() => removeCard(card)}
                            className="btn danger"
                          />
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
