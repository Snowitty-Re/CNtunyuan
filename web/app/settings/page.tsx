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
  const [email, setEmail] = useState('')
  const [savingProfile, setSavingProfile] = useState(false)
  const [oldPassword, setOldPassword] = useState('')
  const [newPassword, setNewPassword] = useState('')
  const [confirmPassword, setConfirmPassword] = useState('')
  const [changingPassword, setChangingPassword] = useState(false)

  async function load() {
    setLoading(true)
    setError('')
    try {
      const me = await authService.me()
      setNickname(me.nickname || '')
      setPhone(me.phone || '')
      setEmail(me.email || '')
      setRole(me.role || '')
      setAvatar(me.avatar || '')
      setWxBound(!!me.wx_bound)
    } catch (err) {
      setError(err instanceof Error ? err.message : '加载失败')
    } finally {
      setLoading(false)
    }
  }

  async function onSave(e: FormEvent) {
    e.preventDefault()
    setError('')
    setOk('')
    setSavingProfile(true)
    try {
      const profile = await authService.updateProfile({
        nickname: nickname.trim(),
        email: email.trim(),
      })
      setNickname(profile.nickname || '')
      setEmail(profile.email || '')
      setOk('资料已保存')
    } catch (err) {
      setError(err instanceof Error ? err.message : '保存失败')
    } finally {
      setSavingProfile(false)
    }
  }

  async function onChangePassword(e: FormEvent) {
    e.preventDefault()
    if (!oldPassword.trim() || !newPassword.trim()) {
      setError('请填写完整密码信息')
      return
    }
    if (newPassword.length < 8) {
      setError('新密码至少 8 位')
      return
    }
    if (newPassword !== confirmPassword) {
      setError('两次输入的新密码不一致')
      return
    }

    setError('')
    setOk('')
    setChangingPassword(true)
    try {
      await authService.changePassword(oldPassword, newPassword)
      setOldPassword('')
      setNewPassword('')
      setConfirmPassword('')
      setOk('密码修改成功')
    } catch (err) {
      setError(err instanceof Error ? err.message : '修改密码失败')
    } finally {
      setChangingPassword(false)
    }
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

  async function onUnbindWechat() {
    setError('')
    setOk('')
    try {
      await authService.unbindWechat()
      setWxBound(false)
      setOk('微信解绑成功')
    } catch (err) {
      setError(err instanceof Error ? err.message : '微信解绑失败')
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
          <input className="input" value={phone} onChange={(e) => setPhone(e.target.value)} disabled />
        </label>
        <label>
          <div>邮箱</div>
          <input className="input" value={email} onChange={(e) => setEmail(e.target.value)} disabled={loading || savingProfile} />
        </label>
        <label>
          <div>角色</div>
          <input className="input" value={role} disabled />
        </label>
        <div className="row" style={{ alignItems: 'flex-end' }}>
          <button className="btn primary" type="submit" disabled={loading || savingProfile}>
            {savingProfile ? '保存中...' : '保存'}
          </button>
        </div>
      </form>
      <form className="section-card grid cols-2" style={{ marginTop: 12 }} onSubmit={onChangePassword}>
        <label>
          <div>当前密码</div>
          <input
            className="input"
            type="password"
            value={oldPassword}
            onChange={(e) => setOldPassword(e.target.value)}
            autoComplete="current-password"
          />
        </label>
        <label>
          <div>新密码</div>
          <input
            className="input"
            type="password"
            value={newPassword}
            onChange={(e) => setNewPassword(e.target.value)}
            autoComplete="new-password"
          />
        </label>
        <label>
          <div>确认新密码</div>
          <input
            className="input"
            type="password"
            value={confirmPassword}
            onChange={(e) => setConfirmPassword(e.target.value)}
            autoComplete="new-password"
          />
        </label>
        <div className="row" style={{ alignItems: 'flex-end' }}>
          <button className="btn primary" type="submit" disabled={changingPassword}>
            {changingPassword ? '提交中...' : '修改密码'}
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
        {wxBound ? (
          <button className="btn danger" type="button" onClick={onUnbindWechat}>
            解绑微信
          </button>
        ) : null}
      </form>
      {error ? <div className="alert">{error}</div> : null}
      {ok ? <div style={{ color: '#166534', marginTop: 10 }}>{ok}</div> : null}
    </AppShell>
  )
}
