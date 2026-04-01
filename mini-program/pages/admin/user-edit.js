const userService = require('../../services/user')
const organizationService = require('../../services/organization')
const { showToast, showLoading, hideLoading } = require('../../utils/util')
const app = getApp()

const ROLE_OPTIONS = [
  { key: 'volunteer', label: '志愿者' },
  { key: 'manager', label: '管理者' },
  { key: 'admin', label: '管理员' }
]

const STATUS_OPTIONS = [
  { key: 'active', label: '正常' },
  { key: 'inactive', label: '禁用' },
  { key: 'banned', label: '封禁' }
]

Page({
  data: {
    id: '',
    isEdit: false,
    loading: false,
    saving: false,
    roleOptions: ROLE_OPTIONS,
    roleIndex: 0,
    statusOptions: STATUS_OPTIONS,
    statusIndex: 0,
    orgOptions: [],
    orgIndex: 0,
    form: {
      nickname: '',
      phone: '',
      email: '',
      password: '',
      role: 'volunteer',
      status: 'active',
      org_id: ''
    }
  },

  async onLoad(options) {
    if (!app.ensureAuth || !app.ensureAuth()) return
    if (!app.isAdmin()) {
      showToast('仅管理员可操作')
      wx.navigateBack()
      return
    }

    const id = options && options.id ? options.id : ''
    this.setData({ id, isEdit: !!id })
    wx.setNavigationBarTitle({ title: id ? '编辑用户' : '新建用户' })

    await this.loadOrgOptions()
    if (id) {
      await this.loadDetail(id)
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

      const options = (rows || []).filter(item => item && item.id)
      const defaultOrgID = (options[0] && options[0].id) || ''
      this.setData({
        orgOptions: options,
        orgIndex: 0,
        'form.org_id': this.data.isEdit ? this.data.form.org_id : defaultOrgID
      })
    } catch (e) {
      showToast('加载组织失败')
    }
  },

  async loadDetail(id) {
    this.setData({ loading: true })
    try {
      const user = await userService.getById(id)
      const roleIndex = ROLE_OPTIONS.findIndex(r => r.key === user.role)
      const statusIndex = STATUS_OPTIONS.findIndex(s => s.key === user.status)
      const orgIndex = this.data.orgOptions.findIndex(o => o.id === user.org_id)
      this.setData({
        roleIndex: roleIndex >= 0 ? roleIndex : 0,
        statusIndex: statusIndex >= 0 ? statusIndex : 0,
        orgIndex: orgIndex >= 0 ? orgIndex : 0,
        form: {
          nickname: user.nickname || '',
          phone: user.phone || '',
          email: user.email || '',
          password: '',
          role: user.role || 'volunteer',
          status: user.status || 'active',
          org_id: user.org_id || (this.data.orgOptions[0] && this.data.orgOptions[0].id) || ''
        }
      })
    } catch (e) {
      showToast('加载用户失败')
      wx.navigateBack()
    } finally {
      this.setData({ loading: false })
    }
  },

  onInput(e) {
    const field = e.currentTarget.dataset.field
    this.setData({ [`form.${field}`]: e.detail.value || '' })
  },

  onRoleChange(e) {
    const idx = Number(e.detail.value)
    const option = ROLE_OPTIONS[idx]
    if (!option) return
    this.setData({
      roleIndex: idx,
      'form.role': option.key
    })
  },

  onStatusChange(e) {
    const idx = Number(e.detail.value)
    const option = STATUS_OPTIONS[idx]
    if (!option) return
    this.setData({
      statusIndex: idx,
      'form.status': option.key
    })
  },

  onOrgChange(e) {
    const idx = Number(e.detail.value)
    const option = this.data.orgOptions[idx]
    if (!option) return
    this.setData({
      orgIndex: idx,
      'form.org_id': option.id
    })
  },

  validate() {
    const { form, isEdit } = this.data
    if (!form.nickname.trim()) {
      showToast('昵称不能为空')
      return false
    }
    if (!isEdit) {
      if (!/^1[3-9]\d{9}$/.test(form.phone.trim())) {
        showToast('手机号格式错误')
        return false
      }
      if (!form.password || form.password.length < 8) {
        showToast('初始密码至少8位')
        return false
      }
    }
    if (form.email && !/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(form.email.trim())) {
      showToast('邮箱格式错误')
      return false
    }
    if (!form.org_id) {
      showToast('请选择所属组织')
      return false
    }
    return true
  },

  async onSubmit() {
    if (this.data.saving) return
    if (!this.validate()) return

    const { form, isEdit, id } = this.data
    const email = form.email.trim()
    const payload = {
      nickname: form.nickname.trim(),
      role: form.role,
      org_id: form.org_id,
      status: form.status
    }
    if (email) payload.email = email

    this.setData({ saving: true })
    showLoading(isEdit ? '保存中...' : '创建中...')
    try {
      if (isEdit) {
        await userService.update(id, payload)
      } else {
        const createPayload = {
          nickname: form.nickname.trim(),
          phone: form.phone.trim(),
          password: form.password,
          role: form.role,
          org_id: form.org_id
        }
        if (email) createPayload.email = email
        await userService.create(createPayload)
      }
      hideLoading()
      showToast(isEdit ? '保存成功' : '创建成功')
      setTimeout(() => wx.navigateBack(), 300)
    } catch (e) {
      hideLoading()
      const msg = e && e.message ? String(e.message) : ''
      if (msg.includes('phone already exists')) {
        showToast('手机号已存在')
      } else if (msg.includes('email already exists')) {
        showToast('邮箱已存在')
      } else if (msg.includes('invalid organization id')) {
        showToast('请选择有效组织')
      } else {
        showToast(isEdit ? '保存失败' : '创建失败')
      }
    } finally {
      this.setData({ saving: false })
    }
  }
})
