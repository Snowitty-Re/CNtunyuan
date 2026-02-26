const { post } = require('../../utils/request')
const { showLoading, hideLoading, showSuccess, showError } = require('../../utils/util')

Page({
  data: {
    form: {
      name: '',
      genderIndex: 0,
      age: '',
      height: '',
      caseTypeIndex: 0,
      missingTime: '',
      missingLocation: '',
      address: '',
      latitude: '',
      longitude: '',
      appearance: '',
      clothing: '',
      specialFeatures: '',
      missingDetail: '',
      contactName: '',
      contactRelation: '',
      contactPhone: '',
      contactAddress: ''
    },
    genderOptions: [
      { value: 'male', label: '男' },
      { value: 'female', label: '女' },
      { value: 'other', label: '其他' }
    ],
    caseTypeOptions: [
      { value: 'elderly', label: '老人走失' },
      { value: 'child', label: '儿童走失' },
      { value: 'adult', label: '成年人走失' },
      { value: 'disability', label: '残障人士走失' },
      { value: 'other', label: '其他' }
    ],
    dateTimeArray: [],
    dateTime: [0, 0, 0, 0, 0],
    photos: [],
    submitting: false
  },

  onLoad() {
    this.initDateTimePicker()
  },

  initDateTimePicker() {
    const now = new Date()
    const years = []
    const months = []
    const days = []
    const hours = []
    const minutes = []

    for (let i = now.getFullYear(); i >= now.getFullYear() - 5; i--) {
      years.push(i + '年')
    }
    for (let i = 1; i <= 12; i++) {
      months.push(i + '月')
    }
    for (let i = 1; i <= 31; i++) {
      days.push(i + '日')
    }
    for (let i = 0; i < 24; i++) {
      hours.push(i + '时')
    }
    for (let i = 0; i < 60; i++) {
      minutes.push(i + '分')
    }

    this.setData({
      dateTimeArray: [years, months, days, hours, minutes]
    })
  },

  onInput(e) {
    const { field } = e.currentTarget.dataset
    const { value } = e.detail
    this.setData({ [`form.${field}`]: value })
  },

  onGenderChange(e) {
    this.setData({ 'form.genderIndex': e.detail.value })
  },

  onCaseTypeChange(e) {
    this.setData({ 'form.caseTypeIndex': e.detail.value })
  },

  onDateTimeChange(e) {
    const { value } = e.detail
    const { dateTimeArray } = this.data
    const dateStr = `${dateTimeArray[0][value[0]]}${dateTimeArray[1][value[1]]}${dateTimeArray[2][value[2]]} ${dateTimeArray[3][value[3]]}${dateTimeArray[4][value[4]]}`
    this.setData({
      'form.missingTime': dateStr,
      dateTime: value
    })
  },

  chooseLocation() {
    wx.chooseLocation({
      success: (res) => {
        this.setData({
          'form.address': res.address,
          'form.latitude': res.latitude,
          'form.longitude': res.longitude
        })
      }
    })
  },

  choosePhoto() {
    wx.chooseMedia({
      count: 9 - this.data.photos.length,
      mediaType: ['image'],
      sourceType: ['album', 'camera'],
      success: (res) => {
        const newPhotos = res.tempFiles.map(file => file.tempFilePath)
        this.setData({
          photos: [...this.data.photos, ...newPhotos]
        })
      }
    })
  },

  deletePhoto(e) {
    const { index } = e.currentTarget.dataset
    const photos = [...this.data.photos]
    photos.splice(index, 1)
    this.setData({ photos })
  },

  async submit() {
    const { form, genderOptions, caseTypeOptions, photos } = this.data

    if (!form.name) {
      showError('请输入姓名')
      return
    }
    if (!form.missingTime) {
      showError('请选择走失时间')
      return
    }
    if (!form.missingLocation) {
      showError('请输入走失地点')
      return
    }
    if (!form.contactPhone) {
      showError('请输入联系电话')
      return
    }

    this.setData({ submitting: true })
    showLoading('提交中...')

    try {
      // 上传照片（简化处理，实际应上传到服务器）
      const photoUrls = photos.map((p, i) => ({ url: p, sort: i }))

      const data = {
        name: form.name,
        gender: genderOptions[form.genderIndex]?.value,
        age: parseInt(form.age) || 0,
        height: parseInt(form.height) || 0,
        case_type: caseTypeOptions[form.caseTypeIndex]?.value,
        missing_time: new Date().toISOString(), // 简化处理
        missing_location: form.missingLocation,
        missing_longitude: form.longitude,
        missing_latitude: form.latitude,
        appearance: form.appearance,
        clothing: form.clothing,
        special_features: form.specialFeatures,
        missing_detail: form.missingDetail,
        contact_name: form.contactName,
        contact_relation: form.contactRelation,
        contact_phone: form.contactPhone,
        contact_address: form.contactAddress,
        photos: photoUrls
      }

      await post('/missing-persons', data)
      hideLoading()
      showSuccess('登记成功')
      setTimeout(() => {
        wx.navigateBack()
      }, 1500)
    } catch (error) {
      hideLoading()
      this.setData({ submitting: false })
      showError('提交失败')
    }
  }
})
