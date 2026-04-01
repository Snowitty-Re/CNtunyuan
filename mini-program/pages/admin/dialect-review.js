const dialectService = require('../../services/dialect')
const { showToast, showLoading, hideLoading, formatTimeAgo } = require('../../utils/util')
const app = getApp()

Page({
  data: {
    tabs: [
      { key: 'pending', label: '待审批' },
      { key: 'inactive', label: '已驳回' },
      { key: 'all', label: '全部' }
    ],
    activeTab: 'pending',
    list: [],
    page: 1,
    pageSize: 20,
    loading: false,
    noMore: false
  },

  onLoad() {
    if (!app.ensureAuth || !app.ensureAuth()) return
    if (!app.isManager()) {
      showToast('无权限访问')
      wx.navigateBack()
      return
    }
    this.loadList(true)
  },

  onPullDownRefresh() {
    this.loadList(true).finally(() => wx.stopPullDownRefresh())
  },

  onReachBottom() {
    if (!this.data.loading && !this.data.noMore) {
      this.loadList(false)
    }
  },

  async loadList(reset) {
    if (this.data.loading) return
    const page = reset ? 1 : this.data.page + 1
    const params = { page, page_size: this.data.pageSize }
    if (this.data.activeTab !== 'all') {
      params.status = this.data.activeTab
    }

    this.setData({ loading: true })
    try {
      const res = await dialectService.getList(params)
      const rows = (res.list || []).map(item => ({
        ...item,
        uploaderName: (item.uploader && item.uploader.nickname) || '未知',
        createdAtText: formatTimeAgo(item.created_at)
      }))
      this.setData({
        page,
        list: reset ? rows : this.data.list.concat(rows),
        noMore: rows.length < this.data.pageSize
      })
    } catch (e) {
      showToast('加载审批列表失败')
    } finally {
      this.setData({ loading: false })
    }
  },

  onTabTap(e) {
    const key = e.currentTarget.dataset.key
    if (!key || key === this.data.activeTab) return
    this.setData({ activeTab: key, list: [], page: 1, noMore: false })
    this.loadList(true)
  },

  goDetail(e) {
    const id = e.currentTarget.dataset.id
    if (!id) return
    wx.navigateTo({ url: `/pages/dialect/detail?id=${id}` })
  },

  async updateStatus(e) {
    const { id, status } = e.currentTarget.dataset
    if (!id || !status) return
    const text = status === 'active' ? '通过' : '驳回'
    showLoading(`提交${text}中...`)
    try {
      await dialectService.updateStatus(id, status)
      hideLoading()
      showToast(`已${text}`)
      this.loadList(true)
    } catch (err) {
      hideLoading()
      showToast(`${text}失败`)
    }
  }
})
