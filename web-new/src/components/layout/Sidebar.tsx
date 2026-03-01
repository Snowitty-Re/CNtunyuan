import { useMemo } from 'react';
import { Menu } from 'antd';
import type { MenuProps } from 'antd';
import {
  HomeOutlined,
  SearchOutlined,
  FileTextOutlined,
  SoundOutlined,
  ApartmentOutlined,
  TeamOutlined,
} from '@ant-design/icons';
import { useNavigate, useLocation } from 'react-router-dom';
import { usePermission } from '@/utils/permission';

interface SidebarProps {
  collapsed: boolean;
}

interface MenuItem {
  key: string;
  icon?: React.ReactNode;
  label: string;
  path: string;
  minRole?: string;
  children?: MenuItem[];
}

export default function Sidebar({ collapsed }: SidebarProps) {
  const navigate = useNavigate();
  const location = useLocation();
  const { isAdmin, isManager } = usePermission();

  // 菜单配置 - 简洁办公风
  const menuItems: MenuItem[] = useMemo(
    () => [
      {
        key: '/dashboard',
        icon: <HomeOutlined />,
        label: '工作台',
        path: '/dashboard',
      },
      {
        key: '/cases',
        icon: <SearchOutlined />,
        label: '寻人案件',
        path: '/cases',
      },
      {
        key: '/tasks',
        icon: <FileTextOutlined />,
        label: '任务管理',
        path: '/tasks',
      },
      {
        key: '/dialects',
        icon: <SoundOutlined />,
        label: '方言管理',
        path: '/dialects',
      },
      {
        key: '/organizations',
        icon: <ApartmentOutlined />,
        label: '组织架构',
        path: '/organizations',
        minRole: 'admin',
      },
      {
        key: '/volunteers',
        icon: <TeamOutlined />,
        label: '志愿者管理',
        path: '/volunteers',
        minRole: 'admin',
      },
    ],
    []
  );

  // 过滤菜单
  const filteredMenus = useMemo(() => {
    const canAccess = (minRole?: string): boolean => {
      if (!minRole) return true;
      if (minRole === 'super_admin') return isAdmin;
      if (minRole === 'admin') return isAdmin;
      if (minRole === 'manager') return isManager;
      return true;
    };
    
    return menuItems.filter((item) => canAccess(item.minRole));
  }, [menuItems, isAdmin, isManager]);

  // 转换Ant Design菜单格式
  const convertMenuItems = (items: MenuItem[]): MenuProps['items'] => {
    return items.map((item) => ({
      key: item.key,
      icon: item.icon,
      label: item.label,
      onClick: () => navigate(item.path),
    }));
  };

  // 获取当前选中的菜单
  const selectedKeys = useMemo(() => {
    const path = location.pathname;
    for (const item of menuItems) {
      if (path === item.path || path.startsWith(item.path + '/')) {
        return [item.key];
      }
    }
    return [path];
  }, [location.pathname]);

  return (
    <div style={{ 
      height: '100%', 
      display: 'flex', 
      flexDirection: 'column',
      background: '#fff',
    }}>
      {/* Logo区域 - 简洁温馨 */}
      <div
        style={{
          height: 64,
          display: 'flex',
          alignItems: 'center',
          justifyContent: collapsed ? 'center' : 'flex-start',
          padding: collapsed ? 0 : '0 20px',
          borderBottom: '1px solid #e8e9eb',
        }}
      >
        {collapsed ? (
          <div
            style={{
              width: 36,
              height: 36,
              borderRadius: 8,
              background: 'linear-gradient(135deg, #e67e22 0%, #f39c12 100%)',
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
              color: '#fff',
              fontSize: 18,
              fontWeight: 700,
            }}
          >
            团
          </div>
        ) : (
          <div style={{ display: 'flex', alignItems: 'center', gap: 12 }}>
            <div
              style={{
                width: 36,
                height: 36,
                borderRadius: 8,
                background: 'linear-gradient(135deg, #e67e22 0%, #f39c12 100%)',
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'center',
                color: '#fff',
                fontSize: 18,
                fontWeight: 700,
              }}
            >
              团
            </div>
            <div>
              <div style={{ 
                fontSize: 16, 
                fontWeight: 700, 
                color: '#1f2329',
                lineHeight: '22px',
              }}>
                团圆寻亲
              </div>
              <div style={{ 
                fontSize: 12, 
                color: '#8f959e',
                lineHeight: '16px',
              }}>
                志愿者系统
              </div>
            </div>
          </div>
        )}
      </div>

      {/* 菜单 */}
      <div style={{ flex: 1, overflow: 'auto', padding: '12px 0' }}>
        <Menu
          mode="inline"
          theme="light"
          selectedKeys={selectedKeys}
          items={convertMenuItems(filteredMenus)}
          style={{
            background: 'transparent',
            border: 'none',
          }}
          inlineCollapsed={collapsed}
        />
      </div>

      {/* 底部信息 */}
      {!collapsed && (
        <div 
          style={{ 
            padding: '12px 20px', 
            borderTop: '1px solid #e8e9eb',
          }}
        >
          <div style={{ 
            fontSize: 12, 
            color: '#8f959e',
            textAlign: 'center',
          }}>
            © 2024 团圆寻亲系统
          </div>
          <div style={{ 
            fontSize: 12, 
            color: '#8f959e',
            textAlign: 'center',
            marginTop: 4,
          }}>
            版本 2.0.0
          </div>
        </div>
      )}
    </div>
  );
}
