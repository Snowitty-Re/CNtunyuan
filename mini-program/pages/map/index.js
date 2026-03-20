const services = require('../../services/index')
const { showError, showToast, joinLocation } = require('../../utils/util')

const STATUS_COLOR = {
  missing:   '#ff4d4f',
  searching: '#faad14',
  found:     '#52c41a',
  reunited:  '#52c41a',
  closed:    '#999999'
}

const STATUS_TEXT = {
  missing:   '失踪中',
  searching: '寻找中',
  found:     '已找到',
  reunited:  '已团圆',
  closed:    '已结案'
}

// Default map center: Beijing
const DEFAULT_LAT  = 39.9042
const DEFAULT_LNG  = 116.4074
const DEFAULT_SCALE = 5

Page({
  data: {
    cases:        [],
    markers:      [],
    latitude:     DEFAULT_LAT,
    longitude:    DEFAULT_LNG,
    scale:        DEFAULT_SCALE,
    selectedCase: null,
    loading:      false,
    showCaseList: false,
    status:       '',
    keyword:      ''
  },

  onLoad() {
    this.loadCases()
    this.locateCurrentPosition()
  },

  // ── Data Loading ──────────────────────────────────────────────────────────

  async loadCases() {
    this.setData({ loading: true })
    try {
      const result = await services.missingPerson.getList({
        page: 1,
        page_size: 100,
        status: this.data.status || undefined
      })
      const cases = (result.list || []).map(item => ({
        ...item,
        status_text:     STATUS_TEXT[item.status] || item.status,
        displayLocation: joinLocation(item)
      }))
      const markers = this._buildMarkers(cases)
      this.setData({ cases, markers, loading: false })
    } catch (error) {
      console.error('加载案件失败:', error)
      this.setData({ loading: false })
      showError('加载失败')
    }
  },

  // ── Map Helpers ───────────────────────────────────────────────────────────

  _buildMarkers(cases) {
    return cases
      .filter(item => item.missing_latitude && item.missing_longitude)
      .map((item, index) => ({
        id:        index,
        latitude:  item.missing_latitude,
        longitude: item.missing_longitude,
        iconPath:  '/assets/icons/marker.png',
        width:     36,
        height:    36,
        callout: {
          content:      item.name || '走失人员',
          color:        '#333333',
          fontSize:     12,
          borderRadius: 4,
          bgColor:      '#ffffff',
          padding:      5,
          display:      'BYCLICK'
        },
        // Non-standard: store case reference for panel list
        data: item
      }))
  },

  // ── Map Events ────────────────────────────────────────────────────────────

  locateCurrentPosition() {
    wx.getLocation({
      type: 'gcj02',
      success: (res) => {
        this.setData({
          latitude:  res.latitude,
          longitude: res.longitude,
          scale:     12
        })
      },
      fail() {
        // Permission denied — keep default center
      }
    })
  },

  onMarkerTap(e) {
    const marker = this.data.markers[e.markerId]
    if (marker) {
      this.setData({ selectedCase: marker.data, showCaseList: true })
    }
  },

  onRegionChange() {
    // No action needed for basic usage
  },

  onMapTap() {
    this.setData({ showCaseList: false, selectedCase: null })
  },

  // ── Panel ─────────────────────────────────────────────────────────────────

  showCaseList() {
    this.setData({ showCaseList: true })
  },

  hideCaseList() {
    this.setData({ showCaseList: false, selectedCase: null })
  },

  selectCase(e) {
    const id = e.currentTarget.dataset.id
    const caseItem = this.data.cases.find(c => c.id === id)
    if (caseItem) {
      this.setData({ selectedCase: caseItem, showCaseList: true })
    }
  },

  // ── Filter & Search ───────────────────────────────────────────────────────

  onStatusChange(e) {
    const status = e.currentTarget.dataset.status
    this.setData({ status }, () => { this.loadCases() })
  },

  onSearchInput(e) {
    this.setData({ keyword: e.detail.value })
  },

  onSearch() {
    const { keyword, cases } = this.data
    if (!keyword) {
      this.loadCases()
      return
    }
    const filtered = cases.filter(item =>
      (item.name && item.name.includes(keyword)) ||
      (item.displayLocation && item.displayLocation.includes(keyword))
    )
    this.setData({ cases: filtered, markers: this._buildMarkers(filtered), showCaseList: true })
  },

  // ── Navigation ────────────────────────────────────────────────────────────

  goToCaseDetail(e) {
    const id = e.currentTarget.dataset.id
    wx.navigateTo({ url: `/pages/cases/detail?id=${id}` })
  },

  goToCreateCase() {
    wx.navigateTo({ url: '/pages/cases/create' })
  },

  onRefresh() {
    this.loadCases()
  },

  navigateToLocation(e) {
    const item = e.currentTarget.dataset.item
    if (!item) return
    if (item.missing_latitude && item.missing_longitude) {
      wx.openLocation({
        latitude:  item.missing_latitude,
        longitude: item.missing_longitude,
        name:      item.name || '走失地点',
        address:   item.displayLocation || ''
      })
    } else {
      wx.showToast({ title: '暂无坐标信息', icon: 'none' })
    }
  },

  // ── Phone Call ────────────────────────────────────────────────────────────

  makePhoneCall(e) {
    const phone = e.currentTarget.dataset.phone
    if (!phone) {
      showToast('暂无联系电话')
      return
    }
    wx.showModal({
      title:   '拨打电话',
      content: `确认拨打 ${phone}？`,
      success: (res) => {
        if (res.confirm) {
          wx.makePhoneCall({ phoneNumber: phone })
        }
      }
    })
  }
})
