'use client'

import { ChangeEvent, FormEvent, useEffect, useMemo, useState } from 'react'
import { AppShell } from '@/components/layout/AppShell'
import { ModuleHeader } from '@/components/shared/ModuleHeader'
import { NoticeBar, type Notice } from '@/components/shared/NoticeBar'
import { PageState } from '@/components/shared/PageState'
import { useAuthGuard } from '@/hooks/useAuthGuard'
import { ACTIONS, hasPermission } from '@/lib/rbac'
import { systemService } from '@/services/system'

type FieldType = 'text' | 'number' | 'password' | 'bool'
type FieldDef = { key: string; label: string; type: FieldType; placeholder?: string }
type SectionDef = { title: string; fields: FieldDef[] }

const sections: SectionDef[] = [
  {
    title: 'server',
    fields: [
      { key: 'server.port', label: '端口', type: 'number' },
      { key: 'server.mode', label: '模式(debug/release)', type: 'text' },
      { key: 'server.domain', label: '服务域名', type: 'text' },
      { key: 'server.read_timeout', label: '读超时(秒)', type: 'number' },
      { key: 'server.write_timeout', label: '写超时(秒)', type: 'number' },
      { key: 'server.max_header_bytes', label: '最大请求头字节', type: 'number' },
      { key: 'server.cors_origins', label: 'CORS允许域名(逗号分隔)', type: 'text' },
    ],
  },
  {
    title: 'database',
    fields: [
      { key: 'database.type', label: '数据库类型', type: 'text' },
      { key: 'database.host', label: '主机', type: 'text' },
      { key: 'database.port', label: '端口', type: 'number' },
      { key: 'database.user', label: '用户名', type: 'text' },
      { key: 'database.password', label: '密码', type: 'password' },
      { key: 'database.database', label: '数据库名', type: 'text' },
      { key: 'database.ssl_mode', label: 'SSL模式', type: 'text' },
      { key: 'database.timezone', label: '时区', type: 'text' },
      { key: 'database.charset', label: '字符集', type: 'text' },
      { key: 'database.max_idle_conns', label: '最大空闲连接', type: 'number' },
      { key: 'database.max_open_conns', label: '最大打开连接', type: 'number' },
      { key: 'database.conn_max_lifetime', label: '连接生命周期(秒)', type: 'number' },
    ],
  },
  {
    title: 'redis',
    fields: [
      { key: 'redis.host', label: '主机', type: 'text' },
      { key: 'redis.port', label: '端口', type: 'number' },
      { key: 'redis.password', label: '密码', type: 'password' },
      { key: 'redis.db', label: 'DB', type: 'number' },
      { key: 'redis.pool_size', label: '连接池大小', type: 'number' },
      { key: 'redis.min_idle_conns', label: '最小空闲连接', type: 'number' },
    ],
  },
  {
    title: 'jwt',
    fields: [
      { key: 'jwt.secret', label: 'JWT密钥', type: 'password' },
      { key: 'jwt.expire_time', label: '访问令牌过期(秒)', type: 'number' },
      { key: 'jwt.refresh_time', label: '刷新令牌过期(秒)', type: 'number' },
    ],
  },
  {
    title: 'wechat',
    fields: [
      { key: 'wechat.app_id', label: 'AppID', type: 'text' },
      { key: 'wechat.app_secret', label: 'AppSecret', type: 'password' },
      { key: 'wechat.enable_login', label: '启用微信登录', type: 'bool' },
      { key: 'wechat.mch_id', label: '商户号', type: 'text' },
      { key: 'wechat.api_key', label: '支付API密钥', type: 'password' },
      { key: 'wechat.notify_url', label: '支付回调地址', type: 'text' },
    ],
  },
  {
    title: 'storage',
    fields: [
      { key: 'storage.type', label: '存储类型(local/oss/cos)', type: 'text' },
      { key: 'storage.local_path', label: '本地路径', type: 'text' },
      { key: 'storage.base_url', label: '访问基地址', type: 'text' },
      { key: 'storage.max_file_size', label: '最大文件大小(字节)', type: 'number' },
      { key: 'storage.allowed_types', label: '允许格式', type: 'text' },
      { key: 'storage.oss_access_key_id', label: 'OSS AccessKeyId', type: 'text' },
      { key: 'storage.oss_access_key_secret', label: 'OSS AccessKeySecret', type: 'password' },
      { key: 'storage.oss_endpoint', label: 'OSS Endpoint', type: 'text' },
      { key: 'storage.oss_bucket', label: 'OSS Bucket', type: 'text' },
      { key: 'storage.oss_region', label: 'OSS Region', type: 'text' },
      { key: 'storage.cos_secret_id', label: 'COS SecretId', type: 'text' },
      { key: 'storage.cos_secret_key', label: 'COS SecretKey', type: 'password' },
      { key: 'storage.cos_bucket', label: 'COS Bucket', type: 'text' },
      { key: 'storage.cos_region', label: 'COS Region', type: 'text' },
    ],
  },
  {
    title: 'sms',
    fields: [
      { key: 'sms.provider', label: '短信服务商', type: 'text' },
      { key: 'sms.sign_name', label: '短信签名', type: 'text' },
      { key: 'sms.dev_mode', label: '开发模式(不真实发送)', type: 'bool' },
      { key: 'sms.code_expiry', label: '验证码过期(秒)', type: 'number' },
      { key: 'sms.template_verify_code', label: '注册验证码模板ID', type: 'text' },
      { key: 'sms.template_reset_password', label: '重置密码模板ID', type: 'text' },
      { key: 'sms.template_change_phone', label: '换绑手机号模板ID', type: 'text' },
      { key: 'sms.aliyun_access_key_id', label: '阿里云AccessKeyId', type: 'text' },
      { key: 'sms.aliyun_access_key_secret', label: '阿里云AccessKeySecret', type: 'password' },
      { key: 'sms.tencent_secret_id', label: '腾讯SecretId', type: 'text' },
      { key: 'sms.tencent_secret_key', label: '腾讯SecretKey', type: 'password' },
      { key: 'sms.tencent_app_id', label: '腾讯AppId', type: 'text' },
    ],
  },
  {
    title: 'email',
    fields: [
      { key: 'email.enabled', label: '启用邮件', type: 'bool' },
      { key: 'email.smtp_host', label: 'SMTP Host', type: 'text' },
      { key: 'email.smtp_port', label: 'SMTP Port', type: 'number' },
      { key: 'email.smtp_user', label: 'SMTP 用户名', type: 'text' },
      { key: 'email.smtp_password', label: 'SMTP 密码', type: 'password' },
      { key: 'email.from_name', label: '发件人名称', type: 'text' },
      { key: 'email.use_tls', label: '启用TLS', type: 'bool' },
    ],
  },
  {
    title: 'map',
    fields: [
      { key: 'map.provider', label: '地图服务商', type: 'text' },
      { key: 'map.key', label: '默认Key', type: 'password' },
      { key: 'map.tencent_key', label: '腾讯Key', type: 'password' },
      { key: 'map.amap_key', label: '高德Key', type: 'password' },
      { key: 'map.baidu_key', label: '百度Key', type: 'password' },
    ],
  },
  {
    title: 'log',
    fields: [
      { key: 'log.level', label: '日志级别', type: 'text' },
      { key: 'log.format', label: '日志格式', type: 'text' },
      { key: 'log.output_path', label: '输出路径', type: 'text' },
      { key: 'log.file_name', label: '文件名', type: 'text' },
      { key: 'log.max_size', label: '单文件大小MB', type: 'number' },
      { key: 'log.max_backups', label: '最大备份数', type: 'number' },
      { key: 'log.max_age', label: '保留天数', type: 'number' },
      { key: 'log.compress', label: '启用压缩', type: 'bool' },
    ],
  },
  {
    title: 'notification',
    fields: [
      { key: 'notification.push_enabled', label: '启用推送', type: 'bool' },
      { key: 'notification.getui_app_id', label: '个推AppID', type: 'text' },
      { key: 'notification.getui_app_key', label: '个推AppKey', type: 'password' },
      { key: 'notification.getui_master_secret', label: '个推MasterSecret', type: 'password' },
      { key: 'notification.jpush_app_key', label: '极光AppKey', type: 'text' },
      { key: 'notification.jpush_master_secret', label: '极光MasterSecret', type: 'password' },
    ],
  },
  {
    title: 'system',
    fields: [
      { key: 'system.default_org_name', label: '默认组织名称', type: 'text' },
      { key: 'system.default_org_code', label: '默认组织编码', type: 'text' },
      { key: 'system.enable_register', label: '开放注册', type: 'bool' },
      { key: 'system.enable_wechat_login', label: '启用微信登录(兼容总开关)', type: 'bool' },
      { key: 'system.enable_wechat_login_web', label: '启用微信登录(Web)', type: 'bool' },
      { key: 'system.enable_wechat_login_mini_program', label: '启用微信登录(小程序)', type: 'bool' },
      { key: 'system.enable_sms_login', label: '启用短信登录', type: 'bool' },
      { key: 'system.authz_policy_change_requires_approval', label: '权限策略变更需审批', type: 'bool' },
      { key: 'system.authz_policy_change_approval_code', label: '权限策略审批码', type: 'password' },
      { key: 'system.authz_policy_request_expire_hours', label: '权限策略申请过期小时', type: 'number' },
      { key: 'system.admin_ips', label: '管理员IP白名单', type: 'text' },
      { key: 'system.rate_limit', label: '每分钟限流', type: 'number' },
    ],
  },
  {
    title: 'security',
    fields: [
      { key: 'security.max_login_attempts', label: '最大登录失败次数', type: 'number' },
      { key: 'security.lockout_duration', label: '账号锁定时长(秒)', type: 'number' },
    ],
  },
  {
    title: 'backup',
    fields: [
      { key: 'backup.enabled', label: '启用自动备份', type: 'bool' },
      { key: 'backup.backup_dir', label: '备份目录', type: 'text' },
      { key: 'backup.retention', label: '备份保留天数', type: 'number' },
    ],
  },
]

function initialConfig(): Record<string, any> {
  return {
    'server.port': 8080,
    'server.mode': 'release',
    'server.domain': 'http://localhost:8080',
    'server.read_timeout': 30,
    'server.write_timeout': 30,
    'server.max_header_bytes': 1048576,
    'server.cors_origins': 'http://localhost:3000,http://localhost:5173',
    'database.type': 'postgres',
    'database.host': 'localhost',
    'database.port': 5432,
    'database.user': 'postgres',
    'database.password': '',
    'database.database': 'cntuanyuan',
    'database.ssl_mode': 'disable',
    'database.timezone': 'Asia/Shanghai',
    'database.charset': 'UTF8',
    'database.max_idle_conns': 10,
    'database.max_open_conns': 100,
    'database.conn_max_lifetime': 3600,
    'redis.host': 'localhost',
    'redis.port': 6379,
    'redis.password': '',
    'redis.db': 0,
    'redis.pool_size': 10,
    'redis.min_idle_conns': 2,
    'jwt.secret': '',
    'jwt.expire_time': 604800,
    'jwt.refresh_time': 2592000,
    'wechat.app_id': '',
    'wechat.app_secret': '',
    'wechat.enable_login': true,
    'wechat.mch_id': '',
    'wechat.api_key': '',
    'wechat.notify_url': '',
    'storage.type': 'local',
    'storage.local_path': '/abs/path/backend/uploads',
    'storage.base_url': 'http://localhost:8080/uploads',
    'storage.max_file_size': 52428800,
    'storage.allowed_types': 'jpg,jpeg,png,gif,webp,mp4,mp3,wav',
    'storage.oss_access_key_id': '',
    'storage.oss_access_key_secret': '',
    'storage.oss_endpoint': '',
    'storage.oss_bucket': '',
    'storage.oss_region': '',
    'storage.cos_secret_id': '',
    'storage.cos_secret_key': '',
    'storage.cos_bucket': '',
    'storage.cos_region': '',
    'sms.provider': 'aliyun',
    'sms.sign_name': '助力团圆',
    'sms.dev_mode': false,
    'sms.code_expiry': 300,
    'sms.template_verify_code': '',
    'sms.template_reset_password': '',
    'sms.template_change_phone': '',
    'sms.aliyun_access_key_id': '',
    'sms.aliyun_access_key_secret': '',
    'sms.tencent_secret_id': '',
    'sms.tencent_secret_key': '',
    'sms.tencent_app_id': '',
    'email.enabled': false,
    'email.smtp_host': 'smtp.qq.com',
    'email.smtp_port': 587,
    'email.smtp_user': '',
    'email.smtp_password': '',
    'email.from_name': '助力团圆',
    'email.use_tls': true,
    'map.provider': 'tencent',
    'map.key': '',
    'map.tencent_key': '',
    'map.amap_key': '',
    'map.baidu_key': '',
    'log.level': 'info',
    'log.format': 'json',
    'log.output_path': './logs',
    'log.file_name': 'app.log',
    'log.max_size': 100,
    'log.max_backups': 10,
    'log.max_age': 30,
    'log.compress': true,
    'notification.push_enabled': false,
    'notification.getui_app_id': '',
    'notification.getui_app_key': '',
    'notification.getui_master_secret': '',
    'notification.jpush_app_key': '',
    'notification.jpush_master_secret': '',
    'system.default_org_name': '助力团圆志愿者协会',
    'system.default_org_code': 'ROOT',
    'system.enable_register': true,
    'system.enable_wechat_login': false,
    'system.enable_wechat_login_web': false,
    'system.enable_wechat_login_mini_program': false,
    'system.enable_sms_login': false,
    'system.authz_policy_change_requires_approval': false,
    'system.authz_policy_change_approval_code': '',
    'system.authz_policy_request_expire_hours': 72,
    'system.admin_ips': '',
    'system.rate_limit': 100,
    'security.max_login_attempts': 5,
    'security.lockout_duration': 1800,
    'backup.enabled': false,
    'backup.backup_dir': './backups',
    'backup.retention': 7,
  }
}

export default function SiteSettingsPage() {
  const { ready, user } = useAuthGuard()
  const [notice, setNotice] = useState<Notice | null>(null)
  const [config, setConfig] = useState<Record<string, any>>(initialConfig())
  const [loading, setLoading] = useState(false)
  const sectionCount = useMemo(() => sections.reduce((s, x) => s + x.fields.length, 0), [])

  async function loadConfig() {
    setLoading(true)
    try {
      const res = await systemService.getSiteConfig()
      setConfig({ ...initialConfig(), ...(res.config || {}) })
    } catch (err) {
      setNotice({ type: 'error', text: err instanceof Error ? err.message : '加载配置失败' })
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    if (!ready || !hasPermission(user, ACTIONS.USER_MODIFY)) return
    loadConfig()
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [ready, user?.id])

  function setValue(key: string, value: any) {
    setConfig((prev) => ({ ...prev, [key]: value }))
  }

  async function save(e: FormEvent) {
    e.preventDefault()
    setLoading(true)
    try {
      const res = await systemService.updateSiteConfig(config)
      setConfig({ ...initialConfig(), ...(res.config || {}) })
      setNotice({ type: 'success', text: '配置已保存到后端 config.yaml（建议重启后端生效）' })
    } catch (err) {
      setNotice({ type: 'error', text: err instanceof Error ? err.message : '保存配置失败' })
    } finally {
      setLoading(false)
    }
  }

  function resetDefaults() {
    setConfig(initialConfig())
    setNotice({ type: 'info', text: '已恢复默认值，点击保存后生效' })
  }

  function exportJson() {
    const content = JSON.stringify(config, null, 2)
    const blob = new Blob([content], { type: 'application/json;charset=utf-8;' })
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = `site-settings-${Date.now()}.json`
    a.click()
    URL.revokeObjectURL(url)
    setNotice({ type: 'success', text: '已导出设置 JSON' })
  }

  function importJson(e: ChangeEvent<HTMLInputElement>) {
    const file = e.target.files?.[0]
    if (!file) return
    const reader = new FileReader()
    reader.onload = () => {
      try {
        const parsed = JSON.parse(String(reader.result || '{}'))
        if (!parsed || typeof parsed !== 'object') throw new Error('invalid')
        setConfig((prev) => ({ ...prev, ...parsed }))
        setNotice({ type: 'success', text: '设置已导入，请检查后保存' })
      } catch {
        setNotice({ type: 'error', text: '导入失败：JSON 格式不正确' })
      }
    }
    reader.readAsText(file)
  }

  function renderField(f: FieldDef) {
    const value = config[f.key]
    if (f.type === 'bool') {
      return (
        <label key={f.key}>
          <div>{f.label}</div>
          <label><input type="checkbox" checked={Boolean(value)} onChange={(e) => setValue(f.key, e.target.checked)} /> 启用</label>
        </label>
      )
    }
    return (
      <label key={f.key}>
        <div>{f.label}</div>
        <input
          className="input"
          type={f.type === 'number' ? 'number' : f.type === 'password' ? 'password' : 'text'}
          placeholder={f.placeholder || ''}
          value={value ?? ''}
          onChange={(e) => setValue(f.key, f.type === 'number' ? Number(e.target.value || 0) : e.target.value)}
        />
      </label>
    )
  }

  if (!ready) return null
  if (!hasPermission(user, ACTIONS.USER_MODIFY)) {
    return (
      <AppShell>
        <ModuleHeader title="网站设置" desc="对齐 config.yaml 的系统配置中心" />
        <PageState error="当前账号无权限访问该页面（需要 user:modify 权限）" />
      </AppShell>
    )
  }

  return (
    <AppShell>
      <ModuleHeader
        title="网站设置"
        desc={`对齐 config.yaml 的系统配置中心（共 ${sections.length} 组 / ${sectionCount} 项）`}
        right={
          <div className="row wrap">
            <button className="btn" type="button" onClick={exportJson}>
              导出JSON
            </button>
            <label className="btn ghost" style={{ cursor: 'pointer' }}>
              导入JSON
              <input type="file" accept="application/json" style={{ display: 'none' }} onChange={importJson} />
            </label>
          </div>
        }
      />
      <NoticeBar notice={notice} onClose={() => setNotice(null)} />
      <div className="panel" style={{ marginBottom: 12, color: '#7a3e00', background: '#fff7e6', borderColor: '#ffd591' }}>
        当前页面已接入后端配置读写，密钥类字段会以掩码展示。保持掩码不改动即可保留原值。
      </div>
      <form className="grid" onSubmit={save}>
        {sections.map((section) => (
          <div className="section-card" key={section.title}>
            <b>{section.title}</b>
            <div className="grid cols-3" style={{ marginTop: 10 }}>
              {section.fields.map(renderField)}
            </div>
          </div>
        ))}
        <div className="panel row wrap">
          <button className="btn primary" type="submit">
            {loading ? '保存中...' : '保存设置'}
          </button>
          <button className="btn ghost" type="button" onClick={resetDefaults}>
            恢复默认
          </button>
        </div>
      </form>
    </AppShell>
  )
}
