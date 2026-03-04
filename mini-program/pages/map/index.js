const services = require('../../services/index')
const { showError, showToast } = require('../../utils/util')

// 地图上下文
let mapContext = null

Page({
  data: {
    // 地图配置
    latitude: 39.9042,  // 默认北京
    longitude: 116.4074,
    scale: 10,
    
    // 标记点
    markers: [],
    
    // 案件列表（用于地图标记）
    cases: [],
    
    // 当前选中的案件
    selectedCase: null,
    
    // 页面状态
    loading: false,
    showCaseList: false,
    
    // 筛选条件
    status: '',
    keyword: ''
  },

  onLoad(options) {
    // 获取当前位置
    this.getCurrentLocation()
    
    // 加载案件数据
    this.loadCases()
  },

  onReady() {
    // 创建地图上下文
    mapContext = wx.createMapContext('map')
  },

  // 获取当前位置
  getCurrentLocation() {
    wx.getLocation({
      type: 'gcj02',
      success: (res) => {
        this.setData({
          latitude: res.latitude,
          longitude: res.longitude
        })
      },
      fail: () => {
        showToast('获取位置失败，请检查权限设置')
      }
    })
  },

  // 加载案件数据
  async loadCases() {
    this.setData({ loading: true })
    
    try {
      const params = {
        page: 1,
        page_size: 100,
        status: this.data.status || undefined
      }
      
      const result = await services.missingPerson.getList(params)
      const cases = result.list || []
      
      // 生成地图标记
      const markers = this.generateMarkers(cases)
      
      this.setData({
        cases,
        markers,
        loading: false
      })
    } catch (error) {
      console.error('加载案件失败:', error)
      this.setData({ loading: false })
      showError('加载失败')
    }
  },

  // 生成地图标记
  generateMarkers(cases) {
    return cases
      .filter(item => item.last_seen_latitude && item.last_seen_longitude)
      .map((item, index) => ({
        id: index,
        latitude: parseFloat(item.last_seen_latitude),
        longitude: parseFloat(item.last_seen_longitude),
        title: item.name,
        iconPath: this.getMarkerIcon(item.status),
        width: 40,
        height: 40,
        callout: {
          content: `${item.name}\n${item.last_seen_location || '未知位置'}`,
          color: '#333',
          fontSize: 14,
          borderRadius: 8,
          bgColor: '#fff',
          padding: 10,
          display: 'BYCLICK'
        },
        data: item
      }))
  },

  // 获取标记图标
  getMarkerIcon(status) {
    const iconMap = {
      'missing': '/assets/icons/marker_red.png',
      'searching': '/assets/icons/marker_orange.png',
      'found': '/assets/icons/marker_green.png',
      'reunited': '/assets/icons/marker_blue.png'
    }
    return iconMap[status] || '/assets/icons/marker_red.png'
  },

  // 标记点击事件
  onMarkerTap(e) {
    const markerId = e.detail.markerId
    const marker = this.data.markers[markerId]
    
    if (marker && marker.data) {
      this.setData({
        selectedCase: marker.data,
        showCaseList: true
      })
    }
  },

  // 地图点击事件
  onMapTap() {
    this.setData({
      selectedCase: null,
      showCaseList: false
    })
  },

  // 视野改变事件
  onRegionChange(e) {
    if (e.type === 'end') {
      // 可以在这里根据视野范围加载数据
    }
  },

  // 定位到当前位置
  locateCurrentPosition() {
    this.getCurrentLocation()
    mapContext.moveToLocation()
  },

  // 显示案件列表
  showCaseList() {
    this.setData({ showCaseList: true })
  },

  // 隐藏案件列表
  hideCaseList() {
    this.setData({ showCaseList: false })
  },

  // 筛选状态改变
  onStatusChange(e) {
    const status = e.currentTarget.dataset.status
    this.setData({ status }, () => {
      this.loadCases()
    })
  },

  // 搜索输入
  onSearchInput(e) {
    this.setData({ keyword: e.detail.value })
  },

  // 搜索
  onSearch() {
    const { keyword, cases } = this.data
    
    if (!keyword) {
      this.loadCases()
      return
    }
    
    // 本地筛选
    const filtered = cases.filter(item => 
      item.name.includes(keyword) || 
      (item.last_seen_location && item.last_seen_location.includes(keyword))
    )
    
    const markers = this.generateMarkers(filtered)
    
    this.setData({
      markers,
      showCaseList: true
    })
  },

  // 导航到案件位置
  navigateToLocation(e) {
    const item = e.currentTarget.dataset.item
    
    if (!item.last_seen_latitude || !item.last_seen_longitude) {
      showToast('该案件没有位置信息')
      return
    }
    
    wx.openLocation({
      latitude: parseFloat(item.last_seen_latitude),
      longitude: parseFloat(item.last_seen_longitude),
      name: item.name,
      address: item.last_seen_location || '未知位置'
    })
  },

  // 跳转到案件详情
  goToCaseDetail(e) {
    const id = e.currentTarget.dataset.id
    wx.navigateTo({
      url: `/pages/cases/detail?id=${id}`
    })
  },

  // 跳转到创建案件
  goToCreateCase() {
    wx.navigateTo({
      url: '/pages/cases/create'
    })
  },

  // 刷新数据
  onRefresh() {
    this.loadCases()
  }
})
