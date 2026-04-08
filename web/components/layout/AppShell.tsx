'use client'

import Link from 'next/link'
import { usePathname, useRouter } from 'next/navigation'
import { PropsWithChildren, useEffect, useState } from 'react'
import { clearAuth, getCurrentUser } from '@/lib/auth'
import { ACTIONS, hasPermission } from '@/lib/rbac'

const items = [
  { href: '/dashboard', label: '工作台' },
  { href: '/cases', label: '案件中心', action: ACTIONS.MISSING_MODIFY },
  { href: '/tasks', label: '任务中心', action: ACTIONS.TASK_VIEW },
  { href: '/dialects', label: '方言中心', action: ACTIONS.DIALECT_MODIFY },
  { href: '/dialects/cards', label: '方言卡片', action: ACTIONS.DIALECT_MANAGE },
  { href: '/attachments', label: '附件管理', action: ACTIONS.USER_MODIFY },
  { href: '/site-settings', label: '网站设置', action: ACTIONS.USER_MODIFY },
  { href: '/feature-settings', label: '功能设置', action: ACTIONS.USER_MODIFY },
  { href: '/monitor', label: '服务监控', action: ACTIONS.USER_MODIFY },
  { href: '/organizations', label: '组织管理', action: ACTIONS.ORG_MANAGE },
  { href: '/users', label: '人员管理', action: ACTIONS.USER_VIEW },
  { href: '/audit', label: '审计中心', action: ACTIONS.USER_MODIFY },
  { href: '/settings', label: '个人设置' },
]

export function AppShell({ children }: PropsWithChildren) {
  const pathname = usePathname()
  const router = useRouter()
  const user = getCurrentUser()
  const [menuOpen, setMenuOpen] = useState(false)

  useEffect(() => {
    setMenuOpen(false)
  }, [pathname])

  return (
    <div className="shell">
      <aside className={`sidebar ${menuOpen ? 'open' : ''}`}>
        <div className="brand">
          助力团圆 Web
          <small>走失人员寻亲协作平台</small>
        </div>
        <nav className="nav-list">
          {items
            .filter((item) => !item.action || hasPermission(user, item.action))
            .map((item) => {
            const active = pathname === item.href || pathname.startsWith(`${item.href}/`)
            return (
              <Link key={item.href} href={item.href} className={`nav-item ${active ? 'active' : ''}`}>
                {item.label}
              </Link>
            )
          })}
        </nav>
      </aside>
      {menuOpen ? <div className="sidebar-mask" onClick={() => setMenuOpen(false)} /> : null}
      <main className="main">
        <header className="topbar">
          <button className="btn ghost menu-btn" type="button" onClick={() => setMenuOpen((v) => !v)}>
            菜单
          </button>
          <div>
            <h1 className="top-title">团圆寻亲志愿者系统</h1>
            <p className="top-subtitle">以温暖协作连接每一次线索与团圆</p>
          </div>
          <div className="top-user">
            <img
              className="top-avatar"
              src={user?.avatar || '/default-avatar.svg'}
              alt={user?.nickname || 'avatar'}
              onError={(e) => {
                ;(e.currentTarget as HTMLImageElement).src = '/default-avatar.svg'
              }}
            />
            <span>{user?.nickname || user?.phone || '未登录'}</span>
            <span className="role-badge">{user?.role || '-'}</span>
            <button
              className="btn ghost"
              type="button"
              onClick={() => {
                clearAuth()
                router.replace('/login')
              }}
            >
              退出
            </button>
          </div>
        </header>
        <section className="content">{children}</section>
      </main>
    </div>
  )
}
