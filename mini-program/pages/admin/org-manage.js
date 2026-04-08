const organizationService = require('../../services/organization')
const { showToast, showLoading, hideLoading } = require('../../utils/util')
const { ACTIONS } = require('../../utils/permission')
const app = getApp()

Page({
  data: {
    list: [],
    loading: false,
    page: 1,
    pageSize: 30,
    noMore: false
  },

  onLoad() {
    if (!app.ensureAuth || !app.ensureAuth()) return
    if (!app.hasPermission(ACTIONS.ORG_MANAGE)) {
      showToast('无权限访问')
      wx.navigateBack()
      return
    }
    this.loadList(true)
  },

  onShow() {
    if (!app.ensureAuth || !app.ensureAuth()) return
    if (!app.hasPermission(ACTIONS.ORG_MANAGE)) return
    if (this.data.list.length > 0 && !this.data.loading) {
      this.loadList(true)
    }
  },

  onPullDownRefresh() {
    this.loadList(true).finally(() => wx.stopPullDownRefresh())
  },

  onReachBottom() {
    if (!this.data.loading && !this.data.noMore) this.loadList(false)
  },

  async loadList(reset) {
    if (this.data.loading) return
    const page = reset ? 1 : this.data.page + 1
    this.setData({ loading: true })
    try {
      const res = await organizationService.getList({ page, page_size: this.data.pageSize })
      const rows = res.list || []
      this.setData({
        page,
        list: reset ? rows : this.data.list.concat(rows),
        noMore: rows.length < this.data.pageSize
      })
    } catch (e) {
      showToast('加载组织失败')
    } finally {
      this.setData({ loading: false })
    }
  },

  onCreateOrg() {
    if (!app.hasPermission(ACTIONS.ORG_MANAGE)) {
      showToast('仅管理员可创建组织')
      return
    }
    wx.navigateTo({ url: '/pages/admin/org-edit' })
  },

  onOrgActions(e) {
    const index = Number(e.currentTarget.dataset.index)
    const item = this.data.list[index]
    if (!item || !item.id) return
    const options = [
      { key: 'edit', label: '编辑详情' },
      { key: 'rename', label: '修改名称' },
      { key: 'toggle', label: item.status === 'active' ? '设为停用' : '设为启用' }
    ]
    if (app.hasPermission(ACTIONS.ORG_MANAGE)) options.push({ key: 'delete', label: '删除组织' })

    wx.showActionSheet({
      itemList: options.map(o => o.label),
      success: ({ tapIndex }) => {
        const opt = options[tapIndex]
        if (!opt) return
        if (opt.key === 'edit') return this.editOrg(item)
        if (opt.key === 'rename') return this.renameOrg(item)
        if (opt.key === 'toggle') return this.toggleStatus(item)
        if (opt.key === 'delete') return this.deleteOrg(item)
      }
    })
  },

  editOrg(item) {
    if (!item || !item.id) return
    wx.navigateTo({ url: `/pages/admin/org-edit?id=${item.id}` })
  },

  renameOrg(item) {
    wx.showModal({
      title: '修改组织名称',
      editable: true,
      placeholderText: '请输入新名称',
      success: async (res) => {
        const name = (res.content || '').trim()
        if (!res.confirm || !name) return
        showLoading('保存中...')
        try {
          await organizationService.update(item.id, { name })
          hideLoading()
          showToast('已更新')
          this.loadList(true)
        } catch (e) {
          hideLoading()
          showToast('更新失败')
        }
      }
    })
  },

  async toggleStatus(item) {
    const status = item.status === 'active' ? 'inactive' : 'active'
    showLoading('更新中...')
    try {
      await organizationService.update(item.id, { status })
      hideLoading()
      showToast('状态已更新')
      this.loadList(true)
    } catch (e) {
      hideLoading()
      showToast('状态更新失败')
    }
  },

  deleteOrg(item) {
    wx.showModal({
      title: '删除组织',
      content: `确认删除组织「${item.name}」吗？`,
      success: async (res) => {
        if (!res.confirm) return
        showLoading('删除中...')
        try {
          await organizationService.delete(item.id)
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
