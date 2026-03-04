const services = require('../../services/index')
const { showSuccess, showError, showLoading, hideLoading } = require('../../utils/util')

Page({
  data: {
    taskId: '',
    task: null,
    feedback: '',
    result: '',
    photos: [],
    loading: false
  },

  onLoad(options) {
    if (options.id) {
      this.setData({ taskId: options.id })
      this.loadTaskDetail(options.id)
    }
  },

  // 加载任务详情
  async loadTaskDetail(id) {
    try {
      const task = await services.task.getById(id)
      this.setData({ task })
    } catch (error) {
      console.error('加载任务失败:', error)
      showError('加载任务失败')
    }
  },

  // 反馈内容输入
  onFeedbackInput(e) {
    this.setData({ feedback: e.detail.value })
  },

  // 结果输入
  onResultInput(e) {
    this.setData({ result: e.detail.value })
  },

  // 选择图片
  chooseImage() {
    const { photos } = this.data
    const remainCount = 9 - photos.length
    
    if (remainCount <= 0) {
      showError('最多上传9张图片')
      return
    }

    wx.chooseMedia({
      count: remainCount,
      mediaType: ['image'],
      sourceType: ['album', 'camera'],
      success: (res) => {
        const newPhotos = res.tempFiles.map(file => file.tempFilePath)
        this.setData({
          photos: [...photos, ...newPhotos]
        })
      }
    })
  },

  // 删除图片
  deleteImage(e) {
    const index = e.currentTarget.dataset.index
    const photos = this.data.photos.filter((_, i) => i !== index)
    this.setData({ photos })
  },

  // 预览图片
  previewImage(e) {
    const url = e.currentTarget.dataset.url
    wx.previewImage({
      urls: this.data.photos,
      current: url
    })
  },

  // 提交反馈
  async submitFeedback() {
    const { taskId, feedback, result, photos } = this.data
    
    if (!feedback.trim()) {
      showError('请输入反馈内容')
      return
    }

    this.setData({ loading: true })
    showLoading('提交中...')

    try {
      // 先上传图片
      let photoUrls = []
      if (photos.length > 0) {
        const uploadResults = await Promise.all(
          photos.map(path => services.upload.upload(path))
        )
        photoUrls = uploadResults.map(r => r.url)
      }

      // 提交任务完成
      await services.task.complete(taskId, {
        result: result || feedback,
        feedback,
        attachments: photoUrls
      })

      hideLoading()
      showSuccess('提交成功')
      
      setTimeout(() => {
        wx.navigateBack()
      }, 1500)
    } catch (error) {
      hideLoading()
      console.error('提交失败:', error)
      showError(error.message || '提交失败')
    } finally {
      this.setData({ loading: false })
    }
  }
})
