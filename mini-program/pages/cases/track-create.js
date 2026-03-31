const missingPersonService = require('../../services/missingPerson')
const { showLoading, hideLoading, showSuccess, showToast, joinLocation, formatDate } = require('../../utils/util')
const app = getApp()

Page({
  data: {
    caseId: '',
    caseData: null,
    submitting: false,
    _lastAutoLocation: '',
    form: {
      description: '',
      isKeyPoint: false,
      date: '',
      time: '',
      province: '',
      city: '',
      district: '',
      address: '',
      location: '',
      lat: '',
      lng: ''
    }
  },

  onLoad(options) {
    if (!app.ensureAuth || !app.ensureAuth()) return
    const id = options && options.id
    if (!id) {
      showToast('参数错误')
      wx.navigateBack()
      return
    }
    this.setData({ caseId: id })
    this.setDefaultDateTime()
    this.loadCaseDetail(id)
  },

  setDefaultDateTime() {
    const now = new Date()
    this.setData({
      'form.date': formatDate(now, 'YYYY-MM-DD'),
      'form.time': formatDate(now, 'HH:mm')
    })
  },

  async loadCaseDetail(id) {
    try {
      const data = await missingPersonService.getById(id)
      this.setData({
        caseData: data || null,
        'form.province': data && data.province ? data.province : '',
        'form.city': data && data.city ? data.city : '',
        'form.district': data && data.district ? data.district : '',
        'form.address': data && data.address ? data.address : ''
      })
      this.refreshLocationText()
    } catch (error) {
      showToast('加载案件失败')
    }
  },

  onInput(e) {
    const field = e.currentTarget.dataset.field
    const value = e.detail.value
    this.setData({ [`form.${field}`]: value }, () => {
      if (field === 'province' || field === 'city' || field === 'district' || field === 'address') {
        this.refreshLocationText()
      }
    })
  },

  onRegionChange(e) {
    const value = e.detail.value || []
    this.setData({
      'form.province': value[0] || '',
      'form.city': value[1] || '',
      'form.district': value[2] || ''
    }, () => this.refreshLocationText())
  },

  onDateChange(e) {
    this.setData({ 'form.date': e.detail.value })
  },

  onTimeChange(e) {
    this.setData({ 'form.time': e.detail.value })
  },

  onKeyPointChange(e) {
    this.setData({ 'form.isKeyPoint': !!e.detail.value })
  },

  chooseLocation() {
    wx.chooseLocation({
      success: (res) => {
        const parsed = this.parseAddress(res.address || '')
        const address = res.address
          ? (res.name && !res.address.includes(res.name) ? `${res.address}${res.name}` : res.address)
          : (res.name || '')
        this.setData({
          'form.province': parsed.province || this.data.form.province,
          'form.city': parsed.city || this.data.form.city,
          'form.district': parsed.district || this.data.form.district,
          'form.address': address,
          'form.lat': String(res.latitude || ''),
          'form.lng': String(res.longitude || '')
        }, () => this.refreshLocationText())
      },
      fail: (err) => {
        if (err && err.errMsg && err.errMsg.includes('cancel')) return
        showToast('选择位置失败')
      }
    })
  },

  parseAddress(address) {
    const result = { province: '', city: '', district: '' }
    if (!address) return result
    const region = address.match(/^(.+?(省|自治区|行政区|特别行政区|市))(.+?(市|自治州|地区|盟))(.+?(区|县|旗))/)
    if (region) {
      result.province = region[1] || ''
      result.city = region[3] || ''
      result.district = region[5] || ''
    }
    return result
  },

  refreshLocationText() {
    const form = this.data.form
    const autoText = joinLocation({
      province: form.province,
      city: form.city,
      district: form.district,
      address: form.address
    }, '')
    if (!form.location || form.location === '' || form.location === this.data._lastAutoLocation) {
      this.setData({
        'form.location': autoText || '',
        _lastAutoLocation: autoText || ''
      })
      return
    }
    this.setData({ _lastAutoLocation: autoText || '' })
  },

  buildTrackTime(date, time) {
    if (!date || !time) return null
    const localDate = new Date(`${date}T${time}:00`)
    if (Number.isNaN(localDate.getTime())) return null
    return localDate.toISOString()
  },

  validateForm() {
    const { form, caseData } = this.data
    const description = (form.description || '').trim()
    if (!description) {
      showToast('请填写线索描述')
      return false
    }
    if (description.length < 2) {
      showToast('线索描述至少2个字')
      return false
    }

    const trackTime = this.buildTrackTime(form.date, form.time)
    if (!trackTime) {
      showToast('请选择有效的线索时间')
      return false
    }
    if (new Date(trackTime).getTime() > Date.now()) {
      showToast('线索时间不能晚于当前时间')
      return false
    }
    if (caseData && caseData.missing_time) {
      const missingAt = new Date(caseData.missing_time).getTime()
      if (!Number.isNaN(missingAt) && new Date(trackTime).getTime() < missingAt) {
        showToast('线索时间不能早于走失时间')
        return false
      }
    }

    const location = (form.location || '').trim()
    if (!location) {
      showToast('请补充线索地点')
      return false
    }

    if ((form.lat && !form.lng) || (!form.lat && form.lng)) {
      showToast('经纬度请成对填写')
      return false
    }
    if (form.lat && form.lng) {
      const lat = Number(form.lat)
      const lng = Number(form.lng)
      if (!Number.isFinite(lat) || !Number.isFinite(lng)) {
        showToast('经纬度格式无效')
        return false
      }
      if (lat < -90 || lat > 90 || lng < -180 || lng > 180) {
        showToast('经纬度超出范围')
        return false
      }
    }

    return true
  },

  async submit() {
    if (this.data.submitting) return
    if (!this.validateForm()) return

    const { form, caseId } = this.data
    const payload = {
      description: form.description.trim(),
      is_key_point: !!form.isKeyPoint,
      time: this.buildTrackTime(form.date, form.time),
      location: (form.location || '').trim(),
      province: (form.province || '').trim(),
      city: (form.city || '').trim(),
      district: (form.district || '').trim(),
      address: (form.address || '').trim()
    }
    if (form.lat && form.lng) {
      payload.lat = Number(form.lat)
      payload.lng = Number(form.lng)
    }

    this.setData({ submitting: true })
    showLoading('提交中...')
    try {
      await missingPersonService.addTrack(caseId, payload)
      hideLoading()
      showSuccess('线索提交成功')
      setTimeout(() => wx.navigateBack(), 500)
    } catch (error) {
      hideLoading()
      showToast('线索提交失败')
    } finally {
      this.setData({ submitting: false })
    }
  }
})
