const taskService = require('../../services/task')
const { formatDate, formatTimeAgo, showSuccess, showToast, showConfirm } = require('../../utils/util')

Page({
  data: {
    taskId: '',
    task: null,
    logs: [],
    loading: false,
    actionLoading: false,
    showProgressModal: false,
    progressInput: 50,
    currentUser: null,
    isManager: false,
    isAssignee: false,
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
    }
  },

  onLoad(options) {
    const userInfo = wx.getStorageSync('userInfo') || {}
    this.setData({ 
      taskId: options.id,
      currentUser: userInfo,
      isManager: ['super_admin', 'admin', 'manager'].includes(userInfo.role)
    })
    
    if (options.id) {
      this.loadTaskDetail()
      this.loadTaskLogs()
    }
  },

  onShow() {
    if (this.data.taskId) {
      this.loadTaskDetail()
      this.loadTaskLogs()
    }
  },

  // 加载任务详情
  async loadTaskDetail() {
    this.setData({ loading: true })
    try {
      const task = await taskService.getById(this.data.taskId)
      task.deadline = task.deadline ? formatDate(task.deadline) : null
      task.created_at = task.created_at ? formatDate(task.created_at, 'YYYY-MM-DD HH:mm') : null
      
      // 检查当前用户是否是执行人
      const isAssignee = task.assignee_id && String(task.assignee_id) === String(this.data.currentUser.id)
      
      // 更新页面标题
      wx.setNavigationBarTitle({ title: task.title || '任务详情' })
      
      this.setData({ 
        task, 
        isAssignee,
        loading: false 
      })
    } catch (error) {
      console.error('加载任务详情失败:', error)
      showToast('加载失败')
      this.setData({ loading: false })
    }
  },

  // 加载任务日志
  async loadTaskLogs() {
    try {
      const logs = await taskService.getLogs(this.data.taskId)
      const formattedLogs = logs.map(log => ({
        ...log,
        created_at: formatTimeAgo(log.created_at)
      }))
      this.setData({ logs: formattedLogs })
    } catch (error) {
      console.error('加载任务日志失败:', error)
    }
  },

  // 分配任务（管理者）
  async assignTask() {
    if (!this.data.isManager) {
      showToast('无权限操作')
      return
    }

    wx.navigateTo({
      url: `/pages/tasks/assign?id=${this.data.taskId}`
    })
  },

  // 开始任务（执行人）
  async startTask() {
    if (!this.data.isAssignee) {
      showToast('只有执行人可以开始任务')
      return
    }

    this.setData({ actionLoading: true })
    try {
      await taskService.start(this.data.taskId)
      showSuccess('任务已开始')
      this.loadTaskDetail()
      this.loadTaskLogs()
    } catch (error) {
      console.error('开始任务失败:', error)
      showToast('操作失败')
    } finally {
      this.setData({ actionLoading: false })
    }
  },

  // 显示进度更新弹窗
  showProgressModal() {
    if (!this.data.isAssignee) {
      showToast('只有执行人可以更新进度')
      return
    }
    const currentProgress = this.data.task?.progress || 0
    this.setData({ 
      showProgressModal: true, 
      progressInput: currentProgress 
    })
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
    this.setData({ actionLoading: true })
    try {
      await taskService.updateProgress(
        this.data.taskId, 
        this.data.progressInput, 
        `更新进度至${this.data.progressInput}%`
      )
      showSuccess('进度更新成功')
      this.setData({ showProgressModal: false })
      this.loadTaskDetail()
      this.loadTaskLogs()
    } catch (error) {
      console.error('更新进度失败:', error)
      showToast('更新失败')
    } finally {
      this.setData({ actionLoading: false })
    }
  },

  // 完成任务（执行人）
  async completeTask() {
    if (!this.data.isAssignee) {
      showToast('只有执行人可以完成任务')
      return
    }

    const confirmed = await showConfirm('确认完成', '确定要将此任务标记为完成吗？')
    if (!confirmed) return

    this.setData({ actionLoading: true })
    try {
      await taskService.complete(this.data.taskId, { 
        result: '任务已完成',
        feedback: ''
      })
      showSuccess('任务已完成')
      this.loadTaskDetail()
      this.loadTaskLogs()
    } catch (error) {
      console.error('完成任务失败:', error)
      showToast('操作失败')
    } finally {
      this.setData({ actionLoading: false })
    }
  },

  // 取消任务（管理者）
  async cancelTask() {
    if (!this.data.isManager) {
      showToast('无权限操作')
      return
    }

    const confirmed = await showConfirm('确认取消', '确定要取消此任务吗？')
    if (!confirmed) return

    this.setData({ actionLoading: true })
    try {
      await taskService.cancel(this.data.taskId, '管理员取消')
      showSuccess('任务已取消')
      this.loadTaskDetail()
      this.loadTaskLogs()
    } catch (error) {
      console.error('取消任务失败:', error)
      showToast('操作失败')
    } finally {
      this.setData({ actionLoading: false })
    }
  },

  // 跳转到关联案件
  goToCase() {
    const { task } = this.data
    if (task && task.missing_person_id) {
      wx.navigateTo({
        url: `/pages/cases/detail?id=${task.missing_person_id}`
      })
    }
  },

  // 查看执行人信息
  viewAssignee() {
    const { task } = this.data
    if (task && task.assignee) {
      wx.navigateTo({
        url: `/pages/volunteer/profile?id=${task.assignee_id}`
      })
    }
  },

  // 查看位置
  viewLocation() {
    const { task } = this.data
    if (!task || !task.latitude || !task.longitude) {
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

  // 下拉刷新
  onPullDownRefresh() {
    Promise.all([
      this.loadTaskDetail(),
      this.loadTaskLogs()
    ]).finally(() => {
      wx.stopPullDownRefresh()
    })
  },

  // 阻止冒泡
  stopPropagation() {}
})
