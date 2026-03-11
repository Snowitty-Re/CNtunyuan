import { useMemo } from 'react';
import { Outlet, useNavigate, useLocation } from 'react-router-dom';
import { Layout, Menu, Avatar, Dropdown, Typography, theme } from 'antd';
import {
  DashboardOutlined,
  TeamOutlined,
  OrderedListOutlined,
  AudioOutlined,
  UserOutlined,
  ApartmentOutlined,
  FolderOutlined,
  AuditOutlined,
  MonitorOutlined,
  MenuFoldOutlined,
  MenuUnfoldOutlined,
  LogoutOutlined,
  ProfileOutlined,
} from '@ant-design/icons';
import { useAppStore } from '@/stores/appStore';
import { useAuthStore } from '@/stores/authStore';
import { hasRole } from '@/utils/permission';
import type { UserRole } from '@/types';

const { Header, Sider, Content, Footer } = Layout;

interface MenuItem {
  key: string;
  icon: React.ReactNode;
  label: string;
  minRole: UserRole;
}

const menuItems: MenuItem[] = [
  { key: '/dashboard', icon: <DashboardOutlined />, label: '工作台', minRole: 'volunteer' },
  { key: '/missing-persons', icon: <TeamOutlined />, label: '走失人员', minRole: 'volunteer' },
  { key: '/tasks', icon: <OrderedListOutlined />, label: '任务管理', minRole: 'volunteer' },
  { key: '/dialects', icon: <AudioOutlined />, label: '方言库', minRole: 'volunteer' },
  { key: '/users', icon: <UserOutlined />, label: '用户管理', minRole: 'admin' },
  { key: '/organizations', icon: <ApartmentOutlined />, label: '组织管理', minRole: 'admin' },
  { key: '/files', icon: <FolderOutlined />, label: '文件管理', minRole: 'manager' },
  { key: '/audit', icon: <AuditOutlined />, label: '审计日志', minRole: 'manager' },
  { key: '/system/health', icon: <MonitorOutlined />, label: '系统监控', minRole: 'admin' },
];

export default function AdminLayout() {
  const navigate = useNavigate();
  const location = useLocation();
  const { collapsed, toggleCollapsed } = useAppStore();
  const { user, logout } = useAuthStore();
  const { token: themeToken } = theme.useToken();

  const userRole = (user?.role || 'volunteer') as UserRole;

  const visibleMenuItems = useMemo(
    () =>
      menuItems
        .filter((item) => hasRole(userRole, item.minRole))
        .map(({ key, icon, label }) => ({ key, icon, label })),
    [userRole]
  );

  const selectedKey = useMemo(() => {
    const path = location.pathname;
    const found = menuItems.find((m) => path.startsWith(m.key));
    return found?.key || '/dashboard';
  }, [location.pathname]);

  const handleLogout = async () => {
    await logout();
    navigate('/login');
  };

  const dropdownItems = {
    items: [
      { key: 'profile', icon: <ProfileOutlined />, label: '个人资料' },
      { type: 'divider' as const },
      { key: 'logout', icon: <LogoutOutlined />, label: '退出登录', danger: true },
    ],
    onClick: ({ key }: { key: string }) => {
      if (key === 'logout') handleLogout();
      else if (key === 'profile') navigate('/profile');
    },
  };

  return (
    <Layout style={{ minHeight: '100vh' }}>
      <Sider
        trigger={null}
        collapsible
        collapsed={collapsed}
        style={{
          overflow: 'auto',
          height: '100vh',
          position: 'fixed',
          left: 0,
          top: 0,
          bottom: 0,
          zIndex: 100,
        }}
      >
        <div
          style={{
            height: 64,
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            borderBottom: `1px solid ${themeToken.colorBorderSecondary}`,
          }}
        >
          <Typography.Title
            level={4}
            style={{ color: '#fff', margin: 0, whiteSpace: 'nowrap' }}
          >
            {collapsed ? '团' : '团圆寻亲'}
          </Typography.Title>
        </div>
        <Menu
          theme="dark"
          mode="inline"
          selectedKeys={[selectedKey]}
          items={visibleMenuItems}
          onClick={({ key }) => navigate(key)}
        />
      </Sider>
      <Layout style={{ marginLeft: collapsed ? 80 : 200, transition: 'margin-left 0.2s' }}>
        <Header
          style={{
            padding: '0 24px',
            background: themeToken.colorBgContainer,
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'space-between',
            borderBottom: `1px solid ${themeToken.colorBorderSecondary}`,
            position: 'sticky',
            top: 0,
            zIndex: 99,
          }}
        >
          <div
            onClick={toggleCollapsed}
            style={{ fontSize: 18, cursor: 'pointer', padding: '0 8px' }}
          >
            {collapsed ? <MenuUnfoldOutlined /> : <MenuFoldOutlined />}
          </div>
          <Dropdown menu={dropdownItems} placement="bottomRight">
            <div style={{ cursor: 'pointer', display: 'flex', alignItems: 'center', gap: 8 }}>
              <Avatar icon={<UserOutlined />} src={user?.avatar} />
              <span>{user?.nickname || '用户'}</span>
            </div>
          </Dropdown>
        </Header>
        <Content style={{ margin: 24, minHeight: 280 }}>
          <Outlet />
        </Content>
        <Footer style={{ textAlign: 'center', color: themeToken.colorTextSecondary }}>
          团圆寻亲志愿者系统 &copy; {new Date().getFullYear()}
        </Footer>
      </Layout>
    </Layout>
  );
}
