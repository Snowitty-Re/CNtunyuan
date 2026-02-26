const { get } = require('../../utils/request')
const { formatDate } = require('../../utils/util')

Page({
  data: {
    userInfo: {},
    stats: {
      taskCount: 0,
      caseCount: 0,
      dialectCount: 0,
      score: 0
    },
    todos: [],
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
    this.loadStats()
    this.loadTodos()
    this.loadActivities()
  },

  onShow() {
    this.loadStats()
    this.loadTodos()
  },

  async loadUserInfo() {
    try {
      const userInfo = wx.getStorageSync('userInfo')
      if (userInfo) {
        this.setData({ userInfo })
      }
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

  async loadTodos() {
    try {
      const result = await get('/tasks', { status: 'assigned', page: 1, page_size: 5 })
      const todos = result.list.map(item => ({
        ...item,
        deadline: item.deadline ? formatDate(item.deadline) : null
      }))
      this.setData({ todos })
    } catch (error) {
      console.error('加载待办失败:', error)
    }
  },

  async loadActivities() {
    try {
      // 模拟动态数据
      this.setData({
        activities: [
          { id: 1, content: '完成了任务"实地寻访张大爷"', time: '2小时前' },
          { id: 2, content: '领取了新任务"电话核实李奶奶信息"', time: '5小时前' },
          { id: 3, content: '上传了方言录音"四川话-成都地区"', time: '1天前' },
          { id: 4, content: '参与了案件"儿童走失-小明"的寻找', time: '2天前' }
        ]
      })
    } catch (error) {
      console.error('加载动态失败:', error)
    }
  },

  goToMyTasks() {
    wx.switchTab({ url: '/pages/tasks/list' })
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

  startTask(e) {
    e.stopPropagation()
    const id = e.currentTarget.dataset.id
    wx.navigateTo({ url: `/pages/tasks/detail?id=${id}` })
  }
})
