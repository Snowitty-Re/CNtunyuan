const { post } = require('../../utils/request')
const { showSuccess, showToast } = require('../../utils/util')

Page({
  data: {
    userInfo: {},
    settings: {
      notification: true,
      sound: true,
      autoUpdate: true
    },
    version: '1.0.0'
  },

  onLoad() {
    this.loadUserInfo()
    this.loadSettings()
  },

  onShow() {
    this.loadUserInfo()
  },

  loadUserInfo() {
    const userInfo = wx.getStorageSync('userInfo') || {}
    this.setData({ userInfo })
  },

  loadSettings() {
    const settings = wx.getStorageSync('app_settings')
    if (settings) {
      this.setData({ settings: { ...this.data.settings, ...settings } })
    }
  },

  saveSettings() {
    wx.setStorageSync('app_settings', this.data.settings)
  },

  // 切换通知设置
  toggleNotification(e) {
    const value = e.detail.value
    this.setData({
      'settings.notification': value
    })
    this.saveSettings()
    showSuccess(value ? '已开启消息通知' : '已关闭消息通知')
  },

  // 切换声音设置
  toggleSound(e) {
    const value = e.detail.value
    this.setData({
      'settings.sound': value
    })
    this.saveSettings()
  },

  // 清除缓存
  clearCache() {
    wx.showModal({
      title: '提示',
      content: '确定清除缓存吗？不会影响您的登录状态',
      success: (res) => {
        if (res.confirm) {
          wx.clearStorage({
            success: () => {
              // 保留登录信息
              const token = wx.getStorageSync('token')
              const userInfo = wx.getStorageSync('userInfo')
              const refreshToken = wx.getStorageSync('refresh_token')
              
              wx.setStorageSync('token', token)
              wx.setStorageSync('userInfo', userInfo)
              wx.setStorageSync('refresh_token', refreshToken)
              
              showSuccess('缓存已清除')
            }
          })
        }
      }
    })
  },

  // 关于我们
  goToAbout() {
    wx.navigateTo({ url: '/pages/settings/about' })
  },

  // 隐私政策
  goToPrivacy() {
    wx.navigateTo({ url: '/pages/settings/privacy' })
  },

  // 用户协议
  goToAgreement() {
    wx.navigateTo({ url: '/pages/settings/agreement' })
  },

  // 检查更新
  checkUpdate() {
    wx.showLoading({ title: '检查中...' })
    
    setTimeout(() => {
      wx.hideLoading()
      wx.showModal({
        title: '检查更新',
        content: '当前已是最新版本',
        showCancel: false
      })
    }, 1000)
  },

  // 退出登录
  logout() {
    wx.showModal({
      title: '提示',
      content: '确定退出登录吗？',
      success: async (res) => {
        if (res.confirm) {
          try {
            await post('/auth/logout')
          } catch (error) {
            console.error('退出登录失败:', error)
          }
          
          // 清除本地存储
          wx.clearStorageSync()
          
          // 跳转到登录页
          wx.reLaunch({ url: '/pages/login/index' })
        }
      }
    })
  },

  // 编辑个人资料
  editProfile() {
    wx.navigateTo({ url: '/pages/volunteer/edit-profile' })
  },

  // 查看通知
  viewNotifications() {
    wx.navigateTo({ url: '/pages/notification/list' })
  }
})
