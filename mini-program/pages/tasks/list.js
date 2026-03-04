const taskService = require('../../services/task')
const { formatDate, showSuccess, showToast, showConfirm } = require('../../utils/util')

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
      processing: 0,
      completed: 0,
      cancelled: 0
    },
    statusMap: {
      pending: '待分配',
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
    priorityColorMap: {
      urgent: '#ff4d4f',
      high: '#faad14',
      normal: '#1890ff',
      low: '#52c41a'
    },
    canCreate: false,
    userRole: ''
  },

  onLoad() {
    this.checkPermission()
    this.loadStats()
    this.loadTasks()
  },

  onShow() {
    this.checkPermission()
    this.loadStats()
    this.loadTasks()
  },

  // 检查权限
  checkPermission() {
    const userInfo = wx.getStorageSync('userInfo') || {}
    this.setData({
      canCreate: ['super_admin', 'admin', 'manager'].includes(userInfo.role),
      userRole: userInfo.role || ''
    })
  },

  // 加载统计数据
  async loadStats() {
    try {
      const stats = await taskService.getStats()
      this.setData({ stats })
    } catch (error) {
      console.error('加载统计失败:', error)
    }
  },

  // 加载任务列表
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

      const result = await taskService.getList(params)
      const list = result.list || result || []

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

  // 按状态筛选
  filterByStatus(e) {
    const status = e.currentTarget.dataset.status
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
  onRefresh() {
    this.setData({ refreshing: true, page: 1, hasMore: true })
    this.loadStats()
    this.loadTasks()
  },

  // 加载更多
  onLoadMore() {
    if (!this.data.hasMore || this.data.loadingMore) return
    this.loadTasks(true)
  },

  // 跳转到详情页
  goToDetail(e) {
    const id = e.currentTarget.dataset.id
    wx.navigateTo({ url: `/pages/tasks/detail?id=${id}` })
  },

  // 快速领取任务
  async quickClaim(e) {
    e.stopPropagation()
    const id = e.currentTarget.dataset.id
    const userInfo = wx.getStorageSync('userInfo') || {}
    
    if (!userInfo.id) {
      showToast('请先登录')
      return
    }

    try {
      await taskService.assign(id, String(userInfo.id))
      showSuccess('领取成功')
      this.loadTasks()
      this.loadStats()
    } catch (error) {
      console.error('领取失败:', error)
      showToast('领取失败')
    }
  },

  // 跳转到创建页
  goToCreate() {
    wx.navigateTo({ url: '/pages/tasks/create' })
  },

  // 阻止冒泡
  stopPropagation() {}
})
