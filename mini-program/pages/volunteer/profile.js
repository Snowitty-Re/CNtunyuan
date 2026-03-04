const userService = require('../../services/user')
const taskService = require('../../services/task')
const { showConfirm, showSuccess, showToast } = require('../../utils/util')
const app = getApp()

Page({
  data: {
    // 用户信息
    userInfo: {
      id: '',
      avatar: '',
      nickname: '',
      realName: '',
      role: 'volunteer',
      points: 0,
      org: null
    },
    
    // 角色映射
    roleMap: {
      super_admin: '超级管理员',
      admin: '管理员',
      manager: '管理者',
      volunteer: '志愿者'
    },
    
    // 角色等级颜色
    roleColorMap: {
      super_admin: '#E74C3C',
      admin: '#E67E22',
      manager: '#3498DB',
      volunteer: '#27AE60'
    },
    
    // 统计数据
    stats: {
      taskCount: 0,
      caseCount: 0,
      dialectCount: 0,
      points: 0
    },
    
    // 功能菜单
    menuList: [
      { icon: 'edit', text: '编辑资料', url: '/pages/volunteer/edit-profile', type: 'navigate' },
      { icon: 'task', text: '我的任务', url: '/pages/tasks/my', type: 'navigate' },
      { icon: 'notification', text: '消息通知', url: '/pages/notification/list', type: 'navigate', badge: 0 },
      { icon: 'certificate', text: '志愿者证书', url: '', type: 'toast' },
      { icon: 'settings', text: '设置', url: '/pages/settings/index', type: 'navigate' }
    ]
  },

  onLoad() {
    this.loadData()
  },

  onShow() {
    this.loadData()
  },

  onPullDownRefresh() {
    this.loadData().finally(() => {
      wx.stopPullDownRefresh()
    })
  },

  // 加载所有数据
  async loadData() {
    try {
      await Promise.all([
        this.loadUserInfo(),
        this.loadStats()
      ])
    } catch (error) {
      console.error('加载数据失败:', error)
    }
  },

  // 加载用户信息
  async loadUserInfo() {
    try {
      const userInfo = await app.getUserInfo() || wx.getStorageSync('userInfo') || {}
      const profile = await userService.getProfile().catch(() => ({}))
      
      const mergedUserInfo = {
        ...userInfo,
        ...profile,
        id: userInfo.id || profile.id || '',
        avatar: userInfo.avatar || profile.avatar || '/assets/images/avatar-default.png',
        nickname: userInfo.nickname || profile.nickname || '志愿者',
        realName: userInfo.real_name || profile.real_name || '',
        role: userInfo.role || profile.role || 'volunteer',
        points: userInfo.points || profile.points || 0,
        org: userInfo.org || profile.org || null
      }
      
      this.setData({ userInfo: mergedUserInfo })
      wx.setStorageSync('userInfo', mergedUserInfo)
    } catch (error) {
      console.error('加载用户信息失败:', error)
    }
  },

  // 加载统计数据
  async loadStats() {
    try {
      // 并行获取各项统计
      const [userStats, taskStats] = await Promise.all([
        userService.getStats().catch(() => ({})),
        taskService.getStats().catch(() => ({}))
      ])
      
      this.setData({
        stats: {
          taskCount: userStats.task_count || taskStats.total || 0,
          caseCount: userStats.case_count || 0,
          dialectCount: userStats.dialect_count || 0,
          points: userStats.points || this.data.userInfo.points || 0
        }
      })
    } catch (error) {
      console.error('加载统计失败:', error)
    }
  },

  // 点击统计卡片
  onStatTap(e) {
    const { type } = e.currentTarget.dataset
    switch (type) {
      case 'task':
        wx.navigateTo({ url: '/pages/tasks/my' })
        break
      case 'case':
        wx.switchTab({ url: '/pages/cases/list' })
        break
      case 'dialect':
        wx.navigateTo({ url: '/pages/dialect/list' })
        break
      case 'points':
        showToast('积分可用于兑换志愿者福利')
        break
    }
  },

  // 菜单点击
  onMenuTap(e) {
    const { index } = e.currentTarget.dataset
    const menu = this.data.menuList[index]
    
    if (!menu) return
    
    switch (menu.type) {
      case 'navigate':
        if (menu.url) {
          wx.navigateTo({ url: menu.url })
        }
        break
      case 'toast':
        showToast('功能开发中，敬请期待')
        break
      case 'switchTab':
        if (menu.url) {
          wx.switchTab({ url: menu.url })
        }
        break
    }
  },

  // 复制用户ID
  copyUserId() {
    const { id } = this.data.userInfo
    if (!id) {
      showToast('用户ID获取失败')
      return
    }
    
    wx.setClipboardData({
      data: id,
      success: () => {
        showSuccess('已复制用户ID')
      }
    })
  },

  // 退出登录
  async logout() {
    const confirm = await showConfirm('确认退出', '退出后需要重新登录')
    if (!confirm) return
    
    try {
      // 调用退出登录接口
      await userService.logout?.().catch(() => {})
    } catch (error) {
      console.error('退出登录接口调用失败:', error)
    }
    
    // 清除本地数据
    wx.clearStorageSync()
    app.globalData.token = null
    app.globalData.userInfo = null
    
    // 跳转到登录页
    wx.reLaunch({ url: '/pages/login/index' })
  }
})
