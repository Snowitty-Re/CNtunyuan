const taskService = require('../../services/task')
const { formatTimeAgo, showSuccess, showToast, showLoading, hideLoading } = require('../../utils/util')
const app = getApp()

Page({
  data: {
    // 用户信息
    userInfo: {},
    currentDate: '',
    
    // 今日统计
    todayStats: {
      pendingCount: 0,    // 待处理任务
      processingCount: 0, // 进行中任务
      helpedCount: 0      // 已帮助案件
    },
    
    // 快捷入口
    quickActions: [
      { key: 'myTasks', icon: 'task', label: '我的任务', color: '#FF8C42' },
      { key: 'createCase', icon: 'case', label: '发布案件', color: '#3498DB' },
      { key: 'recordDialect', icon: 'mic', label: '录制方言', color: '#27AE60' },
      { key: 'pendingAssign', icon: 'assign', label: '待分配', color: '#9B59B6', managerOnly: true }
    ],
    
    // 最近任务列表
    recentTasks: [],
    
    // 状态映射
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
    statusMap: {
      pending: '待处理',
      assigned: '已分配',
      processing: '进行中',
      completed: '已完成',
      cancelled: '已取消'
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
    this.setCurrentDate()
  },

  onShow() {
    this.loadUserInfo()
    this.loadTodayStats()
    this.loadRecentTasks()
  },

  onPullDownRefresh() {
    Promise.all([
      this.loadUserInfo(),
      this.loadTodayStats(),
      this.loadRecentTasks()
    ]).finally(() => {
      wx.stopPullDownRefresh()
    })
  },

  // 设置当前日期
  setCurrentDate() {
    const date = new Date()
    const weekDays = ['周日', '周一', '周二', '周三', '周四', '周五', '周六']
    const month = date.getMonth() + 1
    const day = date.getDate()
    const weekDay = weekDays[date.getDay()]
    this.setData({
      currentDate: `${month}月${day}日 ${weekDay}`
    })
  },

  // 加载用户信息
  async loadUserInfo() {
    try {
      const userInfo = await app.getUserInfo() || wx.getStorageSync('userInfo') || {}
      this.setData({ 
        userInfo: {
          ...userInfo,
          avatar: userInfo.avatar || '/assets/images/avatar-default.png',
          nickname: userInfo.nickname || '志愿者',
          role: userInfo.role || 'volunteer'
        }
      })
    } catch (error) {
      console.error('加载用户信息失败:', error)
    }
  },

  // 加载今日统计
  async loadTodayStats() {
    try {
      showLoading('加载中...')
      const stats = await taskService.getStats()
      this.setData({
        todayStats: {
          pendingCount: stats.pending_count || 0,
          processingCount: stats.processing_count || 0,
          helpedCount: stats.helped_count || stats.completed_count || 0
        }
      })
    } catch (error) {
      console.error('加载统计失败:', error)
      showToast('统计加载失败')
    } finally {
      hideLoading()
    }
  },

  // 加载最近任务
  async loadRecentTasks() {
    try {
      const result = await taskService.getMyTasks({ 
        page: 1, 
        page_size: 5 
      })
      const tasks = (result.list || result || []).map(item => ({
        ...item,
        timeAgo: formatTimeAgo(item.created_at || item.updated_at)
      }))
      this.setData({ recentTasks: tasks })
    } catch (error) {
      console.error('加载最近任务失败:', error)
    }
  },

  // 快捷入口点击
  onQuickActionTap(e) {
    const { key } = e.currentTarget.dataset
    switch (key) {
      case 'myTasks':
        wx.navigateTo({ url: '/pages/tasks/my' })
        break
      case 'createCase':
        wx.navigateTo({ url: '/pages/cases/create' })
        break
      case 'recordDialect':
        wx.navigateTo({ url: '/pages/dialect/create' })
        break
      case 'pendingAssign':
        wx.navigateTo({ url: '/pages/tasks/list?status=pending' })
        break
    }
  },

  // 查看全部任务
  goToMyTasks() {
    wx.navigateTo({ url: '/pages/tasks/my' })
  },

  // 任务详情
  goToTaskDetail(e) {
    const { id } = e.currentTarget.dataset
    wx.navigateTo({ url: `/pages/tasks/detail?id=${id}` })
  },

  // 开始任务
  async startTask(e) {
    e.stopPropagation()
    const { id } = e.currentTarget.dataset
    try {
      showLoading('处理中...')
      await taskService.start(id)
      showSuccess('任务已开始')
      this.loadRecentTasks()
      this.loadTodayStats()
    } catch (error) {
      showToast(error.message || '开始任务失败')
    } finally {
      hideLoading()
    }
  },

  // 跳转到任务列表
  goToTaskList() {
    wx.navigateTo({ url: '/pages/tasks/list' })
  },

  // 判断是否显示管理者专属入口
  isManager() {
    const { role } = this.data.userInfo
    return ['super_admin', 'admin', 'manager'].includes(role)
  }
})
