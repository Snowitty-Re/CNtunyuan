const services = require('../../services')
const { validatePhone, showError, showSuccess, showLoading, hideLoading } = require('../../utils/util')

Page({
  data: {
    loading: false,
    loginType: 'wechat', // wechat, phone
    phone: '',
    password: '',
    smsCode: '',
    counting: false,
    countDown: 60,
    canSendCode: false,
    isBinding: false, // 是否处于绑定手机号流程
    tempUserInfo: null // 临时用户信息（微信登录后未绑定手机号）
  },

  onLoad(options) {
    // 检查是否已登录
    const token = wx.getStorageSync('token')
    if (token) {
      wx.switchTab({ url: '/pages/index/index' })
      return
    }

    // 从跳转参数获取信息
    if (options.binding === '1') {
      this.setData({ 
        isBinding: true,
        loginType: 'phone'
      })
    }
  },

  // 切换登录方式
  switchLoginType(e) {
    const type = e.currentTarget.dataset.type
    this.setData({ loginType: type })
  },

  // ==================== 微信登录 ====================

  // 微信一键登录
  async handleWechatLogin() {
    if (this.data.loading) return
    
    this.setData({ loading: true })
    showLoading('登录中...')

    try {
      // 获取微信登录码
      const wxLoginRes = await wx.login()
      
      if (!wxLoginRes.code) {
        throw new Error('获取微信登录码失败')
      }

      // 调用后端微信登录
      const result = await services.auth.wechatLogin(wxLoginRes.code)
      
      hideLoading()

      // 判断是否需要绑定手机号
      if (result.need_bind_phone) {
        this.setData({
          isBinding: true,
          loginType: 'phone',
          tempUserInfo: result.user
        })
        wx.showModal({
          title: '绑定手机号',
          content: '为了您的账号安全，请先绑定手机号',
          showCancel: false
        })
        return
      }

      // 保存登录信息
      this.setLoginData(result)
      showSuccess('登录成功')
      
      // 延迟跳转
      setTimeout(() => {
        wx.switchTab({ url: '/pages/index/index' })
      }, 1500)

    } catch (error) {
      hideLoading()
      console.error('微信登录失败:', error)
      showError(error.message || '登录失败')
    } finally {
      this.setData({ loading: false })
    }
  },

  // ==================== 手机号登录 ====================

  // 手机号输入
  onPhoneInput(e) {
    const phone = e.detail.value
    this.setData({ 
      phone,
      canSendCode: validatePhone(phone) && !this.data.counting
    })
  },

  // 密码输入
  onPasswordInput(e) {
    this.setData({ password: e.detail.value })
  },

  // 验证码输入
  onCodeInput(e) {
    this.setData({ smsCode: e.detail.value })
  },

  // 发送验证码
  async sendVerifyCode() {
    const { phone, counting } = this.data
    
    if (counting) return
    if (!validatePhone(phone)) {
      showError('请输入正确的手机号')
      return
    }

    this.setData({ loading: true })
    showLoading('发送中...')

    try {
      await services.auth.sendVerifyCode(phone)
      hideLoading()
      showSuccess('验证码已发送')
      
      // 开始倒计时
      this.startCountDown()
    } catch (error) {
      hideLoading()
      showError(error.message || '发送失败')
    } finally {
      this.setData({ loading: false })
    }
  },

  // 开始倒计时
  startCountDown() {
    this.setData({ 
      counting: true,
      canSendCode: false,
      countDown: 60 
    })

    const timer = setInterval(() => {
      let countDown = this.data.countDown - 1
      
      if (countDown <= 0) {
        clearInterval(timer)
        this.setData({
          counting: false,
          canSendCode: validatePhone(this.data.phone),
          countDown: 60
        })
      } else {
        this.setData({ countDown })
      }
    }, 1000)
  },

  // 手机号登录
  async handlePhoneLogin() {
    const { phone, password, isBinding, smsCode } = this.data
    
    if (!validatePhone(phone)) {
      showError('请输入正确的手机号')
      return
    }

    // 绑定手机号需要验证码
    if (isBinding && !smsCode) {
      showError('请输入验证码')
      return
    }

    // 普通登录需要密码
    if (!isBinding && !password) {
      showError('请输入密码')
      return
    }

    this.setData({ loading: true })
    showLoading('登录中...')

    try {
      let result

      if (isBinding) {
        // 绑定手机号
        result = await services.auth.bindPhone(phone, smsCode)
      } else {
        // 手机号密码登录
        result = await services.auth.login(phone, password)
      }

      hideLoading()
      this.setLoginData(result)
      showSuccess(isBinding ? '绑定成功' : '登录成功')

      setTimeout(() => {
        wx.switchTab({ url: '/pages/index/index' })
      }, 1500)

    } catch (error) {
      hideLoading()
      console.error('登录失败:', error)
      showError(error.message || '登录失败')
    } finally {
      this.setData({ loading: false })
    }
  },

  // ==================== 通用方法 ====================

  // 设置登录数据
  setLoginData(data) {
    const app = getApp()
    if (app && app.setLoginData) {
      app.setLoginData(data)
    } else {
      // 备用方案
      const { access_token, refresh_token, user } = data
      if (access_token) {
        wx.setStorageSync('token', access_token)
      }
      if (refresh_token) {
        wx.setStorageSync('refresh_token', refresh_token)
      }
      if (user) {
        wx.setStorageSync('userInfo', user)
      }
    }
  },

  // 用户协议
  goToAgreement() {
    wx.navigateTo({
      url: '/pages/settings/agreement'
    })
  },

  // 隐私政策
  goToPrivacy() {
    wx.navigateTo({
      url: '/pages/settings/privacy'
    })
  }
})
