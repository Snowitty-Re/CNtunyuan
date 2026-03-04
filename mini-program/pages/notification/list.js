Page({
  data: {
    notifications: [],
    loading: false,
    hasMore: true,
    page: 1,
    pageSize: 20,
    unreadCount: 0
  },

  onLoad() {
    this.loadNotifications()
  },

  onShow() {
    this.loadNotifications()
  },

  // 加载通知列表
  async loadNotifications(refresh = false) {
    if (this.data.loading) return
    
    const page = refresh ? 1 : this.data.page
    
    this.setData({ loading: true })
    
    try {
      // TODO: 接入后端通知API
      // const result = await services.notification.getList({
      //   page,
      //   page_size: this.data.pageSize
      // })
      
      // 模拟数据
      const mockData = this.getMockNotifications()
      
      this.setData({
        notifications: refresh ? mockData : [...this.data.notifications, ...mockData],
        page: page + 1,
        hasMore: mockData.length === this.data.pageSize,
        loading: false,
        unreadCount: mockData.filter(n => !n.is_read).length
      })
    } catch (error) {
      console.error('加载通知失败:', error)
      this.setData({ loading: false })
    }
  },

  // 模拟数据
  getMockNotifications() {
    const types = ['task', 'case', 'system']
    const titles = ['新任务分配', '案件状态更新', '系统通知']
    const contents = [
      '您有一个新的寻人任务待处理',
      '您关注的案件有了新的进展',
      '欢迎使用团圆寻亲志愿者系统'
    ]
    
    return Array.from({ length: 10 }, (_, i) => ({
      id: `notification_${i}`,
      type: types[i % 3],
      title: titles[i % 3],
      content: contents[i % 3],
      is_read: i > 2,
      created_at: new Date(Date.now() - i * 3600000).toISOString()
    }))
  },

  // 下拉刷新
  onPullDownRefresh() {
    this.loadNotifications(true).finally(() => {
      wx.stopPullDownRefresh()
    })
  },

  // 加载更多
  onReachBottom() {
    if (this.data.hasMore && !this.data.loading) {
      this.loadNotifications()
    }
  },

  // 点击通知
  onNotificationTap(e) {
    const notification = e.currentTarget.dataset.item
    
    // 标记为已读
    this.markAsRead(notification.id)
    
    // 根据类型跳转
    switch (notification.type) {
      case 'task':
        wx.navigateTo({ url: '/pages/tasks/my' })
        break
      case 'case':
        wx.navigateTo({ url: '/pages/cases/list' })
        break
    }
  },

  // 标记为已读
  async markAsRead(id) {
    const notifications = this.data.notifications.map(n => {
      if (n.id === id) {
        return { ...n, is_read: true }
      }
      return n
    })
    
    this.setData({ 
      notifications,
      unreadCount: notifications.filter(n => !n.is_read).length
    })
    
    // TODO: 调用后端API标记已读
  },

  // 全部标记为已读
  async markAllAsRead() {
    const notifications = this.data.notifications.map(n => ({
      ...n,
      is_read: true
    }))
    
    this.setData({ 
      notifications,
      unreadCount: 0
    })
    
    wx.showToast({
      title: '已标记为已读',
      icon: 'success'
    })
  },

  // 删除通知
  async deleteNotification(e) {
    const id = e.currentTarget.dataset.id
    
    const confirm = await wx.showModal({
      title: '确认删除',
      content: '确定要删除这条通知吗？'
    })
    
    if (confirm.confirm) {
      const notifications = this.data.notifications.filter(n => n.id !== id)
      this.setData({ notifications })
    }
  },

  // 清空所有通知
  async clearAll() {
    const confirm = await wx.showModal({
      title: '确认清空',
      content: '确定要清空所有通知吗？'
    })
    
    if (confirm.confirm) {
      this.setData({ 
        notifications: [],
        unreadCount: 0
      })
    }
  }
})
