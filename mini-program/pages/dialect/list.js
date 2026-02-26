const { get, post } = require('../../utils/request')
const { showLoading, hideLoading } = require('../../utils/util')

const innerAudioContext = wx.createInnerAudioContext()

Page({
  data: {
    dialects: [],
    keyword: '',
    regionArray: [
      ['全部省份', '北京市', '上海市', '广东省', '四川省', '河南省'],
      ['全部城市'],
      ['全部区县']
    ],
    regionIndex: [0, 0, 0],
    selectedRegion: '',
    page: 1,
    pageSize: 20,
    loading: false,
    noMore: false,
    playingId: null
  },

  onLoad() {
    this.loadDialects()
    
    innerAudioContext.onEnded(() => {
      this.setData({ playingId: null })
    })
  },

  onUnload() {
    innerAudioContext.stop()
    innerAudioContext.destroy()
  },

  onPullDownRefresh() {
    this.setData({ page: 1, noMore: false })
    this.loadDialects().finally(() => {
      wx.stopPullDownRefresh()
    })
  },

  onReachBottom() {
    if (!this.data.noMore && !this.data.loading) {
      this.loadMore()
    }
  },

  async loadDialects() {
    this.setData({ loading: true })
    showLoading()

    try {
      const params = {
        page: this.data.page,
        page_size: this.data.pageSize,
        keyword: this.data.keyword
      }

      if (this.data.selectedRegion) {
        params.province = this.data.selectedRegion.split(' ')[0]
      }

      const result = await get('/dialects', params)
      
      this.setData({
        dialects: this.data.page === 1 ? result.list : [...this.data.dialects, ...result.list],
        noMore: result.list.length < this.data.pageSize
      })
    } catch (error) {
      console.error('加载方言失败:', error)
    } finally {
      hideLoading()
      this.setData({ loading: false })
    }
  },

  loadMore() {
    this.setData({ page: this.data.page + 1 })
    this.loadDialects()
  },

  onSearchInput(e) {
    this.setData({ keyword: e.detail.value })
    // 防抖搜索
    clearTimeout(this.searchTimer)
    this.searchTimer = setTimeout(() => {
      this.setData({ page: 1, dialects: [] })
      this.loadDialects()
    }, 500)
  },

  onRegionChange(e) {
    const { value } = e.detail
    const { regionArray } = this.data
    const region = value.map((i, idx) => regionArray[idx][i]).filter(r => r && !r.startsWith('全部')).join(' ')
    this.setData({
      selectedRegion: region,
      regionIndex: value,
      page: 1,
      dialects: []
    })
    this.loadDialects()
  },

  onRegionColumnChange(e) {
    // 处理地区级联
  },

  playAudio(e) {
    e.stopPropagation()
    const { id, url } = e.currentTarget.dataset
    
    if (this.data.playingId === id) {
      innerAudioContext.stop()
      this.setData({ playingId: null })
    } else {
      innerAudioContext.src = url
      innerAudioContext.play()
      this.setData({ playingId: id })
      
      // 记录播放
      post(`/dialects/${id}/play`).catch(() => {})
    }
  },

  formatDuration(seconds) {
    const mins = Math.floor(seconds / 60)
    const secs = seconds % 60
    return `${mins}:${secs.toString().padStart(2, '0')}`
  },

  goToDetail(e) {
    const id = e.currentTarget.dataset.id
    wx.navigateTo({ url: `/pages/dialect/detail?id=${id}` })
  },

  goToCreate() {
    wx.navigateTo({ url: '/pages/dialect/create' })
  }
})
