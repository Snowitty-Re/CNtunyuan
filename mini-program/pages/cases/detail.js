const missingPersonService = require('../../services/missingPerson')
const taskService = require('../../services/task')
const { formatDate, showConfirm, showSuccess, showLoading, hideLoading, joinLocation, normalizeMediaUrl, normalizeAge } = require('../../utils/util')
const app = getApp()
const CASES_LIST_DIRTY_KEY = 'cases_list_dirty'

// 状态映射
const STATUS_MAP = {
  'missing': { label: '失踪中', color: '#ff4d4f' },
  'searching': { label: '寻找中', color: '#1890ff' },
  'found': { label: '已找到', color: '#52c41a' },
  'reunited': { label: '已团圆', color: '#52c41a' },
  'closed': { label: '已结案', color: '#999' }
}

// 性别映射
const GENDER_MAP = {
  'male': '男',
  'female': '女',
  'other': '其他'
}

const TASK_STATUS_MAP = {
  draft: '草稿',
  pending: '待分配',
  assigned: '已分配',
  processing: '进行中',
  completed: '已完成',
  cancelled: '已取消',
  overdue: '已逾期'
}

Page({
  data: {
    id: '',
    caseData: {},
    tracks: [],
    relatedTasks: [],
    loading: false,
    isManager: false, // 是否为管理者
    
    // 映射常量
    statusMap: STATUS_MAP,
    genderMap: GENDER_MAP,
    taskStatusMap: TASK_STATUS_MAP
  },

  onLoad(options) {
    if (!app.ensureAuth || !app.ensureAuth()) return
    const { id } = options
    if (!id) {
      wx.showToast({ title: '参数错误', icon: 'error' })
      wx.navigateBack()
      return
    }
    
    this.setData({ id })
    
    // 检查用户权限（从本地存储或全局数据获取）
    this.checkUserRole()
    
    this.loadCaseDetail()
    this.loadTracks()
    this.loadRelatedTasks()
  },

  onShow() {
    if (!app.ensureAuth || !app.ensureAuth()) return
    // 返回时刷新数据
    if (this.data.id) {
      Promise.all([this.loadCaseDetail(), this.loadTracks(), this.loadRelatedTasks()]).catch(err => console.error('刷新详情失败:', err))
    }
  },

  /**
   * 检查用户角色
   */
  checkUserRole() {
    this.setData({ isManager: app.isManager() })
  },

  /**
   * 加载案件详情
   */
  async loadCaseDetail() {
    showLoading()
    try {
      const data = await missingPersonService.getById(this.data.id)
      
      // 处理数据 - 使用后端实际返回的字段
      data.missing_time = formatDate(data.missing_time)
      const createdAtRaw = data.created_at
      data.created_at = formatDate(createdAtRaw)
      data.created_at_full = formatDate(createdAtRaw, 'YYYY-MM-DD HH:mm:ss')
      data.photos = (data.photos || []).map((photo) => {
        if (photo && typeof photo === 'object') {
          return { ...photo, url: normalizeMediaUrl(photo.url) || photo.url }
        }
        return normalizeMediaUrl(photo) || photo
      })
      data.photoUrl = this.getFirstPhoto(data.photos, data.photo_url)
      
      data.displayLocation = joinLocation(data, '未知')
      data.displayAge = normalizeAge(data.age)
      data.displayAgeText = data.displayAge ? `${data.displayAge}岁` : '未知'
      
      // 仅展示真实录入值，不再注入默认文案
      data.clothes = (data.clothes || '').trim()
      data.features = (data.features || '').trim()
      // 详细描述
      data.description = data.description || '暂无描述'
      
      // 找到信息
      if (data.found_time) {
        data.found_time = formatDate(data.found_time)
      }

      this.setData({ 
        caseData: data
      })
    } catch (error) {
      console.error('加载案件详情失败:', error)
      wx.showToast({ title: '加载失败', icon: 'none' })
    } finally {
      hideLoading()
    }
  },

  /**
   * 加载轨迹记录
   */
  async loadTracks() {
    try {
      const tracks = await missingPersonService.getTracks(this.data.id)
      this.setData({
        tracks: (tracks || []).map(t => ({
          ...t,
          displayLocation: t.location || joinLocation(t, '未知地点'),
          // 后端返回的字段是 time，不是 track_time
          displayTime: formatDate(t.time || t.created_at)
        }))
      })
    } catch (error) {
      console.error('加载轨迹失败:', error)
    }
  },

  /**
   * 加载关联任务
   */
  async loadRelatedTasks() {
    try {
      const result = await taskService.getList({
        page: 1,
        page_size: 20,
        missing_person_id: this.data.id
      })
      const list = (result && result.list) || []
      const relatedTasks = list.map(item => ({
        ...item,
        assigneeName: (item.assignee && (item.assignee.nickname || item.assignee.name)) || '未分配',
        updatedAtText: formatDate(item.updated_at || item.created_at, 'YYYY-MM-DD HH:mm')
      }))
      this.setData({ relatedTasks })
    } catch (error) {
      console.error('加载关联任务失败:', error)
    }
  },

  /**
   * 获取第一张图片
   */
  getFirstPhoto(photos, fallbackPhotoUrl = '') {
    if (photos && photos.length > 0) {
      return normalizeMediaUrl(photos[0].url || photos[0]) || '/assets/images/default-avatar.png'
    }
    if (fallbackPhotoUrl) {
      return normalizeMediaUrl(fallbackPhotoUrl) || '/assets/images/default-avatar.png'
    }
    return '/assets/images/default-avatar.png'
  },

  /**
   * 预览图片
   */
  previewImage(e) {
    const { url } = e.currentTarget.dataset
    const urls = (this.data.caseData.photos || []).map(p => normalizeMediaUrl(p.url || p)).filter(Boolean)
    if (urls.length === 0 && this.data.caseData.photoUrl) {
      urls.push(this.data.caseData.photoUrl)
    }
    wx.previewImage({
      current: normalizeMediaUrl(url) || url,
      urls
    })
  },

  /**
   * 拨打电话
   */
  makePhoneCall() {
    const phone = this.data.caseData.contact_phone
    if (!phone) {
      wx.showToast({ title: '暂无联系电话', icon: 'none' })
      return
    }
    wx.showModal({
      title:   '拨打电话',
      content: `确认拨打 ${phone}？`,
      success: (res) => {
        if (res.confirm) {
          wx.makePhoneCall({
            phoneNumber: phone,
            fail: () => { wx.showToast({ title: '呼叫失败', icon: 'none' }) }
          })
        }
      }
    })
  },

  /**
   * 添加轨迹
   */
  addTrack() {
    const { id } = this.data
    if (!id) {
      wx.showToast({ title: '案件ID无效', icon: 'none' })
      return
    }
    wx.navigateTo({
      url: `/pages/cases/track-create?id=${id}`
    })
  },

  /**
   * 更新状态
   */
  async updateStatus() {
    const { caseData } = this.data
    const statusOptions = [
      { value: 'missing', label: '失踪中' },
      { value: 'searching', label: '寻找中' },
      { value: 'found', label: '已找到' },
      { value: 'reunited', label: '已团圆' }
    ]
    
    const items = statusOptions.map(item => item.label)
    
    wx.showActionSheet({
      itemList: items,
      success: async (res) => {
        const newStatus = statusOptions[res.tapIndex].value
        if (newStatus === caseData.status) return
        
        showLoading('更新中...')
        try {
          await missingPersonService.updateStatus(caseData.id, newStatus)
          wx.setStorageSync(CASES_LIST_DIRTY_KEY, 1)
          hideLoading()
          showSuccess('状态更新成功')
          this.loadCaseDetail()
        } catch (error) {
          hideLoading()
          wx.showToast({ title: '更新失败', icon: 'none' })
        }
      }
    })
  },

  /**
   * 标记已找到
   */
  async markFound() {
    const confirmed = await showConfirm('确认标记', '确认将该案件标记为已找到？')
    if (!confirmed) return

    showLoading('处理中...')
    try {
      const { caseData } = this.data
      // 使用后端要求的字段名：location 和 note
      await missingPersonService.markFound(this.data.id, {
        location: '',  // 找到地点由后端默认处理，此处不传走失地点
        note: '通过志愿者帮助找到'
      })
      wx.setStorageSync(CASES_LIST_DIRTY_KEY, 1)
      hideLoading()
      showSuccess('标记成功')
      this.loadCaseDetail()
    } catch (error) {
      hideLoading()
      wx.showToast({ title: '操作失败', icon: 'none' })
    }
  },

  /**
   * 标记已团圆
   */
  async markReunited() {
    const confirmed = await showConfirm('确认标记', '确认将该案件标记为已团圆？')
    if (!confirmed) return

    showLoading('处理中...')
    try {
      await missingPersonService.markReunited(this.data.id)
      wx.setStorageSync(CASES_LIST_DIRTY_KEY, 1)
      hideLoading()
      showSuccess('标记成功')
      this.loadCaseDetail()
    } catch (error) {
      hideLoading()
      wx.showToast({ title: '操作失败', icon: 'none' })
    }
  },

  /**
   * 删除案件
   */
  async deleteCase() {
    const confirmed = await showConfirm('确认删除', '删除后无法恢复，确认删除该案件吗？')
    if (!confirmed) return

    showLoading('删除中...')
    try {
      await missingPersonService.delete(this.data.id)
      wx.setStorageSync(CASES_LIST_DIRTY_KEY, 1)
      hideLoading()
      showSuccess('删除成功')
      setTimeout(() => {
        wx.navigateBack()
      }, 600)
    } catch (error) {
      hideLoading()
      wx.showToast({ title: '删除失败', icon: 'none' })
    }
  },

  /**
   * 创建任务
   */
  createTask() {
    const { id } = this.data
    wx.navigateTo({
      url: `/pages/tasks/create?caseId=${id}`
    })
  },

  /**
   * 跳转任务详情
   */
  goTaskDetail(e) {
    const { id } = e.currentTarget.dataset
    if (!id) return
    wx.navigateTo({
      url: `/pages/tasks/detail?id=${id}`
    })
  },

  /**
   * 编辑案件
   */
  editCase() {
    const { id } = this.data
    wx.navigateTo({
      url: `/pages/cases/create?id=${id}`
    })
  },

  /**
   * 打开地图导航
   */
  openLocation() {
    const { caseData } = this.data

    if (caseData.missing_latitude && caseData.missing_longitude) {
      wx.openLocation({
        latitude:  caseData.missing_latitude,
        longitude: caseData.missing_longitude,
        name:      caseData.name ? `${caseData.name}走失地点` : '走失地点',
        address:   caseData.displayLocation || ''
      })
      return
    }

    if (!caseData.province && !caseData.city && !caseData.address) {
      wx.showToast({ title: '暂无位置信息', icon: 'none' })
      return
    }

    wx.showModal({
      title: '走失地点',
      content: caseData.displayLocation,
      showCancel: false
    })
  },

  /**
   * 分享
   */
  onShareAppMessage() {
    const { caseData } = this.data
    return {
      title: `寻亲：${caseData.name}，${caseData.age}岁，${caseData.displayLocation || '未知地点'}`,
      path: `/pages/cases/detail?id=${caseData.id}`,
      imageUrl: caseData.photoUrl || '/assets/images/share-default.png'
    }
  },

  /**
   * 分享到朋友圈
   */
  onShareTimeline() {
    const { caseData } = this.data
    return {
      title: `寻亲：${caseData.name}，${caseData.age}岁`,
      query: `id=${caseData.id}`,
      imageUrl: caseData.photoUrl || '/assets/images/share-default.png'
    }
  }
})
