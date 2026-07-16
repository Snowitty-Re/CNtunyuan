const taskService = require('../../services/task')
const { formatDate, formatTimeAgo, showSuccess, showToast } = require('../../utils/util')
const { ACTIONS } = require('../../utils/permission')
const app = getApp()

Page({
  data: {
    taskId: '',
    followUpId: '',
    followUp: null,
    comments: [],
    commentInput: '',
    loading: false,
    submitting: false,
    canReview: false
  },

  onLoad(options = {}) {
    if (!app.ensurePhoneBound || !app.ensurePhoneBound({ message: '查看任务跟进需绑定手机号' })) return
    const { taskId, followUpId } = options
    if (!taskId || !followUpId) {
      showToast('记录参数错误')
      wx.navigateBack()
      return
    }
    this.setData({
      taskId,
      followUpId,
      canReview: app.hasPermission(ACTIONS.TASK_MANAGE)
    })
    this.loadData()
  },

  onShow() {
    if (!app.ensurePhoneBound || !app.ensurePhoneBound({ message: '查看任务跟进需绑定手机号' })) return
    if (this.data.taskId && this.data.followUpId) {
      this.loadData()
    }
  },

  async loadData() {
    this.setData({ loading: true })
    try {
      await Promise.all([this.loadFollowUp(), this.loadComments()])
    } finally {
      this.setData({ loading: false })
    }
  },

  async loadFollowUp() {
    try {
      const followUp = await taskService.getFollowUpById(this.data.taskId, this.data.followUpId)
      const data = {
        ...followUp,
        createdAtText: followUp.created_at ? formatDate(followUp.created_at, 'YYYY-MM-DD HH:mm') : '',
        reviewedAtText: followUp.reviewed_at ? formatDate(followUp.reviewed_at, 'YYYY-MM-DD HH:mm') : '',
        statusText: followUp.status === 'pending' ? '待审核' : (followUp.status === 'approved' ? '已通过' : '已驳回'),
        attachments: Array.isArray(followUp.attachments) ? followUp.attachments : []
      }
      this.setData({ followUp: data })
    } catch (error) {
      showToast(error.message || '加载记录失败')
    }
  },

  async loadComments() {
    try {
      const result = await taskService.getFollowUpComments(this.data.taskId, this.data.followUpId, { page: 1, page_size: 100 })
      const list = (result.list || []).map(item => ({
        ...item,
        createdAtText: formatTimeAgo(item.created_at),
        isReviewComment: app.hasPermission(ACTIONS.TASK_MANAGE, item.user || {}),
        commentTagText: app.hasPermission(ACTIONS.TASK_MANAGE, item.user || {}) ? '审核意见' : '执行备注'
      }))
      this.setData({ comments: list })
    } catch (error) {
      showToast(error.message || '加载评论失败')
    }
  },

  previewAttachment(e) {
    const current = e.currentTarget.dataset.url
    const urls = this.data.followUp ? this.data.followUp.attachments || [] : []
    if (/\.(png|jpg|jpeg|gif|webp)$/i.test(current)) {
      wx.previewImage({ current, urls: urls.length > 0 ? urls : [current] })
      return
    }
    wx.setClipboardData({
      data: current,
      success: () => showSuccess('附件链接已复制')
    })
  },

  onCommentInput(e) {
    this.setData({ commentInput: e.detail.value || '' })
  },

  async submitComment() {
    const content = (this.data.commentInput || '').trim()
    if (!content) {
      showToast('评论不能为空')
      return
    }
    this.setData({ submitting: true })
    try {
      await taskService.addFollowUpComment(this.data.taskId, this.data.followUpId, { content })
      showSuccess('评论成功')
      this.setData({ commentInput: '' })
      this.loadComments()
      this.loadFollowUp()
    } catch (error) {
      showToast(error.message || '评论失败')
    } finally {
      this.setData({ submitting: false })
    }
  },

  reviewFollowUp(e) {
    if (!this.data.canReview) {
      showToast('仅管理员可审核')
      return
    }
    const approve = !!e.currentTarget.dataset.approve
    const title = approve ? '审核通过' : '审核驳回'

    wx.showModal({
      title,
      editable: true,
      placeholderText: approve ? '审核备注（可选）' : '请输入驳回原因',
      success: async (res) => {
        if (!res.confirm) return
        try {
          await taskService.reviewFollowUp(this.data.taskId, this.data.followUpId, {
            approve,
            remark: res.content || ''
          })
          showSuccess(approve ? '已通过' : '已驳回')
          this.loadFollowUp()
          this.loadComments()
        } catch (error) {
          showToast(error.message || '审核失败')
        }
      }
    })
  },

  onPullDownRefresh() {
    this.loadData().finally(() => wx.stopPullDownRefresh())
  }
})
