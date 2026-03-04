// 应用配置
export const config = {
  // API 基础地址
  apiBaseUrl: import.meta.env.VITE_API_BASE_URL || '/api/v1',
  
  // 应用域名
  appDomain: import.meta.env.VITE_APP_DOMAIN || window.location.origin,
  
  // 应用名称
  appName: import.meta.env.VITE_APP_NAME || '团圆寻亲',
  
  // 应用版本
  appVersion: import.meta.env.VITE_APP_VERSION || '1.0.0',
  
  // 上传文件基础URL
  uploadBaseUrl: `${import.meta.env.VITE_APP_DOMAIN || window.location.origin}/uploads`,
};

export default config;
