const dialectService = require('../../services/dialect')
const uploadService = require('../../services/upload')
const missingPersonService = require('../../services/missingPerson')
const { showLoading, hideLoading, showSuccess, showError, showToast } = require('../../utils/util')
const app = getApp()

const MIN_DURATION = 2
const MAX_DURATION = 8

let recorderManager = null
let innerAudioContext = null

Page({
  data: {
    form: {
      description: '',
      region: '',
      province: '',
      city: '',
      district: '',
      collect_address: '',
      collect_latitude: 0,
      collect_longitude: 0,
      missing_person_id: ''
    },

    minDuration: MIN_DURATION,
    maxDuration: MAX_DURATION,

    cardGroups: [],
    cards: [],
    currentCardIndex: 0,
    currentCard: null,

    recordings: {},
    completedCount: 0,
    allCardsRecorded: false,

    isRecording: false,
    hasRecorded: false,
    recordTime: 0,
    recordTimeText: '0:00',
    recordDuration: 0,
    recordDurationText: '0:00',
    tempFilePath: '',

    isPlaying: false,
    playProgress: 0,
    playCurrentTime: 0,
    playCurrentTimeText: '0:00',

    missingPersonList: [],
    showMissingPersonPicker: false,
    selectedMissingPerson: null,

    isUploading: false,
    uploadProgress: 0
  },

  async onLoad() {
    if (!app.ensureAuth || !app.ensureAuth()) return
    this.initRecorder()
    await Promise.all([this.loadCardTemplate(), this.loadMissingPersons()])
  },

  onUnload() {
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

  async loadCardTemplate() {
    try {
      const result = await dialectService.getCardTemplate()
      const groups = (result.groups || result.list || []).filter(Boolean)
      const cards = []
      groups.forEach((group) => {
        const groupCards = (group.cards || []).filter(item => item && item.id)
        groupCards.forEach((card) => {
          cards.push({
            ...card,
            group_name: group.name || ''
          })
        })
      })

      if (!cards.length) {
        showError('暂无可录入方言卡片，请联系管理员配置')
      }

      this.setData({
        cardGroups: groups,
        cards,
        currentCardIndex: 0,
        currentCard: cards[0] || null
      })
      this.refreshCompletedCount(this.data.recordings || {})
      this.syncCurrentRecordingState()
    } catch (error) {
      console.error('加载方言卡片模板失败:', error)
      showError('加载录入卡片失败')
    }
  },

  syncCurrentRecordingState() {
    const card = this.data.currentCard
    if (!card || !card.id) {
      this.setData({
        hasRecorded: false,
        tempFilePath: '',
        recordDuration: 0,
        recordDurationText: '0:00',
        playProgress: 0,
        playCurrentTime: 0,
        playCurrentTimeText: '0:00',
        isPlaying: false
      })
      return
    }

    const rec = this.data.recordings[card.id]
    if (!rec) {
      this.setData({
        hasRecorded: false,
        tempFilePath: '',
        recordDuration: 0,
        recordDurationText: '0:00',
        playProgress: 0,
        playCurrentTime: 0,
        playCurrentTimeText: '0:00',
        isPlaying: false
      })
      return
    }

    this.setData({
      hasRecorded: true,
      tempFilePath: rec.tempFilePath,
      recordDuration: rec.duration,
      recordDurationText: this._formatTime(rec.duration),
      playProgress: 0,
      playCurrentTime: 0,
      playCurrentTimeText: '0:00',
      isPlaying: false
    })
  },

  refreshCompletedCount(nextRecordings) {
    const cards = this.data.cards || []
    let completed = 0
    cards.forEach((card) => {
      if (nextRecordings[card.id]) completed += 1
    })
    this.setData({
      completedCount: completed,
      allCardsRecorded: cards.length > 0 && completed === cards.length
    })
  },

  initRecorder() {
    recorderManager = wx.getRecorderManager()

    recorderManager.onStart(() => {
      this.setData({ isRecording: true })
      this.startRecordTimer()
    })

    recorderManager.onStop((res) => {
      this.stopRecordTimer()
      const duration = Math.floor((res.duration || 0) / 1000)

      if (duration < MIN_DURATION) {
        showError(`录音时长不足${MIN_DURATION}秒，请重新录制`)
        this.setData({
          isRecording: false,
          recordTime: 0,
          recordTimeText: '0:00'
        })
        return
      }

      const card = this.data.currentCard
      if (!card || !card.id) {
        this.setData({ isRecording: false })
        return
      }

      const recordings = { ...this.data.recordings }
      recordings[card.id] = {
        tempFilePath: res.tempFilePath,
        duration
      }

      this.setData({
        recordings,
        isRecording: false,
        hasRecorded: true,
        tempFilePath: res.tempFilePath,
        recordDuration: duration,
        recordDurationText: this._formatTime(duration)
      })
      this.refreshCompletedCount(recordings)
      showSuccess(`卡片录音完成 ${this._formatTime(duration)}`)
    })

    recorderManager.onError((err) => {
      console.error('录音错误:', err)
      this.stopRecordTimer()
      this.setData({ isRecording: false })
      showError('录音失败，请重试')
    })
  },

  startRecordTimer() {
    this.setData({ recordTime: 0, recordTimeText: '0:00' })
    this.recordTimer = setInterval(() => {
      const recordTime = this.data.recordTime + 1
      this.setData({ recordTime, recordTimeText: this._formatTime(recordTime) })
      if (recordTime >= MAX_DURATION) {
        this.stopRecord()
        showToast(`已达到最大录音时长${MAX_DURATION}秒`, 'none')
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
    if (!this.data.currentCard || !this.data.currentCard.id) {
      showError('暂无可录制卡片')
      return
    }
    if (this.data.isRecording) return
    if (this.data.hasRecorded) {
      wx.showModal({
        title: '提示',
        content: '重新录制将覆盖当前卡片录音，是否继续？',
        success: (res) => {
          if (!res.confirm) return
          this.resetCurrentRecord()
          this.doStartRecord()
        }
      })
      return
    }
    this.doStartRecord()
  },

  doStartRecord() {
    recorderManager.start({
      duration: MAX_DURATION * 1000,
      sampleRate: 44100,
      numberOfChannels: 1,
      encodeBitRate: 192000,
      format: 'mp3'
    })
  },

  stopRecord() {
    if (!this.data.isRecording) return
    if (this.data.recordTime < MIN_DURATION) {
      showError(`录音至少需要${MIN_DURATION}秒`)
      return
    }
    recorderManager.stop()
  },

  resetCurrentRecord() {
    const card = this.data.currentCard
    if (!card || !card.id) return
    this.stopPlay()
    const recordings = { ...this.data.recordings }
    delete recordings[card.id]
    this.setData({
      recordings,
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
    this.refreshCompletedCount(recordings)
  },

  initAudioPlayer() {
    if (innerAudioContext) return
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
        playCurrentTimeText: '0:00'
      })
    })
    innerAudioContext.onError(() => {
      this.setData({ isPlaying: false })
      showToast('播放失败', 'none')
    })
  },

  togglePlay() {
    if (!this.data.hasRecorded || !this.data.tempFilePath) return
    this.initAudioPlayer()
    if (this.data.isPlaying) {
      innerAudioContext.pause()
      this.setData({ isPlaying: false })
      return
    }
    innerAudioContext.src = this.data.tempFilePath
    innerAudioContext.play()
    this.setData({ isPlaying: true })
  },

  stopPlay() {
    if (innerAudioContext) innerAudioContext.stop()
    this.setData({
      isPlaying: false,
      playProgress: 0,
      playCurrentTime: 0,
      playCurrentTimeText: '0:00'
    })
  },

  onPlayProgressChange(e) {
    if (!innerAudioContext) return
    const value = e.detail.value
    const seekTime = Math.floor((value / 100) * this.data.recordDuration)
    innerAudioContext.seek(seekTime)
    this.setData({
      playProgress: value,
      playCurrentTime: seekTime,
      playCurrentTimeText: this._formatTime(seekTime)
    })
  },

  goPrevCard() {
    if (this.data.currentCardIndex <= 0) return
    this.stopPlay()
    const nextIndex = this.data.currentCardIndex - 1
    this.setData({
      currentCardIndex: nextIndex,
      currentCard: this.data.cards[nextIndex] || null
    })
    this.syncCurrentRecordingState()
  },

  goNextCard() {
    if (!this.data.hasRecorded) {
      showToast('请先完成当前卡片录音', 'none')
      return
    }
    if (this.data.currentCardIndex >= this.data.cards.length - 1) return
    this.stopPlay()
    const nextIndex = this.data.currentCardIndex + 1
    this.setData({
      currentCardIndex: nextIndex,
      currentCard: this.data.cards[nextIndex] || null
    })
    this.syncCurrentRecordingState()
  },

  switchCard(e) {
    const index = Number(e.currentTarget.dataset.index)
    if (Number.isNaN(index)) return
    if (index < 0 || index >= this.data.cards.length) return
    this.stopPlay()
    this.setData({
      currentCardIndex: index,
      currentCard: this.data.cards[index] || null
    })
    this.syncCurrentRecordingState()
  },

  onDescInput(e) {
    this.setData({ 'form.description': e.detail.value || '' })
  },

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
    this.ensureLocationPermission().then(() => this.openLocationPicker()).catch(() => {})
  },

  ensureLocationPermission() {
    return new Promise((resolve, reject) => {
      wx.getSetting({
        success: (res) => {
          const setting = res.authSetting || {}
          if (setting['scope.userLocation'] === true) return resolve()
          if (setting['scope.userLocation'] === false) {
            wx.showModal({
              title: '需要位置权限',
              content: '请在设置中允许位置权限后再选择采集地址',
              success: (modalRes) => {
                if (!modalRes.confirm) return reject(new Error('permission denied'))
                wx.openSetting({
                  success: (openRes) => {
                    if (openRes.authSetting && openRes.authSetting['scope.userLocation']) resolve()
                    else {
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
          success: () => this.openMapPickerFallback()
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

  showMissingPersonSelector() {
    this.setData({ showMissingPersonPicker: true })
  },

  hideMissingPersonSelector() {
    this.setData({ showMissingPersonPicker: false })
  },

  selectMissingPerson(e) {
    const index = Number(e.currentTarget.dataset.index)
    const person = this.data.missingPersonList[index]
    if (!person) return
    this.setData({
      selectedMissingPerson: person,
      'form.missing_person_id': person.id,
      showMissingPersonPicker: false
    })
  },

  clearMissingPerson() {
    this.setData({
      selectedMissingPerson: null,
      'form.missing_person_id': ''
    })
  },

  validateForm() {
    if (!this.data.cards.length) {
      showError('暂无录入卡片模板')
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
    if (this.data.completedCount !== this.data.cards.length) {
      showError('请完成全部卡片录音后再提交')
      return false
    }
    return true
  },

  async submitForm() {
    if (!this.validateForm()) return
    if (this.data.isUploading) return

    this.setData({ isUploading: true, uploadProgress: 0 })
    showLoading('上传中...')

    try {
      const payloadItems = []
      const cards = this.data.cards || []
      for (let i = 0; i < cards.length; i++) {
        const card = cards[i]
        const record = this.data.recordings[card.id]
        if (!record || !record.tempFilePath) {
          throw new Error('存在未完成卡片录音')
        }
        const uploadRes = await uploadService.upload(record.tempFilePath, {
          type: 'audio',
          entity_type: 'dialect'
        })
        const audioURL = uploadRes.url || (uploadRes.data && uploadRes.data.url) || ''
        if (!audioURL) {
          throw new Error('录音上传失败')
        }
        payloadItems.push({
          card_id: card.id,
          audio_url: audioURL,
          duration: record.duration,
          file_size: 0,
          format: 'mp3'
        })
        this.setData({ uploadProgress: Math.floor(((i + 1) / cards.length) * 100) })
      }

      const regionText = [this.data.form.province, this.data.form.city, this.data.form.district].filter(Boolean).join(' ')
      const submitData = {
        region: regionText,
        province: this.data.form.province,
        city: this.data.form.city,
        district: this.data.form.district,
        description: (this.data.form.description || '').trim(),
        collect_address: this.data.form.collect_address,
        collect_latitude: this.data.form.collect_latitude,
        collect_longitude: this.data.form.collect_longitude,
        recordings: payloadItems
      }
      if (this.data.form.missing_person_id) {
        submitData.missing_person_id = this.data.form.missing_person_id
      }

      await dialectService.createBatch(submitData)
      hideLoading()
      showSuccess('卡片方言录入提交成功')
      setTimeout(() => wx.navigateBack(), 1200)
    } catch (error) {
      hideLoading()
      console.error('提交失败:', error)
      showError(error.message || '提交失败，请重试')
      this.setData({ isUploading: false })
      return
    }

    this.setData({ isUploading: false })
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
        if (districtMatch) result.district = districtMatch[1]
      }
    }
    return result
  },

  _formatTime(seconds) {
    if (!seconds || Number.isNaN(seconds)) return '0:00'
    const mins = Math.floor(seconds / 60)
    const secs = Math.floor(seconds % 60)
    return `${mins}:${secs.toString().padStart(2, '0')}`
  }
})
