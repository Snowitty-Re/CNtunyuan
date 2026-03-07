import axios, { AxiosError, AxiosRequestConfig, AxiosResponse } from 'axios';
import { message } from 'antd';

// 从 auth-storage 解析 token
function getTokenFromStorage(): string | null {
  try {
    const authStorage = localStorage.getItem('auth-storage');
    if (authStorage) {
      const parsed = JSON.parse(authStorage);
      return parsed.state?.token || null;
    }
  } catch (e) {
    console.error('解析 auth-storage 失败:', e);
  }
  return null;
}

// 从 auth-storage 解析 refresh token
function getRefreshTokenFromStorage(): string | null {
  try {
    const authStorage = localStorage.getItem('auth-storage');
    if (authStorage) {
      const parsed = JSON.parse(authStorage);
      return parsed.state?.refreshToken || null;
    }
  } catch (e) {
    console.error('解析 auth-storage 失败:', e);
  }
  return null;
}

// 更新存储中的 token
function updateTokens(accessToken: string, refreshToken: string) {
  try {
    const authStorage = localStorage.getItem('auth-storage');
    if (authStorage) {
      const parsed = JSON.parse(authStorage);
      parsed.state.token = accessToken;
      parsed.state.refreshToken = refreshToken;
      localStorage.setItem('auth-storage', JSON.stringify(parsed));
      return true;
    }
  } catch (e) {
    console.error('更新 token 失败:', e);
  }
  return false;
}

// 是否正在刷新 token
let isRefreshing = false;
// 等待 token 刷新的请求队列
let refreshSubscribers: Array<(token: string) => void> = [];

// 订阅 token 刷新
function subscribeTokenRefresh(callback: (token: string) => void) {
  refreshSubscribers.push(callback);
}

// 通知所有订阅者新 token
function onTokenRefreshed(newToken: string) {
  refreshSubscribers.forEach((callback) => callback(newToken));
  refreshSubscribers = [];
}

// 创建 axios 实例
const http = axios.create({
  baseURL: import.meta.env.VITE_API_BASE_URL || '/api/v1',
  timeout: 30000,
  headers: {
    'Content-Type': 'application/json',
  },
});

// 请求拦截器
http.interceptors.request.use(
  (config) => {
    // 从 zustand persist 存储中获取 token
    const token = getTokenFromStorage();
    if (token) {
      config.headers.Authorization = `Bearer ${token}`;
    }
    
    // 添加请求 ID
    config.headers['X-Request-ID'] = generateRequestId();
    
    return config;
  },
  (error) => {
    return Promise.reject(error);
  }
);

// 响应拦截器
http.interceptors.response.use(
  (response: AxiosResponse) => {
    // 统一处理响应
    const { data } = response;
    
    // 处理业务错误
    if (data.code !== 0 && data.code !== 200) {
      return Promise.reject(new Error(data.message || '请求失败'));
    }
    
    return data.data;
  },
  async (error: AxiosError) => {
    const config = error.config as AxiosRequestConfig & { retryCount?: number; _retry?: boolean };
    
    if (!config) {
      return Promise.reject(error);
    }
    
    const status = error.response?.status;
    
    // 401 未授权，尝试刷新 token
    if (status === 401 && !config._retry) {
      // 如果是刷新 token 的请求本身失败，直接跳转登录
      if (config.url?.includes('/auth/refresh')) {
        handleAuthError();
        return Promise.reject(error);
      }
      
      config._retry = true;
      
      // 如果正在刷新，等待刷新完成
      if (isRefreshing) {
        return new Promise((resolve) => {
          subscribeTokenRefresh((newToken: string) => {
            config.headers = config.headers || {};
            config.headers.Authorization = `Bearer ${newToken}`;
            resolve(http(config));
          });
        });
      }
      
      isRefreshing = true;
      
      try {
        const refreshToken = getRefreshTokenFromStorage();
        if (!refreshToken) {
          throw new Error('No refresh token');
        }
        
        // 调用刷新 token 接口
        const response = await axios.post(
          `${import.meta.env.VITE_API_BASE_URL || '/api/v1'}/auth/refresh`,
          { refresh_token: refreshToken },
          { headers: { 'Content-Type': 'application/json' } }
        );
        
        if (response.data.code === 0 || response.data.code === 200) {
          const { access_token, refresh_token } = response.data.data;
          
          // 更新存储
          updateTokens(access_token, refresh_token);
          
          // 通知等待的请求
          onTokenRefreshed(access_token);
          
          // 重试原请求
          config.headers = config.headers || {};
          config.headers.Authorization = `Bearer ${access_token}`;
          return http(config);
        }
      } catch (refreshError) {
        // 刷新失败，清除登录状态并跳转
        handleAuthError();
        return Promise.reject(refreshError);
      } finally {
        isRefreshing = false;
      }
    }
    
    // 重试逻辑（最多 3 次）
    config.retryCount = config.retryCount || 0;
    const maxRetries = 3;
    
    if (config.retryCount < maxRetries && shouldRetry(error)) {
      config.retryCount++;
      
      // 延迟重试
      await new Promise(resolve => setTimeout(resolve, 1000 * config.retryCount!));
      
      console.log(`Retrying request (${config.retryCount}/${maxRetries}):`, config.url);
      return http(config);
    }
    
    // 处理错误
    return handleError(error);
  }
);

// 处理认证错误
function handleAuthError() {
  message.error('登录已过期，请重新登录');
  localStorage.removeItem('auth-storage');
  window.location.href = '/login';
}

// 判断是否应该重试
function shouldRetry(error: AxiosError): boolean {
  const status = error.response?.status;
  // 仅在网络错误或 5xx 错误时重试
  return !status || status >= 500 || error.code === 'ECONNABORTED';
}

// 错误处理
function handleError(error: AxiosError) {
  const status = error.response?.status;
  const data = error.response?.data as any;
  
  switch (status) {
    case 400:
      message.error(data?.message || '请求参数错误');
      break;
    case 403:
      message.error(data?.message || '没有权限执行此操作');
      break;
    case 404:
      message.error(data?.message || '请求的资源不存在');
      break;
    case 429:
      message.error('请求过于频繁，请稍后再试');
      break;
    case 500:
    case 502:
    case 503:
      message.error('服务器错误，请稍后重试');
      break;
    default:
      if (error.code === 'ECONNABORTED') {
        message.error('请求超时，请检查网络连接');
      } else {
        message.error(data?.message || '网络错误');
      }
  }
  
  return Promise.reject(error);
}

// 生成请求 ID
function generateRequestId(): string {
  return `${Date.now()}-${Math.random().toString(36).substr(2, 9)}`;
}

// 防抖请求
export function debounceRequest<T>(
  config: AxiosRequestConfig,
  wait: number = 500
): Promise<T> {
  return new Promise((resolve, reject) => {
    const key = `${config.method}-${config.url}`;
    
    // 清除之前的定时器
    if ((debounceRequest as any).timers?.[key]) {
      clearTimeout((debounceRequest as any).timers[key]);
    }
    
    // 设置新的定时器
    const timer = setTimeout(() => {
      http.request(config)
        .then((data: any) => resolve(data as T))
        .catch(reject);
    }, wait);
    
    // 保存定时器
    if (!(debounceRequest as any).timers) {
      (debounceRequest as any).timers = {};
    }
    (debounceRequest as any).timers[key] = timer;
  });
}

export default http;
export { http };
