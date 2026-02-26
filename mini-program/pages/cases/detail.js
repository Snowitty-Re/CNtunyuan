const { get, post } = require('../../utils/request')
const { formatDate, showConfirm, showSuccess, showLoading, hideLoading } = require('../../utils/util')

Page({
  data: {
    caseData: {},
    tracks: [],
    markers: [],
    statusMap: {
      missing: '失踪中',
      searching: '寻找中',
      found: '已找到',
      reunited: '已团圆',
      closed: '已结案'
    },
    caseTypeMap: {
      elderly: '老人走失',
      child: '儿童走失',
      adult: '成年人走失',
      disability: '残障人士走失',
      other: '其他'
    },
    genderMap: {
      male: '男',
      female: '女',
      other: '其他'
    }
  },

  onLoad(options) {
    const { id } = options
    if (id) {
      this.loadCaseDetail(id)
      this.loadTracks(id)
    }
  },

  async loadCaseDetail(id) {
    showLoading()
    try {
      const data = await get(`/missing-persons/${id}`)
      data.missing_time = formatDate(data.missing_time)
      
      const markers = []
      if (data.missing_latitude && data.missing_longitude) {
        markers.push({
          id: 1,
          latitude: data.missing_latitude,
          longitude: data.missing_longitude,
          title: '走失地点',
          iconPath: '/assets/marker.png',
          width: 30,
          height: 30
        })
      }

      this.setData({ 
        caseData: data,
        markers
      })
    } catch (error) {
      console.error('加载案件详情失败:', error)
    } finally {
      hideLoading()
    }
  },

  async loadTracks(id) {
    try {
      const tracks = await get(`/missing-persons/${id}/tracks`)
      this.setData({
        tracks: tracks.map(t => ({
          ...t,
          track_time: formatDate(t.track_time)
        }))
      })
    } catch (error) {
      console.error('加载轨迹失败:', error)
    }
  },

  previewImage(e) {
    const { url } = e.currentTarget.dataset
    const urls = this.data.caseData.photos.map(p => p.url)
    wx.previewImage({
      current: url,
      urls
    })
  },

  makePhoneCall() {
    const phone = this.data.caseData.contact_phone
    if (phone) {
      wx.makePhoneCall({ phoneNumber: phone })
    }
  },

  async addTrack() {
    const id = this.data.caseData.id
    wx.navigateTo({
      url: `/pages/map/index?caseId=${id}&mode=track`
    })
  },

  createTask() {
    const id = this.data.caseData.id
    wx.navigateTo({
      url: `/pages/tasks/create?caseId=${id}`
    })
  },

  onShareAppMessage() {
    const { caseData } = this.data
    return {
      title: `寻亲：${caseData.name}，${caseData.age}岁，${caseData.missing_location}`,
      path: `/pages/cases/detail?id=${caseData.id}`,
      imageUrl: caseData.photos[0]?.url || '/assets/share-default.png'
    }
  }
})
