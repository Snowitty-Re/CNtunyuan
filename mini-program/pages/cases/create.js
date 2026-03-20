const missingPersonService = require('../../services/missingPerson')
const uploadService = require('../../services/upload')
const { showLoading, hideLoading, showSuccess, showError, validatePhone, formatDate } = require('../../utils/util')

// 性别选项
const GENDER_OPTIONS = [
  { value: 'male', label: '男' },
  { value: 'female', label: '女' },
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
      missingTime: '',
      // 位置信息（前端展示用）
      province: '',
      city: '',
      district: '',
      address: '',
      // 详细描述（包含外貌、衣着、特征等）
      description: '',
      // 联系人信息
      contactName: '',
      contactRel: '',
      contactPhone: '',
      altContact: ''
    },
    
    // 选项数据
    genderOptions: GENDER_OPTIONS,
    genderLabel: '男',
    
    // 照片
    photos: [], // 本地临时文件路径
    uploadedPhotos: [], // 已上传的文件URL
    
    // 状态
    submitting: false,
    uploadProgress: 0,

    // 编辑模式
    isEdit: false,
    editId: ''
  },

  onLoad(options) {
    if (options.id) {
      // 编辑模式
      this.setData({ isEdit: true, editId: options.id })
      wx.setNavigationBarTitle({ title: '编辑案件' })
      this.loadCaseData(options.id)
    } else {
      // 设置默认失踪时间为当前时间
      this.setDefaultMissingTime()
    }
  },

  /**
   * 加载案件数据（编辑模式）
   */
  async loadCaseData(id) {
    try {
      showLoading('加载中...')
      const data = await missingPersonService.getById(id)
      hideLoading()

      const genderIndex = GENDER_OPTIONS.findIndex(g => g.value === data.gender)

      // 将后端的省市区地址拼接为地址字符串（用于展示）
      const locationParts = [data.province, data.city, data.district, data.address].filter(Boolean)

      this.setData({
        form: {
          name: data.name || '',
          gender: data.gender || 'male',
          age: data.age ? String(data.age) : '',
          height: data.height ? String(data.height) : '',
          missingTime: data.missing_time ? formatDate(data.missing_time, 'YYYY-MM-DD HH:mm') : '',
          // 位置信息
          province: data.province || '',
          city: data.city || '',
          district: data.district || '',
          address: data.address || '',
          // 详细描述（后端存储的是合并后的描述）
          description: data.description || '',
          // 联系人信息（注意字段名映射）
          contactName: data.contact_name || '',
          contactRel: data.contact_rel || '',  // 后端返回 contact_rel
          contactPhone: data.contact_phone || '',
          altContact: data.alt_contact || ''
        },
        // 如果有照片，设置到photos中
        uploadedPhotos: data.photos || [],
        genderLabel: genderIndex >= 0 ? GENDER_OPTIONS[genderIndex].label : '男'
      })
    } catch (error) {
      hideLoading()
      showError('加载案件信息失败')
      console.error('加载案件失败:', error)
    }
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
   * 失踪时间选择
   */
  onMissingTimeChange(e) {
    this.setData({ 
      'form.missingTime': e.detail.value 
    })
  },

  /**
   * 选择位置（使用微信小程序选择位置API）
   */
  chooseLocation() {
    wx.chooseLocation({
      success: (res) => {
        // 解析地址字符串，尝试提取省市区
        const address = res.address || res.name || ''
        const parts = this.parseAddress(address)
        
        this.setData({
          'form.province': parts.province,
          'form.city': parts.city,
          'form.district': parts.district,
          'form.address': address
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
   * 解析地址字符串为省市区
   * 这是一个简化的实现，实际可能需要更复杂的地址解析
   */
  parseAddress(address) {
    const parts = { province: '', city: '', district: '' }
    if (!address) return parts
    
    // 简单解析：假设地址格式为"XX省XX市XX区..."
    const provinceMatch = address.match(/([^省市自治区]+(?:省|市|自治区))/)
    const cityMatch = address.match(/([^市区县]+(?:市|区))/g)
    
    if (provinceMatch) {
      parts.province = provinceMatch[1]
    }
    
    // 这里使用简化的逻辑，实际可能需要调用地址解析服务
    const addrParts = address.split(/[省市区县]/)
    if (addrParts.length >= 3) {
      parts.province = addrParts[0] + (address.includes('省') ? '省' : address.includes('自治区') ? '自治区' : '市')
      parts.city = addrParts[1] + (address.includes('市') ? '市' : '区')
      if (addrParts[2]) {
        parts.district = addrParts[2] + (address.includes('县') ? '县' : '区')
      }
    } else {
      // 如果无法解析，将完整地址放入address字段
      parts.address = address
    }
    
    return parts
  },

  /**
   * 手动输入位置
   */
  onLocationInput(e) {
    const { field } = e.currentTarget.dataset
    const { value } = e.detail
    this.setData({ [`form.${field}`]: value })
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
        uploadedUrls.push(result.url || result)
        
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
    
    if (!form.province && !form.city && !form.address) {
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
      
      // 构建提交数据（按照后端API要求的字段名）
      const submitData = {
        name: form.name.trim(),
        gender: form.gender,
        age: parseInt(form.age) || 0,
        height: parseInt(form.height) || 0,
        // 走失时间
        missing_time: (() => { try { return new Date(form.missingTime).toISOString() } catch (e) { return '' } })(),
        // 位置信息（分别提交）
        province: form.province.trim(),
        city: form.city.trim(),
        district: form.district.trim(),
        address: form.address.trim(),
        // 详细描述
        description: form.description.trim(),
        // 联系人信息（注意字段名与后端一致）
        contact_name: form.contactName.trim(),
        contact_rel: form.contactRel.trim(),      // 使用 contact_rel 而非 contact_relation
        contact_phone: form.contactPhone.trim(),
        alt_contact: form.altContact.trim(),
        // 照片URL（后端只接受单个字符串，取第一张）
        photo_url: photoUrls.length > 0 ? photoUrls[0] : ''
      }

      if (this.data.isEdit) {
        await missingPersonService.update(this.data.editId, submitData)
      } else {
        await missingPersonService.create(submitData)
      }

      hideLoading()
      showSuccess(this.data.isEdit ? '更新成功' : '登记成功')
      
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
              missingTime: '',
              province: '',
              city: '',
              district: '',
              address: '',
              description: '',
              contactName: '',
              contactRel: '',
              contactPhone: '',
              altContact: ''
            },
            photos: []
          })
          this.setDefaultMissingTime()
        }
      }
    })
  }
})
