const { get, post } = require('../../utils/request')
const { formatDate, showSuccess } = require('../../utils/util')

Page({
  data: {
    userInfo: {},
    stats: {
      taskCount: 0,
      caseCount: 0,
      dialectCount: 0,
      completedCount: 0
    },
    todos: [],
    myTasks: [],
    activities: [],
    roleMap: {
      super_admin: '超级管理员',
      admin: '管理员',
      manager: '管理者',
      volunteer: '志愿者'
    },
    priorityMap: {
      urgent: '紧急',
      high: '高',
      normal: '普通',
      low: '低'
    },
    taskTypeMap: {
      search: '实地寻访',
      call: '电话核实',
      info_collect: '信息收集',
      dialect_record: '方言录制',
      coordination: '协调沟通',
      other: '其他'
    }
  },

  onLoad() {
    this.loadUserInfo()
  },

  onShow() {
    this.loadStats()
    this.loadMyTasks()
    this.loadActivities()
  },

  async loadUserInfo() {
    try {
      let userInfo = wx.getStorageSync('userInfo') || {}
      // 获取最新用户信息
      const freshUserInfo = await get('/auth/me')
      if (freshUserInfo) {
        userInfo = { ...userInfo, ...freshUserInfo }
        wx.setStorageSync('userInfo', userInfo)
      }
      userInfo.avatar = userInfo.avatar || 'https://picsum.photos/100/100'
      userInfo.nickname = userInfo.nickname || '志愿者'
      this.setData({ userInfo })
    } catch (error) {
      console.error('加载用户信息失败:', error)
    }
  },

  async loadStats() {
    try {
      // 获取任务统计
      const taskStats = await get('/tasks/statistics')
      this.setData({
        stats: {
          taskCount: taskStats.total || 0,
          completedCount: taskStats.completed || 0,
          pendingCount: taskStats.pending || 0,
          processingCount: taskStats.processing || 0
        }
      })
    } catch (error) {
      console.error('加载统计失败:', error)
    }
  },

  async loadMyTasks() {
    try {
      const result = await get('/tasks/my', { status: 'assigned' })
      const myTasks = result.slice(0, 5).map(item => ({
        ...item,
        deadline: item.deadline ? formatDate(item.deadline) : null
      }))
      this.setData({ myTasks })
    } catch (error) {
      console.error('加载我的任务失败:', error)
    }
  },

  async loadActivities() {
    try {
      // 获取任务日志作为动态
      const result = await get('/tasks/my', { status: 'completed' })
      const activities = result.slice(0, 5).map(item => ({
        id: item.id,
        content: `完成了任务"${item.title}"`,
        time: formatDate(item.completed_time || item.updated_at)
      }))
      this.setData({ activities })
    } catch (error) {
      console.error('加载动态失败:', error)
    }
  },

  // 快速操作
  quickAction(e) {
    const type = e.currentTarget.dataset.type
    switch(type) {
      case 'task':
        wx.navigateTo({ url: '/pages/tasks/list' })
        break
      case 'case':
        wx.switchTab({ url: '/pages/cases/list' })
        break
      case 'dialect':
        wx.navigateTo({ url: '/pages/dialect/create' })
        break
      case 'map':
        wx.navigateTo({ url: '/pages/map/index' })
        break
    }
  },

  goToMyTasks() {
    wx.navigateTo({ url: '/pages/tasks/my' })
  },

  goToCreateTask() {
    wx.navigateTo({ url: '/pages/tasks/create' })
  },

  goToCreateCase() {
    wx.navigateTo({ url: '/pages/cases/create' })
  },

  goToCreateDialect() {
    wx.navigateTo({ url: '/pages/dialect/create' })
  },

  goToMap() {
    wx.navigateTo({ url: '/pages/map/index' })
  },

  goToTaskDetail(e) {
    const id = e.currentTarget.dataset.id
    wx.navigateTo({ url: `/pages/tasks/detail?id=${id}` })
  },

  // 开始任务
  async startTask(e) {
    e.stopPropagation()
    const id = e.currentTarget.dataset.id
    try {
      await post(`/tasks/${id}/progress`, { progress: 1 })
      showSuccess('开始执行')
      this.loadMyTasks()
    } catch (error) {
      console.error('开始任务失败:', error)
    }
  },

  // 完成任务
  completeTask(e) {
    e.stopPropagation()
    const id = e.currentTarget.dataset.id
    wx.navigateTo({ url: `/pages/tasks/feedback?id=${id}&type=complete` })
  }
})
