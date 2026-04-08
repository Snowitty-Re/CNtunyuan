const dialectService = require('../../services/dialect')
const uploadService = require('../../services/upload')
const { showToast, showLoading, hideLoading } = require('../../utils/util')
const { ACTIONS } = require('../../utils/permission')
const app = getApp()

Page({
  data: {
    groups: [],
    activeGroupId: '',
    cards: [],
    loading: false
  },

  onLoad() {
    if (!app.ensureAuth || !app.ensureAuth()) return
    if (!app.hasPermission(ACTIONS.DIALECT_MANAGE)) {
      showToast('无权限访问')
      wx.navigateBack()
      return
    }
    this.loadData()
  },

  onPullDownRefresh() {
    this.loadData().finally(() => wx.stopPullDownRefresh())
  },

  async loadData() {
    this.setData({ loading: true })
    try {
      const data = await dialectService.getCardGroups()
      const groups = (data.groups || []).map(item => ({
        ...item,
        cardCount: (item.cards || []).length
      }))
      let activeGroupId = this.data.activeGroupId
      if (!activeGroupId || !groups.find(item => item.id === activeGroupId)) {
        activeGroupId = groups[0] ? groups[0].id : ''
      }
      const cards = this.extractCards(groups, activeGroupId)
      this.setData({ groups, activeGroupId, cards })
    } catch (err) {
      showToast('加载卡片模板失败')
    } finally {
      this.setData({ loading: false })
    }
  },

  extractCards(groups, groupID) {
    const group = (groups || []).find(item => item.id === groupID)
    return group && Array.isArray(group.cards) ? group.cards : []
  },

  onSwitchGroup(e) {
    const id = e.currentTarget.dataset.id
    if (!id || id === this.data.activeGroupId) return
    this.setData({
      activeGroupId: id,
      cards: this.extractCards(this.data.groups, id)
    })
  },

  async promptText(title, placeholder, value = '') {
    return new Promise((resolve) => {
      wx.showModal({
        title,
        editable: true,
        placeholderText: placeholder,
        content: value || '',
        success: (res) => {
          if (!res.confirm) return resolve('')
          resolve((res.content || '').trim())
        },
        fail: () => resolve('')
      })
    })
  },

  async addGroup() {
    const name = await this.promptText('新建分组', '请输入分组名称')
    if (!name) return
    showLoading('创建中...')
    try {
      await dialectService.createCardGroup({ name, status: 'active' })
      hideLoading()
      showToast('分组创建成功')
      this.loadData()
    } catch (err) {
      hideLoading()
      showToast('分组创建失败')
    }
  },

  async editGroup(e) {
    const index = Number(e.currentTarget.dataset.index)
    const group = this.data.groups[index]
    if (!group || !group.id) return
    const name = await this.promptText('编辑分组', '请输入分组名称', group.name || '')
    if (!name) return
    showLoading('保存中...')
    try {
      await dialectService.updateCardGroup(group.id, { name })
      hideLoading()
      showToast('分组已更新')
      this.loadData()
    } catch (err) {
      hideLoading()
      showToast('更新失败')
    }
  },

  toggleGroupStatus(e) {
    const index = Number(e.currentTarget.dataset.index)
    const group = this.data.groups[index]
    if (!group || !group.id) return
    const nextStatus = group.status === 'active' ? 'inactive' : 'active'
    showLoading('处理中...')
    dialectService.updateCardGroup(group.id, { status: nextStatus })
      .then(() => {
        hideLoading()
        showToast(nextStatus === 'active' ? '已启用' : '已停用')
        this.loadData()
      })
      .catch(() => {
        hideLoading()
        showToast('操作失败')
      })
  },

  deleteGroup(e) {
    const index = Number(e.currentTarget.dataset.index)
    const group = this.data.groups[index]
    if (!group || !group.id) return
    wx.showModal({
      title: '删除分组',
      content: `确认删除分组「${group.name || ''}」吗？`,
      confirmColor: '#FF4D4F',
      success: async (res) => {
        if (!res.confirm) return
        showLoading('删除中...')
        try {
          await dialectService.deleteCardGroup(group.id)
          hideLoading()
          showToast('分组已删除')
          this.loadData()
        } catch (err) {
          hideLoading()
          showToast('删除失败，请先删除分组下卡片')
        }
      }
    })
  },

  async addCard() {
    if (!this.data.activeGroupId) {
      showToast('请先创建分组')
      return
    }
    const name = await this.promptText('新建卡片', '请输入卡片名称（如：鸡）')
    if (!name) {
      showToast('请先填写卡片名称')
      return
    }
    const tempPath = await this.chooseImage()
    if (!tempPath) {
      showToast('请上传卡片图片')
      return
    }
    showLoading('上传中...')
    try {
      const uploadRes = await uploadService.upload(tempPath, {
        type: 'image',
        entity_type: 'dialect_card'
      })
      const imageURL = uploadRes.url || (uploadRes.data && uploadRes.data.url) || ''
      if (!imageURL) {
        throw new Error('上传失败')
      }

      await dialectService.createCard({
        group_id: this.data.activeGroupId,
        content: name,
        image_url: imageURL,
        status: 'active',
        required: true
      })
      hideLoading()
      showToast('卡片创建成功')
      this.loadData()
    } catch (err) {
      hideLoading()
      showToast('卡片创建失败')
    }
  },

  async editCard(e) {
    const index = Number(e.currentTarget.dataset.index)
    const card = this.data.cards[index]
    if (!card || !card.id) return
    wx.showActionSheet({
      itemList: ['替换图片', '修改名称'],
      success: async ({ tapIndex }) => {
        if (tapIndex === 0) {
          await this.replaceCardImage(card)
          return
        }
        const content = await this.promptText('编辑卡片名称', '请输入卡片名称', card.content || '')
        if (!content) return
        showLoading('保存中...')
        try {
          await dialectService.updateCard(card.id, { content })
          hideLoading()
          showToast('卡片已更新')
          this.loadData()
        } catch (err) {
          hideLoading()
          showToast('更新失败')
        }
      }
    })
  },

  async replaceCardImage(card) {
    const tempPath = await this.chooseImage()
    if (!tempPath) return
    showLoading('上传中...')
    try {
      const uploadRes = await uploadService.upload(tempPath, {
        type: 'image',
        entity_type: 'dialect_card'
      })
      const imageURL = uploadRes.url || (uploadRes.data && uploadRes.data.url) || ''
      if (!imageURL) {
        throw new Error('上传失败')
      }
      const payload = { image_url: imageURL }
      if (!card.content) {
        payload.content = this.generateCardContentByURL(imageURL)
      }
      await dialectService.updateCard(card.id, payload)
      hideLoading()
      showToast('图片已更新')
      this.loadData()
    } catch (err) {
      hideLoading()
      showToast('上传失败')
    }
  },

  chooseImage() {
    return new Promise((resolve) => {
      wx.chooseMedia({
        count: 1,
        mediaType: ['image'],
        sourceType: ['album', 'camera'],
        sizeType: ['compressed'],
        success: (res) => {
          const item = (res.tempFiles || [])[0]
          resolve(item ? item.tempFilePath : '')
        },
        fail: () => resolve('')
      })
    })
  },

  generateCardContentByURL(imageURL) {
    const raw = String(imageURL || '').split('?')[0]
    const fileName = raw.split('/').pop() || ''
    const pure = fileName.replace(/\.[^.]+$/, '').trim()
    if (pure) return pure.slice(0, 32)
    return `图片卡片_${Date.now().toString().slice(-6)}`
  },

  toggleCardStatus(e) {
    const index = Number(e.currentTarget.dataset.index)
    const card = this.data.cards[index]
    if (!card || !card.id) return
    const nextStatus = card.status === 'active' ? 'inactive' : 'active'
    showLoading('处理中...')
    dialectService.updateCard(card.id, { status: nextStatus })
      .then(() => {
        hideLoading()
        showToast(nextStatus === 'active' ? '已启用' : '已停用')
        this.loadData()
      })
      .catch(() => {
        hideLoading()
        showToast('操作失败')
      })
  },

  deleteCard(e) {
    const index = Number(e.currentTarget.dataset.index)
    const card = this.data.cards[index]
    if (!card || !card.id) return
    wx.showModal({
      title: '删除卡片',
      content: `确认删除卡片「${card.content || ''}」吗？`,
      confirmColor: '#FF4D4F',
      success: async (res) => {
        if (!res.confirm) return
        showLoading('删除中...')
        try {
          await dialectService.deleteCard(card.id)
          hideLoading()
          showToast('卡片已删除')
          this.loadData()
        } catch (err) {
          hideLoading()
          showToast('删除失败')
        }
      }
    })
  }
})
