/**
 * 请求工具
 * 统一处理请求、响应、错误和 Token 刷新
 */

// 生产环境 API 地址 - 统一配置在这里
const API_BASE_URL = 'https://cntuanyuan.com/api/v1'

// 请求队列（用于 token 刷新时暂存请求）
let requestQueue = []
let isRefreshing = false

// 基础配置
const BASE_CONFIG = {
  timeout: 30000,
  header: {
    'Content-Type': 'application/json'
  }
}

/**
 * 获取 token
 */
function getToken() {
  return wx.getStorageSync('token') || ''
}

/**
 * 获取 refresh token
 */
function getRefreshToken() {
  return wx.getStorageSync('refresh_token') || ''
}

/**
 * 更新 tokens
 */
function updateTokens(accessToken, refreshToken) {
  wx.setStorageSync('token', accessToken)
  wx.setStorageSync('refresh_token', refreshToken)
  const app = getApp()
  if (app) {
    app.globalData.token = accessToken
    app.globalData.refreshToken = refreshToken
  }
}

/**
 * 处理认证失败
 */
function handleAuthFail() {
  wx.removeStorageSync('token')
  wx.removeStorageSync('refresh_token')
  wx.removeStorageSync('userInfo')
  
  const app = getApp()
  if (app) {
    app.globalData.token = null
    app.globalData.refreshToken = null
    app.globalData.userInfo = null
    app.globalData.isLogin = false
  }
  
  wx.showToast({
    title: '登录已过期',
    icon: 'none'
  })
  
  setTimeout(() => {
    wx.reLaunch({ url: '/pages/login/index' })
  }, 1500)
}

/**
 * 刷新 token
 */
function refreshToken() {
  return new Promise((resolve, reject) => {
    const refreshToken = getRefreshToken()
    
    if (!refreshToken) {
      reject(new Error('No refresh token'))
      return
    }

    wx.request({
      url: `${API_BASE_URL}/auth/refresh`,
      method: 'POST',
      data: { refresh_token: refreshToken },
      header: {
        'Content-Type': 'application/json'
      },
      success: (res) => {
        if (res.statusCode === 200 && (res.data.code === 0 || res.data.code === 200)) {
          const { access_token, refresh_token } = res.data.data
          updateTokens(access_token, refresh_token)
          resolve(access_token)
        } else {
          reject(new Error('Refresh failed'))
        }
      },
      fail: reject
    })
  })
}

/**
 * 统一请求封装
 */
const request = (options) => {
  return new Promise((resolve, reject) => {
    const token = getToken()
    
    // 合并配置
    const config = {
      ...BASE_CONFIG,
      url: `${API_BASE_URL}${options.url}`,
      method: options.method || 'GET',
      data: options.data,
      header: {
        ...BASE_CONFIG.header,
        ...options.header,
        'Authorization': token ? `Bearer ${token}` : ''
      },
      timeout: options.timeout || BASE_CONFIG.timeout
    }

    // 显示加载中
    if (options.loading !== false) {
      wx.showLoading({ 
        title: options.loadingText || '加载中...',
        mask: true
      })
    }

    wx.request({
      ...config,
      success: (res) => {
        // 隐藏加载
        if (options.loading !== false) {
          wx.hideLoading()
        }

        // 处理 HTTP 状态码
        if (res.statusCode === 200 || res.statusCode === 201) {
          const data = res.data
          
          // 处理业务状态码
          if (data.code === 0 || data.code === 200) {
            resolve(data.data)
          } else if (data.code === 401) {
            // Token 过期，尝试刷新
            handleTokenExpired(options, resolve, reject)
          } else {
            // 业务错误
            const errorMsg = data.message || '请求失败'
            
            if (options.silent !== true) {
              wx.showToast({
                title: errorMsg,
                icon: 'none',
                duration: 2000
              })
            }
            
            reject(new Error(errorMsg))
          }
        } else if (res.statusCode === 401) {
          // HTTP 401，尝试刷新 token
          handleTokenExpired(options, resolve, reject)
        } else {
          // HTTP 错误
          const errorMsg = `请求失败: ${res.statusCode}`
          
          if (options.silent !== true) {
            wx.showToast({
              title: errorMsg,
              icon: 'none'
            })
          }
          
          reject(new Error(errorMsg))
        }
      },
      fail: (err) => {
        // 隐藏加载
        if (options.loading !== false) {
          wx.hideLoading()
        }

        // 网络错误处理
        let errorMsg = '网络错误'
        
        if (err.errMsg && err.errMsg.includes('timeout')) {
          errorMsg = '请求超时，请稍后重试'
        } else if (err.errMsg && err.errMsg.includes('fail')) {
          errorMsg = '网络连接失败，请检查网络'
        }

        if (options.silent !== true) {
          wx.showToast({
            title: errorMsg,
            icon: 'none',
            duration: 2000
          })
        }

        reject(new Error(errorMsg))
      }
    })
  })
}

/**
 * 处理 Token 过期
 */
function handleTokenExpired(failedRequest, resolve, reject) {
  // 将失败请求加入队列
  requestQueue.push({ failedRequest, resolve, reject })
  
  if (isRefreshing) {
    return
  }
  
  isRefreshing = true
  
  refreshToken().then((newToken) => {
    // 重试队列中的请求
    requestQueue.forEach(({ failedRequest, resolve }) => {
      // 更新请求头中的 token
      failedRequest.header = failedRequest.header || {}
      failedRequest.header['Authorization'] = `Bearer ${newToken}`
      request(failedRequest).then(resolve).catch(() => {})
    })
    requestQueue = []
  }).catch(() => {
    // 刷新失败，清除登录状态并跳转登录页
    requestQueue.forEach(({ reject }) => {
      reject(new Error('登录已过期'))
    })
    requestQueue = []
    
    handleAuthFail()
  }).finally(() => {
    isRefreshing = false
  })
}

/**
 * 上传文件
 */
const uploadFile = (url, filePath, name = 'file', formData = {}) => {
  return new Promise((resolve, reject) => {
    const token = getToken()
    
    wx.showLoading({ title: '上传中...', mask: true })

    wx.uploadFile({
      url: `${API_BASE_URL}${url}`,
      filePath,
      name,
      formData,
      header: {
        'Authorization': token ? `Bearer ${token}` : ''
      },
      success: (res) => {
        wx.hideLoading()
        
        if (res.statusCode === 200 || res.statusCode === 201) {
          try {
            const data = JSON.parse(res.data)
            if (data.code === 0 || data.code === 200) {
              resolve(data.data)
            } else if (data.code === 401) {
              // Token 过期
              handleAuthFail()
              reject(new Error('登录已过期'))
            } else {
              wx.showToast({
                title: data.message || '上传失败',
                icon: 'none'
              })
              reject(new Error(data.message))
            }
          } catch (e) {
            resolve(res.data)
          }
        } else if (res.statusCode === 401) {
          handleAuthFail()
          reject(new Error('登录已过期'))
        } else {
          wx.showToast({
            title: `上传失败: ${res.statusCode}`,
            icon: 'none'
          })
          reject(new Error(`Upload failed: ${res.statusCode}`))
        }
      },
      fail: (err) => {
        wx.hideLoading()
        wx.showToast({
          title: '上传失败，请检查网络',
          icon: 'none'
        })
        reject(err)
      }
    })
  })
}

/**
 * 批量上传文件
 */
const uploadFiles = (files, url = '/upload/batch', formData = {}) => {
  const uploadPromises = files.map(filePath => uploadFile(url, filePath, 'files', formData))
  return Promise.all(uploadPromises)
}

// HTTP 方法封装
const get = (url, params = {}, options = {}) => {
  return request({ url, method: 'GET', data: params, ...options })
}

const post = (url, data = {}, options = {}) => {
  return request({ url, method: 'POST', data, ...options })
}

const put = (url, data = {}, options = {}) => {
  return request({ url, method: 'PUT', data, ...options })
}

const del = (url, params = {}, options = {}) => {
  return request({ url, method: 'DELETE', data: params, ...options })
}

module.exports = {
  request,
  get,
  post,
  put,
  del,
  uploadFile,
  uploadFiles
}
