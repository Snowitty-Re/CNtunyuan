const { get } = require('../../utils/request')
const { formatDate } = require('../../utils/util')

Page({
  data: {
    stats: {
      totalCases: 0,
      resolvedCases: 0,
      volunteers: 0,
      dialects: 0
    },
    recentCases: [],
    recentDialects: [],
    statusText: {
      missing: '失踪中',
      searching: '寻找中',
      found: '已找到',
      reunited: '已团圆',
      closed: '已结案'
    }
  },

  onLoad() {
    this.loadData()
  },

  onShow() {
    this.loadData()
  },

  onPullDownRefresh() {
    this.loadData().finally(() => {
      wx.stopPullDownRefresh()
    })
  },

  async loadData() {
    try {
      // 加载统计数据
      this.loadStats()
      // 加载最新案件
      this.loadRecentCases()
      // 加载最新方言
      this.loadRecentDialects()
    } catch (error) {
      console.error('加载数据失败:', error)
    }
  },

  async loadStats() {
    try {
      const [caseStats, userStats, dialectStats] = await Promise.all([
        get('/missing-persons/statistics'),
        get('/users/statistics'),
        get('/dialects/statistics')
      ])
      
      this.setData({
        stats: {
          totalCases: caseStats.total || 0,
          resolvedCases: caseStats.resolved || 0,
          volunteers: userStats.total || 0,
          dialects: dialectStats.total || 0
        }
      })
    } catch (error) {
      console.error('加载统计失败:', error)
    }
  },

  async loadRecentCases() {
    try {
      const result = await get('/missing-persons', { page: 1, page_size: 5 })
      const cases = result.list.map(item => ({
        ...item,
        missing_time: formatDate(item.missing_time)
      }))
      this.setData({ recentCases: cases })
    } catch (error) {
      console.error('加载案件失败:', error)
    }
  },

  async loadRecentDialects() {
    try {
      const result = await get('/dialects', { page: 1, page_size: 5 })
      this.setData({ recentDialects: result.list })
    } catch (error) {
      console.error('加载方言失败:', error)
    }
  },

  onSearchInput(e) {
    // 搜索逻辑
    console.log('搜索:', e.detail.value)
  },

  goToCreateCase() {
    wx.navigateTo({ url: '/pages/cases/create' })
  },

  goToDialect() {
    wx.navigateTo({ url: '/pages/dialect/create' })
  },

  goToMap() {
    wx.navigateTo({ url: '/pages/map/index' })
  },

  goToTasks() {
    wx.switchTab({ url: '/pages/volunteer/workbench' })
  },

  goToCases() {
    wx.switchTab({ url: '/pages/cases/list' })
  },

  goToCaseDetail(e) {
    const id = e.currentTarget.dataset.id
    wx.navigateTo({ url: `/pages/cases/detail?id=${id}` })
  },

  goToDialectList() {
    wx.navigateTo({ url: '/pages/dialect/list' })
  },

  playDialect(e) {
    const id = e.currentTarget.dataset.id
    wx.navigateTo({ url: `/pages/dialect/detail?id=${id}` })
  }
})
