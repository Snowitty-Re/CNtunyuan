const { get } = require('../../utils/request')
const { getLocation, showLoading, hideLoading } = require('../../utils/util')

Page({
  data: {
    latitude: 39.9042,
    longitude: 116.4074,
    scale: 14,
    markers: [],
    nearbyCases: [],
    showPanel: true,
    mode: 'case', // case | dialect
    statusMap: {
      missing: '失踪中',
      searching: '寻找中',
      found: '已找到'
    }
  },

  onLoad() {
    this.locate()
  },

  async locate() {
    showLoading('定位中...')
    try {
      const location = await getLocation()
      this.setData({
        latitude: location.latitude,
        longitude: location.longitude
      })
      this.loadNearbyCases(location.latitude, location.longitude)
    } catch (error) {
      console.error('定位失败:', error)
      wx.showToast({
        title: '定位失败，使用默认位置',
        icon: 'none'
      })
      this.loadNearbyCases(this.data.latitude, this.data.longitude)
    } finally {
      hideLoading()
    }
  },

  async loadNearbyCases(lat, lng) {
    showLoading()
    try {
      const result = await get('/missing-persons/nearby', {
        lat,
        lng,
        radius: 10
      })

      const cases = result.map(item => ({
        ...item,
        distance: this.calculateDistance(lat, lng, item.missing_latitude, item.missing_longitude).toFixed(1)
      }))

      const markers = cases.map((item, index) => ({
        id: index,
        latitude: item.missing_latitude,
        longitude: item.missing_longitude,
        title: item.name,
        iconPath: item.status === 'missing' ? '/assets/marker-red.png' : '/assets/marker-blue.png',
        width: 40,
        height: 40,
        callout: {
          content: `${item.name} ${item.age}岁`,
          color: '#333',
          fontSize: 14,
          borderRadius: 8,
          bgColor: '#fff',
          padding: 10,
          display: 'ALWAYS'
        }
      }))

      this.setData({
        nearbyCases: cases,
        markers
      })
    } catch (error) {
      console.error('加载附近案件失败:', error)
    } finally {
      hideLoading()
    }
  },

  calculateDistance(lat1, lng1, lat2, lng2) {
    const radLat1 = lat1 * Math.PI / 180
    const radLat2 = lat2 * Math.PI / 180
    const a = radLat1 - radLat2
    const b = lng1 * Math.PI / 180 - lng2 * Math.PI / 180
    const s = 2 * Math.asin(Math.sqrt(Math.pow(Math.sin(a / 2), 2) + Math.cos(radLat1) * Math.cos(radLat2) * Math.pow(Math.sin(b / 2), 2)))
    const earthRadius = 6378.137
    return s * earthRadius
  },

  onMarkerTap(e) {
    const { markerId } = e.detail
    const caseItem = this.data.nearbyCases[markerId]
    if (caseItem) {
      this.goToCaseDetail({ currentTarget: { dataset: { id: caseItem.id } } })
    }
  },

  closePanel() {
    this.setData({ showPanel: false })
  },

  goToCaseDetail(e) {
    const id = e.currentTarget.dataset.id
    wx.navigateTo({ url: `/pages/cases/detail?id=${id}` })
  },

  switchMode(e) {
    const mode = e.currentTarget.dataset.mode
    this.setData({ mode })
    // 根据模式加载不同数据
    if (mode === 'dialect') {
      this.loadNearbyDialects()
    } else {
      this.locate()
    }
  },

  async loadNearbyDialects() {
    // 加载附近方言
  }
})
