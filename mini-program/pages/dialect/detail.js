const dialectService = require('../../services/dialect')
const { showLoading, hideLoading, showSuccess, showError, showToast, formatTimeAgo, formatDate } = require('../../utils/util')

// 音频上下文
let innerAudioContext = null

Page({
  data: {
    // 方言ID
    id: '',
    
    // 方言详情
    dialect: null,
    
    // 音频播放
    isPlaying: false,
    currentTime: 0,
    duration: 0,
    progress: 0,
    
    // 点赞状态
    isLiked: false,
    likeLoading: false,
    
    // 评论
    comments: [],
    commentPage: 1,
    commentPageSize: 10,
    commentLoading: false,
    commentNoMore: false,
    commentContent: '',
    
    // 关联走失人员
    missingPerson: null,
    
    // 页面状态
    loading: true
  },

  onLoad(options) {
    const { id } = options
    if (!id) {
      showError('参数错误')
      wx.navigateBack()
      return
    }
    
    this.setData({ id })
    this.initAudioContext()
    this.loadDialectDetail()
    this.loadComments()
  },

  onUnload() {
    if (innerAudioContext) {
      innerAudioContext.stop()
      innerAudioContext.destroy()
      innerAudioContext = null
    }
  },

  onPullDownRefresh() {
    this.setData({ 
      commentPage: 1, 
      commentNoMore: false,
      comments: []
    })
    Promise.all([
      this.loadDialectDetail(),
      this.loadComments()
    ]).finally(() => {
      wx.stopPullDownRefresh()
    })
  },

  onReachBottom() {
    if (!this.data.commentNoMore && !this.data.commentLoading) {
      this.loadMoreComments()
    }
  },

  // 初始化音频上下文
  initAudioContext() {
    innerAudioContext = wx.createInnerAudioContext()
    
    innerAudioContext.onCanplay(() => {
      const duration = innerAudioContext.duration || 0
      this.setData({ duration: Math.floor(duration) })
    })

    innerAudioContext.onTimeUpdate(() => {
      const currentTime = innerAudioContext.currentTime || 0
      const duration = innerAudioContext.duration || 1
      this.setData({
        currentTime: Math.floor(currentTime),
        progress: (currentTime / duration) * 100
      })
    })

    innerAudioContext.onEnded(() => {
      this.setData({ 
        isPlaying: false,
        currentTime: 0,
        progress: 0
      })
    })

    innerAudioContext.onError((err) => {
      console.error('播放错误:', err)
      this.setData({ isPlaying: false })
      showToast('播放失败', 'none')
    })
  },

  // 加载方言详情
  async loadDialectDetail() {
    this.setData({ loading: true })
    
    try {
      const dialect = await dialectService.getById(this.data.id)
      
      // 处理数据
      dialect.created_at_text = formatTimeAgo(dialect.created_at)
      dialect.formatted_date = formatDate(dialect.created_at, 'YYYY-MM-DD HH:mm')
      
      // 设置音频源
      if (dialect.audio_url) {
        innerAudioContext.src = dialect.audio_url
        this.setData({ duration: dialect.duration || 0 })
      }
      
      // 处理关联走失人员
      if (dialect.missing_person) {
        this.setData({ missingPerson: dialect.missing_person })
      }
      
      this.setData({ 
        dialect,
        isLiked: dialect.is_liked || false
      })
      
      // 记录播放
      dialectService.recordPlay(this.data.id).catch(() => {})
    } catch (error) {
      console.error('加载方言详情失败:', error)
      showToast('加载失败', 'none')
    } finally {
      this.setData({ loading: false })
    }
  },

  // 播放/暂停
  togglePlay() {
    if (!this.data.dialect?.audio_url) {
      showToast('音频文件不存在', 'none')
      return
    }

    if (this.data.isPlaying) {
      innerAudioContext.pause()
      this.setData({ isPlaying: false })
    } else {
      innerAudioContext.play()
      this.setData({ isPlaying: true })
    }
  },

  // 拖动进度条
  onSliderChange(e) {
    const value = e.detail.value
    const duration = this.data.duration || 1
    const seekTime = (value / 100) * duration
    innerAudioContext.seek(seekTime)
    this.setData({ 
      progress: value,
      currentTime: Math.floor(seekTime)
    })
  },

  // 拖动中
  onSliderChanging(e) {
    const value = e.detail.value
    const duration = this.data.duration || 1
    this.setData({
      currentTime: Math.floor((value / 100) * duration)
    })
  },

  // 格式化时间
  formatTime(seconds) {
    if (!seconds || isNaN(seconds)) return '0:00'
    const mins = Math.floor(seconds / 60)
    const secs = Math.floor(seconds % 60)
    return `${mins}:${secs.toString().padStart(2, '0')}`
  },

  // 点赞/取消点赞
  async toggleLike() {
    if (this.data.likeLoading) return
    
    this.setData({ likeLoading: true })
    
    try {
      if (this.data.isLiked) {
        await dialectService.unlike(this.data.id)
        this.setData({
          isLiked: false,
          'dialect.like_count': Math.max(0, (this.data.dialect.like_count || 0) - 1)
        })
      } else {
        await dialectService.like(this.data.id)
        this.setData({
          isLiked: true,
          'dialect.like_count': (this.data.dialect.like_count || 0) + 1
        })
        showToast('点赞成功', 'none')
      }
    } catch (error) {
      console.error('点赞操作失败:', error)
      showToast('操作失败', 'none')
    } finally {
      this.setData({ likeLoading: false })
    }
  },

  // 加载评论列表
  async loadComments() {
    if (this.data.commentLoading) return
    
    this.setData({ commentLoading: true })
    
    try {
      const result = await dialectService.getComments(this.data.id, {
        page: this.data.commentPage,
        page_size: this.data.commentPageSize
      })
      
      const newComments = result.list || result.data || []
      
      // 处理评论时间
      newComments.forEach(comment => {
        comment.time_text = formatTimeAgo(comment.created_at)
      })
      
      this.setData({
        comments: this.data.commentPage === 1 ? newComments : [...this.data.comments, ...newComments],
        commentNoMore: newComments.length < this.data.commentPageSize
      })
    } catch (error) {
      console.error('加载评论失败:', error)
    } finally {
      this.setData({ commentLoading: false })
    }
  },

  // 加载更多评论
  loadMoreComments() {
    this.setData({ commentPage: this.data.commentPage + 1 })
    this.loadComments()
  },

  // 评论输入
  onCommentInput(e) {
    this.setData({ commentContent: e.detail.value })
  },

  // 发表评论
  async submitComment() {
    const content = this.data.commentContent.trim()
    if (!content) {
      showToast('请输入评论内容', 'none')
      return
    }
    
    showLoading('发送中...')
    
    try {
      await dialectService.addComment(this.data.id, { content })
      
      this.setData({
        commentContent: '',
        commentPage: 1,
        comments: []
      })
      
      await this.loadComments()
      showSuccess('评论成功')
    } catch (error) {
      console.error('发表评论失败:', error)
      showToast('评论失败', 'none')
    } finally {
      hideLoading()
    }
  },

  // 跳转到走失人员详情
  goToMissingPerson() {
    if (this.data.missingPerson?.id) {
      wx.navigateTo({
        url: `/pages/missing/detail?id=${this.data.missingPerson.id}`
      })
    }
  },

  // 跳转到采集者主页
  goToCollectorProfile() {
    const collectorId = this.data.dialect?.collector?.id
    if (collectorId) {
      wx.navigateTo({
        url: `/pages/volunteer/profile?id=${collectorId}`
      })
    }
  },

  // 分享
  onShareAppMessage() {
    const dialect = this.data.dialect
    return {
      title: `${dialect?.title || '方言录音'} - 团圆寻亲志愿者`,
      path: `/pages/dialect/detail?id=${this.data.id}`,
      imageUrl: '/images/share-dialect.png'
    }
  },

  // 播放次数格式化
  formatPlayCount(count) {
    if (!count) return '0'
    if (count < 1000) return count.toString()
    if (count < 10000) return (count / 1000).toFixed(1) + 'k'
    return (count / 10000).toFixed(1) + 'w'
  }
})
