const dialectService = require('../../services/dialect')
const { showLoading, hideLoading, formatTimeAgo, showToast } = require('../../utils/util')

// 创建 InnerAudioContext 实例
let innerAudioContext = null

Page({
  data: {
    // 方言列表数据
    dialects: [],
    
    // 分页参数
    page: 1,
    pageSize: 10,
    loading: false,
    noMore: false,
    
    // 地区筛选
    regionOptions: [
      { id: '', name: '全部地区' },
      { id: '北京市', name: '北京市' },
      { id: '上海市', name: '上海市' },
      { id: '广东省', name: '广东省' },
      { id: '四川省', name: '四川省' },
      { id: '河南省', name: '河南省' },
      { id: '山东省', name: '山东省' },
      { id: '浙江省', name: '浙江省' },
      { id: '江苏省', name: '江苏省' },
      { id: '湖南省', name: '湖南省' },
      { id: '湖北省', name: '湖北省' },
      { id: '福建省', name: '福建省' },
      { id: '陕西省', name: '陕西省' },
      { id: '辽宁省', name: '辽宁省' }
    ],
    selectedRegionIndex: 0,
    selectedRegion: '',
    
    // 播放状态
    playingId: null,
    
    // 创建按钮动画
    animateCreateBtn: false
  },

  onLoad() {
    this.loadDialects()
    this.initAudioContext()
  },

  onUnload() {
    // 页面卸载时销毁音频上下文
    if (innerAudioContext) {
      innerAudioContext.stop()
      innerAudioContext.destroy()
      innerAudioContext = null
    }
  },

  onPullDownRefresh() {
    this.setData({ 
      page: 1, 
      noMore: false,
      dialects: []
    })
    this.loadDialects().finally(() => {
      wx.stopPullDownRefresh()
    })
  },

  onReachBottom() {
    if (!this.data.noMore && !this.data.loading) {
      this.loadMore()
    }
  },

  // 初始化音频上下文
  initAudioContext() {
    if (!innerAudioContext) {
      innerAudioContext = wx.createInnerAudioContext()
      
      innerAudioContext.onEnded(() => {
        this.setData({ playingId: null })
      })

      innerAudioContext.onError(() => {
        this.setData({ playingId: null })
        showToast('播放失败', 'none')
      })
    }
  },

  // 加载方言列表
  async loadDialects() {
    if (this.data.loading) return
    
    this.setData({ loading: true })
    
    try {
      const params = {
        page: this.data.page,
        page_size: this.data.pageSize
      }

      // 添加地区筛选
      if (this.data.selectedRegion) {
        params.region = this.data.selectedRegion
      }

      const result = await dialectService.getList(params)
      
      const newDialects = result.list || result.data || []
      
      this.setData({
        dialects: this.data.page === 1 ? newDialects : [...this.data.dialects, ...newDialects],
        noMore: newDialects.length < this.data.pageSize
      })
    } catch (error) {
      console.error('加载方言列表失败:', error)
      showToast('加载失败', 'none')
    } finally {
      this.setData({ loading: false })
    }
  },

  // 加载更多
  loadMore() {
    this.setData({ page: this.data.page + 1 })
    this.loadDialects()
  },

  // 地区筛选变化
  onRegionChange(e) {
    const index = parseInt(e.detail.value)
    const region = this.data.regionOptions[index].id
    
    this.setData({
      selectedRegionIndex: index,
      selectedRegion: region,
      page: 1,
      dialects: [],
      noMore: false
    })
    
    this.loadDialects()
  },

  // 播放/暂停录音
  togglePlay(e) {
    const { id, url } = e.currentTarget.dataset
    
    if (!url) {
      showToast('音频文件不存在', 'none')
      return
    }

    // 如果点击的是当前正在播放的，则停止
    if (this.data.playingId === id) {
      innerAudioContext.stop()
      this.setData({ playingId: null })
      return
    }

    // 停止之前的播放
    innerAudioContext.stop()
    
    // 设置新的音频源并播放
    innerAudioContext.src = url
    innerAudioContext.play()
    
    this.setData({ playingId: id })
    
    // 记录播放次数
    dialectService.recordPlay(id).catch(() => {})
  },

  // 跳转到详情页
  goToDetail(e) {
    const { id } = e.currentTarget.dataset
    wx.navigateTo({
      url: `/pages/dialect/detail?id=${id}`
    })
  },

  // 跳转到创建页
  goToCreate() {
    // 添加按钮动画效果
    this.setData({ animateCreateBtn: true })
    setTimeout(() => {
      this.setData({ animateCreateBtn: false })
      wx.navigateTo({
        url: '/pages/dialect/create'
      })
    }, 200)
  },

  // 格式化播放次数
  formatPlayCount(count) {
    if (!count) return '0'
    if (count < 1000) return count.toString()
    if (count < 10000) return (count / 1000).toFixed(1) + 'k'
    return (count / 10000).toFixed(1) + 'w'
  },

  // 格式化时间
  formatTime(seconds) {
    if (!seconds) return '0:00'
    const mins = Math.floor(seconds / 60)
    const secs = seconds % 60
    return `${mins}:${secs.toString().padStart(2, '0')}`
  },

  // 格式化相对时间
  formatTimeAgo(date) {
    return formatTimeAgo(date)
  }
})
