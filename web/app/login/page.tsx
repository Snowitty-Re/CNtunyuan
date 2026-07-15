'use client'

import { LockOutlined, UserOutlined, WechatOutlined } from '@ant-design/icons'
import { Alert, App, Button, Card, Form, Input, Space, Tabs, Typography } from 'antd'
import { useRouter } from 'next/navigation'
import { useEffect, useMemo, useState } from 'react'
import { useAuth } from '@/components/providers/AuthProvider'
import { SafeImage } from '@/components/shared/SafeImage'
import { useSiteBrand } from '@/hooks/useSiteBrand'
import { saveAuth } from '@/lib/auth'
import { authService } from '@/services/auth'
import { systemService } from '@/services/system'

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
  const brand = useSiteBrand()
  const { message } = App.useApp()
  const { setUser } = useAuth()
  const [mode, setMode] = useState<'password' | 'wechat'>('password')
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')
  const [oauthCode, setOauthCode] = useState('')
  const [oauthState, setOauthState] = useState('')
  const [allowWechatLogin, setAllowWechatLogin] = useState(false)

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

  useEffect(() => {
    ;(async () => {
      try {
        const bootstrap = await systemService.bootstrapStatus()
        if (bootstrap && bootstrap.initialized === false) {
          router.replace('/init')
          return
        }
        const site = bootstrap?.site || {}
        const enabled =
          site.enable_wechat_login_web !== undefined
            ? Boolean(site.enable_wechat_login_web)
            : Boolean(site.enable_wechat_login)
        setAllowWechatLogin(enabled)
        if (enabled) setMode('wechat')
      } catch {
        // ignore
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
      try {
        const data = await authService.wechatWebLogin(oauthCode)
        saveAuth(data)
        setUser(data.user)
        message.success('登录成功')
        router.replace('/dashboard')
      } catch (err) {
        setError(err instanceof Error ? err.message : '微信登录失败')
      } finally {
        setLoading(false)
      }
    })()
  }, [oauthCode, router, setUser, message])

  useEffect(() => {
    if (!allowWechatLogin || mode !== 'wechat' || oauthCode) return
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

  async function onPassword(values: { username: string; password: string }) {
    setLoading(true)
    setError('')
    try {
      const data = await authService.login(values.username.trim(), values.password)
      saveAuth(data)
      setUser(data.user)
      message.success('登录成功')
      router.replace('/dashboard')
    } catch (err) {
      setError(err instanceof Error ? err.message : '登录失败')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div
      style={{
        minHeight: '100vh',
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
        padding: 24,
        background:
          'radial-gradient(circle at 12% 8%, rgba(255, 193, 117, 0.25) 0, transparent 30%), linear-gradient(180deg, #fff9f3 0%, #fff6ee 100%)',
      }}
    >
      <Card style={{ width: 420, maxWidth: '100%', boxShadow: '0 12px 40px rgba(120,70,20,0.12)' }}>
        <Space direction="vertical" size="middle" style={{ width: '100%', textAlign: 'center' }}>
          {brand.logoUrl ? (
            <SafeImage src={brand.logoUrl} alt={brand.orgName} width={64} height={64} style={{ borderRadius: 14, margin: '0 auto' }} />
          ) : (
            <div
              style={{
                width: 64,
                height: 64,
                margin: '0 auto',
                borderRadius: 14,
                background: '#d97706',
                color: '#fff',
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'center',
                fontSize: 28,
                fontWeight: 700,
              }}
            >
              {brand.orgName.slice(0, 1)}
            </div>
          )}
          <div>
            <Typography.Title level={3} style={{ margin: 0 }}>
              {brand.orgName}
            </Typography.Title>
            <Typography.Text type="secondary">{brand.subtitle || '志愿者协作管理端'}</Typography.Text>
          </div>

          {error ? <Alert type="error" showIcon message={error} /> : null}

          <Tabs
            activeKey={mode}
            onChange={(k) => setMode(k as 'password' | 'wechat')}
            items={[
              {
                key: 'password',
                label: '账号登录',
                children: (
                  <Form layout="vertical" onFinish={onPassword} requiredMark={false}>
                    <Form.Item name="username" label="手机号 / 用户名" rules={[{ required: true, message: '请输入账号' }]}>
                      <Input size="large" prefix={<UserOutlined />} placeholder="请输入账号" autoComplete="username" />
                    </Form.Item>
                    <Form.Item name="password" label="密码" rules={[{ required: true, message: '请输入密码' }]}>
                      <Input.Password size="large" prefix={<LockOutlined />} placeholder="请输入密码" autoComplete="current-password" />
                    </Form.Item>
                    <Button type="primary" htmlType="submit" size="large" block loading={loading}>
                      登录
                    </Button>
                  </Form>
                ),
              },
              ...(allowWechatLogin
                ? [
                    {
                      key: 'wechat',
                      label: (
                        <span>
                          <WechatOutlined /> 微信扫码
                        </span>
                      ),
                      children: (
                        <div>
                          <div id="wechat-scan-container" style={{ minHeight: 280, display: 'flex', justifyContent: 'center' }} />
                          {loading ? <Typography.Text type="secondary">正在完成微信登录…</Typography.Text> : null}
                        </div>
                      ),
                    },
                  ]
                : []),
            ]}
          />
        </Space>
      </Card>
    </div>
  )
}
