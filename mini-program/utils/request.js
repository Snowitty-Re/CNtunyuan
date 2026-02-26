const app = getApp()

// 请求封装
const request = (options) => {
  return new Promise((resolve, reject) => {
    const token = wx.getStorageSync('token')
    
    wx.request({
      url: `${app.globalData.apiBaseUrl}${options.url}`,
      method: options.method || 'GET',
      data: options.data,
      header: {
        'Content-Type': 'application/json',
        'Authorization': token ? `Bearer ${token}` : ''
      },
      success: (res) => {
        if (res.statusCode === 401) {
          // Token过期
          wx.navigateTo({ url: '/pages/login/index' })
          reject(new Error('登录已过期'))
        } else if (res.data.code === 200 || res.data.code === 201) {
          resolve(res.data.data)
        } else {
          wx.showToast({
            title: res.data.message || '请求失败',
            icon: 'none'
          })
          reject(new Error(res.data.message))
        }
      },
      fail: (err) => {
        wx.showToast({
          title: '网络错误',
          icon: 'none'
        })
        reject(err)
      }
    })
  })
}

// GET请求
const get = (url, params = {}) => {
  return request({ url, method: 'GET', data: params })
}

// POST请求
const post = (url, data = {}) => {
  return request({ url, method: 'POST', data })
}

// PUT请求
const put = (url, data = {}) => {
  return request({ url, method: 'PUT', data })
}

// DELETE请求
const del = (url, params = {}) => {
  return request({ url, method: 'DELETE', data: params })
}

module.exports = {
  request,
  get,
  post,
  put,
  del
}
