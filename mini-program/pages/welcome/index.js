const app = getApp()

Page({
  onShow() {
    // 已登录则进入业务首页
    if (app.isLoggedIn && app.isLoggedIn()) {
      wx.switchTab({ url: '/pages/index/index' })
    }
  },

  goLogin() {
    wx.navigateTo({ url: '/pages/login/index' })
  },

  goAgreement() {
    wx.navigateTo({ url: '/pages/settings/agreement' })
  },

  goPrivacy() {
    wx.navigateTo({ url: '/pages/settings/privacy' })
  },

  exitApp() {
    wx.showModal({
      title: '提示',
      content: '本小程序仅供认证志愿者使用。非目标用户可关闭小程序。',
      confirmText: '知道了',
      showCancel: true,
      cancelText: '关闭小程序',
      success: (res) => {
        if (res.cancel && wx.exitMiniProgram) {
          wx.exitMiniProgram({ fail: () => {} })
        }
      }
    })
  }
})
