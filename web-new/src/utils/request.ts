import axios, { AxiosError, AxiosRequestConfig, AxiosResponse } from 'axios';
import { message } from 'antd';
import { useAuthStore } from '@/stores/auth';

// 创建axios实例
const request = axios.create({
  baseURL: import.meta.env.VITE_API_BASE_URL || '/api/v1',
  timeout: 30000,
  headers: {
    'Content-Type': 'application/json',
  },
});

// 请求拦截器
request.interceptors.request.use(
  (config) => {
    const token = useAuthStore.getState().token;
    if (token) {
      config.headers.Authorization = `Bearer ${token}`;
    }
    return config;
  },
  (error) => {
    return Promise.reject(error);
  }
);

// 响应拦截器
request.interceptors.response.use(
  (response: AxiosResponse) => {
    const { data } = response;
    
    // 如果响应中有code字段，按照统一格式处理
    if (data.code !== undefined) {
      if (data.code === 200) {
        return data.data;
      } else {
        message.error(data.message || '请求失败');
        return Promise.reject(new Error(data.message));
      }
    }
    
    return data;
  },
  (error: AxiosError) => {
    const { response } = error;
    
    if (response) {
      const { status, data } = response as AxiosResponse;
      
      switch (status) {
        case 401:
          message.error('登录已过期，请重新登录');
          useAuthStore.getState().logout();
          window.location.href = '/login';
          break;
        case 403:
          message.error('没有权限访问该资源');
          break;
        case 404:
          message.error('请求的资源不存在');
          break;
        case 500:
          message.error('服务器内部错误');
          break;
        default:
          message.error((data as any)?.message || '请求失败');
      }
    } else {
      message.error('网络连接失败，请检查网络');
    }
    
    return Promise.reject(error);
  }
);

// 封装请求方法
export const http = {
  get: <T>(url: string, config?: AxiosRequestConfig) =>
    request.get<T, T>(url, config),
  
  post: <T>(url: string, data?: unknown, config?: AxiosRequestConfig) =>
    request.post<T, T>(url, data, config),
  
  put: <T>(url: string, data?: unknown, config?: AxiosRequestConfig) =>
    request.put<T, T>(url, data, config),
  
  patch: <T>(url: string, data?: unknown, config?: AxiosRequestConfig) =>
    request.patch<T, T>(url, data, config),
  
  delete: <T>(url: string, config?: AxiosRequestConfig) =>
    request.delete<T, T>(url, config),
};

export default request;
