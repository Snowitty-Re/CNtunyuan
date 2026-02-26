App({
  globalData: {
    userInfo: null,
    token: null,
    apiBaseUrl: 'https://api.cntunyuan.com/api/v1'
  },

  onLaunch() {
    // 初始化云开发
    if (!wx.cloud) {
      console.error('请使用 2.2.3 或以上的基础库以使用云能力')
    } else {
      wx.cloud.init({
        env: 'your-cloud-env-id',
        traceUser: true,
      })
    }

    // 检查登录状态
    this.checkLoginStatus()
  },

  // 检查登录状态
  checkLoginStatus() {
    const token = wx.getStorageSync('token')
    if (token) {
      this.globalData.token = token
      this.getUserInfo()
    }
  },

  // 微信登录
  wxLogin() {
    return new Promise((resolve, reject) => {
      wx.login({
        success: (res) => {
          if (res.code) {
            // 调用后端登录接口
            wx.request({
              url: `${this.globalData.apiBaseUrl}/auth/wechat-login`,
              method: 'POST',
              data: {
                code: res.code
              },
              success: (result) => {
                if (result.data.code === 200) {
                  const { token, refresh_token } = result.data.data
                  wx.setStorageSync('token', token)
                  wx.setStorageSync('refresh_token', refresh_token)
                  this.globalData.token = token
                  resolve(result.data.data)
                } else {
                  reject(new Error(result.data.message))
                }
              },
              fail: reject
            })
          } else {
            reject(new Error('登录失败'))
          }
        },
        fail: reject
      })
    })
  },

  // 获取用户信息
  getUserInfo() {
    return new Promise((resolve, reject) => {
      wx.request({
        url: `${this.globalData.apiBaseUrl}/auth/me`,
        header: {
          'Authorization': `Bearer ${this.globalData.token}`
        },
        success: (res) => {
          if (res.data.code === 200) {
            this.globalData.userInfo = res.data.data
            resolve(res.data.data)
          } else {
            reject(new Error(res.data.message))
          }
        },
        fail: reject
      })
    })
  },

  // 请求封装
  request(options) {
    return new Promise((resolve, reject) => {
      const token = this.globalData.token
      wx.request({
        url: `${this.globalData.apiBaseUrl}${options.url}`,
        method: options.method || 'GET',
        data: options.data,
        header: {
          'Content-Type': 'application/json',
          'Authorization': token ? `Bearer ${token}` : ''
        },
        success: (res) => {
          if (res.statusCode === 401) {
            // Token过期，尝试刷新
            this.refreshToken().then(() => {
              // 重试原请求
              this.request(options).then(resolve).catch(reject)
            }).catch(() => {
              // 刷新失败，跳转登录
              wx.navigateTo({ url: '/pages/login/index' })
              reject(new Error('登录已过期'))
            })
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
  },

  // 刷新Token
  refreshToken() {
    return new Promise((resolve, reject) => {
      const refreshToken = wx.getStorageSync('refresh_token')
      if (!refreshToken) {
        reject(new Error('No refresh token'))
        return
      }

      wx.request({
        url: `${this.globalData.apiBaseUrl}/auth/refresh`,
        method: 'POST',
        data: { refresh_token: refreshToken },
        success: (res) => {
          if (res.data.code === 200) {
            const { token, refresh_token } = res.data.data
            wx.setStorageSync('token', token)
            wx.setStorageSync('refresh_token', refresh_token)
            this.globalData.token = token
            resolve(res.data.data)
          } else {
            reject(new Error(res.data.message))
          }
        },
        fail: reject
      })
    })
  }
})
