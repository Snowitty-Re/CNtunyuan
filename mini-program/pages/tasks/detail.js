const taskService = require('../../services/task')
const { formatDate, formatTimeAgo, showSuccess, showToast, showConfirm, showLoading, hideLoading } = require('../../utils/util')
const { TASK_STATUS_MAP, TASK_PRIORITY_MAP, TASK_PRIORITY_COLOR } = require('../../utils/constants')
const app = getApp()

Page({
  data: {
    taskId: '',
    task: null,
    logs: [],
    followUps: [],
    showFollowUpComments: false,
    activeFollowUpId: '',
    activeFollowUpTitle: '',
    followUpComments: [],
    followUpCommentsLoading: false,
    followUpCommentInput: '',
    loading: false,
    actionLoading: false,
    showProgressModal: false,
    progressInput: 50,
    currentUser: null,
    isManager: false,
    isAssignee: false,
    statusMap:        TASK_STATUS_MAP,
    priorityMap:      TASK_PRIORITY_MAP,
    priorityColorMap: TASK_PRIORITY_COLOR
  },

  onLoad(options) {
    if (!app.ensureAuth || !app.ensureAuth()) return
    const userInfo = wx.getStorageSync('userInfo') || {}
    this.setData({ 
      taskId: options.id,
      currentUser: userInfo,
      isManager: app.isManager()
    })
    
    if (options.id) {
      this.loadTaskDetail()
      this.loadTaskLogs()
      this.loadFollowUps()
    }
  },

  onShow() {
    if (!app.ensureAuth || !app.ensureAuth()) return
    if (this.data.taskId) {
      this.loadTaskDetail()
      this.loadTaskLogs()
      this.loadFollowUps()
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

  // 加载任务跟进
  async loadFollowUps() {
    try {
      const result = await taskService.getFollowUps(this.data.taskId, { page: 1, page_size: 20 })
      const followUps = (result.list || []).map(item => ({
        ...item,
        createdAtText: formatTimeAgo(item.created_at),
        attachments: Array.isArray(item.attachments) ? item.attachments : []
      }))
      this.setData({ followUps })
    } catch (error) {
      console.error('加载任务跟进失败:', error)
    }
  },

  // 分配任务（管理者）- 使用页面内选择而非跳转
  async assignTask() {
    if (!this.data.isManager) {
      showToast('无权限操作')
      return
    }

    // 显示输入框让用户输入执行人ID
    wx.showModal({
      title: '分配任务',
      editable: true,
      placeholderText: '请输入执行人ID',
      success: async (res) => {
        if (res.confirm && res.content) {
          try {
            showLoading('分配中...')
            await taskService.assign(this.data.taskId, res.content)
            showSuccess('分配成功')
            this.loadTaskDetail()
            this.loadTaskLogs()
          } catch (error) {
            showToast('分配失败')
          }
        }
      }
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
      this.loadFollowUps()
    } catch (error) {
      console.error('开始任务失败:', error)
      showToast(error.message || '开始任务失败')
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
      this.loadFollowUps()
    } catch (error) {
      console.error('更新进度失败:', error)
      showToast(error.message || '更新失败')
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
      this.loadFollowUps()
    } catch (error) {
      console.error('完成任务失败:', error)
      showToast(error.message || '完成任务失败')
    } finally {
      this.setData({ actionLoading: false })
    }
  },

  // 跳转到反馈页（反馈+附件后完成任务）
  submitFeedback() {
    if (!this.data.isAssignee) {
      showToast('只有执行人可以提交反馈')
      return
    }
    wx.navigateTo({ url: `/pages/tasks/feedback?id=${this.data.taskId}` })
  },

  // 新增任务跟进
  addFollowUp() {
    if (!this.data.isAssignee && !this.data.isManager) {
      showToast('无权限操作')
      return
    }
    wx.navigateTo({ url: `/pages/tasks/follow-up-create?id=${this.data.taskId}` })
  },

  // 审核跟进
  reviewFollowUp(e) {
    if (!this.data.isManager) {
      showToast('仅管理员可审核')
      return
    }
    const id = e.currentTarget.dataset.id
    const approve = !!e.currentTarget.dataset.approve
    const title = approve ? '审核通过' : '审核驳回'

    wx.showModal({
      title,
      editable: true,
      placeholderText: approve ? '审核备注（可选）' : '请输入驳回原因',
      success: async (res) => {
        if (!res.confirm) return
        try {
          await taskService.reviewFollowUp(this.data.taskId, id, {
            approve,
            remark: res.content || ''
          })
          showSuccess(approve ? '已通过' : '已驳回')
          this.loadFollowUps()
          this.loadTaskLogs()
        } catch (error) {
          showToast(error.message || '审核失败')
        }
      }
    })
  },

  // 评论跟进
  addFollowUpComment(e) {
    const id = e.currentTarget.dataset.id
    const author = e.currentTarget.dataset.author || '执行人'
    this.setData({
      showFollowUpComments: true,
      activeFollowUpId: id,
      activeFollowUpTitle: author,
      followUpComments: [],
      followUpCommentInput: ''
    })
    this.loadFollowUpComments()
  },

  async loadFollowUpComments() {
    const followUpID = this.data.activeFollowUpId
    if (!followUpID) return
    this.setData({ followUpCommentsLoading: true })
    try {
      const result = await taskService.getFollowUpComments(this.data.taskId, followUpID, { page: 1, page_size: 50 })
      const list = (result.list || []).map(item => ({
        ...item,
        createdAtText: formatTimeAgo(item.created_at),
        isReviewComment: ['super_admin', 'admin', 'manager'].includes((item.user && item.user.role) || ''),
        commentTagText: ['super_admin', 'admin', 'manager'].includes((item.user && item.user.role) || '')
          ? '审核意见'
          : '执行备注'
      }))
      this.setData({ followUpComments: list })
    } catch (error) {
      showToast(error.message || '加载评论失败')
    } finally {
      this.setData({ followUpCommentsLoading: false })
    }
  },

  closeFollowUpComments() {
    this.setData({
      showFollowUpComments: false,
      activeFollowUpId: '',
      activeFollowUpTitle: '',
      followUpComments: [],
      followUpCommentInput: ''
    })
  },

  onFollowUpCommentInput(e) {
    this.setData({ followUpCommentInput: e.detail.value || '' })
  },

  async submitFollowUpComment() {
    const followUpID = this.data.activeFollowUpId
    if (!followUpID) return
    const content = (this.data.followUpCommentInput || '').trim()
    if (!content) {
      showToast('评论不能为空')
      return
    }
    try {
      await taskService.addFollowUpComment(this.data.taskId, followUpID, { content })
      showSuccess('评论成功')
      this.setData({ followUpCommentInput: '' })
      this.loadFollowUpComments()
      this.loadFollowUps()
    } catch (error) {
      showToast(error.message || '评论失败')
    }
  },

  previewFollowUpAttachment(e) {
    const current = e.currentTarget.dataset.url
    const rawUrls = e.currentTarget.dataset.urls
    const urls = Array.isArray(rawUrls) ? rawUrls : [current]
    if (/\.(png|jpg|jpeg|gif|webp)$/i.test(current)) {
      wx.previewImage({ current, urls })
      return
    }
    wx.setClipboardData({
      data: current,
      success: () => showSuccess('附件链接已复制')
    })
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
      this.loadFollowUps()
    } catch (error) {
      console.error('取消任务失败:', error)
      showToast(error.message || '取消任务失败')
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
    // 后端返回的是 lat/lng 而非 latitude/longitude
    if (!task || !task.lat || !task.lng) {
      showToast('暂无位置信息')
      return
    }
    wx.openLocation({
      latitude: parseFloat(task.lat),
      longitude: parseFloat(task.lng),
      name: task.address || '任务位置',
      address: task.address
    })
  },

  // 下拉刷新
  onPullDownRefresh() {
    Promise.all([
      this.loadTaskDetail(),
      this.loadTaskLogs(),
      this.loadFollowUps()
    ]).finally(() => {
      wx.stopPullDownRefresh()
    })
  },

  // 阻止冒泡
  stopPropagation() {}
})
