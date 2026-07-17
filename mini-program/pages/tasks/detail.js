const taskService = require('../../services/task')
const userService = require('../../services/user')
const { formatDate, formatTimeAgo, showSuccess, showToast, showConfirm, showLoading, hideLoading } = require('../../utils/util')
const { TASK_STATUS_MAP, TASK_PRIORITY_MAP, TASK_PRIORITY_COLOR } = require('../../utils/constants')
const { ACTIONS } = require('../../utils/permission')
const app = getApp()
const TASK_LIST_DIRTY_KEY = 'tasks_list_dirty'

Page({
  data: {
    taskId: '',
    task: null,
    logs: [],
    followUps: [],
    loading: false,
    actionLoading: false,
    showProgressModal: false,
    progressInput: 50,
    currentUser: null,
    canManageTask: false,
    canEditTask: false,
    canExecuteTask: false,
    isAssignee: false,
    canOperateAsAssignee: false,
    assignUsers: [],
    statusMap:        TASK_STATUS_MAP,
    priorityMap:      TASK_PRIORITY_MAP,
    priorityColorMap: TASK_PRIORITY_COLOR
  },

  onLoad(options) {
    if (!app.ensurePhoneBound || !app.ensurePhoneBound({ message: '查看任务详情需绑定手机号' })) return
    const userInfo = wx.getStorageSync('userInfo') || {}
    this.setData({ 
      taskId: options.id,
      currentUser: userInfo,
      canManageTask: app.hasPermission(ACTIONS.TASK_MANAGE, userInfo),
      canEditTask: app.hasPermission(ACTIONS.TASK_EDIT, userInfo),
      canExecuteTask: app.hasPermission(ACTIONS.TASK_EXECUTE, userInfo)
    })
    
    if (options.id) {
      this.loadTaskDetail()
      this.loadTaskLogs()
      this.loadFollowUps()
    }
  },

  onShow() {
    if (!app.ensurePhoneBound || !app.ensurePhoneBound({ message: '查看任务详情需绑定手机号' })) return
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
      const canOperateAsAssignee = isAssignee && this.data.canExecuteTask
      
      // 更新页面标题
      wx.setNavigationBarTitle({ title: task.title || '任务详情' })
      
      this.setData({ 
        task, 
        isAssignee,
        canOperateAsAssignee,
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
        statusText: item.status === 'pending' ? '待审核' : (item.status === 'approved' ? '已通过' : '已驳回'),
        reviewerName: item.reviewer ? (item.reviewer.nickname || item.reviewer.name || '') : '',
        reviewedAtText: item.reviewed_at ? formatTimeAgo(item.reviewed_at) : ''
      }))
      this.setData({ followUps })
    } catch (error) {
      console.error('加载任务跟进失败:', error)
    }
  },

  // 分配任务（管理者）- 人员选择器
  async assignTask() {
    if (!this.data.canManageTask) {
      showToast('无权限操作')
      return
    }

    try {
      showLoading('加载人员...')
      let users = this.data.assignUsers
      if (!users.length) {
        const res = await userService.getList({ page: 1, page_size: 50, status: 'active' })
        users = (res.list || []).filter((u) => u && u.id)
        this.setData({ assignUsers: users })
      }
      hideLoading()
      if (!users.length) {
        showToast('暂无可分配人员')
        return
      }
      const names = users.map((u) => u.nickname || u.phone || u.id)
      wx.showActionSheet({
        itemList: names.slice(0, 6),
        success: async ({ tapIndex }) => {
          const user = users[tapIndex]
          if (!user) return
          try {
            showLoading('分配中...')
            await taskService.assign(this.data.taskId, user.id)
            hideLoading()
            showSuccess('分配成功')
            this.loadTaskDetail()
            this.loadTaskLogs()
          } catch (error) {
            hideLoading()
            showToast((error && error.message) || '分配失败')
          }
        }
      })
    } catch (error) {
      hideLoading()
      showToast((error && error.message) || '加载人员失败')
    }
  },

  // 开始任务（执行人）
  async startTask() {
    if (!this.data.canOperateAsAssignee) {
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
    if (!this.data.isAssignee || !this.data.canEditTask) {
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
    if (!this.data.canOperateAsAssignee) {
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
    if (!this.data.canOperateAsAssignee) {
      showToast('只有执行人可以提交反馈')
      return
    }
    wx.navigateTo({ url: `/pages/tasks/feedback?id=${this.data.taskId}` })
  },

  // 新增任务跟进
  addFollowUp() {
    if (!this.data.canOperateAsAssignee && !this.data.canManageTask) {
      showToast('无权限操作')
      return
    }
    wx.navigateTo({ url: `/pages/tasks/follow-up-create?id=${this.data.taskId}` })
  },

  // 打开跟进详情
  goToFollowUpDetail(e) {
    const followUpID = e.currentTarget.dataset.id
    wx.navigateTo({
      url: `/pages/tasks/follow-up-detail?taskId=${this.data.taskId}&followUpId=${followUpID}`
    })
  },

  // 取消任务（管理者）
  async cancelTask() {
    if (!this.data.canManageTask) {
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

  // 删除任务（管理者）
  async deleteTask() {
    if (!this.data.canManageTask) {
      showToast('无权限操作')
      return
    }

    const confirmed = await showConfirm('确认删除', '删除后不可恢复，是否继续？')
    if (!confirmed) return

    this.setData({ actionLoading: true })
    try {
      await taskService.delete(this.data.taskId)
      showSuccess('任务已删除')
      wx.setStorageSync(TASK_LIST_DIRTY_KEY, 1)
      setTimeout(() => wx.navigateBack(), 400)
    } catch (error) {
      console.error('删除任务失败:', error)
      showToast(error.message || '删除失败')
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

  // 查看执行人信息（资料页仅支持当前用户，改为弹窗展示）
  viewAssignee() {
    const { task } = this.data
    if (!task || !task.assignee) return
    const name = task.assignee.nickname || task.assignee.name || '未命名'
    const phone = task.assignee.phone || '未绑定手机'
    wx.showModal({
      title: '执行人',
      content: `${name}\n${phone}`,
      showCancel: false
    })
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
