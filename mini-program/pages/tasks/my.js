const { get } = require('../../utils/request')
const { formatDate } = require('../../utils/util')

Page({
  data: {
    tasks: [],
    status: '',
    page: 1,
    pageSize: 20,
    loading: false,
    loadingMore: false,
    hasMore: true,
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
    tabs: [
      { key: '', label: '全部' },
      { key: 'assigned', label: '待处理' },
      { key: 'processing', label: '进行中' },
      { key: 'completed', label: '已完成' }
    ]
  },

  onLoad() {
    this.loadTasks()
  },

  onPullDownRefresh() {
    this.setData({ page: 1, tasks: [] })
    this.loadTasks().finally(() => {
      wx.stopPullDownRefresh()
    })
  },

  onReachBottom() {
    if (!this.data.loadingMore && this.data.hasMore) {
      this.setData({ page: this.data.page + 1 })
      this.loadTasks(true)
    }
  },

  async loadTasks(loadMore = false) {
    if (loadMore) {
      this.setData({ loadingMore: true })
    } else {
      this.setData({ loading: true })
    }

    try {
      const result = await get('/tasks/my', {
        page: this.data.page,
        page_size: this.data.pageSize,
        status: this.data.status || undefined
      })
      
      const tasks = result.map(item => ({
        ...item,
        deadline: item.deadline ? formatDate(item.deadline) : null,
        isOverdue: item.deadline && new Date(item.deadline) < new Date() && item.status !== 'completed'
      }))

      this.setData({
        tasks: loadMore ? [...this.data.tasks, ...tasks] : tasks,
        hasMore: tasks.length === this.data.pageSize,
        loading: false,
        loadingMore: false
      })
    } catch (error) {
      console.error('加载任务失败:', error)
      this.setData({ loading: false, loadingMore: false })
    }
  },

  switchTab(e) {
    const status = e.currentTarget.dataset.key
    this.setData({ status, page: 1, tasks: [] })
    this.loadTasks()
  },

  goToDetail(e) {
    const id = e.currentTarget.dataset.id
    wx.navigateTo({ url: `/pages/tasks/detail?id=${id}` })
  }
})
