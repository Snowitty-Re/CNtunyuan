import { useEffect, useState } from 'react';
import { useNavigate, useLocation } from 'react-router-dom';
import { useAuthStore } from '@/stores/auth';
import { Spin } from 'antd';

interface RouteGuardProps {
  children: React.ReactNode;
}

export default function RouteGuard({ children }: RouteGuardProps) {
  const navigate = useNavigate();
  const location = useLocation();
  const { isAuthenticated, isHydrated } = useAuthStore();
  const [isReady, setIsReady] = useState(false);

  // 等待状态从存储恢复
  useEffect(() => {
    // 如果状态已经恢复，标记为 ready
    if (isHydrated) {
      setIsReady(true);
    }
  }, [isHydrated]);

  useEffect(() => {
    // 等待状态恢复后再执行路由守卫逻辑
    if (!isReady) return;

    // 如果未登录且不在登录页，跳转到登录页
    if (!isAuthenticated && location.pathname !== '/login') {
      navigate('/login', { replace: true });
    }
    
    // 如果已登录且在登录页，跳转到工作台
    if (isAuthenticated && location.pathname === '/login') {
      navigate('/dashboard', { replace: true });
    }
  }, [isAuthenticated, location.pathname, navigate, isReady]);

  // 等待状态恢复时显示 loading
  if (!isReady) {
    return (
      <div style={{ 
        display: 'flex', 
        justifyContent: 'center', 
        alignItems: 'center', 
        height: '100vh',
        background: '#f5f7fa'
      }}>
        <Spin size="large" tip="加载中..." />
      </div>
    );
  }

  return <>{children}</>;
}
