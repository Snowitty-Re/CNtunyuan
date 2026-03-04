const { showConfirm } = require('../../utils/util')

Page({
  data: {
    settings: [
      {
        group: '通用',
        items: [
          { icon: '&#xe6b0;', title: '消息通知', type: 'switch', key: 'notification' },
          { icon: '&#xe6b1;', title: '清除缓存', type: 'action', key: 'clearCache' }
        ]
      },
      {
        group: '关于',
        items: [
          { icon: '&#xe6b2;', title: '帮助中心', type: 'navigate', url: '/pages/settings/help' },
          { icon: '&#xe6b3;', title: '关于我们', type: 'navigate', url: '/pages/settings/about' },
          { icon: '&#xe6b4;', title: '用户协议', type: 'navigate', url: '/pages/settings/agreement' },
          { icon: '&#xe6b5;', title: '隐私政策', type: 'navigate', url: '/pages/settings/privacy' }
        ]
      }
    ],
    cacheSize: '0MB'
  },

  onLoad() {
    this.calculateCacheSize()
  },

  onShow() {
    this.calculateCacheSize()
  },

  // 计算缓存大小
  calculateCacheSize() {
    try {
      const info = wx.getStorageInfoSync()
      const size = (info.currentSize / 1024).toFixed(2)
      this.setData({ cacheSize: `${size}MB` })
    } catch (e) {
      console.error('获取缓存信息失败:', e)
    }
  },

  // 设置项点击
  onSettingTap(e) {
    const item = e.currentTarget.dataset.item
    
    switch (item.type) {
      case 'navigate':
        wx.navigateTo({ url: item.url })
        break
      case 'action':
        this.handleAction(item.key)
        break
      case 'switch':
        // 处理开关
        break
    }
  },

  // 处理操作
  async handleAction(key) {
    switch (key) {
      case 'clearCache':
        const confirm = await showConfirm('清除缓存', '确定要清除本地缓存吗？')
        if (confirm) {
          try {
            wx.clearStorageSync()
            wx.showToast({
              title: '清除成功',
              icon: 'success'
            })
            this.setData({ cacheSize: '0MB' })
          } catch (e) {
            wx.showToast({
              title: '清除失败',
              icon: 'none'
            })
          }
        }
        break
    }
  },

  // 退出登录
  async logout() {
    const confirm = await showConfirm('退出登录', '确定要退出登录吗？')
    if (!confirm) return

    try {
      const app = getApp()
      if (app.logout) {
        await app.logout()
      }
      
      wx.reLaunch({ url: '/pages/login/index' })
    } catch (error) {
      console.error('退出失败:', error)
    }
  }
})
