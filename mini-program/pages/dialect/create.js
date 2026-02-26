const { post, get } = require('../../utils/request')
const { showLoading, hideLoading, showSuccess, showError } = require('../../utils/util')

const MAX_DURATION = 20 // 最大录音时长
const MIN_DURATION = 15 // 最小录音时长

Page({
  data: {
    form: {
      title: '',
      description: '',
      address: '',
      detailAddress: '',
      latitude: 0,
      longitude: 0
    },
    isRecording: false,
    hasRecorded: false,
    recordTime: 0,
    recordDuration: 0,
    playProgress: 0,
    currentTime: 0,
    tempFilePath: '',
    uploading: false,
    myDialects: []
  },

  recorderManager: null,
  innerAudioContext: null,
  recordTimer: null,

  onLoad() {
    this.initRecorder()
    this.loadMyDialects()
  },

  onUnload() {
    if (this.innerAudioContext) {
      this.innerAudioContext.destroy()
    }
  },

  initRecorder() {
    this.recorderManager = wx.getRecorderManager()
    
    this.recorderManager.onStart(() => {
      console.log('录音开始')
      this.setData({ isRecording: true })
      this.startRecordTimer()
    })

    this.recorderManager.onStop((res) => {
      console.log('录音结束', res)
      this.stopRecordTimer()
      
      if (res.duration < MIN_DURATION * 1000) {
        showError(`录音时长不足${MIN_DURATION}秒`)
        this.setData({ isRecording: false })
        return
      }

      this.setData({
        isRecording: false,
        hasRecorded: true,
        tempFilePath: res.tempFilePath,
        recordDuration: Math.floor(res.duration / 1000)
      })
    })

    this.recorderManager.onError((err) => {
      console.error('录音错误:', err)
      this.stopRecordTimer()
      this.setData({ isRecording: false })
      showError('录音失败')
    })
  },

  startRecordTimer() {
    this.setData({ recordTime: 0 })
    this.recordTimer = setInterval(() => {
      const recordTime = this.data.recordTime + 1
      this.setData({ recordTime })
      
      if (recordTime >= MAX_DURATION) {
        this.stopRecord()
      }
    }, 1000)
  },

  stopRecordTimer() {
    if (this.recordTimer) {
      clearInterval(this.recordTimer)
      this.recordTimer = null
    }
  },

  startRecord() {
    if (this.data.isRecording) return
    
    this.recorderManager.start({
      duration: MAX_DURATION * 1000,
      sampleRate: 44100,
      numberOfChannels: 1,
      encodeBitRate: 192000,
      format: 'mp3'
    })
  },

  stopRecord() {
    if (!this.data.isRecording) return
    this.recorderManager.stop()
  },

  toggleRecord() {
    if (this.data.isRecording) {
      this.stopRecord()
    } else if (this.data.hasRecorded) {
      this.playRecord()
    }
  },

  playRecord() {
    if (!this.innerAudioContext) {
      this.innerAudioContext = wx.createInnerAudioContext()
      
      this.innerAudioContext.onTimeUpdate(() => {
        const currentTime = Math.floor(this.innerAudioContext.currentTime)
        const progress = (currentTime / this.data.recordDuration) * 100
        this.setData({
          currentTime,
          playProgress: progress
        })
      })

      this.innerAudioContext.onEnded(() => {
        this.setData({ playProgress: 0, currentTime: 0 })
      })
    }

    this.innerAudioContext.src = this.data.tempFilePath
    this.innerAudioContext.play()
  },

  reRecord() {
    this.setData({
      hasRecorded: false,
      tempFilePath: '',
      recordTime: 0,
      playProgress: 0,
      currentTime: 0
    })
    if (this.innerAudioContext) {
      this.innerAudioContext.stop()
    }
  },

  async uploadRecord() {
    if (!this.data.form.title) {
      showError('请输入标题')
      return
    }

    this.setData({ uploading: true })
    showLoading('上传中...')

    try {
      // 先上传录音文件
      // const uploadRes = await uploadFile('/upload/audio', this.data.tempFilePath)
      const audioUrl = 'https://example.com/audio.mp3' // 模拟上传后的URL

      // 创建方言记录
      await post('/dialects', {
        title: this.data.form.title,
        description: this.data.form.description,
        audio_url: audioUrl,
        duration: this.data.recordDuration,
        province: '',
        city: '',
        district: '',
        address: this.data.form.detailAddress,
        latitude: this.data.form.latitude,
        longitude: this.data.form.longitude
      })

      hideLoading()
      showSuccess('上传成功')
      
      // 重置表单
      this.setData({
        form: { title: '', description: '', address: '', detailAddress: '' },
        hasRecorded: false,
        tempFilePath: '',
        recordTime: 0,
        uploading: false
      })

      this.loadMyDialects()
    } catch (error) {
      hideLoading()
      this.setData({ uploading: false })
      showError('上传失败')
    }
  },

  async loadMyDialects() {
    try {
      const result = await get('/dialects', { page: 1, page_size: 5 })
      this.setData({ myDialects: result.list })
    } catch (error) {
      console.error('加载我的录音失败:', error)
    }
  },

  chooseLocation() {
    wx.chooseLocation({
      success: (res) => {
        this.setData({
          'form.address': res.name,
          'form.detailAddress': res.address,
          'form.latitude': res.latitude,
          'form.longitude': res.longitude
        })
      }
    })
  },

  onTitleInput(e) {
    this.setData({ 'form.title': e.detail.value })
  },

  onDescInput(e) {
    this.setData({ 'form.description': e.detail.value })
  },

  onAddressInput(e) {
    this.setData({ 'form.detailAddress': e.detail.value })
  }
})
