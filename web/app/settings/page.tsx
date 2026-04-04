'use client'

import { FormEvent, useEffect, useState } from 'react'
import { AppShell } from '@/components/layout/AppShell'
import { useAuthGuard } from '@/hooks/useAuthGuard'
import { authService } from '@/services/auth'
import { ModuleHeader } from '@/components/shared/ModuleHeader'

export default function SettingsPage() {
  const { ready } = useAuthGuard()
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [ok, setOk] = useState('')
  const [nickname, setNickname] = useState('')
  const [phone, setPhone] = useState('')
  const [role, setRole] = useState('')
  const [avatar, setAvatar] = useState('')
  const [wxBound, setWxBound] = useState(false)
  const [wechatCode, setWechatCode] = useState('')

  async function load() {
    setLoading(true)
    setError('')
    try {
      const me = await authService.me()
      setNickname(me.nickname || '')
      setPhone(me.phone || '')
      setRole(me.role || '')
      setAvatar(me.avatar || '')
      setWxBound(!!me.wx_bound)
    } catch (err) {
      setError(err instanceof Error ? err.message : '加载失败')
    } finally {
      setLoading(false)
    }
  }

  function onSave(e: FormEvent) {
    e.preventDefault()
    setOk('当前后端未提供 profile 写接口到本项目 Web 端，已保留展示和扩展位。')
    setTimeout(() => setOk(''), 2200)
  }

  async function onBindWechat(e: FormEvent) {
    e.preventDefault()
    if (!wechatCode.trim()) {
      setError('请输入微信 code')
      return
    }
    setError('')
    setOk('')
    try {
      await authService.bindWechat(wechatCode.trim())
      setOk('微信绑定成功')
      setWechatCode('')
      setWxBound(true)
    } catch (err) {
      setError(err instanceof Error ? err.message : '微信绑定失败')
    }
  }

  useEffect(() => {
    if (ready) load()
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [ready])

  if (!ready) return null

  return (
    <AppShell>
      <ModuleHeader title="个人设置" desc="查看账号信息与后续扩展配置" />
      <form className="section-card grid cols-2" onSubmit={onSave}>
        <div className="row" style={{ gridColumn: '1 / -1' }}>
          <img
            className="profile-avatar"
            src={avatar || '/default-avatar.svg'}
            alt={nickname || 'avatar'}
            onError={(e) => {
              ;(e.currentTarget as HTMLImageElement).src = '/default-avatar.svg'
            }}
          />
          <div>
            <div style={{ fontWeight: 600 }}>{nickname || phone || '当前用户'}</div>
            <div className="hint">{wxBound ? '微信已绑定' : '微信未绑定'}</div>
          </div>
        </div>
        <label>
          <div>昵称</div>
          <input className="input" value={nickname} onChange={(e) => setNickname(e.target.value)} disabled={loading} />
        </label>
        <label>
          <div>手机号</div>
          <input className="input" value={phone} onChange={(e) => setPhone(e.target.value)} disabled={loading} />
        </label>
        <label>
          <div>角色</div>
          <input className="input" value={role} disabled />
        </label>
        <div className="row" style={{ alignItems: 'flex-end' }}>
          <button className="btn primary" type="submit">
            保存
          </button>
        </div>
      </form>
      <form className="section-card row wrap" style={{ marginTop: 12 }} onSubmit={onBindWechat}>
        <b>绑定微信账号</b>
        <input
          className="input"
          style={{ minWidth: 280 }}
          value={wechatCode}
          onChange={(e) => setWechatCode(e.target.value)}
          placeholder="输入微信登录 code"
        />
        <button className="btn primary" type="submit">绑定微信</button>
      </form>
      {error ? <div className="alert">{error}</div> : null}
      {ok ? <div style={{ color: '#166534', marginTop: 10 }}>{ok}</div> : null}
    </AppShell>
  )
}
