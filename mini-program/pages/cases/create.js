const missingPersonService = require('../../services/missingPerson')
const uploadService = require('../../services/upload')
const { showLoading, hideLoading, showSuccess, showError, validatePhone, formatDate } = require('../../utils/util')
const app = getApp()
const CASES_LIST_DIRTY_KEY = 'cases_list_dirty'

// 性别选项
const GENDER_OPTIONS = [
  { value: 'male', label: '男' },
  { value: 'female', label: '女' },
  { value: 'other', label: '其他' }
]

// 案件类型选项（前端分类展示用）
const CASE_TYPE_OPTIONS = [
  { value: 'elderly', label: '老人' },
  { value: 'child', label: '儿童' },
  { value: 'adult', label: '成人' },
  { value: 'disability', label: '残障' },
  { value: 'other', label: '其他' }
]

Page({
  data: {
    // 表单数据
    form: {
      name: '',
      gender: 'male',
      caseType: 'other',
      age: '',
      height: '',
      missingTime: '',
      // 位置信息（前端展示用）
      province: '',
      city: '',
      district: '',
      address: '',
      missingLatitude: '',
      missingLongitude: '',
      // 详细描述（包含外貌、衣着、特征等）
      description: '',
      // 外貌特征
      appearance: '',
      clothing: '',
      specialFeatures: '',
      // 联系人信息
      contactName: '',
      contactRel: '',
      contactPhone: '',
      altContact: ''
    },
    
    // 选项数据
    genderOptions: GENDER_OPTIONS,
    genderLabel: '男',
    caseTypeOptions: CASE_TYPE_OPTIONS,
    caseTypeLabel: '其他',
    caseTypeIndex: 4,
    
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
    if (!app.ensureAuth || !app.ensureAuth()) return
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

      const caseTypeValue = data.case_type || 'other'
      const caseTypeIndex = CASE_TYPE_OPTIONS.findIndex(c => c.value === caseTypeValue)
      this.setData({
        form: {
          name: data.name || '',
          gender: data.gender || 'male',
          caseType: caseTypeValue,
          age: data.age ? String(data.age) : '',
          height: data.height ? String(data.height) : '',
          missingTime: data.missing_time ? formatDate(data.missing_time, 'YYYY-MM-DD') : '',
          // 位置信息
          province: data.province || '',
          city: data.city || '',
          district: data.district || '',
          address: data.address || '',
          missingLatitude: data.missing_latitude || '',
          missingLongitude: data.missing_longitude || '',
          // 详细描述（后端存储的是合并后的描述）
          description: data.description || '',
          appearance: '',
          clothing: data.clothes || '',
          specialFeatures: data.features || '',
          // 联系人信息（注意字段名映射）
          contactName: data.contact_name || '',
          contactRel: data.contact_rel || '',  // 后端返回 contact_rel
          contactPhone: data.contact_phone || '',
          altContact: data.alt_contact || ''
        },
        // 如果有照片，设置到photos中
        uploadedPhotos: data.photos || [],
        genderLabel: genderIndex >= 0 ? GENDER_OPTIONS[genderIndex].label : '男',
        caseTypeLabel: (CASE_TYPE_OPTIONS.find(c => c.value === caseTypeValue) || CASE_TYPE_OPTIONS[4]).label,
        caseTypeIndex: caseTypeIndex >= 0 ? caseTypeIndex : 4
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
    
    this.setData({
      'form.missingTime': `${year}-${month}-${day}`
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
   * 案件类型选择
   */
  onCaseTypeChange(e) {
    const index = parseInt(e.detail.value, 10)
    const item = CASE_TYPE_OPTIONS[index]
    if (!item) return
    this.setData({
      'form.caseType': item.value,
      caseTypeLabel: item.label,
      caseTypeIndex: index
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
   * 选择位置（使用小程序选择位置 API）
   */
  chooseLocation() {
    wx.chooseLocation({
      success: (res) => {
        // 解析地址字符串，尝试提取省市区
        const address = res.address
          ? (res.name && !res.address.includes(res.name) ? `${res.address}${res.name}` : res.address)
          : (res.name || '')
        const parts = this.parseAddress(address)
        
        this.setData({
          'form.province': parts.province,
          'form.city': parts.city,
          'form.district': parts.district,
          'form.address': address,
          'form.missingLatitude': res.latitude || '',
          'form.missingLongitude': res.longitude || ''
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

    const latRaw = (form.missingLatitude === undefined || form.missingLatitude === null)
      ? ''
      : String(form.missingLatitude).trim()
    const lngRaw = (form.missingLongitude === undefined || form.missingLongitude === null)
      ? ''
      : String(form.missingLongitude).trim()
    const hasLat = latRaw !== ''
    const hasLng = lngRaw !== ''

    if (hasLat !== hasLng) {
      showError('经纬度请成对填写')
      return false
    }

    if (hasLat && hasLng) {
      const lat = Number(latRaw)
      const lng = Number(lngRaw)
      if (Number.isNaN(lat) || Number.isNaN(lng)) {
        showError('经纬度格式无效')
        return false
      }
      if (lat < -90 || lat > 90 || lng < -180 || lng > 180) {
        showError('经纬度超出范围')
        return false
      }
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
      const latRaw = (form.missingLatitude === undefined || form.missingLatitude === null)
        ? ''
        : String(form.missingLatitude).trim()
      const lngRaw = (form.missingLongitude === undefined || form.missingLongitude === null)
        ? ''
        : String(form.missingLongitude).trim()
      const hasCoordinate = latRaw !== '' && lngRaw !== ''
      
      // 构建提交数据（按照后端API要求的字段名）
      const submitData = {
        name: form.name.trim(),
        gender: form.gender,
        case_type: form.caseType || 'other',
        age: parseInt(form.age) || 0,
        height: parseInt(form.height) || 0,
        // 走失时间
        missing_time: (() => {
          const d = this.parseLocalDateTime(form.missingTime)
          if (!d) {
            throw new Error('走失时间格式无效，请重新选择')
          }
          return d.toISOString()
        })(),
        // 位置信息（分别提交）
        province: form.province.trim(),
        city: form.city.trim(),
        district: form.district.trim(),
        address: form.address.trim(),
        missing_latitude: hasCoordinate ? Number(latRaw) : null,
        missing_longitude: hasCoordinate ? Number(lngRaw) : null,
        // 详细描述
        description: form.description.trim(),
        // 外貌特征（对齐后端字段）
        clothes: form.clothing.trim(),
        features: (() => {
          const appearance = (form.appearance || '').trim()
          const specialFeatures = (form.specialFeatures || '').trim()
          if (appearance && specialFeatures) {
            return `体貌特征：${appearance}\n特殊特征：${specialFeatures}`
          }
          return appearance || specialFeatures
        })(),
        // 联系人信息（注意字段名与后端一致）
        contact_name: form.contactName.trim(),
        contact_rel: form.contactRel.trim(),      // 使用 contact_rel 而非 contact_relation
        contact_phone: form.contactPhone.trim(),
        alt_contact: form.altContact.trim(),
        // 照片URL（后端只接受单个字符串，取第一张）
        photo_url: photoUrls.length > 0 ? photoUrls[0] : '',
        // 紧急程度（与后端字段名保持一致）
        urgency_level: form.urgencyLevel || 'medium'
      }

      if (this.data.isEdit) {
        await missingPersonService.update(this.data.editId, submitData)
      } else {
        await missingPersonService.create(submitData)
      }
      wx.setStorageSync(CASES_LIST_DIRTY_KEY, 1)

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
   * 解析本地日期时间字符串（兼容 iOS）
   * 支持: YYYY-MM-DD / YYYY-MM-DD HH:mm / YYYY-MM-DDTHH:mm / YYYY-MM-DD HH:mm:ss / YYYY-MM-DDTHH:mm:ss
   */
  parseLocalDateTime(value) {
    if (!value || typeof value !== 'string') return null
    const matched = value.trim().match(/^(\d{4})-(\d{2})-(\d{2})(?:[ T](\d{2}):(\d{2})(?::(\d{2}))?)?$/)
    if (!matched) return null

    const year = Number(matched[1])
    const month = Number(matched[2])
    const day = Number(matched[3])
    const hour = matched[4] ? Number(matched[4]) : 0
    const minute = matched[5] ? Number(matched[5]) : 0
    const second = matched[6] ? Number(matched[6]) : 0

    const d = new Date(year, month - 1, day, hour, minute, second)
    if (isNaN(d.getTime())) return null
    return d
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
              caseType: 'other',
              age: '',
              height: '',
              missingTime: '',
              province: '',
              city: '',
              district: '',
              address: '',
              missingLatitude: '',
              missingLongitude: '',
              description: '',
              appearance: '',
              clothing: '',
              specialFeatures: '',
              contactName: '',
              contactRel: '',
              contactPhone: '',
              altContact: ''
            },
            photos: [],
            caseTypeLabel: '其他',
            caseTypeIndex: 4
          })
          this.setDefaultMissingTime()
        }
      }
    })
  }
})
