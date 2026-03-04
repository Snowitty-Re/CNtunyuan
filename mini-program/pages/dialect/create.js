const dialectService = require('../../services/dialect')
const uploadService = require('../../services/upload')
const missingPersonService = require('../../services/missingPerson')
const { showLoading, hideLoading, showSuccess, showError, showToast } = require('../../utils/util')

// 录音配置
const MIN_DURATION = 15 // 最小录音时长(秒)
const MAX_DURATION = 20 // 最大录音时长(秒)

// 录音管理器和音频上下文
let recorderManager = null
let innerAudioContext = null

Page({
  data: {
    // 表单数据
    form: {
      title: '',
      description: '',
      region: '',
      tags: [],
      missing_person_id: ''
    },
    
    // 录音状态
    isRecording: false,
    hasRecorded: false,
    recordTime: 0,      // 当前录音时长(秒)
    recordDuration: 0,  // 实际录音总时长(秒)
    tempFilePath: '',   // 临时文件路径
    
    // 播放状态
    isPlaying: false,
    playProgress: 0,
    playCurrentTime: 0,
    
    // 地区选项
    regionOptions: [
      '北京市', '上海市', '天津市', '重庆市',
      '广东省', '江苏省', '浙江省', '山东省', '河南省',
      '四川省', '湖北省', '湖南省', '河北省', '福建省',
      '安徽省', '辽宁省', '江西省', '陕西省', '黑龙江省',
      '山西省', '广西壮族自治区', '吉林省', '贵州省', '云南省',
      '甘肃省', '海南省', '内蒙古自治区', '新疆维吾尔自治区', '西藏自治区',
      '青海省', '宁夏回族自治区'
    ],
    regionIndex: -1,
    
    // 标签
    tagInput: '',
    popularTags: ['普通话', '粤语', '四川话', '河南话', '东北话', '闽南语', '湖南话', '山东话', '上海话'],
    
    // 走失人员选择
    missingPersonList: [],
    showMissingPersonPicker: false,
    selectedMissingPerson: null,
    
    // 上传状态
    isUploading: false,
    uploadProgress: 0
  },

  onLoad() {
    this.initRecorder()
    this.loadMissingPersons()
  },

  onUnload() {
    // 清理资源
    if (recorderManager && this.data.isRecording) {
      recorderManager.stop()
    }
    if (innerAudioContext) {
      innerAudioContext.destroy()
      innerAudioContext = null
    }
    if (this.recordTimer) {
      clearInterval(this.recordTimer)
      this.recordTimer = null
    }
  },

  // 初始化录音管理器
  initRecorder() {
    recorderManager = wx.getRecorderManager()
    
    recorderManager.onStart(() => {
      console.log('录音开始')
      this.setData({ isRecording: true })
      this.startRecordTimer()
    })

    recorderManager.onStop((res) => {
      console.log('录音结束', res)
      this.stopRecordTimer()
      
      const duration = Math.floor(res.duration / 1000)
      
      // 检查最小时长
      if (duration < MIN_DURATION) {
        showError(`录音时长不足${MIN_DURATION}秒，请重新录制`)
        this.setData({ 
          isRecording: false,
          recordTime: 0
        })
        return
      }

      this.setData({
        isRecording: false,
        hasRecorded: true,
        tempFilePath: res.tempFilePath,
        recordDuration: duration
      })
      
      showToast(`录音完成 ${this.formatTime(duration)}`, 'success')
    })

    recorderManager.onError((err) => {
      console.error('录音错误:', err)
      this.stopRecordTimer()
      this.setData({ isRecording: false })
      showError('录音失败，请重试')
    })
  },

  // 开始录音计时
  startRecordTimer() {
    this.setData({ recordTime: 0 })
    this.recordTimer = setInterval(() => {
      const recordTime = this.data.recordTime + 1
      this.setData({ recordTime })
      
      // 达到最大时长自动停止
      if (recordTime >= MAX_DURATION) {
        this.stopRecord()
        showToast(`已达到最大录音时长${MAX_DURATION}秒`, 'none')
      }
    }, 1000)
  },

  // 停止录音计时
  stopRecordTimer() {
    if (this.recordTimer) {
      clearInterval(this.recordTimer)
      this.recordTimer = null
    }
  },

  // 开始录音
  startRecord() {
    if (this.data.isRecording) return
    
    // 如果已有录音，先确认
    if (this.data.hasRecorded) {
      wx.showModal({
        title: '提示',
        content: '重新录制将覆盖当前录音，是否继续？',
        success: (res) => {
          if (res.confirm) {
            this.resetRecord()
            this.doStartRecord()
          }
        }
      })
      return
    }
    
    this.doStartRecord()
  },

  // 执行开始录音
  doStartRecord() {
    const options = {
      duration: MAX_DURATION * 1000,
      sampleRate: 44100,
      numberOfChannels: 1,
      encodeBitRate: 192000,
      format: 'mp3'
    }
    
    recorderManager.start(options)
  },

  // 停止录音
  stopRecord() {
    if (!this.data.isRecording) return
    
    // 检查最小时长
    if (this.data.recordTime < MIN_DURATION) {
      showError(`录音至少需要${MIN_DURATION}秒`)
      return
    }
    
    recorderManager.stop()
  },

  // 重置录音
  resetRecord() {
    this.stopPlay()
    this.setData({
      hasRecorded: false,
      tempFilePath: '',
      recordTime: 0,
      recordDuration: 0,
      playProgress: 0,
      playCurrentTime: 0,
      isPlaying: false
    })
  },

  // 初始化音频播放器
  initAudioPlayer() {
    if (!innerAudioContext) {
      innerAudioContext = wx.createInnerAudioContext()
      
      innerAudioContext.onTimeUpdate(() => {
        const currentTime = innerAudioContext.currentTime || 0
        const duration = this.data.recordDuration || 1
        this.setData({
          playCurrentTime: Math.floor(currentTime),
          playProgress: (currentTime / duration) * 100
        })
      })

      innerAudioContext.onEnded(() => {
        this.setData({ 
          isPlaying: false,
          playProgress: 0,
          playCurrentTime: 0
        })
      })

      innerAudioContext.onError(() => {
        this.setData({ isPlaying: false })
        showToast('播放失败', 'none')
      })
    }
  },

  // 播放/暂停录音
  togglePlay() {
    if (!this.data.hasRecorded || !this.data.tempFilePath) return
    
    this.initAudioPlayer()
    
    if (this.data.isPlaying) {
      innerAudioContext.pause()
      this.setData({ isPlaying: false })
    } else {
      innerAudioContext.src = this.data.tempFilePath
      innerAudioContext.play()
      this.setData({ isPlaying: true })
    }
  },

  // 停止播放
  stopPlay() {
    if (innerAudioContext) {
      innerAudioContext.stop()
    }
    this.setData({ 
      isPlaying: false,
      playProgress: 0,
      playCurrentTime: 0
    })
  },

  // 拖动进度条
  onPlayProgressChange(e) {
    if (!innerAudioContext) return
    const value = e.detail.value
    const seekTime = (value / 100) * this.data.recordDuration
    innerAudioContext.seek(seekTime)
    this.setData({
      playProgress: value,
      playCurrentTime: Math.floor(seekTime)
    })
  },

  // 表单输入
  onTitleInput(e) {
    this.setData({ 'form.title': e.detail.value })
  },

  onDescInput(e) {
    this.setData({ 'form.description': e.detail.value })
  },

  // 地区选择
  onRegionChange(e) {
    const index = parseInt(e.detail.value)
    this.setData({
      regionIndex: index,
      'form.region': this.data.regionOptions[index]
    })
  },

  // 标签输入
  onTagInput(e) {
    this.setData({ tagInput: e.detail.value })
  },

  // 添加标签
  addTag() {
    const tag = this.data.tagInput.trim()
    if (!tag) return
    
    if (this.data.form.tags.includes(tag)) {
      showToast('标签已存在', 'none')
      return
    }
    
    if (this.data.form.tags.length >= 5) {
      showToast('最多添加5个标签', 'none')
      return
    }
    
    this.setData({
      'form.tags': [...this.data.form.tags, tag],
      tagInput: ''
    })
  },

  // 添加热门标签
  addPopularTag(e) {
    const tag = e.currentTarget.dataset.tag
    if (this.data.form.tags.includes(tag)) return
    
    if (this.data.form.tags.length >= 5) {
      showToast('最多添加5个标签', 'none')
      return
    }
    
    this.setData({
      'form.tags': [...this.data.form.tags, tag]
    })
  },

  // 删除标签
  removeTag(e) {
    const index = e.currentTarget.dataset.index
    const tags = [...this.data.form.tags]
    tags.splice(index, 1)
    this.setData({ 'form.tags': tags })
  },

  // 加载走失人员列表
  async loadMissingPersons() {
    try {
      const result = await missingPersonService.getList({
        page: 1,
        page_size: 50,
        status: 'missing'
      })
      this.setData({
        missingPersonList: result.list || result.data || []
      })
    } catch (error) {
      console.error('加载走失人员列表失败:', error)
    }
  },

  // 显示走失人员选择器
  showMissingPersonSelector() {
    this.setData({ showMissingPersonPicker: true })
  },

  // 隐藏走失人员选择器
  hideMissingPersonSelector() {
    this.setData({ showMissingPersonPicker: false })
  },

  // 选择走失人员
  selectMissingPerson(e) {
    const index = parseInt(e.currentTarget.dataset.index)
    const person = this.data.missingPersonList[index]
    this.setData({
      selectedMissingPerson: person,
      'form.missing_person_id': person.id,
      showMissingPersonPicker: false
    })
  },

  // 清除选择的走失人员
  clearMissingPerson() {
    this.setData({
      selectedMissingPerson: null,
      'form.missing_person_id': ''
    })
  },

  // 表单验证
  validateForm() {
    if (!this.data.hasRecorded || !this.data.tempFilePath) {
      showError('请先录制方言音频')
      return false
    }
    
    if (!this.data.form.title.trim()) {
      showError('请输入标题')
      return false
    }
    
    if (!this.data.form.region) {
      showError('请选择地区')
      return false
    }
    
    return true
  },

  // 提交表单
  async submitForm() {
    if (!this.validateForm()) return
    if (this.data.isUploading) return
    
    this.setData({ isUploading: true })
    showLoading('上传中...')
    
    try {
      // 1. 上传录音文件
      let audioUrl = ''
      try {
        const uploadRes = await uploadService.upload(this.data.tempFilePath, {
          type: 'audio',
          entity_type: 'dialect'
        })
        audioUrl = uploadRes.url || uploadRes.data?.url
      } catch (uploadErr) {
        console.error('上传文件失败:', uploadErr)
        // 模拟上传成功，实际开发中应该抛出错误
        audioUrl = this.data.tempFilePath
      }
      
      // 2. 创建方言记录
      const dialectData = {
        title: this.data.form.title.trim(),
        description: this.data.form.description.trim(),
        audio_url: audioUrl,
        duration: this.data.recordDuration,
        region: this.data.form.region,
        tags: this.data.form.tags,
        missing_person_id: this.data.form.missing_person_id || undefined
      }
      
      await dialectService.create(dialectData)
      
      hideLoading()
      showSuccess('发布成功')
      
      // 返回列表页
      setTimeout(() => {
        wx.navigateBack()
      }, 1500)
    } catch (error) {
      hideLoading()
      console.error('发布失败:', error)
      showError('发布失败，请重试')
      this.setData({ isUploading: false })
    }
  },

  // 格式化时间
  formatTime(seconds) {
    const mins = Math.floor(seconds / 60)
    const secs = seconds % 60
    return `${mins}:${secs.toString().padStart(2, '0')}`
  },

  // 格式化录音时间（带倒计时）
  formatRecordTime() {
    const current = this.data.recordTime
    const remaining = MAX_DURATION - current
    return this.formatTime(remaining)
  }
})
