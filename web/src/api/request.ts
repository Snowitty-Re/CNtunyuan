import axios, { AxiosError } from 'axios';
import { message } from 'antd';
import { getAccessToken, getRefreshToken, setTokens, clearTokens } from '@/utils/token';
import { ERROR_CODE_MAP } from '@/utils/constants';

const request = axios.create({
  baseURL: import.meta.env.VITE_API_BASE_URL,
  timeout: 30000,
});

let isRefreshing = false;
let pendingRequests: Array<(token: string) => void> = [];

function onTokenRefreshed(token: string) {
  pendingRequests.forEach((cb) => cb(token));
  pendingRequests = [];
}

request.interceptors.request.use((config) => {
  const token = getAccessToken();
  if (token) {
    config.headers.Authorization = `Bearer ${token}`;
  }
  return config;
});

request.interceptors.response.use(
  (response) => {
    if (response.status === 204) {
      return null as never;
    }

    const { code, data, message: msg } = response.data;

    if (code === 0) {
      return data;
    }

    if (code === 401 || code === 1003 || code === 1004) {
      const refreshToken = getRefreshToken();
      if (refreshToken && !isRefreshing) {
        isRefreshing = true;
        return axios
          .post(`${import.meta.env.VITE_API_BASE_URL}/auth/refresh`, {
            refresh_token: refreshToken,
          })
          .then((res) => {
            const loginData = res.data?.data;
            if (loginData?.access_token) {
              setTokens(loginData.access_token, loginData.refresh_token);
              onTokenRefreshed(loginData.access_token);
              response.config.headers.Authorization = `Bearer ${loginData.access_token}`;
              return request(response.config);
            }
            throw new Error('refresh failed');
          })
          .catch(() => {
            clearTokens();
            window.location.href = '/login';
            return Promise.reject(new Error('登录已过期，请重新登录'));
          })
          .finally(() => {
            isRefreshing = false;
          });
      }

      if (isRefreshing) {
        return new Promise((resolve) => {
          pendingRequests.push((token: string) => {
            response.config.headers.Authorization = `Bearer ${token}`;
            resolve(request(response.config));
          });
        });
      }

      clearTokens();
      window.location.href = '/login';
      return Promise.reject(new Error('登录已过期'));
    }

    const errorMsg = ERROR_CODE_MAP[code] || msg || '操作失败';
    message.error(errorMsg);
    return Promise.reject(new Error(errorMsg));
  },
  (error: AxiosError) => {
    if (error.response) {
      const status = error.response.status;
      if (status === 401) {
        clearTokens();
        window.location.href = '/login';
      } else if (status === 403) {
        message.error('没有权限执行此操作');
      } else if (status === 404) {
        message.error('请求的资源不存在');
      } else if (status >= 500) {
        message.error('服务器错误，请稍后重试');
      }
    } else if (error.code === 'ECONNABORTED') {
      message.error('请求超时，请稍后重试');
    } else {
      message.error('网络异常，请检查网络连接');
    }
    return Promise.reject(error);
  }
);

export default request;
