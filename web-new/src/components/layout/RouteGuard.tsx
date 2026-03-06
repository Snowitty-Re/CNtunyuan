import { useEffect, useState, useRef } from 'react';
import { useNavigate, useLocation } from 'react-router-dom';
import { useAuthStore } from '@/stores/auth';
import { Spin } from 'antd';

interface RouteGuardProps {
  children: React.ReactNode;
}

export default function RouteGuard({ children }: RouteGuardProps) {
  const navigate = useNavigate();
  const location = useLocation();
  const { isAuthenticated } = useAuthStore();
  const [isReady, setIsReady] = useState(false);
  const lastRedirectRef = useRef<number>(0);

  useEffect(() => {
    // 延迟一小段时间，让 zustand persist 有时间恢复状态
    const timer = setTimeout(() => {
      setIsReady(true);
    }, 300);

    return () => clearTimeout(timer);
  }, []);

  useEffect(() => {
    if (!isReady) return;

    const currentPath = location.pathname;
    const now = Date.now();
    
    // 防止短时间内重复跳转（500ms内）
    if (now - lastRedirectRef.current < 500) {
      return;
    }

    // 如果未登录且不在登录页，跳转到登录页
    if (!isAuthenticated && currentPath !== '/login') {
      lastRedirectRef.current = now;
      navigate('/login', { replace: true });
      return;
    }
    
    // 如果已登录且在登录页，跳转到工作台
    if (isAuthenticated && currentPath === '/login') {
      lastRedirectRef.current = now;
      navigate('/dashboard', { replace: true });
      return;
    }
  }, [isAuthenticated, location.pathname, navigate, isReady]);

  if (!isReady) {
    return (
      <div style={{ 
        display: 'flex', 
        justifyContent: 'center', 
        alignItems: 'center', 
        height: '100vh',
        background: '#f5f7fa'
      }}>
        <Spin size="large" />
      </div>
    );
  }

  return <>{children}</>;
}
