const { get, post } = require('../../utils/request')
const { showConfirm, showSuccess, showToast } = require('../../utils/util')
const app = getApp()

Page({
  data: {
    userInfo: {},
    stats: {
      taskCount: 0,
      caseCount: 0,
      dialectCount: 0,
      score: 0
    },
    roleMap: {
      super_admin: '超级管理员',
      admin: '管理员',
      manager: '管理者',
      volunteer: '志愿者'
    },
    menuList: [
      { icon: 'task', text: '我的任务', url: '/pages/tasks/my', badge: 0 },
      { icon: 'case', text: '我的案件', url: '/pages/cases/list', badge: 0 },
      { icon: 'dialect', text: '方言录音', url: '/pages/dialect/list', badge: 0 },
      { icon: 'certificate', text: '志愿者证书', url: '', badge: 0 }
    ]
  },

  onLoad() {
    this.loadUserInfo()
    this.loadStats()
  },

  onShow() {
    this.loadUserInfo()
    this.loadStats()
  },

  async loadUserInfo() {
    try {
      let userInfo = await app.getUserInfo() || {}
      userInfo.avatar = userInfo.avatar || 'https://picsum.photos/100/100'
      userInfo.nickname = userInfo.nickname || '志愿者'
      userInfo.phone = userInfo.phone || ''
      userInfo.email = userInfo.email || ''
      userInfo.real_name = userInfo.real_name || ''
      
      // 更新本地存储
      wx.setStorageSync('userInfo', userInfo)
      
      this.setData({ userInfo })
    } catch (error) {
      console.error('加载用户信息失败:', error)
    }
  },

  async loadStats() {
    try {
      // 获取真实统计数据
      const [taskStats, caseStats, dialectStats] = await Promise.all([
        get('/tasks/statistics').catch(() => ({ total: 0 })),
        get('/missing-persons/statistics').catch(() => ({ total: 0 })),
        get('/dialects/statistics').catch(() => ({ total: 0 }))
      ])
      
      this.setData({
        stats: {
          taskCount: taskStats.total || 0,
          caseCount: caseStats.total || 0,
          dialectCount: dialectStats.total || 0,
          score: this.calculateScore(taskStats.total, caseStats.total, dialectStats.total)
        }
      })
      
      // 更新菜单徽章
      this.updateMenuBadge(0, taskStats.pending || 0)
    } catch (error) {
      console.error('加载统计失败:', error)
    }
  },

  calculateScore(tasks, cases, dialects) {
    // 简单的积分计算规则
    return (tasks * 10) + (cases * 20) + (dialects * 15)
  },

  updateMenuBadge(index, count) {
    const menuList = this.data.menuList
    menuList[index].badge = count
    this.setData({ menuList })
  },

  // 更换头像
  changeAvatar() {
    wx.chooseMedia({
      count: 1,
      mediaType: ['image'],
      sourceType: ['album', 'camera'],
      success: async (res) => {
        try {
          const tempFilePath = res.tempFiles[0].tempFilePath
          // 这里应该调用上传接口
          // const result = await uploadFile('/upload', tempFilePath)
          showSuccess('头像更新成功')
        } catch (error) {
          showToast('上传失败')
        }
      }
    })
  },

  // 编辑资料
  editProfile() {
    wx.navigateTo({ url: '/pages/volunteer/edit-profile' })
  },

  // 菜单点击
  onMenuTap(e) {
    const { url } = e.currentTarget.dataset
    if (url) {
      wx.navigateTo({ url })
    } else {
      showToast('功能开发中')
    }
  },

  // 功能菜单点击
  onFunctionTap(e) {
    const { type } = e.currentTarget.dataset
    switch(type) {
      case 'notification':
        wx.navigateTo({ url: '/pages/notification/list' })
        break
      case 'settings':
        wx.navigateTo({ url: '/pages/settings/index' })
        break
      case 'help':
        wx.navigateTo({ url: '/pages/settings/help' })
        break
      case 'about':
        wx.navigateTo({ url: '/pages/settings/about' })
        break
    }
  },

  // 复制ID
  copyId() {
    wx.setClipboardData({
      data: this.data.userInfo.id || '',
      success: () => {
        showSuccess('已复制用户ID')
      }
    })
  },

  async logout() {
    const confirm = await showConfirm('确认退出', '退出后需要重新登录')
    if (confirm) {
      try {
        await post('/auth/logout')
      } catch (error) {
        console.error('退出登录失败:', error)
      }
      
      // 清除登录信息
      wx.clearStorageSync()
      
      app.globalData.token = null
      app.globalData.userInfo = null
      
      // 跳转到登录页
      wx.reLaunch({ url: '/pages/login/index' })
    }
  }
})
