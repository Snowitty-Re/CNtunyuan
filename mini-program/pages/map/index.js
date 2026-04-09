const services = require('../../services/index')
const { showError, showToast, joinLocation, formatDate, normalizeMediaUrl, debounce } = require('../../utils/util')
const app = getApp()

const STATUS_TEXT = {
  missing: '失踪中',
  searching: '寻找中',
  found: '已找到',
  reunited: '已团圆',
  closed: '已结案'
}

const STATUS_MARKER_ICON = {
  missing: '/assets/images/marker_red.png',
  searching: '/assets/images/marker_orange.png',
  found: '/assets/images/marker_green.png',
  reunited: '/assets/images/marker_blue.png',
  closed: '/assets/images/marker.png'
}

const STATUS_LABEL_BG = {
  missing: '#ffe5e5',
  searching: '#fff0df',
  found: '#e6f7e6',
  reunited: '#e6f4ff',
  closed: '#efefef'
}

const STATUS_LABEL_COLOR = {
  missing: '#d9363e',
  searching: '#d46b08',
  found: '#389e0d',
  reunited: '#1677ff',
  closed: '#666666'
}

const DEFAULT_LAT = 39.9042
const DEFAULT_LNG = 116.4074
const DEFAULT_SCALE = 5

Page({
  data: {
    loading: false,
    status: '',
    keyword: '',
    panelExpanded: true,

    latitude: DEFAULT_LAT,
    longitude: DEFAULT_LNG,
    scale: DEFAULT_SCALE,
    markers: [],
    includePoints: [],

    allCases: [],
    filteredCases: [],
    visibleCases: [],
    selectedCase: null,

    totalCount: 0,
    matchedCount: 0,
    visibleCount: 0,
    mappableCount: 0,
    selectedStatusLabel: '全部状态'
  },

  onLoad() {
    if (!app.ensureAuth || !app.ensureAuth()) return
    this.mapCtx = wx.createMapContext('map', this)
    this.debouncedSyncVisibleCases = debounce(() => {
      this.syncVisibleCasesByRegion()
    }, 180)
    this.debouncedApplySearch = debounce(() => {
      this.applyFilters()
    }, 180)

    this.loadCases()
  },

  async loadCases() {
    this.setData({ loading: true })
    try {
      const params = {
        page: 1,
        page_size: 100
      }
      if (this.data.status) {
        params.status = this.data.status
      }
      const result = await services.missingPerson.getList(params)

      const allCases = (result.list || []).map((item) => this.normalizeCase(item))
      this.setData({
        allCases,
        totalCount: allCases.length,
        loading: false
      })
      this.applyFilters()
    } catch (error) {
      console.error('加载案件失败:', error)
      this.setData({ loading: false })
      showError('案件地图加载失败')
    }
  },

  normalizeCase(item) {
    const latitude = Number(item.missing_latitude || 0)
    const longitude = Number(item.missing_longitude || 0)
    const photos = Array.isArray(item.photos) ? item.photos : []
    return {
      ...item,
      missing_latitude: latitude,
      missing_longitude: longitude,
      hasCoordinates: !!(latitude && longitude),
      status_text: STATUS_TEXT[item.status] || item.status || '未知状态',
      displayLocation: joinLocation(item),
      missing_time_text: item.missing_time ? formatDate(item.missing_time, 'YYYY-MM-DD HH:mm') : '未知时间',
      cover: photos[0] && photos[0].url ? normalizeMediaUrl(photos[0].url) : '/assets/images/default-avatar.png'
    }
  },

  applyFilters() {
    const keyword = (this.data.keyword || '').trim().toLowerCase()
    const filteredCases = this.data.allCases.filter((item) => {
      if (!item.hasCoordinates) return false
      if (!keyword) return true
      const name = (item.name || '').toLowerCase()
      const location = (item.displayLocation || '').toLowerCase()
      const detail = (item.address || '').toLowerCase()
      return name.includes(keyword) || location.includes(keyword) || detail.includes(keyword)
    })

    const markers = this.buildMarkers(filteredCases)
    const includePoints = filteredCases.map((item) => ({
      latitude: item.missing_latitude,
      longitude: item.missing_longitude
    }))

    const currentSelectedID = this.data.selectedCase ? this.data.selectedCase.id : ''
    const selectedCase = filteredCases.find((item) => item.id === currentSelectedID) || filteredCases[0] || null

    this.setData({
      filteredCases,
      markers,
      includePoints,
      matchedCount: filteredCases.length,
      mappableCount: filteredCases.length,
      selectedCase,
      selectedStatusLabel: this.getStatusLabel(this.data.status)
    })

    if (!filteredCases.length) {
      this.setData({
        visibleCases: [],
        visibleCount: 0,
        panelExpanded: true
      })
      return
    }

    this.focusOnCases(filteredCases)

    setTimeout(() => {
      this.syncVisibleCasesByRegion()
    }, 220)
  },

  focusOnCases(cases) {
    const targetCases = Array.isArray(cases) ? cases.filter((item) => item.hasCoordinates) : []
    if (!targetCases.length) {
      return
    }

    if (targetCases.length === 1) {
      this.setData({
        latitude: targetCases[0].missing_latitude,
        longitude: targetCases[0].missing_longitude,
        scale: 12
      })
      return
    }

    const summary = targetCases.reduce((acc, item) => {
      acc.lat += item.missing_latitude
      acc.lng += item.missing_longitude
      return acc
    }, { lat: 0, lng: 0 })

    this.setData({
      latitude: summary.lat / targetCases.length,
      longitude: summary.lng / targetCases.length,
      scale: 5
    })
  },

  buildMarkers(cases) {
    this.markerCaseMap = {}
    return cases.map((item, index) => {
      const markerID = index + 1
      this.markerCaseMap[markerID] = item
      return {
        id: markerID,
        latitude: item.missing_latitude,
        longitude: item.missing_longitude,
        iconPath: STATUS_MARKER_ICON[item.status] || STATUS_MARKER_ICON.missing,
        width: 42,
        height: 52,
        anchor: {
          x: 0.5,
          y: 1
        },
        callout: {
          content: item.name || '未命名案件',
          color: STATUS_LABEL_COLOR[item.status] || '#2f241f',
          fontSize: 11,
          borderRadius: 12,
          bgColor: STATUS_LABEL_BG[item.status] || '#fffaf5',
          padding: 5,
          display: 'BYCLICK'
        }
      }
    })
  },

  getStatusLabel(status) {
    if (!status) return '全部状态'
    return STATUS_TEXT[status] || status
  },

  syncVisibleCasesByRegion() {
    if (!this.mapCtx || !this.data.filteredCases.length) {
      return
    }

    this.mapCtx.getRegion({
      success: (region) => {
        const southwest = region && region.southwest ? region.southwest : {}
        const northeast = region && region.northeast ? region.northeast : {}
        const minLat = Number(southwest.latitude || 0)
        const maxLat = Number(northeast.latitude || 0)
        const minLng = Number(southwest.longitude || 0)
        const maxLng = Number(northeast.longitude || 0)

        const visibleCases = this.data.filteredCases.filter((item) => {
          const lat = item.missing_latitude
          const lng = item.missing_longitude
          return lat >= minLat && lat <= maxLat && lng >= minLng && lng <= maxLng
        })

        const nextVisibleCases = visibleCases.length ? visibleCases : this.data.filteredCases.slice(0, 8)
        const currentSelectedID = this.data.selectedCase ? this.data.selectedCase.id : ''
        const selectedCase = nextVisibleCases.find((item) => item.id === currentSelectedID) || nextVisibleCases[0] || null

        this.setData({
          visibleCases: nextVisibleCases,
          visibleCount: visibleCases.length || nextVisibleCases.length,
          selectedCase
        })
      },
      fail: () => {
        const fallbackCases = this.data.filteredCases.slice(0, 8)
        this.setData({
          visibleCases: fallbackCases,
          visibleCount: fallbackCases.length,
          selectedCase: this.data.selectedCase || fallbackCases[0] || null
        })
      }
    })
  },

  locateCurrentPosition() {
    wx.getLocation({
      type: 'gcj02',
      success: (res) => {
        this.setData({
          latitude: res.latitude,
          longitude: res.longitude,
          scale: 11
        })
        setTimeout(() => {
          this.syncVisibleCasesByRegion()
        }, 220)
      }
    })
  },

  onMarkerTap(e) {
    const caseItem = this.markerCaseMap && this.markerCaseMap[e.markerId]
    if (!caseItem) return
    this.setData({
      selectedCase: caseItem,
      panelExpanded: true,
      latitude: caseItem.missing_latitude,
      longitude: caseItem.missing_longitude,
      scale: 15
    })
  },

  onMapTap() {
    this.setData({
      panelExpanded: true
    })
  },

  onRegionChange(e) {
    if (!e || e.type !== 'end') return
    if (this.debouncedSyncVisibleCases) {
      this.debouncedSyncVisibleCases()
    }
  },

  onStatusChange(e) {
    const status = e.currentTarget.dataset.status || ''
    if (status === this.data.status) return
    this.setData({ status })
    this.loadCases()
  },

  onSearchInput(e) {
    this.setData({ keyword: e.detail.value || '' })
    if (this.debouncedApplySearch) {
      this.debouncedApplySearch()
    }
  },

  clearKeyword() {
    this.setData({ keyword: '' })
    this.applyFilters()
  },

  togglePanel() {
    this.setData({
      panelExpanded: !this.data.panelExpanded
    })
  },

  selectCase(e) {
    const id = e.currentTarget.dataset.id
    const caseItem = this.data.filteredCases.find((item) => item.id === id)
    if (!caseItem) return

    this.setData({
      selectedCase: caseItem,
      latitude: caseItem.missing_latitude,
      longitude: caseItem.missing_longitude,
      scale: 15,
      panelExpanded: true
    })
  },

  focusSelectedCase() {
    const caseItem = this.data.selectedCase
    if (!caseItem || !caseItem.hasCoordinates) {
      showToast('当前案件暂无坐标信息')
      return
    }
    this.setData({
      latitude: caseItem.missing_latitude,
      longitude: caseItem.missing_longitude,
      scale: 16
    })
  },

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
    if (!item || !item.hasCoordinates) {
      showToast('暂无坐标信息')
      return
    }
    wx.openLocation({
      latitude: item.missing_latitude,
      longitude: item.missing_longitude,
      name: item.name || '走失地点',
      address: item.displayLocation || ''
    })
  },

  makePhoneCall(e) {
    const phone = e.currentTarget.dataset.phone
    if (!phone) {
      showToast('暂无联系电话')
      return
    }
    wx.showModal({
      title: '拨打电话',
      content: `确认拨打 ${phone}？`,
      success: (res) => {
        if (res.confirm) {
          wx.makePhoneCall({ phoneNumber: phone })
        }
      }
    })
  }
})
