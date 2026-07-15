const userService = require('../../services/user')
const organizationService = require('../../services/organization')
const { showToast, showLoading, hideLoading, formatDate } = require('../../utils/util')
const { ACTIONS, assignableRoles, isManagerRole } = require('../../utils/permission')
const app = getApp()

Page({
  data: {
    keyword: '',
    statusFilter: '',
    roleFilter: '',
    orgFilter: '',
    tabs: [
      { key: '', label: '全部' },
      { key: 'active', label: '正常' },
      { key: 'inactive', label: '待审批' },
      { key: 'banned', label: '封禁' }
    ],
    roleOptions: [
      { key: '', label: '全部角色' },
      { key: 'volunteer', label: '志愿者' },
      { key: 'manager', label: '管理者' },
      { key: 'admin', label: '管理员' },
      { key: 'super_admin', label: '超级管理员' }
    ],
    roleIndex: 0,
    orgOptions: [{ id: '', name: '全部组织' }],
    orgIndex: 0,
    list: [],
    page: 1,
    pageSize: 20,
    loading: false,
    noMore: false,
    canCreateUser: false,
    canModifyUser: false,
    canUpdateStatus: false
  },

  onLoad(options) {
    if (!app.ensureAuth || !app.ensureAuth()) return
    if (!app.hasPermission(ACTIONS.USER_VIEW)) {
      showToast('无权限访问')
      wx.navigateBack()
      return
    }
    const role = options && options.role ? String(options.role) : ''
    const roleIndex = this.data.roleOptions.findIndex(item => item.key === role)
    const userInfo = wx.getStorageSync('userInfo') || app.globalData.userInfo || {}

    this.setData({
      canCreateUser: app.hasPermission(ACTIONS.USER_CREATE),
      canModifyUser: app.hasPermission(ACTIONS.USER_MODIFY),
      // Backend: PUT /users/:id/status requires manager+
      canUpdateStatus: isManagerRole(userInfo),
      roleFilter: roleIndex >= 0 ? role : '',
      roleIndex: roleIndex >= 0 ? roleIndex : 0
    })
    this.loadOrgOptions()
    this.loadList(true)
  },

  onShow() {
    if (!app.ensureAuth || !app.ensureAuth()) return
    if (!app.hasPermission(ACTIONS.USER_VIEW)) return
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

  onRoleChange(e) {
    const idx = Number(e.detail.value)
    const item = this.data.roleOptions[idx]
    this.setData({
      roleIndex: idx,
      roleFilter: item ? item.key : ''
    })
    this.loadList(true)
  },

  onOrgChange(e) {
    const idx = Number(e.detail.value)
    const item = this.data.orgOptions[idx]
    this.setData({
      orgIndex: idx,
      orgFilter: item ? item.id : ''
    })
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
    if (this.data.roleFilter) params.role = this.data.roleFilter
    if (this.data.orgFilter) params.org_id = this.data.orgFilter

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

  async loadOrgOptions() {
    try {
      const rows = []
      let page = 1
      const pageSize = 100
      let keepGoing = true

      while (keepGoing) {
        const res = await organizationService.getList({ page, page_size: pageSize })
        const chunk = res.list || []
        rows.push(...chunk)
        keepGoing = chunk.length === pageSize
        page += 1
        if (page > 20) break
      }

      this.setData({
        orgOptions: [{ id: '', name: '全部组织' }].concat(
          rows.map(item => ({ id: item.id, name: item.name }))
        )
      })
    } catch (e) {
      // ignore
    }
  },

  onCreateUser() {
    if (!this.data.canCreateUser) {
      showToast('仅管理员可创建用户')
      return
    }
    wx.navigateTo({ url: '/pages/admin/user-edit' })
  },

  onUserActions(e) {
    if (!this.data.canModifyUser && !this.data.canUpdateStatus) {
      showToast('无权限操作')
      return
    }
    const index = Number(e.currentTarget.dataset.index)
    const item = this.data.list[index]
    if (!item || !item.id) return

    const options = []
    if (this.data.canModifyUser) options.push({ key: 'edit', label: '编辑用户' })
    if (this.data.canUpdateStatus) {
      if (item.status === 'inactive') {
        options.push({ key: 'active', label: '审批通过（设为正常）' })
      } else if (item.status !== 'active') {
        options.push({ key: 'active', label: '设为正常' })
      }
      if (item.status !== 'inactive') options.push({ key: 'inactive', label: '设为禁用' })
      if (item.status !== 'banned') options.push({ key: 'banned', label: '设为封禁' })
    }
    if (this.data.canModifyUser) options.push({ key: 'role', label: '调整角色' })
    if (this.data.canModifyUser) options.push({ key: 'delete', label: '删除用户' })

    if (options.length === 0) {
      showToast('无可执行操作')
      return
    }

    wx.showActionSheet({
      itemList: options.map(o => o.label),
      success: ({ tapIndex }) => {
        const opt = options[tapIndex]
        if (!opt) return
        if (opt.key === 'edit') return this.editUser(item)
        if (opt.key === 'role') return this.changeRole(item)
        if (opt.key === 'delete') return this.deleteUser(item)
        this.changeStatus(item, opt.key)
      }
    })
  },

  editUser(item) {
    if (!item || !item.id) return
    wx.navigateTo({ url: `/pages/admin/user-edit?id=${item.id}` })
  },

  async changeStatus(item, status) {
    if (!this.data.canUpdateStatus) {
      showToast('无权限操作')
      return
    }
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
    if (!this.data.canModifyUser) {
      showToast('无权限操作')
      return
    }
    const operator = wx.getStorageSync('userInfo') || app.globalData.userInfo || {}
    const roles = assignableRoles(operator)
    if (!roles.length) {
      showToast('当前账号无可分配角色')
      return
    }
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
          showToast((e && e.message) || '角色更新失败')
        }
      }
    })
  },

  async deleteUser(item) {
    if (!this.data.canModifyUser) {
      showToast('无权限操作')
      return
    }
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
