const app = getApp()

Page({
  onShow() {
    if (!app.isLoggedIn || !app.isLoggedIn()) return

    // 已登录且已绑手机 → 业务首页；已登录未绑 → 强制绑定
    if (app.hasPhoneBound && app.hasPhoneBound()) {
      const user = app.getUserInfoSafe && app.getUserInfoSafe()
      const status = user && String(user.status || 'active').toLowerCase()
      if (status && status !== 'active') {
        wx.showModal({
          title: '账号待审批',
          content: '账号未激活，请等待管理员审批',
          showCancel: false,
          success: () => {
            if (app.clearLoginData) app.clearLoginData()
            // 停留说明页
          }
        })
        return
      }
      wx.switchTab({ url: '/pages/index/index' })
      return
    }

    wx.navigateTo({ url: '/pages/login/index?binding=1' })
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
