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
      { key: 'processing', label: '进行中', count: 0 },
      { key: 'completed', label: '已完成', count: 0 }
    ],
    userRole: ''
  },

  onLoad() {
    if (!app.ensureAuth || !app.ensureAuth()) return
    const userInfo = wx.getStorageSync('userInfo') || {}
    this.setData({ userRole: userInfo.role || '' })
    this.loadTasks()
  },

  onShow() {
    if (!app.ensureAuth || !app.ensureAuth()) return
    this.loadTasks()
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

      // 更新标签计数（全量加载时统计各状态）
      const allTasks = loadMore ? [...this.data.tasks, ...tasks] : tasks
      const tabs = this.data.tabs.map(tab => {
        if (tab.key === '') {
          return { ...tab, count: result.total || allTasks.length }
        }
        return { ...tab, count: allTasks.filter(t => t.status === tab.key).length }
      })

      this.setData({
        tasks: loadMore ? [...this.data.tasks, ...tasks] : tasks,
        tabs,
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
    e.stopPropagation()
    const id = e.currentTarget.dataset.id
    
    try {
      await taskService.start(id)
      showSuccess('任务已开始')
      this.loadTasks()
    } catch (error) {
      console.error('开始任务失败:', error)
      showToast('操作失败')
    }
  },

  // 更新进度
  async updateProgress(e) {
    e.stopPropagation()
    const id = e.currentTarget.dataset.id
    
    wx.showActionSheet({
      itemList: ['25%', '50%', '75%', '100%'],
      success: async (res) => {
        const progress = [25, 50, 75, 100][res.tapIndex]
        try {
          await taskService.updateProgress(id, progress, `更新进度至${progress}%`)
          showSuccess('进度更新成功')
          this.loadTasks()
        } catch (error) {
          console.error('更新进度失败:', error)
          showToast('更新失败')
        }
      }
    })
  },

  // 完成任务
  async completeTask(e) {
    e.stopPropagation()
    const id = e.currentTarget.dataset.id
    
    const confirmed = await showConfirm('确认完成', '确定要将此任务标记为完成吗？')
    if (!confirmed) return

    try {
      await taskService.complete(id, { result: '已完成任务', feedback: '' })
      showSuccess('任务已完成')
      this.loadTasks()
    } catch (error) {
      console.error('完成任务失败:', error)
      showToast('操作失败')
    }
  }
})
