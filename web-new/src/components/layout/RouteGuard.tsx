import { useEffect, useRef } from 'react';
import { useNavigate, useLocation } from 'react-router-dom';
import { useAuthStore } from '@/stores/auth';
import { Spin } from 'antd';

interface RouteGuardProps {
  children: React.ReactNode;
}

// 保存日志到 sessionStorage 以便查看
const log = (msg: string, data?: any) => {
  const logs = JSON.parse(sessionStorage.getItem('routeGuardLogs') || '[]');
  logs.push({ time: Date.now(), msg, data });
  sessionStorage.setItem('routeGuardLogs', JSON.stringify(logs));
  console.log(msg, data);
};

export default function RouteGuard({ children }: RouteGuardProps) {
  const navigate = useNavigate();
  const location = useLocation();
  const { isAuthenticated, hasHydrated, token } = useAuthStore();
  const lastRedirectRef = useRef<number>(0);
  const didCheckRef = useRef(false);

  useEffect(() => {
    log('RouteGuard effect', { 
      isAuthenticated, 
      hasHydrated, 
      hasToken: !!token,
      path: location.pathname 
    });
    
    // 只检查一次，避免重复跳转
    if (didCheckRef.current) return;
    
    // 等待 hydrate 完成
    if (!hasHydrated) {
      log('等待 hydrate...');
      return;
    }
    
    didCheckRef.current = true;
    const currentPath = location.pathname;
    const now = Date.now();
    
    // 防止短时间内重复跳转
    if (now - lastRedirectRef.current < 1000) {
      log('短时间内重复跳转，跳过');
      return;
    }

    // 直接从 store 读取最新状态
    const state = useAuthStore.getState();
    log('Store 状态', { 
      isAuthenticated: state.isAuthenticated, 
      hasToken: !!state.token,
      hasHydrated: state.hasHydrated
    });

    // 如果未登录且不在登录页，跳转到登录页
    if (!state.isAuthenticated && !state.token && currentPath !== '/login') {
      log('未登录，准备跳转登录页');
      lastRedirectRef.current = now;
      navigate('/login', { replace: true });
    }
  }, [isAuthenticated, hasHydrated, token, location.pathname, navigate]);

  // 等待状态恢复完成
  if (!hasHydrated) {
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
