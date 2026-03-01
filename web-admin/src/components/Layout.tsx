import { useState, useMemo } from 'react'
import { Layout as AntLayout, Menu, Button, Avatar, Dropdown, Badge } from 'antd'
import type { MenuProps } from 'antd'
import {
  MenuFoldOutlined,
  MenuUnfoldOutlined,
  DashboardOutlined,
  TeamOutlined,
  ApartmentOutlined,
  SearchOutlined,
  SoundOutlined,
  FileTextOutlined,
  NodeCollapseOutlined,
  LogoutOutlined,
  BellOutlined,
  UserOutlined,
  SafetyOutlined,
  SettingOutlined,
} from '@ant-design/icons'
import { Outlet, useNavigate, useLocation } from 'react-router-dom'
import { useAuthStore } from '../stores/auth'
import { usePermission, RoleLevel } from '../hooks/usePermission.tsx'

const { Header, Sider, Content } = AntLayout

interface MenuItem {
  key: string
  icon: React.ReactNode
  label: string
  minRole?: RoleLevel
}

const Layout = () => {
  const [collapsed, setCollapsed] = useState(false)
  const navigate = useNavigate()
  const location = useLocation()
  const { user, logout } = useAuthStore()
  const { roleLevel, isAdmin, isManager, isSuperAdmin } = usePermission()

  const menuItems: MenuItem[] = useMemo(
    () => [
      { key: '/', icon: <DashboardOutlined />, label: '仪表盘' },
      { key: '/users', icon: <TeamOutlined />, label: '志愿者管理', minRole: RoleLevel.Admin },
      { key: '/organizations', icon: <ApartmentOutlined />, label: '组织架构' },
      { key: '/missing-persons', icon: <SearchOutlined />, label: '走失人员' },
      { key: '/dialects', icon: <SoundOutlined />, label: '方言管理' },
      { key: '/tasks', icon: <FileTextOutlined />, label: '任务管理' },
      { key: '/workflows', icon: <NodeCollapseOutlined />, label: '工作流管理' },
      { key: '/logs', icon: <SafetyOutlined />, label: '操作日志', minRole: RoleLevel.SuperAdmin },
      { key: '/settings', icon: <SettingOutlined />, label: '系统设置', minRole: RoleLevel.Admin },
    ],
    []
  )

  // 根据权限过滤菜单
  const filteredMenuItems = useMemo(() => {
    return menuItems
      .filter((item) => !item.minRole || roleLevel >= item.minRole)
      .map((item) => ({
        key: item.key,
        icon: item.icon,
        label: item.label,
      }))
  }, [menuItems, roleLevel])

  const userMenuItems = useMemo(
    () => [
      { key: 'profile', icon: <UserOutlined />, label: '个人中心' },
      { key: 'logout', icon: <LogoutOutlined />, label: '退出登录' },
    ],
    []
  )

  const handleMenuClick: MenuProps['onClick'] = ({ key }) => {
    if (key === 'logout') {
      logout()
      navigate('/login')
    } else if (key === 'profile') {
      navigate('/profile')
    }
  }

  // 获取角色标签
  const getRoleLabel = (role?: string) => {
    const labels: Record<string, string> = {
      super_admin: '超级管理员',
      admin: '管理员',
      manager: '管理者',
      volunteer: '志愿者',
    }
    return labels[role || ''] || '用户'
  }

  // 获取角色颜色
  const getRoleColor = (role?: string) => {
    const colors: Record<string, string> = {
      super_admin: '#f5222d',
      admin: '#fa8c16',
      manager: '#1890ff',
      volunteer: '#52c41a',
    }
    return colors[role || ''] || '#999'
  }

  return (
    <AntLayout style={{ minHeight: '100vh' }}>
      <Sider trigger={null} collapsible collapsed={collapsed} theme="light">
        <div
          style={{
            height: 64,
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            borderBottom: '1px solid #f0f0f0',
          }}
        >
          <h2 style={{ margin: 0, color: '#1890ff' }}>{collapsed ? '团圆' : '团圆寻亲系统'}</h2>
        </div>
        <Menu
          theme="light"
          mode="inline"
          selectedKeys={[location.pathname]}
          items={filteredMenuItems}
          onClick={({ key }) => navigate(key)}
        />
      </Sider>
      <AntLayout>
        <Header
          style={{
            padding: 0,
            background: '#fff',
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'space-between',
          }}
        >
          <Button
            type="text"
            icon={collapsed ? <MenuUnfoldOutlined /> : <MenuFoldOutlined />}
            onClick={() => setCollapsed(!collapsed)}
            style={{ fontSize: 16, width: 64, height: 64 }}
          />
          <div style={{ display: 'flex', alignItems: 'center', gap: 16, paddingRight: 24 }}>
            <Badge count={5} size="small">
              <BellOutlined style={{ fontSize: 20, cursor: 'pointer' }} />
            </Badge>
            <Dropdown menu={{ items: userMenuItems, onClick: handleMenuClick }} placement="bottomRight">
              <div style={{ cursor: 'pointer', display: 'flex', alignItems: 'center', gap: 8 }}>
                <Avatar src={user?.avatar} icon={<UserOutlined />} />
                <div style={{ display: 'flex', flexDirection: 'column', alignItems: 'flex-start' }}>
                  <span style={{ fontSize: 14, fontWeight: 500 }}>
                    {user?.nickname || user?.real_name || '用户'}
                  </span>
                  <span
                    style={{
                      fontSize: 12,
                      color: getRoleColor(user?.role),
                      backgroundColor: `${getRoleColor(user?.role)}15`,
                      padding: '0 6px',
                      borderRadius: 4,
                      lineHeight: '18px',
                    }}
                  >
                    {getRoleLabel(user?.role)}
                  </span>
                </div>
              </div>
            </Dropdown>
          </div>
        </Header>
        <Content
          style={{
            margin: 24,
            padding: 24,
            background: '#fff',
            borderRadius: 8,
            overflow: 'auto',
          }}
        >
          <Outlet />
        </Content>
      </AntLayout>
    </AntLayout>
  )
}

export default Layout
