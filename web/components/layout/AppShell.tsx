'use client'

import Link from 'next/link'
import { usePathname, useRouter } from 'next/navigation'
import { PropsWithChildren, useEffect, useState } from 'react'
import { clearAuth, getCurrentUser } from '@/lib/auth'
import { hasMinRole } from '@/lib/rbac'

const items = [
  { href: '/dashboard', label: '工作台', minRole: 'volunteer' },
  { href: '/cases', label: '案件中心', minRole: 'volunteer' },
  { href: '/tasks', label: '任务中心', minRole: 'volunteer' },
  { href: '/dialects', label: '方言中心', minRole: 'volunteer' },
  { href: '/organizations', label: '组织管理', minRole: 'admin' },
  { href: '/users', label: '人员管理', minRole: 'admin' },
  { href: '/audit', label: '审计中心', minRole: 'admin' },
  { href: '/settings', label: '个人设置', minRole: 'volunteer' },
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
        <div className="brand">助力团圆 Web</div>
        <nav className="nav-list">
          {items.filter((item) => hasMinRole(user, item.minRole)).map((item) => {
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
            <p className="top-subtitle">面向志愿协作的案件与任务闭环管理</p>
          </div>
          <div className="top-user">
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
