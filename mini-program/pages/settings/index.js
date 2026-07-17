const { showConfirm } = require('../../utils/util')
const app = getApp()

Page({
  data: {
    settings: [
      {
        group: '通用',
        items: [
          { icon: 'notification', title: '消息通知（本地摘要）', type: 'switch', key: 'notification', value: true },
          { icon: 'delete', title: '清除缓存', type: 'action', key: 'clearCache' }
        ]
      },
      {
        group: '关于',
        items: [
          { icon: 'help', title: '帮助中心', type: 'navigate', url: '/pages/settings/help' },
          { icon: 'info', title: '关于我们', type: 'navigate', url: '/pages/settings/about' },
          { icon: 'file', title: '用户协议', type: 'navigate', url: '/pages/settings/agreement' },
          { icon: 'settings', title: '隐私政策', type: 'navigate', url: '/pages/settings/privacy' }
        ]
      }
    ],
    cacheSize: '0MB'
  },

  onLoad() {
    if (!app.ensureBusinessAuth || !app.ensureBusinessAuth()) return
    this.syncNotificationSwitch()
    this.calculateCacheSize()
  },

  onShow() {
    if (!app.ensureBusinessAuth || !app.ensureBusinessAuth()) return
    this.syncNotificationSwitch()
    this.calculateCacheSize()
  },

  syncNotificationSwitch() {
    const enabled = wx.getStorageSync('setting_notification')
    const on = enabled === '' || enabled === undefined || enabled === null ? true : !!enabled
    const settings = this.data.settings.map((group) => ({
      ...group,
      items: group.items.map((item) =>
        item.key === 'notification' ? { ...item, value: on } : item
      )
    }))
    this.setData({ settings })
  },

  // 计算缓存大小
  calculateCacheSize() {
    try {
      const info = wx.getStorageInfoSync()
      const kb = info.currentSize
      const display = kb < 1024 ? `${kb}KB` : `${(kb / 1024).toFixed(2)}MB`
      this.setData({ cacheSize: display })
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
        break
    }
  },

  onSwitchChange(e) {
    const key = e.currentTarget.dataset.key
    const next = !!e.detail.value
    if (key === 'notification') {
      wx.setStorageSync('setting_notification', next)
      this.syncNotificationSwitch()
      wx.showToast({
        title: next ? '已开启本地摘要' : '已关闭本地摘要',
        icon: 'none'
      })
    }
  },

  // 处理操作
  async handleAction(key) {
    switch (key) {
      case 'clearCache':
        const confirm = await showConfirm('清除缓存', '确定要清除本地缓存吗？（登录信息将保留）')
        if (confirm) {
          try {
            // 保留登录相关数据
            const token = wx.getStorageSync('token')
            const refreshToken = wx.getStorageSync('refresh_token')
            const userInfo = wx.getStorageSync('userInfo')
            const policyAgreement = wx.getStorageSync('policyAgreement')
            const settingNotification = wx.getStorageSync('setting_notification')
            wx.clearStorageSync()
            // 恢复登录数据与协议勾选
            if (token) wx.setStorageSync('token', token)
            if (refreshToken) wx.setStorageSync('refresh_token', refreshToken)
            if (userInfo) wx.setStorageSync('userInfo', userInfo)
            if (policyAgreement !== '' && policyAgreement !== undefined) {
              wx.setStorageSync('policyAgreement', policyAgreement)
            }
            if (settingNotification !== '' && settingNotification !== undefined) {
              wx.setStorageSync('setting_notification', settingNotification)
            }
            wx.showToast({
              title: '清除成功',
              icon: 'success'
            })
            this.calculateCacheSize()
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
