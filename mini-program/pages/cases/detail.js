const missingPersonService = require('../../services/missingPerson')
const { formatDate, showConfirm, showSuccess, showLoading, hideLoading } = require('../../utils/util')

// 状态映射
const STATUS_MAP = {
  'missing': { label: '失踪中', color: '#ff4d4f' },
  'searching': { label: '寻找中', color: '#1890ff' },
  'found': { label: '已找到', color: '#52c41a' },
  'reunited': { label: '已团圆', color: '#52c41a' },
  'closed': { label: '已结案', color: '#999' }
}

// 案件类型映射
const CASE_TYPE_MAP = {
  'elderly': '老人走失',
  'child': '儿童走失',
  'adult': '成年人走失',
  'disability': '残障人士走失',
  'other': '其他'
}

// 性别映射
const GENDER_MAP = {
  'male': '男',
  'female': '女',
  'other': '其他'
}

Page({
  data: {
    id: '',
    caseData: {},
    tracks: [],
    markers: [],
    loading: false,
    isManager: false, // 是否为管理者
    
    // 映射常量
    statusMap: STATUS_MAP,
    caseTypeMap: CASE_TYPE_MAP,
    genderMap: GENDER_MAP
  },

  onLoad(options) {
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
  },

  onShow() {
    // 返回时刷新数据
    if (this.data.id) {
      this.loadCaseDetail()
      this.loadTracks()
    }
  },

  /**
   * 检查用户角色
   */
  checkUserRole() {
    const userInfo = wx.getStorageSync('userInfo') || {}
    const role = userInfo.role || ''
    // super_admin, admin, manager 视为管理者
    const isManager = ['super_admin', 'admin', 'manager'].includes(role)
    this.setData({ isManager })
  },

  /**
   * 加载案件详情
   */
  async loadCaseDetail() {
    showLoading()
    try {
      const data = await missingPersonService.getById(this.data.id)
      
      // 处理数据
      data.missing_time = formatDate(data.missing_time)
      data.created_at = formatDate(data.created_at)
      data.photoUrl = this.getFirstPhoto(data.photos)
      data.possible_location = data.possible_location || '未知'
      data.appearance = data.appearance || '暂无描述'
      data.clothing = data.clothing || '暂无描述'
      data.special_features = data.special_features || '无'
      
      // 设置地图标记
      const markers = []
      if (data.missing_latitude && data.missing_longitude) {
        markers.push({
          id: 1,
          latitude: data.missing_latitude,
          longitude: data.missing_longitude,
          title: '走失地点',
          iconPath: '/assets/images/marker.png',
          width: 40,
          height: 40
        })
      }

      this.setData({ 
        caseData: data,
        markers
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
          track_time: formatDate(t.track_time)
        }))
      })
    } catch (error) {
      console.error('加载轨迹失败:', error)
    }
  },

  /**
   * 获取第一张图片
   */
  getFirstPhoto(photos) {
    if (photos && photos.length > 0) {
      return photos[0].url || photos[0]
    }
    return '/assets/images/default-avatar.png'
  },

  /**
   * 预览图片
   */
  previewImage(e) {
    const { url } = e.currentTarget.dataset
    const urls = (this.data.caseData.photos || []).map(p => p.url || p)
    wx.previewImage({
      current: url,
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
    wx.makePhoneCall({ 
      phoneNumber: phone,
      fail: () => {
        wx.showToast({ title: '呼叫失败', icon: 'none' })
      }
    })
  },

  /**
   * 添加轨迹
   */
  addTrack() {
    const { id } = this.data
    wx.navigateTo({
      url: `/pages/cases/track?id=${id}`
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
      await missingPersonService.markFound(this.data.id, {
        found_location: this.data.caseData.missing_location,
        found_time: new Date().toISOString(),
        description: '通过志愿者帮助找到'
      })
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
      hideLoading()
      showSuccess('标记成功')
      this.loadCaseDetail()
    } catch (error) {
      hideLoading()
      wx.showToast({ title: '操作失败', icon: 'none' })
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
   * 编辑案件
   */
  editCase() {
    const { id } = this.data
    wx.navigateTo({
      url: `/pages/cases/edit?id=${id}`
    })
  },

  /**
   * 打开地图导航
   */
  openLocation() {
    const { caseData } = this.data
    if (!caseData.missing_latitude || !caseData.missing_longitude) {
      wx.showToast({ title: '暂无位置信息', icon: 'none' })
      return
    }
    
    wx.openLocation({
      latitude: parseFloat(caseData.missing_latitude),
      longitude: parseFloat(caseData.missing_longitude),
      name: '走失地点',
      address: caseData.missing_location
    })
  },

  /**
   * 分享
   */
  onShareAppMessage() {
    const { caseData } = this.data
    return {
      title: `寻亲：${caseData.name}，${caseData.age}岁，${caseData.missing_location}`,
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
