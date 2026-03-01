const { get, put } = require('../../utils/request')
const { formatDate, showSuccess, showToast } = require('../../utils/util')

Page({
  data: {
    form: {
      nickname: '',
      real_name: '',
      phone: '',
      email: '',
      gender: '',
      birth_date: '',
      address: '',
      emergency_contact: '',
      emergency_phone: '',
      skills: '',
      experience: ''
    },
    genderOptions: [
      { value: 'male', label: '男' },
      { value: 'female', label: '女' },
      { value: 'other', label: '其他' }
    ],
    genderIndex: -1,
    loading: false
  },

  onLoad() {
    this.loadUserInfo()
  },

  async loadUserInfo() {
    try {
      const userInfo = wx.getStorageSync('userInfo') || {}
      const profile = await get(`/users/${userInfo.id}`).catch(() => ({}))
      
      const form = {
        nickname: profile.nickname || userInfo.nickname || '',
        real_name: profile.real_name || userInfo.real_name || '',
        phone: profile.phone || userInfo.phone || '',
        email: profile.email || userInfo.email || '',
        gender: profile.gender || '',
        birth_date: profile.birth_date ? formatDate(profile.birth_date) : '',
        address: profile.address || '',
        emergency_contact: profile.emergency_contact || '',
        emergency_phone: profile.emergency_phone || '',
        skills: profile.skills || '',
        experience: profile.experience || ''
      }
      
      // 设置性别选择器索引
      const genderIndex = this.data.genderOptions.findIndex(
        item => item.value === form.gender
      )
      
      this.setData({ form, genderIndex })
    } catch (error) {
      console.error('加载用户信息失败:', error)
    }
  },

  onInput(e) {
    const { field } = e.currentTarget.dataset
    this.setData({ [`form.${field}`]: e.detail.value })
  },

  onGenderChange(e) {
    const index = e.detail.value
    const gender = this.data.genderOptions[index].value
    this.setData({ 
      genderIndex: index,
      'form.gender': gender
    })
  },

  onDateChange(e) {
    this.setData({ 'form.birth_date': e.detail.value })
  },

  // 保存资料
  async saveProfile() {
    const { form } = this.data
    
    if (!form.nickname.trim()) {
      showToast('请输入昵称')
      return
    }
    
    if (!form.real_name.trim()) {
      showToast('请输入真实姓名')
      return
    }
    
    if (!form.phone.trim()) {
      showToast('请输入手机号')
      return
    }
    
    this.setData({ loading: true })
    
    try {
      const userInfo = wx.getStorageSync('userInfo')
      await put(`/users/${userInfo.id}`, form)
      
      // 更新本地存储
      const updatedUserInfo = { ...userInfo, ...form }
      wx.setStorageSync('userInfo', updatedUserInfo)
      
      showSuccess('保存成功')
      setTimeout(() => {
        wx.navigateBack()
      }, 1500)
    } catch (error) {
      console.error('保存失败:', error)
      showToast('保存失败')
    } finally {
      this.setData({ loading: false })
    }
  },

  // 选择位置
  chooseLocation() {
    wx.chooseLocation({
      success: (res) => {
        this.setData({ 'form.address': res.address })
      }
    })
  }
})
