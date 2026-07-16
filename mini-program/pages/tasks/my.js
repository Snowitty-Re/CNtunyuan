const taskService = require('../../services/task')
const { formatDate, showSuccess, showToast, showConfirm } = require('../../utils/util')
const { TASK_STATUS_MAP, TASK_PRIORITY_MAP } = require('../../utils/constants')
const app = getApp()

Page({
  data: {
    tasks: [],
    currentStatus: '',
    page: 1,
    pageSize: 20,
    loading: false,
    loadingMore: false,
    refreshing: false,
    hasMore: true,
    statusMap:   TASK_STATUS_MAP,
    priorityMap: TASK_PRIORITY_MAP,
    tabs: [
      { key: '', label: '全部', count: 0 },
      { key: 'assigned', label: '待开始', count: 0 },
      { key: 'processing', label: '进行中', count: 0 },
      { key: 'completed', label: '已完成', count: 0 }
    ],
    userRole: ''
  },

  onLoad() {
    if (!app.ensurePhoneBound || !app.ensurePhoneBound({ message: '查看我的任务需绑定手机号' })) return
    const userInfo = wx.getStorageSync('userInfo') || {}
    this.setData({ userRole: userInfo.role || '' })
    this.loadStats()
    this.loadTasks()
  },

  onShow() {
    if (!app.ensurePhoneBound || !app.ensurePhoneBound({ message: '查看我的任务需绑定手机号' })) return
    this.loadStats()
    this.loadTasks()
  },

  async loadStats() {
    try {
      const stats = await taskService.getStats()
      const tabs = this.data.tabs.map(tab => {
        if (tab.key === '') return { ...tab, count: stats.my_tasks || 0 }
        if (tab.key === 'assigned') {
          const assigned = Math.max((stats.my_pending || 0) - (stats.my_processing || 0), 0)
          return { ...tab, count: assigned }
        }
        if (tab.key === 'processing') return { ...tab, count: stats.my_processing || 0 }
        if (tab.key === 'completed') return { ...tab, count: stats.my_completed || 0 }
        return tab
      })
      this.setData({ tabs })
    } catch (error) {
      // 统计失败不阻塞列表加载
      console.error('加载任务统计失败:', error)
    }
  },

  // 加载我的任务列表
  async loadTasks(loadMore = false) {
    if (this.data.loading || (loadMore && this.data.loadingMore)) return

    if (loadMore) {
      this.setData({ loadingMore: true })
    } else {
      this.setData({ loading: true })
    }

    try {
      const params = {
        page: loadMore ? this.data.page : 1,
        page_size: this.data.pageSize
      }
      
      if (this.data.currentStatus) {
        params.status = this.data.currentStatus
      }

      const result = await taskService.getMyTasks(params)
      const list = result.list || []

      const tasks = list.map(item => ({
        ...item,
        deadline: item.deadline ? formatDate(item.deadline) : null,
        isOverdue: item.deadline && new Date(item.deadline) < new Date() && 
                   item.status !== 'completed' && item.status !== 'cancelled'
      }))

      this.setData({
        tasks: loadMore ? [...this.data.tasks, ...tasks] : tasks,
        page: loadMore ? this.data.page + 1 : 2,
        hasMore: tasks.length === this.data.pageSize,
        loading: false,
        loadingMore: false,
        refreshing: false
      })
    } catch (error) {
      console.error('加载任务失败:', error)
      showToast('加载失败')
      this.setData({ 
        loading: false, 
        loadingMore: false, 
        refreshing: false 
      })
    }
  },

  // 切换标签
  switchTab(e) {
    const status = e.currentTarget.dataset.key
    if (status === this.data.currentStatus) return
    
    this.setData({
      currentStatus: status,
      page: 1,
      tasks: [],
      hasMore: true
    })
    this.loadTasks()
  },

  // 下拉刷新
  onPullDownRefresh() {
    this.setData({ refreshing: true, page: 1, hasMore: true })
    this.loadTasks().finally(() => {
      wx.stopPullDownRefresh()
      this.setData({ refreshing: false })
    })
  },

  // 加载更多
  onReachBottom() {
    if (!this.data.hasMore || this.data.loadingMore) return
    this.loadTasks(true)
  },

  // 跳转到详情页
  goToDetail(e) {
    const id = e.currentTarget.dataset.id
    wx.navigateTo({ url: `/pages/tasks/detail?id=${id}` })
  },

  // 开始任务
  async startTask(e) {
    const id = e.currentTarget.dataset.id
    
    try {
      await taskService.start(id)
      showSuccess('任务已开始')
      this.loadStats()
      this.loadTasks()
    } catch (error) {
      console.error('开始任务失败:', error)
      showToast(error.message || '开始任务失败')
    }
  },

  // 更新进度
  async updateProgress(e) {
    const id = e.currentTarget.dataset.id
    
    wx.showActionSheet({
      itemList: ['25%', '50%', '75%', '100%'],
      success: async (res) => {
        const progress = [25, 50, 75, 100][res.tapIndex]
        try {
          await taskService.updateProgress(id, progress, `更新进度至${progress}%`)
          showSuccess('进度更新成功')
          this.loadStats()
          this.loadTasks()
        } catch (error) {
          console.error('更新进度失败:', error)
          showToast(error.message || '更新失败')
        }
      }
    })
  },

  // 完成任务
  async completeTask(e) {
    const id = e.currentTarget.dataset.id
    
    const confirmed = await showConfirm('确认完成', '确定要将此任务标记为完成吗？')
    if (!confirmed) return

    try {
      await taskService.complete(id, { result: '已完成任务', feedback: '' })
      showSuccess('任务已完成')
      this.loadStats()
      this.loadTasks()
    } catch (error) {
      console.error('完成任务失败:', error)
      showToast(error.message || '完成任务失败')
    }
  },

  // 填写反馈并完成
  submitFeedback(e) {
    const id = e.currentTarget.dataset.id
    wx.navigateTo({ url: `/pages/tasks/feedback?id=${id}` })
  },

  // 新增跟进
  addFollowUp(e) {
    const id = e.currentTarget.dataset.id
    wx.navigateTo({ url: `/pages/tasks/follow-up-create?id=${id}` })
  }
})
