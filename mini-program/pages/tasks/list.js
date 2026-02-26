const { get } = require('../../utils/request')
const { formatDate, showLoading, hideLoading } = require('../../utils/util')

Page({
  data: {
    tasks: [],
    currentStatus: '',
    page: 1,
    pageSize: 20,
    loading: false,
    noMore: false,
    stats: { total: 0, pending: 0, completed: 0 },
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
    }
  },

  onLoad() {
    this.loadTasks()
    this.loadStats()
  },

  onPullDownRefresh() {
    this.setData({ page: 1, noMore: false })
    this.loadTasks().finally(() => {
      wx.stopPullDownRefresh()
    })
  },

  onReachBottom() {
    if (!this.data.noMore && !this.data.loading) {
      this.loadMore()
    }
  },

  async loadTasks() {
    this.setData({ loading: true })
    showLoading()

    try {
      const params = {
        page: this.data.page,
        page_size: this.data.pageSize,
        status: this.data.currentStatus
      }

      const result = await get('/tasks', params)
      const tasks = result.list.map(item => ({
        ...item,
        deadline: item.deadline ? formatDate(item.deadline) : null,
        isOverdue: item.deadline && new Date(item.deadline) < new Date() && item.status !== 'completed'
      }))

      this.setData({
        tasks: this.data.page === 1 ? tasks : [...this.data.tasks, ...tasks],
        noMore: tasks.length < this.data.pageSize
      })
    } catch (error) {
      console.error('加载任务失败:', error)
    } finally {
      hideLoading()
      this.setData({ loading: false })
    }
  },

  async loadStats() {
    try {
      // 这里应该调用统计接口
      this.setData({
        stats: { total: 156, pending: 23, completed: 98 }
      })
    } catch (error) {
      console.error('加载统计失败:', error)
    }
  },

  loadMore() {
    this.setData({ page: this.data.page + 1 })
    this.loadTasks()
  },

  switchStatus(e) {
    const status = e.currentTarget.dataset.status
    this.setData({ currentStatus: status, page: 1, tasks: [] })
    this.loadTasks()
  },

  goToDetail(e) {
    const id = e.currentTarget.dataset.id
    wx.navigateTo({ url: `/pages/tasks/detail?id=${id}` })
  }
})
