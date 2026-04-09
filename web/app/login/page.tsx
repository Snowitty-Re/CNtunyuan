'use client'

import { FormEvent, useEffect, useMemo, useState } from 'react'
import { useRouter } from 'next/navigation'
import { authService } from '@/services/auth'
import { systemService } from '@/services/system'
import { saveAuth } from '@/lib/auth'
import { SafeImage } from '@/components/shared/SafeImage'
import { useSiteBrand } from '@/hooks/useSiteBrand'

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

const highlights = [
  { value: '案件协同', label: '从登记、线索到团圆闭环' },
  { value: '任务审批', label: '跟进记录、评论、审批全程留痕' },
  { value: '方言采集', label: '分卡片批次录音，提升识别效率' },
]

export default function LoginPage() {
  const router = useRouter()
  const brand = useSiteBrand()
  const [mode, setMode] = useState<'password' | 'wechat-scan'>('password')
  const [username, setUsername] = useState('13800138000')
  const [password, setPassword] = useState('admin123')
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')
  const [tips, setTips] = useState('')
  const [oauthCode, setOauthCode] = useState('')
  const [oauthState, setOauthState] = useState('')
  const [allowWechatLogin, setAllowWechatLogin] = useState(true)

  const wechatAppID = process.env.NEXT_PUBLIC_WECHAT_WEB_APP_ID || ''
  const redirectURI = useMemo(() => {
    const explicit = process.env.NEXT_PUBLIC_WECHAT_WEB_REDIRECT_URI
    if (explicit) return explicit
    if (typeof window !== 'undefined') return `${window.location.origin}/login`
    return ''
  }, [])
  const loginState = useMemo(() => `web_${Date.now()}`, [])

  useEffect(() => {
    document.title = `${brand.orgName} 登录`
  }, [brand.orgName])

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
    ;(async () => {
      try {
        const bootstrap = await systemService.bootstrapStatus()
        if (bootstrap && bootstrap.initialized === false) {
          router.replace('/init')
          return
        }
        const site = bootstrap?.site || {}
        const enabled = site.enable_wechat_login_web !== undefined
          ? Boolean(site.enable_wechat_login_web)
          : Boolean(site.enable_wechat_login)
        setAllowWechatLogin(enabled)
        setMode(enabled ? 'wechat-scan' : 'password')
      } catch {
        // ignore bootstrap status errors on login page
      }
    })()
  }, [router])

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
    if (!allowWechatLogin || mode !== 'wechat-scan' || oauthCode) return
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
  }, [allowWechatLogin, mode, wechatAppID, redirectURI, oauthCode, oauthState, loginState])

  return (
    <div className="auth-page">
      <div className="auth-hero">
        <div className="auth-hero-badge">助力团圆协作中枢</div>
        <div className="auth-brand-lockup">
          {brand.logoUrl ? <SafeImage className="auth-brand-logo" src={brand.logoUrl} alt={brand.orgName} width={72} height={72} /> : <span className="auth-brand-mark">{brand.orgName.slice(0, 1)}</span>}
        </div>
        <h1 className="auth-hero-title">{brand.orgName}</h1>
        <p className="auth-hero-subtitle">{brand.subtitle}</p>
        <div className="auth-highlight-grid">
          {highlights.map((item) => (
            <div className="auth-highlight" key={item.value}>
              <b>{item.value}</b>
              <span>{item.label}</span>
            </div>
          ))}
        </div>
        <div className="auth-hero-panel">
          <div className="auth-hero-panel-title">适用场景</div>
          <ul className="auth-hero-list">
            <li>志愿者组织管理与跨层级协作</li>
            <li>案件、任务、方言线索统一沉淀</li>
            <li>跟进审批、附件留痕、服务监控</li>
          </ul>
        </div>
      </div>

      <div className="auth-card">
        <div className="auth-card-header">
          <div>
            <div className="auth-card-eyebrow">Web Console</div>
            <h2 className="auth-card-title">进入管理后台</h2>
            <p className="auth-card-desc">使用账号密码或微信扫码进入工作台。</p>
          </div>
          {allowWechatLogin ? (
            <div className="auth-switch">
              <button className={`btn ${mode === 'wechat-scan' ? 'primary' : ''}`} type="button" onClick={() => setMode('wechat-scan')}>
                微信扫码
              </button>
              <button className={`btn ${mode === 'password' ? 'primary' : ''}`} type="button" onClick={() => setMode('password')}>
                账号密码
              </button>
            </div>
          ) : (
            <div className="hint">当前系统已关闭微信登录</div>
          )}
        </div>

        {allowWechatLogin && mode === 'wechat-scan' ? (
          <div className="auth-qr-shell">
            <div className="auth-qr-card">
              <div id="wechat-scan-container" />
            </div>
            <div className="hint" style={{ textAlign: 'center' }}>请使用微信扫描二维码并确认登录</div>
          </div>
        ) : (
          <form className="auth-form" onSubmit={onPasswordSubmit}>
            <label>
              <div className="field-label">账号（手机号）</div>
              <input className="input auth-input" value={username} onChange={(e) => setUsername(e.target.value)} />
            </label>
            <label>
              <div className="field-label">密码</div>
              <input className="input auth-input" type="password" value={password} onChange={(e) => setPassword(e.target.value)} />
            </label>
            <button className="btn primary auth-submit" type="submit" disabled={loading}>
              {loading ? '正在登录...' : '登录后台'}
            </button>
          </form>
        )}

        {error ? <div className="alert">{error}</div> : null}
        {tips ? <div style={{ color: '#166534', marginTop: 10 }}>{tips}</div> : null}
      </div>
    </div>
  )
}
