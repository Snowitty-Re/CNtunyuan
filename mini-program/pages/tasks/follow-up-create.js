const taskService = require('../../services/task')
const uploadService = require('../../services/upload')
const { showSuccess, showToast } = require('../../utils/util')
const app = getApp()

Page({
  data: {
    taskId: '',
    content: '',
    progress: 0,
    useProgress: true,
    attachments: [],
    submitting: false
  },

  onLoad(options = {}) {
    if (!app.ensurePhoneBound || !app.ensurePhoneBound({ message: '任务跟进需绑定手机号' })) return
    if (!options.id) {
      showToast('任务ID无效')
      wx.navigateBack()
      return
    }
    this.setData({ taskId: options.id })
  },

  onContentInput(e) {
    this.setData({ content: e.detail.value || '' })
  },

  onProgressChange(e) {
    this.setData({ progress: Number(e.detail.value || 0) })
  },

  onUseProgressChange(e) {
    this.setData({ useProgress: !!e.detail.value.length })
  },

  chooseAttachment() {
    const remain = 9 - this.data.attachments.length
    if (remain <= 0) {
      showToast('最多上传9个附件')
      return
    }

    wx.chooseMedia({
      count: remain,
      mediaType: ['image'],
      sourceType: ['album', 'camera'],
      success: (res) => {
        const next = res.tempFiles.map(item => item.tempFilePath)
        this.setData({ attachments: [...this.data.attachments, ...next] })
      }
    })
  },

  removeAttachment(e) {
    const idx = Number(e.currentTarget.dataset.index)
    const next = this.data.attachments.filter((_, i) => i !== idx)
    this.setData({ attachments: next })
  },

  previewAttachment(e) {
    const url = e.currentTarget.dataset.url
    wx.previewImage({ urls: this.data.attachments, current: url })
  },

  async submit() {
    if (this.data.submitting) return
    const content = (this.data.content || '').trim()
    if (!content) {
      showToast('请填写跟进内容')
      return
    }

    this.setData({ submitting: true })
    try {
      let attachmentUrls = []
      if (this.data.attachments.length > 0) {
        const uploadResults = await Promise.all(this.data.attachments.map(path => uploadService.upload(path)))
        attachmentUrls = uploadResults.map(item => item.url || '').filter(Boolean)
      }

      const payload = {
        content,
        attachments: attachmentUrls
      }
      if (this.data.useProgress) {
        payload.progress = this.data.progress
      }

      await taskService.createFollowUp(this.data.taskId, payload)
      showSuccess('跟进已提交')
      wx.navigateBack()
    } catch (error) {
      showToast(error.message || '提交失败')
    } finally {
      this.setData({ submitting: false })
    }
  }
})
