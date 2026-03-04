import { useEffect, useState } from 'react';
import { useNavigate, useLocation } from 'react-router-dom';
import { useAuthStore } from '@/stores/auth';
import { Spin } from 'antd';
import axios from 'axios';

interface RouteGuardProps {
  children: React.ReactNode;
}

export default function RouteGuard({ children }: RouteGuardProps) {
  const navigate = useNavigate();
  const location = useLocation();
  const { isAuthenticated } = useAuthStore();
  const [isReady, setIsReady] = useState(false);
  const [initialized, setInitialized] = useState<boolean | null>(null);

  useEffect(() => {
    // 检查系统初始化状态
    const checkInitStatus = async () => {
      try {
        const apiUrl = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080/api/v1';
        const res = await axios.get(`${apiUrl}/setup/status`);
        setInitialized(res.data.data?.initialized || false);
      } catch (error) {
        console.error('检查初始化状态失败:', error);
        // 如果请求失败，假设系统已初始化（避免无法访问）
        setInitialized(true);
      } finally {
        setIsReady(true);
      }
    };

    checkInitStatus();
  }, []);

  useEffect(() => {
    if (!isReady || initialized === null) return;

    // 如果系统未初始化且不在初始化页面，跳转到初始化页面
    if (!initialized && location.pathname !== '/setup') {
      navigate('/setup', { replace: true });
      return;
    }

    // 如果系统已初始化但在初始化页面，跳转到登录页
    if (initialized && location.pathname === '/setup') {
      navigate('/login', { replace: true });
      return;
    }

    // 如果未登录且不在登录页，跳转到登录页
    if (!isAuthenticated && location.pathname !== '/login') {
      navigate('/login', { replace: true });
    }
    
    // 如果已登录且在登录页，跳转到工作台
    if (isAuthenticated && location.pathname === '/login') {
      navigate('/dashboard', { replace: true });
    }
  }, [isAuthenticated, location.pathname, navigate, isReady, initialized]);

  if (!isReady) {
    return (
      <div style={{ 
        display: 'flex', 
        justifyContent: 'center', 
        alignItems: 'center', 
        height: '100vh' 
      }}>
        <Spin size="large" />
      </div>
    );
  }

  return <>{children}</>;
}
