import axios from 'axios'
import { message } from 'antd'
import { useAuthStore } from '../stores/auth'

const request = axios.create({
  baseURL: '/api/v1',
  timeout: 30000,
  headers: {
    'Content-Type': 'application/json',
  },
})

// 请求拦截器
request.interceptors.request.use(
  (config) => {
    const token = useAuthStore.getState().token
    if (token) {
      config.headers.Authorization = `Bearer ${token}`
    }
    return config
  },
  (error) => {
    return Promise.reject(error)
  }
)

// 响应拦截器
request.interceptors.response.use(
  (response) => {
    const { code, message: msg, data } = response.data
    if (code !== 200 && code !== 201) {
      message.error(msg || '请求失败')
      return Promise.reject(new Error(msg))
    }
    return data
  },
  (error) => {
    const { response } = error
    if (response) {
      switch (response.status) {
        case 401:
          message.error('登录已过期，请重新登录')
          useAuthStore.getState().logout()
          window.location.href = '/login'
          break
        case 403:
          message.error('没有权限访问')
          break
        case 404:
          message.error('请求的资源不存在')
          break
        case 500:
          message.error('服务器内部错误')
          break
        default:
          message.error(response.data?.message || '请求失败')
      }
    } else {
      message.error('网络错误，请检查网络连接')
    }
    return Promise.reject(error)
  }
)

export default request
