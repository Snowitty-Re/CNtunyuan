const { get, post } = require('../../utils/request')
const { formatDate, showSuccess, showToast } = require('../../utils/util')

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
    stats: {
      total: 0,
      pending: 0,
      assigned: 0,
      processing: 0,
      completed: 0
    },
    statusMap: {
      draft: '草稿',
      pending: '待分配',
      assigned: '已分配',
      processing: '进行中',
      completed: '已完成',
      cancelled: '已取消'
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
    },
    canCreate: false
  },

  onLoad() {
    this.loadStats()
    this.loadTasks()
    this.checkPermission()
  },

  onShow() {
    this.loadStats()
    this.loadTasks()
  },

  // 检查权限
  checkPermission() {
    const userInfo = wx.getStorageSync('userInfo') || {}
    this.setData({
      canCreate: ['super_admin', 'admin', 'manager'].includes(userInfo.role)
    })
  },

  // 加载统计
  async loadStats() {
    try {
      const stats = await get('/tasks/statistics')
      this.setData({ stats })
    } catch (error) {
      console.error('加载统计失败:', error)
    }
  },

  // 加载任务列表
  async loadTasks(loadMore = false) {
    if (loadMore) {
      this.setData({ loadingMore: true })
    } else {
      this.setData({ loading: true })
    }

    try {
      const result = await get('/tasks', {
        page: this.data.page,
        page_size: this.data.pageSize,
        status: this.data.currentStatus || undefined
      })

      const tasks = result.list.map(item => ({
        ...item,
        deadline: item.deadline ? formatDate(item.deadline) : null,
        isOverdue: item.deadline && new Date(item.deadline) < new Date() && item.status !== 'completed'
      }))

      this.setData({
        tasks: loadMore ? [...this.data.tasks, ...tasks] : tasks,
        hasMore: tasks.length === this.data.pageSize,
        loading: false,
        loadingMore: false,
        refreshing: false
      })
    } catch (error) {
      console.error('加载任务失败:', error)
      this.setData({ loading: false, loadingMore: false, refreshing: false })
    }
  },

  // 筛选状态
  filterByStatus(e) {
    const status = e.currentTarget.dataset.status
    this.setData({
      currentStatus: status,
      page: 1,
      tasks: []
    })
    this.loadTasks()
  },

  // 下拉刷新
  onRefresh() {
    this.setData({ refreshing: true, page: 1 })
    this.loadStats()
    this.loadTasks()
  },

  // 加载更多
  onLoadMore() {
    if (!this.data.loadingMore && this.data.hasMore) {
      this.setData({ page: this.data.page + 1 })
      this.loadTasks(true)
    }
  },

  // 跳转到详情
  goToDetail(e) {
    const id = e.currentTarget.dataset.id
    wx.navigateTo({ url: `/pages/tasks/detail?id=${id}` })
  },

  // 快速领取任务
  async quickClaim(e) {
    e.stopPropagation()
    const id = e.currentTarget.dataset.id
    const userInfo = wx.getStorageSync('userInfo') || {}
    
    try {
      await post(`/tasks/${id}/assign`, {
        assignee_id: String(userInfo.id)
      })
      showSuccess('领取成功')
      this.loadTasks()
      this.loadStats()
    } catch (error) {
      console.error('领取失败:', error)
      showToast('领取失败')
    }
  },

  // 创建任务
  goToCreate() {
    wx.navigateTo({ url: '/pages/tasks/create' })
  },

  stopPropagation() {}
})
