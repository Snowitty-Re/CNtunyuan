const taskService = require('../../services/task')
const { formatTimeAgo, showSuccess, showToast, showLoading, hideLoading } = require('../../utils/util')
const { TASK_STATUS_MAP, TASK_PRIORITY_MAP, TASK_TYPE_MAP, ROLE_MAP } = require('../../utils/constants')
const app = getApp()

Page({
  data: {
    // 用户信息
    userInfo: {},
    currentDate: '',
    
    // 今日统计（使用后端实际返回的字段）
    todayStats: {
      myPending: 0,      // 我的待处理任务 (my_pending)
      myProcessing: 0,   // 我的进行中任务 (my_processing)
      myCompleted: 0     // 我的已完成任务 (my_completed)
    },
    
    // 快捷入口
    quickActions: [
      { key: 'myTasks', icon: 'task', label: '我的任务', color: '#FF8C42' },
      { key: 'createCase', icon: 'case', label: '发布案件', color: '#3498DB' },
      { key: 'recordDialect', icon: 'mic', label: '录制方言', color: '#27AE60' },
      { key: 'pendingAssign', icon: 'assign', label: '待分配', color: '#9B59B6', managerOnly: true },
      { key: 'dialectReview', icon: 'task', label: '方言审批', color: '#E67E22', managerOnly: true },
      { key: 'userManage', icon: 'notification', label: '人员管理', color: '#16A085', managerOnly: true },
      { key: 'orgManage', icon: 'settings', label: '组织管理', color: '#2C7BE5', managerOnly: true }
    ],
    
    // 最近任务列表
    recentTasks: [],
    
    roleMap:     ROLE_MAP,
    priorityMap: TASK_PRIORITY_MAP,
    statusMap:   TASK_STATUS_MAP,
    taskTypeMap: TASK_TYPE_MAP
  },

  onLoad() {
    if (!app.ensureAuth || !app.ensureAuth()) return
    this.setCurrentDate()
  },

  onShow() {
    if (!app.ensureAuth || !app.ensureAuth()) return
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
          myPending:    stats.my_pending    || 0,
          myProcessing: stats.my_processing || 0,
          myCompleted:  stats.my_completed  || 0
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
      const tasks = (result.list || []).map(item => ({
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
      case 'dialectReview':
        wx.navigateTo({ url: '/pages/admin/dialect-review' })
        break
      case 'userManage':
        wx.navigateTo({ url: '/pages/admin/user-manage' })
        break
      case 'orgManage':
        wx.navigateTo({ url: '/pages/admin/org-manage' })
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
