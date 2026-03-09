import { useState } from 'react'
import { Outlet, useNavigate, useLocation } from 'react-router-dom'
import { Layout, Menu, Dropdown, Avatar, Button, theme } from 'antd'
import {
  DashboardOutlined, TeamOutlined, ApartmentOutlined, AlertOutlined,
  ScheduleOutlined, AudioOutlined, AuditOutlined, UserOutlined,
  MenuFoldOutlined, MenuUnfoldOutlined, LogoutOutlined, SettingOutlined,
} from '@ant-design/icons'
import { useAuthStore } from '@/store/auth'
import { usePermission } from '@/hooks/usePermission'
import { authApi } from '@/api/auth'
import type { MenuProps } from 'antd'

const { Header, Sider, Content } = Layout

const allMenuItems = [
  { key: '/dashboard', icon: <DashboardOutlined />, label: '工作台', minRole: 'volunteer' as const },
  { key: '/cases', icon: <AlertOutlined />, label: '案件管理', minRole: 'volunteer' as const },
  { key: '/tasks', icon: <ScheduleOutlined />, label: '任务管理', minRole: 'volunteer' as const },
  { key: '/users', icon: <TeamOutlined />, label: '用户管理', minRole: 'manager' as const },
  { key: '/organizations', icon: <ApartmentOutlined />, label: '组织管理', minRole: 'manager' as const },
  { key: '/dialects', icon: <AudioOutlined />, label: '方言管理', minRole: 'volunteer' as const },
  { key: '/audit', icon: <AuditOutlined />, label: '审计日志', minRole: 'admin' as const },
]

export default function MainLayout() {
  const [collapsed, setCollapsed] = useState(false)
  const navigate = useNavigate()
  const location = useLocation()
  const { user, logout } = useAuthStore()
  const { hasRole } = usePermission()
  const { token: { colorBgContainer, borderRadiusLG } } = theme.useToken()

  const menuItems = allMenuItems
    .filter((item) => hasRole(item.minRole))
    .map(({ key, icon, label }) => ({ key, icon, label }))

  const selectedKey = '/' + location.pathname.split('/')[1]

  const handleMenuClick: MenuProps['onClick'] = ({ key }) => navigate(key)

  const handleLogout = async () => {
    try { await authApi.logout() } catch { /* ignore */ }
    logout()
    navigate('/login')
  }

  const userMenu: MenuProps = {
    items: [
      { key: 'profile', icon: <UserOutlined />, label: '个人中心', onClick: () => navigate('/profile') },
      { key: 'settings', icon: <SettingOutlined />, label: '系统设置', onClick: () => navigate('/settings') },
      { type: 'divider' },
      { key: 'logout', icon: <LogoutOutlined />, label: '退出登录', danger: true, onClick: handleLogout },
    ],
  }

  return (
    <Layout style={{ minHeight: '100vh' }}>
      <Sider
        trigger={null}
        collapsible
        collapsed={collapsed}
        theme="light"
        style={{ borderRight: '1px solid #f0f0f0' }}
        width={220}
      >
        <div style={{
          height: 64, display: 'flex', alignItems: 'center', justifyContent: 'center',
          borderBottom: '1px solid #f0f0f0', padding: '0 16px', gap: 8,
        }}>
          <AlertOutlined style={{ fontSize: 24, color: '#e67e22' }} />
          {!collapsed && <span style={{ fontSize: 16, fontWeight: 600, color: '#1f2329', whiteSpace: 'nowrap' }}>团圆寻亲</span>}
        </div>
        <Menu
          mode="inline"
          selectedKeys={[selectedKey]}
          items={menuItems}
          onClick={handleMenuClick}
          style={{ border: 'none' }}
        />
      </Sider>
      <Layout>
        <Header style={{
          padding: '0 24px', background: colorBgContainer,
          display: 'flex', alignItems: 'center', justifyContent: 'space-between',
          borderBottom: '1px solid #f0f0f0', height: 64,
        }}>
          <Button
            type="text"
            icon={collapsed ? <MenuUnfoldOutlined /> : <MenuFoldOutlined />}
            onClick={() => setCollapsed(!collapsed)}
          />
          <Dropdown menu={userMenu} placement="bottomRight">
            <div style={{ cursor: 'pointer', display: 'flex', alignItems: 'center', gap: 8 }}>
              <Avatar size="small" icon={<UserOutlined />} src={user?.avatar} />
              <span style={{ fontSize: 14, color: '#1f2329' }}>{user?.nickname || '用户'}</span>
            </div>
          </Dropdown>
        </Header>
        <Content style={{ margin: 24, padding: 24, background: colorBgContainer, borderRadius: borderRadiusLG, overflow: 'auto' }}>
          <Outlet />
        </Content>
      </Layout>
    </Layout>
  )
}
