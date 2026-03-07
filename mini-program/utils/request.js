const app = getApp()

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
 * 统一请求封装
 * @param {Object} options 请求配置
 * @returns {Promise}
 */
const request = (options) => {
  return new Promise((resolve, reject) => {
    const token = wx.getStorageSync('token')
    // 直接使用生产环境 API 地址
    const baseUrl = 'https://cntuanyuan.com/api/v1'
    
    // 合并配置
    const config = {
      ...BASE_CONFIG,
      url: `${baseUrl}${options.url}`,
      method: options.method || 'GET',
      data: options.data,
      header: {
        ...BASE_CONFIG.header,
        ...options.header,
        'Authorization': token ? `Bearer ${token}` : ''
      },
      timeout: options.timeout || BASE_CONFIG.timeout
    }

    // 显示加载中（如果需要）
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
          // code: 0 或 200 表示成功
          if (data.code === 0 || data.code === 200) {
            resolve(data.data)
          } else {
            // 业务错误
            const errorMsg = data.message || '请求失败'
            
            // 未授权，token 过期
            if (data.code === 401) {
              handleTokenExpired(options, resolve, reject)
              return
            }
            
            // 显示错误提示（如果不是静默请求）
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
          // HTTP 401，token 过期
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
  
  refreshToken().then(() => {
    // 重试队列中的请求
    requestQueue.forEach(({ failedRequest, resolve }) => {
      request(failedRequest).then(resolve).catch(() => {})
    })
    requestQueue = []
  }).catch(() => {
    // 刷新失败，清除登录状态并跳转登录页
    requestQueue.forEach(({ reject }) => {
      reject(new Error('登录已过期'))
    })
    requestQueue = []
    
    // 清除登录数据
    const app = getApp()
    if (app) {
      app.clearLoginData()
    }
    
    // 跳转登录页
    wx.navigateTo({ url: '/pages/login/index' })
  }).finally(() => {
    isRefreshing = false
  })
}

/**
 * 刷新 Token
 */
function refreshToken() {
  return new Promise((resolve, reject) => {
    const refreshToken = wx.getStorageSync('refresh_token')
    
    if (!refreshToken) {
      reject(new Error('No refresh token'))
      return
    }

    const baseUrl = app ? app.globalData.apiBaseUrl : 'http://localhost:8080/api/v1'

    wx.request({
      url: `${baseUrl}/auth/refresh`,
      method: 'POST',
      data: { refresh_token: refreshToken },
      header: {
        'Content-Type': 'application/json'
      },
      success: (res) => {
        if (res.statusCode === 200 && (res.data.code === 0 || res.data.code === 200)) {
          const { access_token, refresh_token } = res.data.data
          
          // 更新存储
          wx.setStorageSync('token', access_token)
          wx.setStorageSync('refresh_token', refresh_token)
          
          // 更新全局数据
          if (app) {
            app.globalData.token = access_token
            app.globalData.refreshToken = refresh_token
          }
          
          resolve(res.data.data)
        } else {
          reject(new Error('Refresh failed'))
        }
      },
      fail: reject
    })
  })
}

/**
 * 上传文件
 * @param {String} url 上传地址
 * @param {String} filePath 文件路径
 * @param {String} name 文件字段名
 * @param {Object} formData 附加表单数据
 * @returns {Promise}
 */
const uploadFile = (url, filePath, name = 'file', formData = {}) => {
  return new Promise((resolve, reject) => {
    const token = wx.getStorageSync('token')
    const baseUrl = app ? app.globalData.apiBaseUrl : 'http://localhost:8080/api/v1'
    
    wx.showLoading({ title: '上传中...', mask: true })

    wx.uploadFile({
      url: `${baseUrl}${url}`,
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
 * @param {Array} files 文件路径数组
 * @param {String} url 上传地址
 * @param {Object} formData 附加表单数据
 * @returns {Promise}
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
