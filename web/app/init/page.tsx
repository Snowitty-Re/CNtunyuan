'use client'

import { FormEvent, useEffect, useState } from 'react'
import { useRouter } from 'next/navigation'
import { systemService } from '@/services/system'

type InitStatus = {
  initialized: boolean
  startup_mode?: string
  checks?: {
    database_connected?: boolean
    schema_ready?: boolean
    settings_storage?: string
    health_status?: string
  }
  database?: {
    type?: string
    host?: string
    port?: number
    user?: string
    database?: string
    ssl_mode?: string
    timezone?: string
  }
  site?: {
    domain?: string
    cors_origins?: string
    default_org_name?: string
    default_org_code?: string
    enable_register?: boolean
    enable_wechat_login?: boolean
    enable_wechat_login_web?: boolean
    enable_wechat_login_mini_program?: boolean
    enable_sms_login?: boolean
  }
}

export default function BootstrapInitPage() {
  const router = useRouter()
  const [loading, setLoading] = useState(true)
  const [submitting, setSubmitting] = useState(false)
  const [testingDB, setTestingDB] = useState(false)
  const [error, setError] = useState('')
  const [success, setSuccess] = useState('')
  const [status, setStatus] = useState<InitStatus | null>(null)

  const [db, setDB] = useState({
    type: 'postgres',
    host: '127.0.0.1',
    port: 5432,
    user: 'postgres',
    password: '',
    database: 'cntuanyuan',
    ssl_mode: 'disable',
    timezone: 'Asia/Shanghai',
  })
  const [site, setSite] = useState({
    domain: '',
    cors_origins: '',
    default_org_name: '默认组织',
    default_org_code: 'DEFAULT',
    enable_register: true,
    enable_wechat_login: false,
    enable_wechat_login_web: false,
    enable_wechat_login_mini_program: true,
    enable_sms_login: false,
  })
  const [admin, setAdmin] = useState({
    nickname: '超级管理员',
    phone: '',
    password: '',
    email: '',
  })

  async function loadStatus() {
    setLoading(true)
    setError('')
    try {
      const data = await systemService.bootstrapStatus()
      setStatus(data)
      if (data?.initialized) {
        router.replace('/login')
        return
      }
      if (data?.database) {
        setDB((prev) => ({
          ...prev,
          type: data.database.type || prev.type,
          host: data.database.host || prev.host,
          port: Number(data.database.port || prev.port),
          user: data.database.user || prev.user,
          database: data.database.database || prev.database,
          ssl_mode: data.database.ssl_mode || prev.ssl_mode,
          timezone: data.database.timezone || prev.timezone,
        }))
      }
      if (data?.site) {
        setSite((prev) => ({
          ...prev,
          domain: data.site.domain || prev.domain,
          cors_origins: data.site.cors_origins || prev.cors_origins,
          default_org_name: data.site.default_org_name || prev.default_org_name,
          default_org_code: data.site.default_org_code || prev.default_org_code,
          enable_register: Boolean(data.site.enable_register),
          enable_wechat_login: Boolean(data.site.enable_wechat_login),
          enable_wechat_login_web:
            data.site.enable_wechat_login_web !== undefined
              ? Boolean(data.site.enable_wechat_login_web)
              : Boolean(data.site.enable_wechat_login),
          enable_wechat_login_mini_program:
            data.site.enable_wechat_login_mini_program !== undefined
              ? Boolean(data.site.enable_wechat_login_mini_program)
              : Boolean(data.site.enable_wechat_login),
          enable_sms_login: Boolean(data.site.enable_sms_login),
        }))
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : '加载初始化状态失败')
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    loadStatus()
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  async function testDatabase() {
    setTestingDB(true)
    setError('')
    setSuccess('')
    try {
      await systemService.validateBootstrapDatabase(db)
      setSuccess('数据库连接验证通过')
    } catch (err) {
      setError(err instanceof Error ? err.message : '数据库连接验证失败')
    } finally {
      setTestingDB(false)
    }
  }

  async function submitInitialize(e: FormEvent) {
    e.preventDefault()
    setSubmitting(true)
    setError('')
    setSuccess('')
    try {
      await systemService.bootstrapInitialize({
        database: db,
        site,
        super_admin: admin,
      })
      setSuccess('系统初始化已完成，正在跳转登录页...')
      setTimeout(() => router.replace('/login'), 1000)
    } catch (err) {
      setError(err instanceof Error ? err.message : '初始化失败')
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <div className="login-page" style={{ padding: 24 }}>
      <form className="login-card" onSubmit={submitInitialize} style={{ maxWidth: 980 }}>
        <h1 className="login-title">系统首次初始化</h1>
        <p className="login-subtitle">完成数据库检测、站点设置与超级管理员创建后即可进入后台</p>

        <div className="grid cols-3" style={{ marginBottom: 12 }}>
          <div className="section-card">
            <b>数据库连接</b>
            <div className="hint" style={{ marginTop: 8 }}>
              {status?.checks?.database_connected ? '已连接' : '未连接'}
            </div>
          </div>
          <div className="section-card">
            <b>Schema 状态</b>
            <div className="hint" style={{ marginTop: 8 }}>
              {status?.checks?.schema_ready ? '已建表' : '待初始化'}
            </div>
          </div>
          <div className="section-card">
            <b>当前启动模式</b>
            <div className="hint" style={{ marginTop: 8 }}>
              {status?.startup_mode === 'full' ? '完整模式' : '初始化模式'}
            </div>
          </div>
        </div>

        <div className="section-card" style={{ marginBottom: 12 }}>
          <b>初始化说明</b>
          <div className="hint" style={{ marginTop: 8, lineHeight: 1.7 }}>
            后端已支持无 `config.yaml` 首次启动。此页面会完成数据库检测、自动建表、站点基础设置与超级管理员创建。
            初始化成功后，请重启后端，使服务从“初始化模式”切换到“完整模式”。
          </div>
        </div>

        <div className="form-section">
          <h3 className="form-section-title">1) 数据库设置（含连通性检测与自动建表）</h3>
          <div className="grid cols-3">
            <label>
              <span className="field-label">类型</span>
              <select className="select" value={db.type} onChange={(e) => setDB((p) => ({ ...p, type: e.target.value }))}>
                <option value="postgres">postgres</option>
                <option value="mysql">mysql</option>
              </select>
            </label>
            <label>
              <span className="field-label">主机</span>
              <input className="input" value={db.host} onChange={(e) => setDB((p) => ({ ...p, host: e.target.value }))} />
            </label>
            <label>
              <span className="field-label">端口</span>
              <input className="input" value={db.port} onChange={(e) => setDB((p) => ({ ...p, port: Number(e.target.value || 0) }))} />
            </label>
            <label>
              <span className="field-label">用户</span>
              <input className="input" value={db.user} onChange={(e) => setDB((p) => ({ ...p, user: e.target.value }))} />
            </label>
            <label>
              <span className="field-label">密码</span>
              <input className="input" type="password" value={db.password} onChange={(e) => setDB((p) => ({ ...p, password: e.target.value }))} />
            </label>
            <label>
              <span className="field-label">数据库名</span>
              <input className="input" value={db.database} onChange={(e) => setDB((p) => ({ ...p, database: e.target.value }))} />
            </label>
          </div>
          <div className="row wrap" style={{ marginTop: 10 }}>
            <button className="btn" type="button" onClick={testDatabase} disabled={testingDB || loading}>
              {testingDB ? '检测中...' : '检测数据库连接'}
            </button>
          </div>
        </div>

        <div className="form-section" style={{ marginTop: 12 }}>
          <h3 className="form-section-title">2) 站点设置</h3>
          <div className="grid cols-2">
            <label>
              <span className="field-label">站点域名</span>
              <input className="input" value={site.domain} onChange={(e) => setSite((p) => ({ ...p, domain: e.target.value }))} placeholder="https://example.com" />
            </label>
            <label>
              <span className="field-label">CORS Origins</span>
              <input className="input" value={site.cors_origins} onChange={(e) => setSite((p) => ({ ...p, cors_origins: e.target.value }))} />
            </label>
            <label>
              <span className="field-label">默认组织名称</span>
              <input className="input" value={site.default_org_name} onChange={(e) => setSite((p) => ({ ...p, default_org_name: e.target.value }))} />
            </label>
            <label>
              <span className="field-label">默认组织编码</span>
              <input className="input" value={site.default_org_code} onChange={(e) => setSite((p) => ({ ...p, default_org_code: e.target.value }))} />
            </label>
          </div>
          <div className="row wrap" style={{ marginTop: 10 }}>
            <label className="row">
              <input type="checkbox" checked={site.enable_register} onChange={(e) => setSite((p) => ({ ...p, enable_register: e.target.checked }))} />
              允许注册
            </label>
            <label className="row">
              <input type="checkbox" checked={site.enable_wechat_login} onChange={(e) => setSite((p) => ({ ...p, enable_wechat_login: e.target.checked }))} />
              微信登录兼容总开关
            </label>
            <label className="row">
              <input
                type="checkbox"
                checked={site.enable_wechat_login_web}
                onChange={(e) => setSite((p) => ({ ...p, enable_wechat_login_web: e.target.checked }))}
              />
              允许微信登录(Web)
            </label>
            <label className="row">
              <input
                type="checkbox"
                checked={site.enable_wechat_login_mini_program}
                onChange={(e) => setSite((p) => ({ ...p, enable_wechat_login_mini_program: e.target.checked }))}
              />
              允许微信登录(小程序)
            </label>
            <label className="row">
              <input type="checkbox" checked={site.enable_sms_login} onChange={(e) => setSite((p) => ({ ...p, enable_sms_login: e.target.checked }))} />
              允许短信验证
            </label>
          </div>
        </div>

        <div className="form-section" style={{ marginTop: 12 }}>
          <h3 className="form-section-title">3) 创建超级管理员</h3>
          <div className="grid cols-2">
            <label>
              <span className="field-label">管理员昵称</span>
              <input className="input" value={admin.nickname} onChange={(e) => setAdmin((p) => ({ ...p, nickname: e.target.value }))} />
            </label>
            <label>
              <span className="field-label">手机号</span>
              <input className="input" value={admin.phone} onChange={(e) => setAdmin((p) => ({ ...p, phone: e.target.value }))} required />
            </label>
            <label>
              <span className="field-label">密码（至少8位）</span>
              <input className="input" type="password" value={admin.password} onChange={(e) => setAdmin((p) => ({ ...p, password: e.target.value }))} required />
            </label>
            <label>
              <span className="field-label">邮箱（可选）</span>
              <input className="input" value={admin.email} onChange={(e) => setAdmin((p) => ({ ...p, email: e.target.value }))} />
            </label>
          </div>
        </div>

        <div className="row wrap" style={{ marginTop: 14 }}>
          <button className="btn primary" type="submit" disabled={submitting || loading}>
            {submitting ? '初始化中...' : '执行初始化'}
          </button>
          <button className="btn" type="button" onClick={loadStatus} disabled={loading}>
            {loading ? '刷新中...' : '刷新状态'}
          </button>
        </div>

        {error ? <div className="alert">{error}</div> : null}
        {success ? <div style={{ marginTop: 10, color: '#166534' }}>{success}</div> : null}
      </form>
    </div>
  )
}
