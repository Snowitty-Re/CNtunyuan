const { get, post } = require('../../utils/request')
const { formatDate, showSuccess } = require('../../utils/util')

Page({
  data: {
    notifications: [],
    page: 1,
    pageSize: 20,
    loading: false,
    loadingMore: false,
    hasMore: true,
    unreadCount: 0,
    typeMap: {
      system: '系统通知',
      task: '任务通知',
      case: '案件通知',
      workflow: '审批通知'
    }
  },

  onLoad() {
    this.loadNotifications()
    this.loadUnreadCount()
  },

  onShow() {
    this.loadUnreadCount()
  },

  onPullDownRefresh() {
    this.setData({ page: 1, notifications: [] })
    this.loadNotifications().finally(() => {
      wx.stopPullDownRefresh()
    })
  },

  onReachBottom() {
    if (!this.data.loadingMore && this.data.hasMore) {
      this.setData({ page: this.data.page + 1 })
      this.loadNotifications(true)
    }
  },

  async loadNotifications(loadMore = false) {
    if (loadMore) {
      this.setData({ loadingMore: true })
    } else {
      this.setData({ loading: true })
    }

    try {
      const result = await get('/notifications', {
        page: this.data.page,
        page_size: this.data.pageSize
      })

      const notifications = result.list.map(item => ({
        ...item,
        created_at: formatDate(item.created_at)
      }))

      this.setData({
        notifications: loadMore ? [...this.data.notifications, ...notifications] : notifications,
        hasMore: notifications.length === this.data.pageSize,
        loading: false,
        loadingMore: false
      })
    } catch (error) {
      console.error('加载通知失败:', error)
      this.setData({ loading: false, loadingMore: false })
    }
  },

  async loadUnreadCount() {
    try {
      const result = await get('/notifications/unread-count')
      this.setData({ unreadCount: result.count || 0 })
    } catch (error) {
      console.error('加载未读数失败:', error)
    }
  },

  // 标记已读
  async markAsRead(e) {
    const id = e.currentTarget.dataset.id
    try {
      await post(`/notifications/${id}/read`)
      
      const notifications = this.data.notifications.map(item => {
        if (item.id === id) {
          return { ...item, is_read: true }
        }
        return item
      })
      
      this.setData({ notifications })
      this.loadUnreadCount()
    } catch (error) {
      console.error('标记已读失败:', error)
    }
  },

  // 标记全部已读
  async markAllAsRead() {
    try {
      await post('/notifications/read-all')
      showSuccess('已标记全部已读')
      
      const notifications = this.data.notifications.map(item => ({
        ...item, is_read: true
      }))
      this.setData({ notifications })
      this.loadUnreadCount()
    } catch (error) {
      console.error('标记全部已读失败:', error)
    }
  },

  // 删除通知
  async deleteNotification(e) {
    const id = e.currentTarget.dataset.id
    
    wx.showModal({
      title: '提示',
      content: '确定删除这条通知吗？',
      success: async (res) => {
        if (res.confirm) {
          try {
            await post(`/notifications/${id}/delete`)
            
            const notifications = this.data.notifications.filter(
              item => item.id !== id
            )
            this.setData({ notifications })
            this.loadUnreadCount()
          } catch (error) {
            console.error('删除通知失败:', error)
          }
        }
      }
    })
  },

  // 点击通知
  onNotificationTap(e) {
    const { id, type, businessid } = e.currentTarget.dataset
    
    this.markAsRead(e)
    
    switch(type) {
      case 'task':
        wx.navigateTo({ url: `/pages/tasks/detail?id=${businessid}` })
        break
      case 'case':
        wx.navigateTo({ url: `/pages/cases/detail?id=${businessid}` })
        break
      default:
        break
    }
  }
})
