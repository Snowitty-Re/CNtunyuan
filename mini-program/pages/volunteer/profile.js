const { get } = require('../../utils/request')
const { showConfirm, showSuccess } = require('../../utils/util')
const app = getApp()

Page({
  data: {
    userInfo: {},
    stats: {
      taskCount: 0,
      caseCount: 0,
      dialectCount: 0,
      score: 0
    }
  },

  onLoad() {
    this.loadUserInfo()
    this.loadStats()
  },

  onShow() {
    this.loadUserInfo()
  },

  async loadUserInfo() {
    try {
      const userInfo = await app.getUserInfo()
      this.setData({ userInfo })
    } catch (error) {
      console.error('加载用户信息失败:', error)
    }
  },

  async loadStats() {
    try {
      // 这里应该调用统计接口
      this.setData({
        stats: {
          taskCount: 12,
          caseCount: 8,
          dialectCount: 5,
          score: 1560
        }
      })
    } catch (error) {
      console.error('加载统计失败:', error)
    }
  },

  changeAvatar() {
    wx.chooseMedia({
      count: 1,
      mediaType: ['image'],
      sourceType: ['album', 'camera'],
      success: (res) => {
        // 上传头像
        showSuccess('头像更新成功')
      }
    })
  },

  goToMyCases() {
    wx.switchTab({ url: '/pages/cases/list' })
  },

  goToMyTasks() {
    wx.switchTab({ url: '/pages/tasks/list' })
  },

  goToMyDialects() {
    wx.navigateTo({ url: '/pages/dialect/list' })
  },

  goToCertificates() {
    wx.showToast({
      title: '功能开发中',
      icon: 'none'
    })
  },

  goToSettings() {
    wx.navigateTo({ url: '/pages/volunteer/settings' })
  },

  goToHelp() {
    wx.navigateTo({ url: '/pages/volunteer/help' })
  },

  goToAbout() {
    wx.navigateTo({ url: '/pages/volunteer/about' })
  },

  async logout() {
    const confirm = await showConfirm('确认退出', '退出后需要重新登录')
    if (confirm) {
      // 清除登录信息
      wx.removeStorageSync('token')
      wx.removeStorageSync('refresh_token')
      wx.removeStorageSync('userInfo')
      
      app.globalData.token = null
      app.globalData.userInfo = null
      
      // 重新登录
      wx.reLaunch({ url: '/pages/index/index' })
    }
  }
})
