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

  async function load() {
    setLoading(true)
    setError('')
    try {
      const me = await authService.me()
      setNickname(me.nickname || '')
      setPhone(me.phone || '')
      setRole(me.role || '')
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

  useEffect(() => {
    if (ready) load()
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [ready])

  if (!ready) return null

  return (
    <AppShell>
      <ModuleHeader title="个人设置" desc="查看账号信息与后续扩展配置" />
      <form className="section-card grid cols-2" onSubmit={onSave}>
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
      {error ? <div className="alert">{error}</div> : null}
      {ok ? <div style={{ color: '#166534', marginTop: 10 }}>{ok}</div> : null}
    </AppShell>
  )
}
