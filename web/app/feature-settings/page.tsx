'use client'

import { FormEvent, useEffect, useState } from 'react'
import { AppShell } from '@/components/layout/AppShell'
import { ModuleHeader } from '@/components/shared/ModuleHeader'
import { NoticeBar, type Notice } from '@/components/shared/NoticeBar'
import { PageState } from '@/components/shared/PageState'
import { useAuthGuard } from '@/hooks/useAuthGuard'
import { ACTIONS, hasPermission } from '@/lib/rbac'

const FEATURE_SETTINGS_KEY = 'web_feature_flags_v1'

type FeatureFlags = {
  cases_enabled: boolean
  tasks_enabled: boolean
  dialects_enabled: boolean
  volunteer_registration: boolean
  upload_enabled: boolean
  audit_enabled: boolean
  monitor_enabled: boolean
}

const defaults: FeatureFlags = {
  cases_enabled: true,
  tasks_enabled: true,
  dialects_enabled: true,
  volunteer_registration: true,
  upload_enabled: true,
  audit_enabled: true,
  monitor_enabled: true,
}

export default function FeatureSettingsPage() {
  const { ready, user } = useAuthGuard()
  const [notice, setNotice] = useState<Notice | null>(null)
  const [flags, setFlags] = useState<FeatureFlags>(defaults)

  useEffect(() => {
    if (!ready || typeof window === 'undefined') return
    try {
      const saved = JSON.parse(localStorage.getItem(FEATURE_SETTINGS_KEY) || '{}')
      setFlags({ ...defaults, ...(saved || {}) })
    } catch {
      setFlags(defaults)
    }
  }, [ready])

  function save(e: FormEvent) {
    e.preventDefault()
    if (typeof window === 'undefined') return
    localStorage.setItem(FEATURE_SETTINGS_KEY, JSON.stringify(flags))
    setNotice({ type: 'success', text: '功能开关已保存（当前为前端本地配置）' })
  }

  function resetDefaults() {
    setFlags(defaults)
    setNotice({ type: 'info', text: '已恢复默认开关，点击保存后生效' })
  }

  if (!ready) return null
  if (!hasPermission(user, ACTIONS.USER_MODIFY)) {
    return (
      <AppShell>
        <ModuleHeader title="功能设置" desc="模块开关、业务能力启停与灰度预留" />
        <PageState error="当前账号无权限访问该页面（需要 user:modify 权限）" />
      </AppShell>
    )
  }

  return (
    <AppShell>
      <ModuleHeader title="功能设置" desc="模块开关、业务能力启停与灰度预留" />
      <NoticeBar notice={notice} onClose={() => setNotice(null)} />
      <form className="section-card grid cols-2" onSubmit={save}>
        <label><input type="checkbox" checked={flags.cases_enabled} onChange={(e) => setFlags((f) => ({ ...f, cases_enabled: e.target.checked }))} /> 案件模块启用</label>
        <label><input type="checkbox" checked={flags.tasks_enabled} onChange={(e) => setFlags((f) => ({ ...f, tasks_enabled: e.target.checked }))} /> 任务模块启用</label>
        <label><input type="checkbox" checked={flags.dialects_enabled} onChange={(e) => setFlags((f) => ({ ...f, dialects_enabled: e.target.checked }))} /> 方言模块启用</label>
        <label><input type="checkbox" checked={flags.volunteer_registration} onChange={(e) => setFlags((f) => ({ ...f, volunteer_registration: e.target.checked }))} /> 志愿者自助注册</label>
        <label><input type="checkbox" checked={flags.upload_enabled} onChange={(e) => setFlags((f) => ({ ...f, upload_enabled: e.target.checked }))} /> 文件上传功能</label>
        <label><input type="checkbox" checked={flags.audit_enabled} onChange={(e) => setFlags((f) => ({ ...f, audit_enabled: e.target.checked }))} /> 审计模块启用</label>
        <label><input type="checkbox" checked={flags.monitor_enabled} onChange={(e) => setFlags((f) => ({ ...f, monitor_enabled: e.target.checked }))} /> 服务监控启用</label>
        <div className="row" style={{ gridColumn: '1 / -1' }}>
          <button className="btn primary" type="submit">
            保存开关
          </button>
          <button className="btn ghost" type="button" onClick={resetDefaults}>
            恢复默认
          </button>
        </div>
      </form>
    </AppShell>
  )
}
