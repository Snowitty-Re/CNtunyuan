const taskService = require('../../services/task')
const { formatTimeAgo, showSuccess, showToast, showLoading, hideLoading } = require('../../utils/util')
const { TASK_STATUS_MAP, TASK_PRIORITY_MAP, TASK_TYPE_MAP, ROLE_MAP } = require('../../utils/constants')
const { ACTIONS } = require('../../utils/permission')
const app = getApp()

Page({
  data: {
    // 用户信息
    userInfo: {},
    currentDate: '',
    phoneBound: false,
    
    // 今日统计（使用后端实际返回的字段）
    todayStats: {
      myPending: 0,      // 我的待处理任务 (my_pending)
      myProcessing: 0,   // 我的进行中任务 (my_processing)
      myCompleted: 0     // 我的已完成任务 (my_completed)
    },
    
    // 快捷入口
    quickActions: [
      { key: 'myTasks', icon: 'task', label: '我的任务', color: '#FF8C42' },
      { key: 'createTask', icon: 'add', label: '创建任务', color: '#E67E22', requiredPermission: ACTIONS.TASK_MANAGE },
      { key: 'createCase', icon: 'case', label: '发布案件', color: '#3498DB', requiredPermission: ACTIONS.MISSING_MODIFY },
      { key: 'recordDialect', icon: 'mic', label: '录制方言', color: '#27AE60', requiredPermission: ACTIONS.DIALECT_MODIFY },
      { key: 'pendingAssign', icon: 'assign', label: '待分配', color: '#9B59B6', requiredPermission: ACTIONS.TASK_MANAGE },
      { key: 'dialectReview', icon: 'task', label: '方言审批', color: '#E67E22', requiredPermission: ACTIONS.DIALECT_MANAGE },
      { key: 'dialectCardManage', icon: 'settings', label: '方言卡片', color: '#6C63FF', requiredPermission: ACTIONS.DIALECT_MANAGE },
      { key: 'userManage', icon: 'notification', label: '人员管理', color: '#16A085', requiredPermission: ACTIONS.USER_VIEW },
      { key: 'orgManage', icon: 'settings', label: '组织管理', color: '#2C7BE5', requiredPermission: ACTIONS.ORG_MANAGE }
    ],
    
    // 最近任务列表
    recentTasks: [],
    statsLoading: false,
    tasksLoading: false,
    tasksError: false,
    
    roleMap:     ROLE_MAP,
    priorityMap: TASK_PRIORITY_MAP,
    statusMap:   TASK_STATUS_MAP,
    taskTypeMap: TASK_TYPE_MAP
  },

  onLoad() {
    if (!app.ensureAuth || !app.ensureAuth()) return
    this.setCurrentDate()
    this.setData({ phoneBound: !!(app.hasPhoneBound && app.hasPhoneBound()) })
  },

  onShow() {
    if (!app.ensureAuth || !app.ensureAuth()) return
    const phoneBound = !!(app.hasPhoneBound && app.hasPhoneBound())
    this.setData({ phoneBound })
    if (!phoneBound) return
    const now = Date.now()
    if (this._lastLoadTime && now - this._lastLoadTime < 15000) return
    this._lastLoadTime = now
    this.refreshPageData()
  },

  onPullDownRefresh() {
    this.refreshPageData().finally(() => {
      wx.stopPullDownRefresh()
    })
  },

  async refreshPageData() {
    await Promise.all([
      this.loadUserInfo(),
      this.loadTodayStats(),
      this.loadRecentTasks()
    ])
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
      this.syncQuickActionsPermission(userInfo)
    } catch (error) {
      console.error('加载用户信息失败:', error)
    }
  },

  syncQuickActionsPermission(userInfo = {}) {
    const quickActions = (this.data.quickActions || []).map(item => {
      if (!item.requiredPermission) return { ...item, visible: true }
      return { ...item, visible: app.hasPermission(item.requiredPermission, userInfo) }
    })
    this.setData({ quickActions })
  },

  // 加载今日统计
  async loadTodayStats() {
    this.setData({ statsLoading: true })
    try {
      const stats = await taskService.getStats()
      this.setData({
        todayStats: {
          myPending:    stats.my_pending    ?? stats.myPending    ?? 0,
          myProcessing: stats.my_processing ?? stats.myProcessing ?? 0,
          myCompleted:  stats.my_completed  ?? stats.myCompleted  ?? 0
        }
      })
    } catch (error) {
      console.error('加载统计失败:', error)
      showToast('统计加载失败')
    } finally {
      this.setData({ statsLoading: false })
    }
  },

  // 加载最近任务
  async loadRecentTasks() {
    this.setData({ tasksLoading: true, tasksError: false })
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
      this.setData({ tasksError: true })
    } finally {
      this.setData({ tasksLoading: false })
    }
  },

  // 快捷入口点击
  onQuickActionTap(e) {
    const { key } = e.currentTarget.dataset
    const actionItem = (this.data.quickActions || []).find(item => item.key === key)
    if (actionItem && actionItem.visible === false) {
      showToast('无权限操作')
      return
    }
    switch (key) {
      case 'myTasks':
        if (!app.ensurePhoneBound || !app.ensurePhoneBound()) return
        wx.navigateTo({ url: '/pages/tasks/my' })
        break
      case 'createTask':
        if (!app.ensurePhoneBound || !app.ensurePhoneBound()) return
        wx.navigateTo({ url: '/pages/tasks/create' })
        break
      case 'createCase':
        wx.navigateTo({ url: '/pages/cases/create' })
        break
      case 'recordDialect':
        wx.navigateTo({ url: '/pages/dialect/create' })
        break
      case 'pendingAssign':
        if (!app.ensurePhoneBound || !app.ensurePhoneBound()) return
        wx.navigateTo({ url: '/pages/tasks/list?status=pending' })
        break
      case 'dialectReview':
        wx.navigateTo({ url: '/pages/admin/dialect-review' })
        break
      case 'dialectCardManage':
        wx.navigateTo({ url: '/pages/admin/dialect-card-manage' })
        break
      case 'userManage':
        wx.navigateTo({ url: '/pages/admin/user-manage' })
        break
      case 'orgManage':
        wx.navigateTo({ url: '/pages/admin/org-manage' })
        break
    }
  },

  goBindPhone() {
    wx.navigateTo({ url: '/pages/login/index?binding=1&reason=task' })
  },

  // 查看全部任务
  goToMyTasks() {
    if (!app.ensurePhoneBound || !app.ensurePhoneBound()) return
    wx.navigateTo({ url: '/pages/tasks/my' })
  },

  // 任务详情
  goToTaskDetail(e) {
    if (!app.ensurePhoneBound || !app.ensurePhoneBound()) return
    const { id } = e.currentTarget.dataset
    wx.navigateTo({ url: `/pages/tasks/detail?id=${id}` })
  },

  // 开始任务
  async startTask(e) {
    if (!app.ensurePhoneBound || !app.ensurePhoneBound()) return
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
    if (!app.ensurePhoneBound || !app.ensurePhoneBound()) return
    wx.navigateTo({ url: '/pages/tasks/list' })
  },

  retryRecentTasks() {
    this.loadRecentTasks()
  }
})
