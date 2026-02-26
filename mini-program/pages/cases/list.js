const { get } = require('../../utils/request')
const { formatDate, showLoading, hideLoading } = require('../../utils/util')

Page({
  data: {
    cases: [],
    keyword: '',
    currentStatus: '',
    page: 1,
    pageSize: 20,
    loading: false,
    noMore: false,
    statusMap: {
      missing: '失踪中',
      searching: '寻找中',
      found: '已找到',
      reunited: '已团圆',
      closed: '已结案'
    },
    caseTypeMap: {
      elderly: '老人',
      child: '儿童',
      adult: '成人',
      disability: '残障',
      other: '其他'
    }
  },

  onLoad() {
    this.loadCases()
  },

  onPullDownRefresh() {
    this.setData({ page: 1, noMore: false })
    this.loadCases().finally(() => {
      wx.stopPullDownRefresh()
    })
  },

  onReachBottom() {
    if (!this.data.noMore && !this.data.loading) {
      this.loadMore()
    }
  },

  async loadCases() {
    this.setData({ loading: true })
    showLoading()

    try {
      const params = {
        page: this.data.page,
        page_size: this.data.pageSize,
        status: this.data.currentStatus,
        keyword: this.data.keyword
      }

      const result = await get('/missing-persons', params)
      const cases = result.list.map(item => ({
        ...item,
        missing_time: formatDate(item.missing_time),
        photoUrl: (item.photos && item.photos[0] && item.photos[0].url) ? item.photos[0].url : '/assets/default-avatar.png'
      }))

      this.setData({
        cases: this.data.page === 1 ? cases : [...this.data.cases, ...cases],
        noMore: cases.length < this.data.pageSize
      })
    } catch (error) {
      console.error('加载案件失败:', error)
    } finally {
      hideLoading()
      this.setData({ loading: false })
    }
  },

  loadMore() {
    this.setData({ page: this.data.page + 1 })
    this.loadCases()
  },

  onSearchInput(e) {
    this.setData({ keyword: e.detail.value })
  },

  onSearch() {
    this.setData({ page: 1, cases: [] })
    this.loadCases()
  },

  switchStatus(e) {
    const status = e.currentTarget.dataset.status
    this.setData({ currentStatus: status, page: 1, cases: [] })
    this.loadCases()
  },

  goToDetail(e) {
    const id = e.currentTarget.dataset.id
    wx.navigateTo({ url: `/pages/cases/detail?id=${id}` })
  },

  goToCreate() {
    wx.navigateTo({ url: '/pages/cases/create' })
  }
})
