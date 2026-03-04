const missingPersonService = require('../../services/missingPerson')
const uploadService = require('../../services/upload')
const { showLoading, hideLoading, showSuccess, showError, validatePhone, formatDate } = require('../../utils/util')

// 性别选项
const GENDER_OPTIONS = [
  { value: 'male', label: '男' },
  { value: 'female', label: '女' },
  { value: 'other', label: '其他' }
]

// 案件类型选项
const CASE_TYPE_OPTIONS = [
  { value: 'elderly', label: '老人走失' },
  { value: 'child', label: '儿童走失' },
  { value: 'adult', label: '成年人走失' },
  { value: 'disability', label: '残障人士走失' },
  { value: 'other', label: '其他' }
]

Page({
  data: {
    // 表单数据
    form: {
      name: '',
      gender: 'male',
      age: '',
      height: '',
      caseType: 'elderly',
      missingTime: '',
      missingLocation: '',
      latitude: '',
      longitude: '',
      description: '',
      appearance: '',
      clothing: '',
      specialFeatures: '',
      contactName: '',
      contactRelation: '',
      contactPhone: ''
    },
    
    // 选项数据
    genderOptions: GENDER_OPTIONS,
    caseTypeOptions: CASE_TYPE_OPTIONS,
    genderLabel: '男',
    caseTypeLabel: '老人走失',
    
    // 照片
    photos: [], // 本地临时文件路径
    uploadedPhotos: [], // 已上传的文件URL
    
    // 状态
    submitting: false,
    uploadProgress: 0
  },

  onLoad() {
    // 设置默认失踪时间为当前时间
    this.setDefaultMissingTime()
  },

  /**
   * 设置默认失踪时间
   */
  setDefaultMissingTime() {
    const now = new Date()
    const year = now.getFullYear()
    const month = String(now.getMonth() + 1).padStart(2, '0')
    const day = String(now.getDate()).padStart(2, '0')
    const hour = String(now.getHours()).padStart(2, '0')
    const minute = String(now.getMinutes()).padStart(2, '0')
    
    this.setData({
      'form.missingTime': `${year}-${month}-${day} ${hour}:${minute}`
    })
  },

  /**
   * 输入处理
   */
  onInput(e) {
    const { field } = e.currentTarget.dataset
    const { value } = e.detail
    this.setData({ [`form.${field}`]: value })
  },

  /**
   * 数字输入处理
   */
  onNumberInput(e) {
    const { field } = e.currentTarget.dataset
    const { value } = e.detail
    // 只允许输入数字
    const numValue = value.replace(/\D/g, '')
    this.setData({ [`form.${field}`]: numValue })
  },

  /**
   * 获取性别标签
   */
  getGenderLabel(value) {
    const item = GENDER_OPTIONS.find(item => item.value === value)
    return item ? item.label : '男'
  },

  /**
   * 获取案件类型标签
   */
  getCaseTypeLabel(value) {
    const item = CASE_TYPE_OPTIONS.find(item => item.value === value)
    return item ? item.label : '老人走失'
  },

  /**
   * 性别选择
   */
  onGenderChange(e) {
    const index = parseInt(e.detail.value)
    const gender = GENDER_OPTIONS[index].value
    this.setData({ 
      'form.gender': gender,
      'genderLabel': GENDER_OPTIONS[index].label
    })
  },

  /**
   * 案件类型选择
   */
  onCaseTypeChange(e) {
    const index = parseInt(e.detail.value)
    const caseType = CASE_TYPE_OPTIONS[index].value
    this.setData({ 
      'form.caseType': caseType,
      'caseTypeLabel': CASE_TYPE_OPTIONS[index].label
    })
  },

  /**
   * 失踪时间选择
   */
  onMissingTimeChange(e) {
    this.setData({ 
      'form.missingTime': e.detail.value 
    })
  },

  /**
   * 选择位置
   */
  chooseLocation() {
    wx.chooseLocation({
      success: (res) => {
        this.setData({
          'form.missingLocation': res.name || res.address,
          'form.latitude': res.latitude,
          'form.longitude': res.longitude
        })
      },
      fail: (err) => {
        if (err.errMsg.includes('cancel')) return
        // 检查权限
        wx.getSetting({
          success: (res) => {
            if (!res.authSetting['scope.userLocation']) {
              wx.showModal({
                title: '需要位置权限',
                content: '请允许使用位置信息以选择走失地点',
                success: (modalRes) => {
                  if (modalRes.confirm) {
                    wx.openSetting()
                  }
                }
              })
            }
          }
        })
      }
    })
  },

  /**
   * 选择照片
   */
  choosePhoto() {
    const maxCount = 9 - this.data.photos.length
    if (maxCount <= 0) {
      showError('最多上传9张照片')
      return
    }

    wx.chooseMedia({
      count: maxCount,
      mediaType: ['image'],
      sourceType: ['album', 'camera'],
      success: (res) => {
        const newPhotos = res.tempFiles.map(file => file.tempFilePath)
        this.setData({
          photos: [...this.data.photos, ...newPhotos]
        })
      },
      fail: (err) => {
        if (err.errMsg.includes('cancel')) return
        console.error('选择照片失败:', err)
      }
    })
  },

  /**
   * 预览照片
   */
  previewPhoto(e) {
    const { index } = e.currentTarget.dataset
    wx.previewImage({
      current: this.data.photos[index],
      urls: this.data.photos
    })
  },

  /**
   * 删除照片
   */
  deletePhoto(e) {
    const { index } = e.currentTarget.dataset
    const photos = [...this.data.photos]
    photos.splice(index, 1)
    this.setData({ photos })
  },

  /**
   * 上传照片
   */
  async uploadPhotos() {
    const { photos } = this.data
    if (photos.length === 0) return []

    const uploadedUrls = []
    
    for (let i = 0; i < photos.length; i++) {
      try {
        const result = await uploadService.upload(photos[i], {
          entity_type: 'missing_person',
          sort: i
        })
        uploadedUrls.push({
          url: result.url || result,
          sort: i
        })
        
        // 更新上传进度
        this.setData({
          uploadProgress: Math.round(((i + 1) / photos.length) * 100)
        })
      } catch (error) {
        console.error(`上传第${i + 1}张照片失败:`, error)
        throw new Error(`上传第${i + 1}张照片失败`)
      }
    }
    
    return uploadedUrls
  },

  /**
   * 表单验证
   */
  validateForm() {
    const { form } = this.data
    
    if (!form.name.trim()) {
      showError('请输入姓名')
      return false
    }
    
    if (!form.missingTime) {
      showError('请选择走失时间')
      return false
    }
    
    if (!form.missingLocation.trim()) {
      showError('请输入走失地点')
      return false
    }
    
    if (!form.contactPhone.trim()) {
      showError('请输入联系电话')
      return false
    }
    
    if (!validatePhone(form.contactPhone)) {
      showError('请输入正确的手机号')
      return false
    }
    
    return true
  },

  /**
   * 提交表单
   */
  async submit() {
    if (this.data.submitting) return
    
    // 表单验证
    if (!this.validateForm()) return

    this.setData({ submitting: true })
    showLoading('提交中...')

    try {
      // 先上传照片
      let photoUrls = []
      if (this.data.photos.length > 0) {
        showLoading(`上传照片 0/${this.data.photos.length}...`)
        photoUrls = await this.uploadPhotos()
      }

      showLoading('保存信息...')

      const { form } = this.data
      
      // 构建提交数据
      const submitData = {
        name: form.name.trim(),
        gender: form.gender,
        age: parseInt(form.age) || 0,
        height: parseInt(form.height) || 0,
        case_type: form.caseType,
        missing_time: new Date(form.missingTime).toISOString(),
        missing_location: form.missingLocation.trim(),
        missing_latitude: form.latitude || null,
        missing_longitude: form.longitude || null,
        missing_detail: form.description.trim(),
        appearance: form.appearance.trim(),
        clothing: form.clothing.trim(),
        special_features: form.specialFeatures.trim(),
        contact_name: form.contactName.trim(),
        contact_relation: form.contactRelation.trim(),
        contact_phone: form.contactPhone.trim(),
        photos: photoUrls
      }

      await missingPersonService.create(submitData)
      
      hideLoading()
      showSuccess('登记成功')
      
      // 延迟返回并刷新列表
      setTimeout(() => {
        const pages = getCurrentPages()
        const prevPage = pages[pages.length - 2]
        if (prevPage && prevPage.loadCases) {
          prevPage.setData({ page: 1, cases: [] })
          prevPage.loadCases()
        }
        wx.navigateBack()
      }, 1500)
      
    } catch (error) {
      hideLoading()
      this.setData({ submitting: false })
      console.error('提交失败:', error)
      showError(error.message || '提交失败，请重试')
    }
  },

  /**
   * 重置表单
   */
  resetForm() {
    wx.showModal({
      title: '确认重置',
      content: '确定要清空所有填写的信息吗？',
      success: (res) => {
        if (res.confirm) {
          this.setData({
            form: {
              name: '',
              gender: 'male',
              age: '',
              height: '',
              caseType: 'elderly',
              missingTime: '',
              missingLocation: '',
              latitude: '',
              longitude: '',
              description: '',
              appearance: '',
              clothing: '',
              specialFeatures: '',
              contactName: '',
              contactRelation: '',
              contactPhone: ''
            },
            photos: []
          })
          this.setDefaultMissingTime()
        }
      }
    })
  }
})
