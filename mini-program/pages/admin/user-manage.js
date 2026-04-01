const userService = require('../../services/user')
const { showToast, showLoading, hideLoading, formatDate } = require('../../utils/util')
const app = getApp()

Page({
  data: {
    keyword: '',
    statusFilter: '',
    tabs: [
      { key: '', label: '全部' },
      { key: 'active', label: '正常' },
      { key: 'inactive', label: '禁用' },
      { key: 'banned', label: '封禁' }
    ],
    list: [],
    page: 1,
    pageSize: 20,
    loading: false,
    noMore: false,
    isAdmin: false
  },

  onLoad() {
    if (!app.ensureAuth || !app.ensureAuth()) return
    if (!app.isManager()) {
      showToast('无权限访问')
      wx.navigateBack()
      return
    }
    this.setData({ isAdmin: app.isAdmin() })
    this.loadList(true)
  },

  onPullDownRefresh() {
    this.loadList(true).finally(() => wx.stopPullDownRefresh())
  },

  onReachBottom() {
    if (!this.data.loading && !this.data.noMore) this.loadList(false)
  },

  onKeywordInput(e) {
    this.setData({ keyword: e.detail.value || '' })
  },

  onSearch() {
    this.loadList(true)
  },

  onTabTap(e) {
    const key = e.currentTarget.dataset.key
    if (key === this.data.statusFilter) return
    this.setData({ statusFilter: key || '' })
    this.loadList(true)
  },

  async loadList(reset) {
    if (this.data.loading) return
    const page = reset ? 1 : this.data.page + 1
    const params = {
      page,
      page_size: this.data.pageSize
    }
    if (this.data.keyword.trim()) params.keyword = this.data.keyword.trim()
    if (this.data.statusFilter) params.status = this.data.statusFilter

    this.setData({ loading: true })
    try {
      const res = await userService.getList(params)
      const rows = (res.list || []).map(item => ({
        ...item,
        orgNameText: item.org_name || item.orgName || '未分配组织',
        createdAtText: formatDate(item.created_at, 'YYYY-MM-DD')
      }))
      this.setData({
        page,
        list: reset ? rows : this.data.list.concat(rows),
        noMore: rows.length < this.data.pageSize
      })
    } catch (e) {
      showToast('加载用户失败')
    } finally {
      this.setData({ loading: false })
    }
  },

  onUserActions(e) {
    const index = Number(e.currentTarget.dataset.index)
    const item = this.data.list[index]
    if (!item || !item.id) return

    const options = []
    if (item.status !== 'active') options.push({ key: 'active', label: '设为正常' })
    if (item.status !== 'inactive') options.push({ key: 'inactive', label: '设为禁用' })
    if (item.status !== 'banned') options.push({ key: 'banned', label: '设为封禁' })
    if (this.data.isAdmin) options.push({ key: 'role', label: '调整角色' })
    if (this.data.isAdmin) options.push({ key: 'delete', label: '删除用户' })

    wx.showActionSheet({
      itemList: options.map(o => o.label),
      success: ({ tapIndex }) => {
        const opt = options[tapIndex]
        if (!opt) return
        if (opt.key === 'role') return this.changeRole(item)
        if (opt.key === 'delete') return this.deleteUser(item)
        this.changeStatus(item, opt.key)
      }
    })
  },

  async changeStatus(item, status) {
    showLoading('更新中...')
    try {
      await userService.updateStatus(item.id, status)
      hideLoading()
      showToast('状态已更新')
      this.loadList(true)
    } catch (e) {
      hideLoading()
      showToast('状态更新失败')
    }
  },

  changeRole(item) {
    const roles = [
      { key: 'volunteer', label: '志愿者' },
      { key: 'manager', label: '管理者' },
      { key: 'admin', label: '管理员' }
    ]
    wx.showActionSheet({
      itemList: roles.map(r => r.label),
      success: async ({ tapIndex }) => {
        const selected = roles[tapIndex]
        if (!selected) return
        showLoading('更新中...')
        try {
          await userService.updateRole(item.id, selected.key)
          hideLoading()
          showToast('角色已更新')
          this.loadList(true)
        } catch (e) {
          hideLoading()
          showToast('角色更新失败')
        }
      }
    })
  },

  async deleteUser(item) {
    wx.showModal({
      title: '删除用户',
      content: `确认删除 ${item.nickname || item.phone || '该用户'} 吗？`,
      success: async (res) => {
        if (!res.confirm) return
        showLoading('删除中...')
        try {
          await userService.delete(item.id)
          hideLoading()
          showToast('已删除')
          this.loadList(true)
        } catch (e) {
          hideLoading()
          showToast('删除失败')
        }
      }
    })
  }
})
