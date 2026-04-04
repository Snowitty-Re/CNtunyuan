'use client'

import { FormEvent, useEffect, useMemo, useState } from 'react'
import { useRouter } from 'next/navigation'
import { authService } from '@/services/auth'
import { saveAuth } from '@/lib/auth'

declare global {
  interface Window {
    WxLogin?: new (options: {
      self_redirect?: boolean
      id: string
      appid: string
      scope: string
      redirect_uri: string
      state: string
      style?: string
      href?: string
    }) => unknown
  }
}

function appendScript(src: string): Promise<void> {
  return new Promise((resolve, reject) => {
    const exists = document.querySelector(`script[src="${src}"]`)
    if (exists) {
      resolve()
      return
    }
    const script = document.createElement('script')
    script.src = src
    script.async = true
    script.onload = () => resolve()
    script.onerror = () => reject(new Error('微信扫码脚本加载失败'))
    document.body.appendChild(script)
  })
}

export default function LoginPage() {
  const router = useRouter()
  const [mode, setMode] = useState<'password' | 'wechat-scan'>('wechat-scan')
  const [username, setUsername] = useState('13800138000')
  const [password, setPassword] = useState('admin123')
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')
  const [tips, setTips] = useState('')
  const [oauthCode, setOauthCode] = useState('')
  const [oauthState, setOauthState] = useState('')

  const wechatAppID = process.env.NEXT_PUBLIC_WECHAT_WEB_APP_ID || ''
  const redirectURI = useMemo(() => {
    const explicit = process.env.NEXT_PUBLIC_WECHAT_WEB_REDIRECT_URI
    if (explicit) return explicit
    if (typeof window !== 'undefined') return `${window.location.origin}/login`
    return ''
  }, [])
  const loginState = useMemo(() => `web_${Date.now()}`, [])

  async function onPasswordSubmit(e: FormEvent) {
    e.preventDefault()
    setLoading(true)
    setError('')
    setTips('')
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

  useEffect(() => {
    if (typeof window === 'undefined') return
    const params = new URLSearchParams(window.location.search)
    setOauthCode(params.get('code') || '')
    setOauthState(params.get('state') || '')
  }, [])

  useEffect(() => {
    if (!oauthCode) return
    ;(async () => {
      setLoading(true)
      setError('')
      setTips('正在完成微信登录...')
      try {
        const data = await authService.wechatWebLogin(oauthCode)
        saveAuth(data)
        router.replace('/dashboard')
      } catch (err) {
        setError(err instanceof Error ? err.message : '微信登录失败')
      } finally {
        setLoading(false)
      }
    })()
  }, [oauthCode, router])

  useEffect(() => {
    if (mode !== 'wechat-scan') return
    if (oauthCode) return
    if (!wechatAppID) {
      setError('缺少 NEXT_PUBLIC_WECHAT_WEB_APP_ID，无法启用微信扫码登录')
      return
    }
    if (!redirectURI) return
    ;(async () => {
      try {
        await appendScript('https://res.wx.qq.com/connect/zh_CN/htmledition/js/wxLogin.js')
        const container = document.getElementById('wechat-scan-container')
        if (container) container.innerHTML = ''
        if (!window.WxLogin) throw new Error('微信扫码组件初始化失败')
        new window.WxLogin({
          self_redirect: true,
          id: 'wechat-scan-container',
          appid: wechatAppID,
          scope: 'snsapi_login',
          redirect_uri: encodeURIComponent(redirectURI),
          state: oauthState || loginState,
          style: 'black',
        })
      } catch (err) {
        setError(err instanceof Error ? err.message : '微信扫码初始化失败')
      }
    })()
  }, [mode, wechatAppID, redirectURI, oauthCode, oauthState, loginState])

  return (
    <div className="login-page">
      <form className="login-card" onSubmit={onPasswordSubmit}>
        <h1 className="login-title">助力团圆 Web</h1>
        <p className="login-subtitle">让每一条线索都更快抵达家人</p>
        <div className="row wrap" style={{ marginBottom: 8 }}>
          <button className={`btn ${mode === 'wechat-scan' ? 'primary' : ''}`} type="button" onClick={() => setMode('wechat-scan')}>
            微信扫码登录
          </button>
          <button className={`btn ${mode === 'password' ? 'primary' : ''}`} type="button" onClick={() => setMode('password')}>
            账号密码登录
          </button>
        </div>

        {mode === 'wechat-scan' ? (
          <div className="section-card">
            <div style={{ textAlign: 'center' }}>
              <div id="wechat-scan-container" />
              <div className="hint" style={{ marginTop: 8 }}>
                请使用微信扫描二维码并确认登录
              </div>
            </div>
          </div>
        ) : (
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
              {loading ? '提交中...' : '登录'}
            </button>
          </div>
        )}

        {error ? <div className="alert">{error}</div> : null}
        {tips ? <div style={{ color: '#166534', marginTop: 10 }}>{tips}</div> : null}
      </form>
    </div>
  )
}
