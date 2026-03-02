import { useState, useEffect } from 'react';
import { Layout, Badge, Avatar, Dropdown, Button } from 'antd';
import {
  MenuFoldOutlined,
  MenuUnfoldOutlined,
  BellOutlined,
  UserOutlined,
  LogoutOutlined,
  SettingOutlined,
} from '@ant-design/icons';
import { Outlet, useNavigate, useLocation } from 'react-router-dom';
import Sidebar from './Sidebar';
import { useAuthStore } from '@/stores/auth';
import { useGlobalStore } from '@/stores/global';
import { usePermission, getRoleLabel, getRoleColor } from '@/utils/permission';

const { Header, Content, Sider } = Layout;

export default function MainLayout() {
  const navigate = useNavigate();
  const { user, logout } = useAuthStore();
  const { sidebarCollapsed, toggleSidebar, unreadCount } = useGlobalStore();
  const { isAdmin } = usePermission();
  
  const [isMobile, setIsMobile] = useState(false);

  // 检测移动端
  useEffect(() => {
    const checkMobile = () => {
      setIsMobile(window.innerWidth < 768);
    };
    checkMobile();
    window.addEventListener('resize', checkMobile);
    return () => window.removeEventListener('resize', checkMobile);
  }, []);

  // 用户菜单
  const userMenuItems = [
    {
      key: 'profile',
      icon: <UserOutlined />,
      label: '个人中心',
      onClick: () => navigate('/profile'),
    },
    ...(isAdmin ? [{
      key: 'settings',
      icon: <SettingOutlined />,
      label: '系统设置',
      onClick: () => navigate('/settings'),
    }] : []),
    {
      type: 'divider' as const,
    },
    {
      key: 'logout',
      icon: <LogoutOutlined />,
      label: '退出登录',
      danger: true,
      onClick: () => {
        logout();
        navigate('/login');
      },
    },
  ];

  const userRoleLabel = getRoleLabel(user?.role || '');
  const userRoleColor = getRoleColor(user?.role || '');

  return (
    <Layout style={{ minHeight: '100vh', background: '#f5f7fa' }}>
      {/* 侧边栏 - 简洁干净 */}
      <Sider
        trigger={null}
        collapsible
        collapsed={sidebarCollapsed}
        collapsedWidth={isMobile ? 0 : 72}
        width={220}
        theme="light"
        style={{
          position: 'fixed',
          left: 0,
          top: 0,
          height: '100%',
          zIndex: 100,
          background: '#fff',
          borderRight: '1px solid #e8e9eb',
          boxShadow: '2px 0 8px rgba(0,0,0,0.02)',
        }}
      >
        <Sidebar collapsed={sidebarCollapsed} />
      </Sider>

      {/* 主内容区 */}
      <Layout
        style={{
          marginLeft: isMobile ? 0 : sidebarCollapsed ? 72 : 220,
          transition: 'margin-left 0.2s ease',
        }}
      >
        {/* 顶部导航 - 简洁办公风 */}
        <Header
          style={{
            position: 'sticky',
            top: 0,
            zIndex: 50,
            height: 64,
            background: '#fff',
            borderBottom: '1px solid #e8e9eb',
            padding: '0 24px',
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'space-between',
          }}
        >
          <div style={{ display: 'flex', alignItems: 'center', gap: 16 }}>
            <Button
              type="text"
              icon={sidebarCollapsed ? <MenuUnfoldOutlined /> : <MenuFoldOutlined />}
              onClick={toggleSidebar}
              style={{ 
                fontSize: 16, 
                color: '#646a73',
                width: 32,
                height: 32,
              }}
            />
            <Breadcrumb />
          </div>

          <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
            {/* 通知 */}
            <Badge count={unreadCount} size="small" offset={[-2, 2]}>
              <Button
                type="text"
                icon={<BellOutlined style={{ fontSize: 18 }} />}
                onClick={() => navigate('/notifications')}
                style={{ 
                  color: '#646a73',
                  width: 40,
                  height: 40,
                }}
              />
            </Badge>

            {/* 用户信息 - 简洁温馨 */}
            <Dropdown 
              menu={{ items: userMenuItems }} 
              placement="bottomRight" 
              arrow
              overlayStyle={{ minWidth: 160 }}
            >
              <div 
                style={{ 
                  display: 'flex', 
                  alignItems: 'center', 
                  gap: 10, 
                  cursor: 'pointer',
                  padding: '6px 12px',
                  borderRadius: 6,
                  transition: 'background-color 0.2s',
                }}
                onMouseEnter={(e) => e.currentTarget.style.backgroundColor = '#f5f7fa'}
                onMouseLeave={(e) => e.currentTarget.style.backgroundColor = 'transparent'}
              >
                <Avatar
                  size={32}
                  src={user?.avatar}
                  icon={<UserOutlined />}
                  style={{ 
                    backgroundColor: userRoleColor,
                    fontSize: 14,
                  }}
                />
                <div style={{ display: isMobile ? 'none' : 'block' }}>
                  <div style={{ 
                    fontSize: 14, 
                    fontWeight: 500, 
                    color: '#1f2329',
                    lineHeight: '20px',
                  }}>
                    {user?.nickname || user?.real_name || '用户'}
                  </div>
                  <div
                    style={{
                      fontSize: 12,
                      color: userRoleColor,
                      lineHeight: '18px',
                    }}
                  >
                    {userRoleLabel}
                  </div>
                </div>
              </div>
            </Dropdown>
          </div>
        </Header>

        {/* 页面内容 */}
        <Content style={{ padding: 24, minHeight: 'calc(100vh - 64px)' }}>
          <Outlet />
        </Content>
      </Layout>
    </Layout>
  );
}

// 面包屑组件
function Breadcrumb() {
  const location = useLocation();
  const pathMap: Record<string, string> = {
    '/': '首页',
    '/dashboard': '工作台',
    '/cases': '寻人案件',
    '/tasks': '任务管理',
    '/volunteers': '志愿者管理',
    '/organizations': '组织架构',
    '/dialects': '方言管理',
    '/profile': '个人中心',
    '/settings': '系统设置',
  };

  const title = pathMap[location.pathname] || '页面';

  return (
    <div style={{ 
      fontSize: 16, 
      fontWeight: 600, 
      color: '#1f2329',
    }}>
      {title}
    </div>
  );
}
