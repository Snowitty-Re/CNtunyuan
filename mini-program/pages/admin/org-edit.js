const organizationService = require('../../services/organization')
const { showToast, showLoading, hideLoading } = require('../../utils/util')
const { ACTIONS } = require('../../utils/permission')
const app = getApp()

const ORG_TYPES = [
  { value: 'root', label: '总部' },
  { value: 'province', label: '省级' },
  { value: 'city', label: '市级' },
  { value: 'district', label: '区县级' },
  { value: 'street', label: '街道级' },
  { value: 'community', label: '社区级' },
  { value: 'team', label: '团队' }
]

Page({
  data: {
    id: '',
    isEdit: false,
    loading: false,
    saving: false,
    typeOptions: ORG_TYPES,
    typeIndex: 6,
    statusOptions: [
      { value: 'active', label: '启用' },
      { value: 'inactive', label: '停用' }
    ],
    statusIndex: 0,
    parentOptions: [{ id: '', name: '无（顶级组织）' }],
    parentIndex: 0,
    originalParentId: '',
    form: {
      name: '',
      code: '',
      type: 'team',
      parent_id: '',
      description: '',
      address: '',
      contact_name: '',
      contact_phone: '',
      sort_order: '0',
      status: 'active'
    }
  },

  async onLoad(options) {
    if (!app.ensureAuth || !app.ensureAuth()) return
    if (!app.hasPermission(ACTIONS.ORG_MANAGE)) {
      showToast('无权限操作')
      wx.navigateBack()
      return
    }

    const id = options && options.id ? options.id : ''
    this.setData({
      id,
      isEdit: !!id
    })
    wx.setNavigationBarTitle({
      title: id ? '编辑组织' : '新建组织'
    })

    await this.loadParentOptions()
    if (id) {
      await this.loadDetail(id)
    }
  },

  async loadParentOptions() {
    try {
      const tree = await organizationService.getTree()
      const options = [{ id: '', name: '无（顶级组织）' }]
      const walk = (node, level) => {
        if (!node || !node.id) return
        const prefix = level > 0 ? `${'　'.repeat(level)}└ ` : ''
        options.push({
          id: node.id,
          name: `${prefix}${node.name} (${node.code})`
        })
        const children = node.children || []
        children.forEach(child => walk(child, level + 1))
      }
      if (Array.isArray(tree)) {
        tree.forEach(root => walk(root, 0))
      } else {
        walk(tree, 0)
      }
      this.setData({ parentOptions: options })
    } catch (e) {
      showToast('加载组织列表失败')
    }
  },

  async loadDetail(id) {
    this.setData({ loading: true })
    try {
      const org = await organizationService.getById(id)
      const typeIndex = ORG_TYPES.findIndex(t => t.value === org.type)
      const statusIndex = this.data.statusOptions.findIndex(s => s.value === org.status)
      const parentIndex = this.data.parentOptions.findIndex(p => p.id === (org.parent_id || ''))
      this.setData({
        typeIndex: typeIndex >= 0 ? typeIndex : 6,
        statusIndex: statusIndex >= 0 ? statusIndex : 0,
        parentIndex: parentIndex >= 0 ? parentIndex : 0,
        originalParentId: org.parent_id || '',
        form: {
          name: org.name || '',
          code: org.code || '',
          type: org.type || 'team',
          parent_id: org.parent_id || '',
          description: org.description || '',
          address: org.address || '',
          contact_name: org.contact_name || '',
          contact_phone: org.contact_phone || '',
          sort_order: String(org.sort_order == null ? 0 : org.sort_order),
          status: org.status || 'active'
        }
      })
    } catch (e) {
      showToast('加载组织详情失败')
      wx.navigateBack()
    } finally {
      this.setData({ loading: false })
    }
  },

  onInput(e) {
    const field = e.currentTarget.dataset.field
    const value = e.detail.value
    this.setData({ [`form.${field}`]: value })
  },

  onTypeChange(e) {
    const idx = Number(e.detail.value)
    const option = this.data.typeOptions[idx]
    if (!option) return
    this.setData({
      typeIndex: idx,
      'form.type': option.value
    })
  },

  onParentChange(e) {
    const idx = Number(e.detail.value)
    const option = this.data.parentOptions[idx]
    if (!option) return
    this.setData({
      parentIndex: idx,
      'form.parent_id': option.id
    })
  },

  onStatusChange(e) {
    const idx = Number(e.detail.value)
    const option = this.data.statusOptions[idx]
    if (!option) return
    this.setData({
      statusIndex: idx,
      'form.status': option.value
    })
  },

  validateForm() {
    const { form } = this.data
    if (!form.name.trim()) {
      showToast('组织名称不能为空')
      return false
    }
    if (!form.code.trim()) {
      showToast('组织编码不能为空')
      return false
    }
    if (!/^[A-Za-z0-9_-]{2,50}$/.test(form.code.trim())) {
      showToast('组织编码仅支持字母、数字、-、_')
      return false
    }
    if (!form.type) {
      showToast('请选择组织类型')
      return false
    }
    if (form.contact_phone && !/^1[3-9]\d{9}$/.test(form.contact_phone.trim())) {
      showToast('联系人手机号格式错误')
      return false
    }
    const sortOrder = Number(form.sort_order)
    if (!Number.isFinite(sortOrder) || sortOrder < 0) {
      showToast('排序值必须是非负数字')
      return false
    }
    if (this.data.id && form.parent_id && form.parent_id === this.data.id) {
      showToast('父组织不能选择自己')
      return false
    }
    return true
  },

  async onSubmit() {
    if (this.data.saving) return
    if (!this.validateForm()) return

    const { form, isEdit, id } = this.data
    let payload
    if (isEdit) {
      // UpdateOrganizationRequest 仅支持以下字段
      payload = {
        name: form.name.trim(),
        code: form.code.trim(),
        description: form.description.trim(),
        address: form.address.trim(),
        contact_name: form.contact_name.trim(),
        contact_phone: form.contact_phone.trim(),
        sort_order: Number(form.sort_order || 0),
        status: form.status || 'active'
      }
    } else {
      // CreateOrganizationRequest
      payload = {
        name: form.name.trim(),
        code: form.code.trim(),
        type: form.type,
        description: form.description.trim(),
        address: form.address.trim(),
        contact_name: form.contact_name.trim(),
        contact_phone: form.contact_phone.trim(),
        sort_order: Number(form.sort_order || 0)
      }
      if (form.parent_id) {
        payload.parent_id = form.parent_id
      }
    }

    this.setData({ saving: true })
    showLoading(isEdit ? '保存中...' : '创建中...')
    try {
      if (isEdit) {
        await organizationService.update(id, payload)
        const nextParent = form.parent_id || ''
        const prevParent = this.data.originalParentId || ''
        if (nextParent !== prevParent) {
          if (!nextParent) {
            showToast('暂不支持移到顶级，请选择其他父组织')
            hideLoading()
            this.setData({ saving: false })
            return
          }
          await organizationService.move(id, nextParent)
        }
      } else {
        await organizationService.create(payload)
      }
      hideLoading()
      showToast(isEdit ? '保存成功' : '创建成功')
      setTimeout(() => wx.navigateBack(), 300)
    } catch (e) {
      hideLoading()
      const msg = (e && e.message) ? String(e.message) : ''
      if (msg.includes('organization code already exists') || msg.includes('编码') || msg.toLowerCase().includes('exists')) {
        showToast('组织编码已存在，请更换编码')
      } else {
        showToast((e && e.message) || (isEdit ? '保存失败' : '创建失败'))
      }
    } finally {
      this.setData({ saving: false })
    }
  }
})
