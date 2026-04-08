const dialectService = require('../../services/dialect')
const { showToast, showLoading, hideLoading, formatTimeAgo } = require('../../utils/util')
const { ACTIONS } = require('../../utils/permission')
const app = getApp()
const DIALECT_LIST_DIRTY_KEY = 'dialect_list_dirty'

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
    noMore: false,
    selectedIds: [],
    selectAll: false
  },

  onLoad() {
    if (!app.ensureAuth || !app.ensureAuth()) return
    if (!app.hasPermission(ACTIONS.DIALECT_MANAGE)) {
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
      this.syncSelectionState()
    } catch (e) {
      showToast('加载审批列表失败')
    } finally {
      this.setData({ loading: false })
    }
  },

  onTabTap(e) {
    const key = e.currentTarget.dataset.key
    if (!key || key === this.data.activeTab) return
    this.setData({ activeTab: key, list: [], page: 1, noMore: false, selectedIds: [], selectAll: false })
    this.loadList(true)
  },

  goDetail(e) {
    const id = e.currentTarget.dataset.id
    if (!id) return
    wx.navigateTo({ url: `/pages/dialect/detail?id=${id}` })
  },

  goCardManage() {
    wx.navigateTo({ url: '/pages/admin/dialect-card-manage' })
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
  },

  async deleteDialect(e) {
    const { id, title } = e.currentTarget.dataset
    if (!id) return

    wx.showModal({
      title: '确认删除',
      content: `确定删除方言《${title || '未命名方言'}》吗？删除后不可恢复。`,
      confirmColor: '#FF4D4F',
      success: async (res) => {
        if (!res.confirm) return
        showLoading('删除中...')
        try {
          await dialectService.delete(id)
          hideLoading()
          showToast('删除成功')
          this.setData({
            list: this.data.list.filter(item => item.id !== id)
          })
          this.syncSelectionState()
          wx.setStorageSync(DIALECT_LIST_DIRTY_KEY, true)
          if (!this.data.list.length) {
            this.loadList(true)
          }
        } catch (err) {
          hideLoading()
          showToast('删除失败')
        }
      }
    })
  },

  toggleSelect(e) {
    const id = e.currentTarget.dataset.id
    if (!id) return
    const selected = new Set(this.data.selectedIds)
    if (selected.has(id)) {
      selected.delete(id)
    } else {
      selected.add(id)
    }
    const selectedIds = Array.from(selected)
    this.setData({
      selectedIds,
      selectAll: this.data.list.length > 0 && selectedIds.length === this.data.list.length
    })
  },

  toggleSelectAll() {
    if (!this.data.list.length) return
    if (this.data.selectAll) {
      this.setData({ selectAll: false, selectedIds: [] })
      return
    }
    const selectedIds = this.data.list.map(item => item.id)
    this.setData({ selectAll: true, selectedIds })
  },

  batchDelete() {
    const ids = this.data.selectedIds || []
    if (!ids.length) {
      showToast('请先选择要删除的方言')
      return
    }
    wx.showModal({
      title: '确认批量删除',
      content: `确定删除选中的 ${ids.length} 条方言吗？删除后不可恢复。`,
      confirmColor: '#FF4D4F',
      success: async (res) => {
        if (!res.confirm) return
        showLoading('批量删除中...')
        const results = await Promise.allSettled(ids.map(id => dialectService.delete(id)))
        hideLoading()
        const successCount = results.filter(r => r.status === 'fulfilled').length
        const failCount = ids.length - successCount
        if (successCount > 0) {
          showToast(failCount > 0 ? `已删${successCount}条，失败${failCount}条` : `已删除${successCount}条`)
          wx.setStorageSync(DIALECT_LIST_DIRTY_KEY, true)
          this.loadList(true)
        } else {
          showToast('批量删除失败')
        }
      }
    })
  },

  syncSelectionState() {
    const currentIds = new Set((this.data.list || []).map(item => item.id))
    const selectedIds = (this.data.selectedIds || []).filter(id => currentIds.has(id))
    this.setData({
      selectedIds,
      selectAll: this.data.list.length > 0 && selectedIds.length === this.data.list.length
    })
  }
})
