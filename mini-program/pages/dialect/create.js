const dialectService = require('../../services/dialect')
const uploadService = require('../../services/upload')
const missingPersonService = require('../../services/missingPerson')
const { showLoading, hideLoading, showSuccess, showError, showToast } = require('../../utils/util')
const app = getApp()

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
      province: '',
      city: '',
      district: '',
      collect_address: '',
      collect_latitude: 0,
      collect_longitude: 0,
      tags: [],
      missing_person_id: ''
    },
    
    // 录音时长限制（同步模块常量供 WXML 使用）
    minDuration: MIN_DURATION,
    maxDuration: MAX_DURATION,

    // 录音状态
    isRecording: false,
    hasRecorded: false,
    recordTime: 0,         // 当前录音时长(秒)
    recordTimeText: '0:00',
    recordDuration: 0,     // 实际录音总时长(秒)
    recordDurationText: '0:00',
    tempFilePath: '',      // 临时文件路径

    // 播放状态
    isPlaying: false,
    playProgress: 0,
    playCurrentTime: 0,
    playCurrentTimeText: '0:00',
    
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
    if (!app.ensureAuth || !app.ensureAuth()) return
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
      this.setData({ isRecording: true })
      this.startRecordTimer()
    })

    recorderManager.onStop((res) => {
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
        recordDuration: duration,
        recordDurationText: this._formatTime(duration),
      })
      
      showSuccess(`录音完成 ${this.formatTime(duration)}`)
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
    this.setData({ recordTime: 0, recordTimeText: '0:00' })
    this.recordTimer = setInterval(() => {
      const recordTime = this.data.recordTime + 1
      this.setData({ recordTime, recordTimeText: this._formatTime(recordTime) })
      
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
      recordTimeText: '0:00',
      recordDuration: 0,
      recordDurationText: '0:00',
      playProgress: 0,
      playCurrentTime: 0,
      playCurrentTimeText: '0:00',
      isPlaying: false
    })
  },

  // 初始化音频播放器
  initAudioPlayer() {
    if (!innerAudioContext) {
      innerAudioContext = wx.createInnerAudioContext()
      
      innerAudioContext.onTimeUpdate(() => {
        const currentTime = Math.floor(innerAudioContext.currentTime || 0)
        const duration = this.data.recordDuration || 1
        this.setData({
          playCurrentTime: currentTime,
          playCurrentTimeText: this._formatTime(currentTime),
          playProgress: (currentTime / duration) * 100
        })
      })

      innerAudioContext.onEnded(() => {
        this.setData({
          isPlaying: false,
          playProgress: 0,
          playCurrentTime: 0,
          playCurrentTimeText: '0:00',
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
      playCurrentTime: 0,
      playCurrentTimeText: '0:00',
    })
  },

  // 拖动进度条
  onPlayProgressChange(e) {
    if (!innerAudioContext) return
    const value = e.detail.value
    const seekTime = Math.floor((value / 100) * this.data.recordDuration)
    innerAudioContext.seek(seekTime)
    this.setData({
      playProgress: value,
      playCurrentTime: seekTime,
      playCurrentTimeText: this._formatTime(seekTime),
    })
  },

  _formatTime(seconds) {
    if (!seconds || isNaN(seconds)) return '0:00'
    const mins = Math.floor(seconds / 60)
    const secs = Math.floor(seconds % 60)
    return `${mins}:${secs.toString().padStart(2, '0')}`
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
    const values = e.detail.value || []
    const province = values[0] || ''
    const city = values[1] || ''
    const district = values[2] || ''
    const region = [province, city, district].filter(Boolean).join(' ')
    this.setData({
      'form.province': province,
      'form.city': city,
      'form.district': district,
      'form.region': region
    })
  },

  chooseCollectLocation() {
    this.ensureLocationPermission()
      .then(() => this.openLocationPicker())
      .catch(() => {})
  },

  ensureLocationPermission() {
    return new Promise((resolve, reject) => {
      wx.getSetting({
        success: (res) => {
          const setting = res.authSetting || {}
          if (setting['scope.userLocation'] === true) {
            resolve()
            return
          }
          if (setting['scope.userLocation'] === false) {
            wx.showModal({
              title: '需要位置权限',
              content: '请在设置中允许位置权限后再选择采集地址',
              success: (modalRes) => {
                if (!modalRes.confirm) {
                  reject(new Error('permission denied'))
                  return
                }
                wx.openSetting({
                  success: (openRes) => {
                    if (openRes.authSetting && openRes.authSetting['scope.userLocation']) {
                      resolve()
                    } else {
                      showError('未开启位置权限')
                      reject(new Error('permission denied'))
                    }
                  },
                  fail: () => {
                    showError('打开设置失败，请手动授权位置权限')
                    reject(new Error('open setting failed'))
                  }
                })
              }
            })
            return
          }
          wx.authorize({
            scope: 'scope.userLocation',
            success: resolve,
            fail: () => {
              showError('需要位置权限才能使用地图选点')
              reject(new Error('authorize failed'))
            }
          })
        },
        fail: () => {
          showError('读取系统权限失败')
          reject(new Error('get setting failed'))
        }
      })
    })
  },

  openLocationPicker() {
    wx.chooseLocation({
      success: (res) => {
        const parts = this.extractRegionFromAddress(res.address || '')
        const region = [parts.province, parts.city, parts.district].filter(Boolean).join(' ')
        this.setData({
          'form.collect_address': [res.address, res.name].filter(Boolean).join(' ').trim(),
          'form.collect_latitude': Number(res.latitude) || 0,
          'form.collect_longitude': Number(res.longitude) || 0,
          'form.province': parts.province || this.data.form.province,
          'form.city': parts.city || this.data.form.city,
          'form.district': parts.district || this.data.form.district,
          'form.region': region || this.data.form.region
        })
      },
      fail: (err) => {
        if (err && err.errMsg && err.errMsg.includes('cancel')) return
        const msg = (err && err.errMsg) || ''
        if (msg.includes('auth deny') || msg.includes('permission')) {
          showError('位置权限不足，请到设置开启')
          return
        }
        wx.showModal({
          title: '地图选点失败',
          content: '当前环境无法打开系统选点，将切换到内置地图选点模式',
          showCancel: false,
          success: () => {
            this.openMapPickerFallback()
          }
        })
      }
    })
  },

  openMapPickerFallback() {
    const form = this.data.form || {}
    const url = `/pages/common/location-picker/index?lat=${form.collect_latitude || ''}&lng=${form.collect_longitude || ''}`
    wx.navigateTo({
      url,
      success: (res) => {
        const channel = res.eventChannel
        channel.on('locationPicked', (payload) => {
          if (!payload) return
          const latitude = Number(payload.latitude) || 0
          const longitude = Number(payload.longitude) || 0
          if (!latitude || !longitude) return
          const address = payload.address || `地图选点(${latitude.toFixed(6)}, ${longitude.toFixed(6)})`
          this.setData({
            'form.collect_address': address,
            'form.collect_latitude': latitude,
            'form.collect_longitude': longitude
          })
        })
      }
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
    
    if (!this.data.form.province || !this.data.form.city) {
      showError('请选择采集地区')
      return false
    }

    if (!this.data.form.collect_address) {
      showError('请使用地图选择采集地址')
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
        throw new Error('录音文件上传失败，请重试')
      }
      
      // 2. 创建方言记录
      const regionText = [this.data.form.province, this.data.form.city, this.data.form.district].filter(Boolean).join(' ')
      const dialectType = this.getDialectType(this.data.form.province || this.data.form.city)
      
      const dialectData = {
        title: this.data.form.title.trim(),
        content: this.data.form.description.trim(),
        description: this.data.form.description.trim(),
        audio_url: audioUrl,
        duration: this.data.recordDuration,
        region: regionText,
        province: this.data.form.province,
        city: this.data.form.city,
        dialect_type: dialectType,
        tags: this.data.form.tags.join(','),  // 将数组转换为逗号分隔的字符串
        file_size: 0,     // 可选字段
        format: 'mp3',    // 可选字段
        collect_address: this.data.form.collect_address,
        collect_latitude: this.data.form.collect_latitude,
        collect_longitude: this.data.form.collect_longitude
      }
      if (this.data.form.missing_person_id) {
        dialectData.missing_person_id = this.data.form.missing_person_id
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

  formatTime(seconds) {
    return this._formatTime(seconds)
  },

  // 格式化录音时间（带倒计时）
  formatRecordTime() {
    const current = this.data.recordTime
    const remaining = MAX_DURATION - current
    return this.formatTime(remaining)
  },

  extractRegionFromAddress(address) {
    const text = (address || '').trim()
    const result = { province: '', city: '', district: '' }
    if (!text) return result

    const provinceMatch = text.match(/^(.*?(省|自治区|行政区|特别行政区|市))/)
    if (provinceMatch) {
      result.province = provinceMatch[1]
      const rest = text.slice(provinceMatch[1].length)
      const cityMatch = rest.match(/^(.*?(市|自治州|地区|盟))/)
      if (cityMatch) {
        result.city = cityMatch[1]
        const districtMatch = rest.slice(cityMatch[1].length).match(/^(.*?(区|县|旗))/)
        if (districtMatch) {
          result.district = districtMatch[1]
        }
      }
    }
    return result
  },

  // 根据地区获取方言类型
  getDialectType(region) {
    const dialectMap = {
      '北京市': 'daily',
      '上海市': 'daily',
      '天津市': 'daily',
      '重庆市': 'daily',
      '广东省': 'daily',
      '江苏省': 'daily',
      '浙江省': 'daily',
      '山东省': 'daily',
      '河南省': 'daily',
      '四川省': 'daily',
      '湖北省': 'daily',
      '湖南省': 'daily',
      '河北省': 'daily',
      '福建省': 'daily',
      '安徽省': 'daily',
      '辽宁省': 'daily',
      '江西省': 'daily',
      '陕西省': 'daily',
      '黑龙江省': 'daily',
      '山西省': 'daily',
      '广西壮族自治区': 'daily',
      '吉林省': 'daily',
      '贵州省': 'daily',
      '云南省': 'daily',
      '甘肃省': 'daily',
      '海南省': 'daily',
      '内蒙古自治区': 'daily',
      '新疆维吾尔自治区': 'daily',
      '西藏自治区': 'daily',
      '青海省': 'daily',
      '宁夏回族自治区': 'daily'
    }
    return dialectMap[region] || 'other'
  }
})
