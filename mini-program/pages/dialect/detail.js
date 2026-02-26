const { get, post } = require('../../utils/request')
const { formatDate, showSuccess } = require('../../utils/util')

const innerAudioContext = wx.createInnerAudioContext()

Page({
  data: {
    dialect: {},
    isPlaying: false,
    currentTime: '0:00',
    currentSeconds: 0
  },

  onLoad(options) {
    const { id } = options
    if (id) {
      this.loadDialectDetail(id)
      this.recordPlay(id)
    }

    innerAudioContext.onTimeUpdate(() => {
      const current = Math.floor(innerAudioContext.currentTime)
      this.setData({
        currentTime: this.formatTime(current),
        currentSeconds: current
      })
    })

    innerAudioContext.onEnded(() => {
      this.setData({ isPlaying: false })
    })

    innerAudioContext.onError(() => {
      this.setData({ isPlaying: false })
      wx.showToast({ title: '播放失败', icon: 'none' })
    })
  },

  onUnload() {
    innerAudioContext.stop()
    innerAudioContext.destroy()
  },

  async loadDialectDetail(id) {
    try {
      const data = await get(`/dialects/${id}`)
      data.created_at = formatDate(data.created_at)
      data.collector = data.collector || { nickname: '未知', avatar: '/assets/default-avatar.png' }
      data.address = data.address || '暂无'
      data.description = data.description || '暂无描述'
      this.setData({ dialect: data })
      innerAudioContext.src = data.audio_url
    } catch (error) {
      console.error('加载方言详情失败:', error)
    }
  },

  recordPlay(id) {
    post(`/dialects/${id}/play`).catch(() => {})
  },

  togglePlay() {
    if (this.data.isPlaying) {
      innerAudioContext.pause()
      this.setData({ isPlaying: false })
    } else {
      innerAudioContext.play()
      this.setData({ isPlaying: true })
    }
  },

  seekAudio(e) {
    const position = e.detail.value
    innerAudioContext.seek(position)
    this.setData({
      currentTime: this.formatTime(position),
      currentSeconds: position
    })
  },

  formatTime(seconds) {
    const mins = Math.floor(seconds / 60)
    const secs = seconds % 60
    return `${mins}:${secs.toString().padStart(2, '0')}`
  },

  formatDuration(seconds) {
    return this.formatTime(seconds)
  },

  async like() {
    try {
      await post(`/dialects/${this.data.dialect.id}/like`)
      showSuccess('点赞成功')
      this.setData({
        'dialect.like_count': this.data.dialect.like_count + 1
      })
    } catch (error) {
      console.error('点赞失败:', error)
    }
  },

  share() {
    // 分享功能
  },

  goToCollector() {
    const collectorId = this.data.dialect.collector?.id
    if (collectorId) {
      wx.navigateTo({ url: `/pages/volunteer/profile?id=${collectorId}` })
    }
  }
})
