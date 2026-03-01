const { post, put, get } = require('../../utils/request')
const { showSuccess, showToast } = require('../../utils/util')

Page({
  data: {
    taskId: '',
    type: 'complete', // complete | transfer
    content: '',
    progress: 100,
    attachments: [],
    transferTo: '',
    transferToName: '',
    users: [],
    loading: false
  },

  onLoad(options) {
    this.setData({
      taskId: options.id,
      type: options.type || 'complete'
    })
    wx.setNavigationBarTitle({
      title: this.data.type === 'complete' ? '任务完成反馈' : '任务转派'
    })
    if (this.data.type === 'transfer') {
      this.loadUsers()
    }
  },

  // 加载可选用户
  async loadUsers() {
    try {
      const users = await get('/users', { page_size: 100 })
      this.setData({ users: users.list || [] })
    } catch (error) {
      console.error('加载用户列表失败:', error)
    }
  },

  // 输入反馈内容
  onContentInput(e) {
    this.setData({ content: e.detail.value })
  },

  // 选择转派人
  onUserChange(e) {
    const index = e.detail.value
    const user = this.data.users[index]
    this.setData({ 
      transferTo: user.id,
      transferToName: user.nickname
    })
  },

  // 进度变化
  onProgressChange(e) {
    this.setData({ progress: e.detail.value })
  },

  // 选择图片
  chooseImage() {
    wx.chooseImage({
      count: 9 - this.data.attachments.length,
      success: (res) => {
        const attachments = [...this.data.attachments, ...res.tempFilePaths]
        this.setData({ attachments })
      }
    })
  },

  // 删除图片
  removeImage(e) {
    const index = e.currentTarget.dataset.index
    const attachments = this.data.attachments.filter((_, i) => i !== index)
    this.setData({ attachments })
  },

  // 提交
  async submit() {
    const { taskId, type, content, progress, attachments, transferTo } = this.data
    
    if (!content.trim()) {
      showToast('请填写反馈内容')
      return
    }

    if (type === 'transfer' && !transferTo) {
      showToast('请选择转派对象')
      return
    }

    this.setData({ loading: true })

    try {
      if (type === 'complete') {
        // 先更新进度
        if (progress < 100) {
          await post(`/tasks/${taskId}/progress`, { progress })
        }
        // 完成任务
        await post(`/tasks/${taskId}/complete`, {
          feedback: content,
          attachments
        })
        showSuccess('任务已完成')
      } else {
        // 转派任务
        await post(`/tasks/${taskId}/transfer`, {
          assignee_id: transferTo,
          reason: content
        })
        showSuccess('转派成功')
      }
      
      // 返回上一页
      wx.navigateBack({
        success: () => {
          // 通知前一页刷新
          const pages = getCurrentPages()
          const prevPage = pages[pages.length - 2]
          if (prevPage && prevPage.loadTaskDetail) {
            prevPage.loadTaskDetail()
          }
        }
      })
    } catch (error) {
      console.error('提交失败:', error)
      showToast('提交失败，请重试')
    } finally {
      this.setData({ loading: false })
    }
  }
})
