const services = require('../../services/index')
const { formatTimeAgo } = require('../../utils/util')

const STORAGE_KEY = 'notification_read_status'

Page({
  data: {
    notifications: [],
    loading: false,
    unreadCount: 0
  },

  onLoad() {
    this.loadNotifications()
  },

  onShow() {
    this.loadNotifications()
  },

  onPullDownRefresh() {
    this.loadNotifications().finally(() => {
      wx.stopPullDownRefresh()
    })
  },

  // 加载通知（从任务和案件动态生成本地通知）
  async loadNotifications() {
    if (this.data.loading) return
    this.setData({ loading: true })

    try {
      const notifications = []
      const readStatus = wx.getStorageSync(STORAGE_KEY) || {}

      // 获取最近任务变更作为通知
      try {
        const taskResult = await services.task.getMyTasks({ page: 1, page_size: 10 })
        const tasks = taskResult.list || taskResult || []
        tasks.forEach(task => {
          notifications.push({
            id: `task_${task.id}`,
            type: 'task',
            title: this.getTaskNotificationTitle(task),
            content: task.title || '任务更新',
            is_read: !!readStatus[`task_${task.id}`],
            created_at: task.updated_at || task.created_at,
            target_url: `/pages/tasks/detail?id=${task.id}`
          })
        })
      } catch (e) {
        console.log('获取任务通知失败:', e)
      }

      // 获取最近案件变更作为通知
      try {
        const caseResult = await services.missingPerson.getList({ page: 1, page_size: 5 })
        const cases = caseResult.list || caseResult || []
        cases.forEach(item => {
          if (item.status === 'found' || item.status === 'reunited') {
            notifications.push({
              id: `case_${item.id}`,
              type: 'case',
              title: item.status === 'found' ? '案件有新进展' : '团圆喜讯',
              content: `${item.name} - ${item.status === 'found' ? '已找到' : '已团圆'}`,
              is_read: !!readStatus[`case_${item.id}`],
              created_at: item.updated_at || item.created_at,
              target_url: `/pages/cases/detail?id=${item.id}`
            })
          }
        })
      } catch (e) {
        console.log('获取案件通知失败:', e)
      }

      // 添加系统欢迎通知
      notifications.push({
        id: 'system_welcome',
        type: 'system',
        title: '系统通知',
        content: '欢迎使用团圆寻亲志愿者系统，通知功能即将全面接入',
        is_read: !!readStatus['system_welcome'],
        created_at: new Date().toISOString(),
        target_url: ''
      })

      // 按时间排序
      notifications.sort((a, b) => new Date(b.created_at) - new Date(a.created_at))

      // 格式化时间
      notifications.forEach(n => {
        n.timeAgo = formatTimeAgo(n.created_at)
      })

      this.setData({
        notifications,
        loading: false,
        unreadCount: notifications.filter(n => !n.is_read).length
      })
    } catch (error) {
      console.error('加载通知失败:', error)
      this.setData({ loading: false })
    }
  },

  // 获取任务通知标题
  getTaskNotificationTitle(task) {
    const statusMap = {
      pending: '新任务待处理',
      assigned: '任务已分配给您',
      processing: '任务进行中',
      completed: '任务已完成',
      cancelled: '任务已取消'
    }
    return statusMap[task.status] || '任务更新'
  },

  // 点击通知
  onNotificationTap(e) {
    const notification = e.currentTarget.dataset.item
    this.markAsRead(notification.id)

    if (notification.target_url) {
      wx.navigateTo({ url: notification.target_url })
    }
  },

  // 标记为已读
  markAsRead(id) {
    const readStatus = wx.getStorageSync(STORAGE_KEY) || {}
    readStatus[id] = true
    wx.setStorageSync(STORAGE_KEY, readStatus)

    const notifications = this.data.notifications.map(n => {
      if (n.id === id) return { ...n, is_read: true }
      return n
    })

    this.setData({
      notifications,
      unreadCount: notifications.filter(n => !n.is_read).length
    })
  },

  // 全部标记为已读
  markAllAsRead() {
    const readStatus = wx.getStorageSync(STORAGE_KEY) || {}
    this.data.notifications.forEach(n => {
      readStatus[n.id] = true
    })
    wx.setStorageSync(STORAGE_KEY, readStatus)

    const notifications = this.data.notifications.map(n => ({ ...n, is_read: true }))
    this.setData({ notifications, unreadCount: 0 })
    wx.showToast({ title: '已全部标记已读', icon: 'success' })
  },

  // 清空所有通知
  async clearAll() {
    const res = await wx.showModal({
      title: '确认清空',
      content: '确定要清空所有通知吗？'
    })
    if (res.confirm) {
      wx.removeStorageSync(STORAGE_KEY)
      this.setData({ notifications: [], unreadCount: 0 })
    }
  }
})
