const DEFAULT_LAT = 39.9042
const DEFAULT_LNG = 116.4074

Page({
  data: {
    latitude: DEFAULT_LAT,
    longitude: DEFAULT_LNG,
    scale: 14,
    marker: null,
    picked: null
  },

  onLoad(options) {
    const lat = Number(options.lat)
    const lng = Number(options.lng)
    if (lat && lng) {
      this.setPickedPoint(lat, lng)
      return
    }
    this.locateCurrent()
  },

  locateCurrent() {
    wx.getLocation({
      type: 'gcj02',
      success: (res) => {
        this.setPickedPoint(res.latitude, res.longitude)
      }
    })
  },

  onMapTap(e) {
    const detail = e && e.detail ? e.detail : {}
    const latitude = Number(detail.latitude) || 0
    const longitude = Number(detail.longitude) || 0
    if (!latitude || !longitude) return
    this.setPickedPoint(latitude, longitude)
  },

  setPickedPoint(latitude, longitude) {
    const marker = {
      id: 1,
      latitude,
      longitude,
      width: 32,
      height: 32,
      iconPath: '/assets/images/marker.png'
    }
    this.setData({
      latitude,
      longitude,
      marker,
      picked: {
        latitude,
        longitude
      }
    })
  },

  confirmPick() {
    const picked = this.data.picked
    if (!picked) {
      wx.showToast({
        title: '请先点击地图选点',
        icon: 'none'
      })
      return
    }
    const eventChannel = this.getOpenerEventChannel()
    eventChannel.emit('locationPicked', {
      latitude: picked.latitude,
      longitude: picked.longitude,
      address: `地图选点(${picked.latitude.toFixed(6)}, ${picked.longitude.toFixed(6)})`
    })
    wx.navigateBack()
  }
})
