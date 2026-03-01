const { get, post, put } = require('../../utils/request')
const { formatDate, showSuccess, showToast } = require('../../utils/util')

Page({
  data: {
    taskId: '',
    task: null,
    comments: [],
    loading: false,
    actionLoading: false,
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
    priorityColorMap: {
      urgent: '#ff4d4f',
      high: '#faad14',
      normal: '#1890ff',
      low: '#52c41a'
    },
    taskTypeMap: {
      search: '实地寻访',
      call: '电话核实',
      info_collect: '信息收集',
      dialect_record: '方言录制',
      coordination: '协调沟通',
      other: '其他'
    },
    currentUserId: '',
    progressInput: 50
  },

  onLoad(options) {
    const userInfo = wx.getStorageSync('userInfo') || {}
    this.setData({ 
      taskId: options.id,
      currentUserId: String(userInfo.id || '')
    })
    if (options.id) {
      this.loadTaskDetail()
    }
  },

  onShow() {
    if (this.data.taskId) {
      this.loadTaskDetail()
    }
  },

  async loadTaskDetail() {
    this.setData({ loading: true })
    try {
      const task = await get(`/tasks/${this.data.taskId}`)
      task.deadline = task.deadline ? formatDate(task.deadline) : null
      task.created_at = task.created_at ? formatDate(task.created_at) : null
      
      // 更新页面标题
      wx.setNavigationBarTitle({ title: task.title || '任务详情' })
      
      this.setData({ task, loading: false })
    } catch (error) {
      console.error('加载任务详情失败:', error)
      showToast('加载失败')
      this.setData({ loading: false })
    }
  },

  // 执行操作
  async handleAction() {
    const { task, currentUserId } = this.data
    if (!task) return

    const statusActions = {
      pending: () => this.claimTask(),
      assigned: () => this.startTask(),
      processing: () => this.showProgressModal()
    }

    const action = statusActions[task.status]
    if (action) {
      await action()
    } else if (task.status === 'completed') {
      showToast('任务已完成')
    }
  },

  // 领取任务
  async claimTask() {
    const { task, currentUserId } = this.data
    this.setData({ actionLoading: true })
    
    try {
      await post(`/tasks/${task.id}/assign`, {
        assignee_id: currentUserId
      })
      showSuccess('领取成功')
      this.loadTaskDetail()
    } catch (error) {
      console.error('领取任务失败:', error)
      showToast('领取失败，请重试')
    } finally {
      this.setData({ actionLoading: false })
    }
  },

  // 开始任务
  async startTask() {
    const { task } = this.data
    this.setData({ actionLoading: true })
    
    try {
      await post(`/tasks/${task.id}/progress`, { progress: 1 })
      showSuccess('开始执行')
      this.loadTaskDetail()
    } catch (error) {
      console.error('开始任务失败:', error)
      showToast('操作失败')
    } finally {
      this.setData({ actionLoading: false })
    }
  },

  // 显示进度更新弹窗
  showProgressModal() {
    this.setData({ showProgressModal: true, progressInput: 50 })
  },

  // 隐藏进度弹窗
  hideProgressModal() {
    this.setData({ showProgressModal: false })
  },

  // 进度滑块变化
  onProgressChange(e) {
    this.setData({ progressInput: e.detail.value })
  },

  // 提交进度
  async submitProgress() {
    const { task, progressInput } = this.data
    this.setData({ actionLoading: true })
    
    try {
      await post(`/tasks/${task.id}/progress`, { progress: progressInput })
      showSuccess('进度更新成功')
      this.setData({ showProgressModal: false })
      this.loadTaskDetail()
    } catch (error) {
      console.error('更新进度失败:', error)
      showToast('更新失败')
    } finally {
      this.setData({ actionLoading: false })
    }
  },

  // 完成任务
  completeTask() {
    const { task } = this.data
    wx.navigateTo({
      url: `/pages/tasks/feedback?id=${task.id}&type=complete`
    })
  },

  // 转派任务
  transferTask() {
    const { task } = this.data
    // 跳转到转派页面，选择新负责人
    wx.navigateTo({
      url: `/pages/tasks/transfer?id=${task.id}`
    })
  },

  // 查看位置
  viewLocation() {
    const { task } = this.data
    if (!task.latitude || !task.longitude) {
      showToast('暂无位置信息')
      return
    }
    wx.openLocation({
      latitude: parseFloat(task.latitude),
      longitude: parseFloat(task.longitude),
      name: task.address || '任务位置',
      address: task.address
    })
  },

  // 查看相关人员信息
  viewAssignee() {
    const { task } = this.data
    if (task.assignee) {
      wx.navigateTo({
        url: `/pages/volunteer/profile?id=${task.assignee_id}`
      })
    }
  },

  // 导航到相关案件
  goToCase() {
    const { task } = this.data
    if (task.missing_person_id) {
      wx.navigateTo({
        url: `/pages/cases/detail?id=${task.missing_person_id}`
      })
    }
  },

  // 下拉刷新
  onPullDownRefresh() {
    this.loadTaskDetail().finally(() => {
      wx.stopPullDownRefresh()
    })
  },

  // 阻止冒泡
  stopPropagation() {},

  // 显示操作菜单
  showActionSheet() {
    const { task, currentUserId } = this.data
    const items = ['更新进度']
    
    if (task.assignee_id === currentUserId) {
      items.push('标记完成')
      items.push('申请转派')
    }
    
    wx.showActionSheet({
      itemList: items,
      success: (res) => {
        switch(res.tapIndex) {
          case 0:
            this.showProgressModal()
            break
          case 1:
            this.completeTask()
            break
          case 2:
            this.transferTask()
            break
        }
      }
    })
  }
})
