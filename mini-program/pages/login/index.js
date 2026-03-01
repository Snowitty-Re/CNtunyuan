const app = getApp()
const { showSuccess, showError } = require('../../utils/util')

Page({
  data: {
    loading: false,
    phone: '',
    password: '',
    isPasswordLogin: false
  },

  onLoad() {
    // 检查是否已登录
    const token = wx.getStorageSync('token')
    if (token) {
      wx.switchTab({ url: '/pages/index/index' })
    }
  },

  onPhoneInput(e) {
    this.setData({ phone: e.detail.value })
  },

  onPasswordInput(e) {
    this.setData({ password: e.detail.value })
  },

  // 微信一键登录
  async wxLogin() {
    this.setData({ loading: true })
    try {
      const result = await app.wxLogin()
      showSuccess('登录成功')
      
      // 获取用户信息
      await app.getUserInfo()
      
      wx.switchTab({ url: '/pages/index/index' })
    } catch (error) {
      console.error('登录失败:', error)
      showError('登录失败')
    } finally {
      this.setData({ loading: false })
    }
  },

  // 账号密码登录
  async passwordLogin() {
    const { phone, password } = this.data
    
    if (!phone) {
      showError('请输入手机号')
      return
    }
    if (!password) {
      showError('请输入密码')
      return
    }

    this.setData({ loading: true })
    
    wx.request({
      url: `${app.globalData.apiBaseUrl}/auth/login`,
      method: 'POST',
      data: { phone, password },
      success: (res) => {
        if (res.data.code === 200) {
          const { token, refresh_token, user } = res.data.data
          wx.setStorageSync('token', token)
          wx.setStorageSync('refresh_token', refresh_token)
          wx.setStorageSync('userInfo', user)
          app.globalData.token = token
          app.globalData.userInfo = user
          
          showSuccess('登录成功')
          wx.switchTab({ url: '/pages/index/index' })
        } else {
          showError(res.data.message || '登录失败')
        }
      },
      fail: () => {
        showError('网络错误')
      },
      complete: () => {
        this.setData({ loading: false })
      }
    })
  },

  // 切换登录方式
  toggleLoginType() {
    this.setData({ isPasswordLogin: !this.data.isPasswordLogin })
  }
})
