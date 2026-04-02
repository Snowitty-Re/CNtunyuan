'use client'

import { FormEvent, useState } from 'react'
import { useRouter } from 'next/navigation'
import { authService } from '@/services/auth'
import { saveAuth } from '@/lib/auth'

export default function LoginPage() {
  const router = useRouter()
  const [username, setUsername] = useState('13800138000')
  const [password, setPassword] = useState('admin123')
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')

  async function onSubmit(e: FormEvent) {
    e.preventDefault()
    setLoading(true)
    setError('')
    try {
      const data = await authService.login(username.trim(), password)
      saveAuth(data)
      router.replace('/dashboard')
    } catch (err) {
      setError(err instanceof Error ? err.message : '登录失败')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="login-page">
      <form className="login-card" onSubmit={onSubmit}>
        <h1 className="login-title">助力团圆 Web</h1>
        <p className="login-subtitle">专业协作管理端</p>
        <div className="grid">
          <label>
            <div>账号（手机号）</div>
            <input className="input" value={username} onChange={(e) => setUsername(e.target.value)} />
          </label>
          <label>
            <div>密码</div>
            <input
              className="input"
              type="password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
            />
          </label>
          <button className="btn primary" type="submit" disabled={loading}>
            {loading ? '登录中...' : '登录'}
          </button>
        </div>
        {error ? <div className="alert">{error}</div> : null}
      </form>
    </div>
  )
}
