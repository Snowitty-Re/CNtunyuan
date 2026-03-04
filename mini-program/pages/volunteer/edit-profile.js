const userService = require('../../services/user')
const uploadService = require('../../services/upload')
const { showSuccess, showToast, showLoading, hideLoading, validatePhone, validateEmail } = require('../../utils/util')
const app = getApp()

Page({
  data: {
    // 表单数据
    form: {
      avatar: '',
      nickname: '',
      realName: '',
      phone: '',
      email: ''
    },
    
    // 原始数据（用于比较变更）
    originalForm: {},
    
    // 密码表单
    passwordForm: {
      oldPassword: '',
      newPassword: '',
      confirmPassword: ''
    },
    
    // 验证码相关
    phoneCode: '',
    emailCode: '',
    countdown: 0,
    codeType: '', // 'phone' 或 'email'
    
    // 页面状态
    loading: false,
    uploadLoading: false,
    showPasswordModal: false
  },

  onLoad() {
    this.loadUserInfo()
  },

  // 加载用户信息
  async loadUserInfo() {
    try {
      showLoading('加载中...')
      const userInfo = await app.getUserInfo() || wx.getStorageSync('userInfo') || {}
      const profile = await userService.getProfile().catch(() => ({}))
      
      const form = {
        avatar: userInfo.avatar || profile.avatar || '/assets/images/avatar-default.png',
        nickname: userInfo.nickname || profile.nickname || '',
        realName: userInfo.real_name || profile.real_name || '',
        phone: userInfo.phone || profile.phone || '',
        email: userInfo.email || profile.email || ''
      }
      
      this.setData({
        form: { ...form },
        originalForm: { ...form }
      })
    } catch (error) {
      console.error('加载用户信息失败:', error)
      showToast('加载失败')
    } finally {
      hideLoading()
    }
  },

  // 选择头像
  async chooseAvatar() {
    try {
      const res = await wx.chooseMedia({
        count: 1,
        mediaType: ['image'],
        sourceType: ['album', 'camera'],
        sizeType: ['compressed']
      })
      
      const tempFilePath = res.tempFiles[0].tempFilePath
      
      this.setData({ uploadLoading: true })
      
      // 上传头像
      const uploadRes = await uploadService.upload(tempFilePath, { type: 'avatar' })
      
      this.setData({
        'form.avatar': uploadRes.url || uploadRes.data?.url || tempFilePath,
        uploadLoading: false
      })
      
      showSuccess('头像上传成功')
    } catch (error) {
      console.error('上传头像失败:', error)
      showToast('上传失败')
      this.setData({ uploadLoading: false })
    }
  },

  // 输入框变化
  onInput(e) {
    const { field } = e.currentTarget.dataset
    const { value } = e.detail
    this.setData({ [`form.${field}`]: value })
  },

  // 密码输入框变化
  onPasswordInput(e) {
    const { field } = e.currentTarget.dataset
    const { value } = e.detail
    this.setData({ [`passwordForm.${field}`]: value })
  },

  // 验证码输入
  onCodeInput(e) {
    const { type } = e.currentTarget.dataset
    const { value } = e.detail
    this.setData({ [`${type}Code`]: value })
  },

  // 发送验证码
  async sendCode(e) {
    const { type } = e.currentTarget.dataset
    
    if (this.data.countdown > 0) return
    
    let target = ''
    if (type === 'phone') {
      target = this.data.form.phone
      if (!validatePhone(target)) {
        showToast('请输入正确的手机号')
        return
      }
    } else if (type === 'email') {
      target = this.data.form.email
      if (!validateEmail(target)) {
        showToast('请输入正确的邮箱')
        return
      }
    }
    
    try {
      showLoading('发送中...')
      // 调用发送验证码接口
      await userService.sendVerifyCode?.(target, type).catch(() => {
        // 模拟发送成功
        showSuccess('验证码已发送')
      })
      
      this.setData({ 
        countdown: 60,
        codeType: type
      })
      
      // 开始倒计时
      this.startCountdown()
    } catch (error) {
      showToast('发送失败')
    } finally {
      hideLoading()
    }
  },

  // 倒计时
  startCountdown() {
    const timer = setInterval(() => {
      if (this.data.countdown <= 1) {
        clearInterval(timer)
        this.setData({ countdown: 0 })
      } else {
        this.setData({ countdown: this.data.countdown - 1 })
      }
    }, 1000)
  },

  // 保存资料
  async saveProfile() {
    const { form, originalForm, phoneCode, emailCode } = this.data
    
    // 表单验证
    if (!form.nickname.trim()) {
      showToast('请输入昵称')
      return
    }
    
    if (!form.realName.trim()) {
      showToast('请输入真实姓名')
      return
    }
    
    if (form.phone && !validatePhone(form.phone)) {
      showToast('请输入正确的手机号')
      return
    }
    
    if (form.email && !validateEmail(form.email)) {
      showToast('请输入正确的邮箱')
      return
    }
    
    // 检查手机号是否变更
    if (form.phone !== originalForm.phone && !phoneCode) {
      showToast('请验证新手机号')
      return
    }
    
    // 检查邮箱是否变更
    if (form.email !== originalForm.email && !emailCode) {
      showToast('请验证新邮箱')
      return
    }
    
    this.setData({ loading: true })
    
    try {
      // 构建提交数据
      const submitData = {
        avatar: form.avatar,
        nickname: form.nickname,
        real_name: form.realName,
        phone: form.phone,
        email: form.email
      }
      
      // 如果有验证码，添加到提交数据
      if (form.phone !== originalForm.phone) {
        submitData.phone_code = phoneCode
      }
      if (form.email !== originalForm.email) {
        submitData.email_code = emailCode
      }
      
      // 调用更新接口
      await userService.updateProfile(submitData)
      
      // 更新本地存储
      const userInfo = wx.getStorageSync('userInfo') || {}
      const updatedUserInfo = { 
        ...userInfo, 
        ...submitData,
        real_name: form.realName
      }
      wx.setStorageSync('userInfo', updatedUserInfo)
      
      // 更新全局数据
      if (app.globalData) {
        app.globalData.userInfo = updatedUserInfo
      }
      
      showSuccess('保存成功')
      
      setTimeout(() => {
        wx.navigateBack()
      }, 1500)
    } catch (error) {
      console.error('保存失败:', error)
      showToast(error.message || '保存失败')
    } finally {
      this.setData({ loading: false })
    }
  },

  // 显示修改密码弹窗
  showPasswordModal() {
    this.setData({
      showPasswordModal: true,
      passwordForm: {
        oldPassword: '',
        newPassword: '',
        confirmPassword: ''
      }
    })
  },

  // 关闭修改密码弹窗
  closePasswordModal() {
    this.setData({ showPasswordModal: false })
  },

  // 保存密码
  async savePassword() {
    const { oldPassword, newPassword, confirmPassword } = this.data.passwordForm
    
    if (!oldPassword) {
      showToast('请输入原密码')
      return
    }
    
    if (!newPassword || newPassword.length < 6) {
      showToast('新密码至少6位')
      return
    }
    
    if (newPassword !== confirmPassword) {
      showToast('两次密码不一致')
      return
    }
    
    this.setData({ loading: true })
    
    try {
      await userService.changePassword(oldPassword, newPassword)
      showSuccess('密码修改成功')
      this.closePasswordModal()
    } catch (error) {
      console.error('修改密码失败:', error)
      showToast(error.message || '修改密码失败')
    } finally {
      this.setData({ loading: false })
    }
  },

  // 阻止冒泡
  stopPropagation() {}
})
