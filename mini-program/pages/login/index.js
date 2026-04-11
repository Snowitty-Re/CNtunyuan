const services = require('../../services/index')
const { validatePhone, showError, showSuccess, showLoading, hideLoading } = require('../../utils/util')

const POLICY_VERSION = '2026-04-10'
const POLICY_AGREEMENT_KEY = 'policyAgreement'

Page({
  data: {
    loading: false,
    loginType: 'wechat', // quick, phone
    agreedToPolicies: false,
    isRegister: false,
    phone: '',
    password: '',
    smsCode: '',
    counting: false,
    countDown: 60,
    canSendCode: false,
    isBinding: false, // 是否处于绑定手机号流程
    isResetPassword: false, // 是否处于重置密码流程
    tempUserInfo: null // 临时用户信息（快捷登录后未绑定手机号）
  },

  onLoad(options) {
    this.restorePolicyAgreement()

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
    this.setData({ 
      loginType: type,
      isRegister: false,
      isResetPassword: false,
      isBinding: false // 切换登录方式时退出绑定模式
    })
  },

  // 返回快捷登录（从绑定页面返回）
  backToWechatLogin() {
    this.setData({
      loginType: 'wechat',
      isBinding: false,
      isResetPassword: false,
      phone: '',
      smsCode: ''
    })
  },

  // ==================== 快捷登录 ====================

  // 快捷登录
  async handleWechatLogin() {
    if (this.data.loading) return
    if (!(await this.ensurePoliciesAgreed('wechat_login'))) return
    
    this.setData({ loading: true })
    showLoading('登录中...')

    try {
      // 获取登录凭证
      const wxLoginRes = await wx.login()
      
      if (!wxLoginRes.code) {
        throw new Error('获取登录凭证失败')
      }

      // 获取用户信息（头像、昵称）
      let userInfo = null
      try {
        const profileRes = await wx.getUserProfile({
          desc: '用于完善用户资料',
          lang: 'zh_CN'
        })
        userInfo = profileRes.userInfo
      } catch (profileErr) {
        // 用户拒绝授权时继续登录，使用后端默认资料
      }

      // 调用后端快捷登录
      const result = await services.auth.wechatLogin(wxLoginRes.code, userInfo)
      
      hideLoading()

      // 判断是否需要绑定手机号
      if (result.need_bind_phone) {
        // 保存临时 token，用于后续绑定请求
        if (result.access_token) {
          wx.setStorageSync('token', result.access_token)
          if (result.refresh_token) {
            wx.setStorageSync('refresh_token', result.refresh_token)
          }
        }

        this.setData({
          isBinding: true,
          loginType: 'phone',
          tempUserInfo: result.user
        })

        wx.showToast({ title: '请输入手机号和验证码完成绑定', icon: 'none', duration: 2500 })
        return
      }

      if (result.need_approval) {
        wx.removeStorageSync('token')
        wx.removeStorageSync('refresh_token')
        showError(result.message || '账号待管理员审批，请稍后再试')
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
      console.error('快捷登录失败:', error)
      showError(error.message || '登录失败')
    } finally {
      this.setData({ loading: false })
    }
  },

  // ==================== 手机号登录 ====================

  // 快捷绑定手机号（优先方案）
  async handleWechatPhoneBind(e) {
    if (!this.data.isBinding || this.data.loading) return
    if (!(await this.ensurePoliciesAgreed('wechat_bind_phone'))) return

    const code = e && e.detail ? e.detail.code : ''
    if (!code || code === 'getPhoneNumber:fail user deny') {
      showError('未授权手机号')
      return
    }

    this.setData({ loading: true })
    showLoading('绑定中...')

    try {
      const result = await services.auth.bindPhoneByWechatCode(code)
      hideLoading()
      if (result && result.need_approval) {
        wx.removeStorageSync('token')
        wx.removeStorageSync('refresh_token')
        showError(result.message || '账号待管理员审批，请等待审核')
        return
      }
      this.setLoginData(result)
      showSuccess('绑定成功')

      setTimeout(() => {
        wx.switchTab({ url: '/pages/index/index' })
      }, 1200)
    } catch (error) {
      hideLoading()
      console.error('手机号快捷绑定失败:', error)
      showError(error.message || '绑定失败')
    } finally {
      this.setData({ loading: false })
    }
  },

  // 注册/重置密码：手机号快捷验证
  async handleWechatPhoneForAuth(e) {
    const { isRegister, isResetPassword, password, loading } = this.data
    if (loading || (!isRegister && !isResetPassword)) return
    if (!(await this.ensurePoliciesAgreed(isRegister ? 'wechat_register' : 'wechat_reset_password'))) return

    if (!password) {
      showError(isResetPassword ? '请输入新密码' : '请输入密码')
      return
    }
    if (password.length < 8) {
      showError('密码至少8位')
      return
    }

    const code = e && e.detail ? e.detail.code : ''
    if (!code || code === 'getPhoneNumber:fail user deny') {
      showError('未授权手机号')
      return
    }

    this.setData({ loading: true })
    showLoading(isResetPassword ? '重置中...' : '注册中...')

    try {
      if (isResetPassword) {
        await services.auth.resetPassword('', '', password, code)
        hideLoading()
        showSuccess('密码重置成功，请使用新密码登录')
        this.setData({
          isResetPassword: false,
          isRegister: false,
          loginType: 'phone',
          password: '',
          smsCode: '',
          phone: ''
        })
        return
      }

      const result = await services.auth.register('', password, '', '', code)
      hideLoading()
      if (result && result.need_approval) {
        showSuccess(result.message || '注册成功，待管理员审批')
      } else {
        showSuccess('注册成功')
      }
      this.setData({
        isRegister: false,
        loginType: 'phone',
        password: '',
        smsCode: '',
        phone: ''
      })
    } catch (error) {
      hideLoading()
      console.error('手机号快捷验证失败:', error)
      showError(error.message || '操作失败')
    } finally {
      this.setData({ loading: false })
    }
  },

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
    const { phone, counting, isBinding, isResetPassword } = this.data
    
    if (counting) return
    if (!validatePhone(phone)) {
      showError('请输入正确的手机号')
      return
    }

    this.setData({ loading: true })
    showLoading('发送中...')

    try {
      const scene = isBinding ? 'change_phone' : (isResetPassword ? 'reset_password' : 'verify')
      await services.auth.sendVerifyCode(phone, scene)
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

    this._countDownTimer = setInterval(() => {
      let countDown = this.data.countDown - 1

      if (countDown <= 0) {
        clearInterval(this._countDownTimer)
        this._countDownTimer = null
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

  onUnload() {
    if (this._countDownTimer) {
      clearInterval(this._countDownTimer)
      this._countDownTimer = null
    }
  },

  restorePolicyAgreement() {
    try {
      const saved = wx.getStorageSync(POLICY_AGREEMENT_KEY)
      const agreed = !!(saved && saved.version === POLICY_VERSION && saved.agreed === true)
      this.setData({ agreedToPolicies: agreed })
    } catch (e) {
      this.setData({ agreedToPolicies: false })
    }
  },

  // 手机号登录
  async handlePhoneLogin() {
    const { phone, password, isBinding, isRegister, isResetPassword } = this.data
    if (isBinding || isRegister || isResetPassword) {
      showError('请使用手机号快捷验证按钮完成验证')
      return
    }
    if (!(await this.ensurePoliciesAgreed('phone_login'))) return

    if (!validatePhone(phone)) {
      showError('请输入正确的手机号')
      return
    }

    // 普通登录/注册/重置需要密码
    if (!isBinding && !password) {
      showError('请输入密码')
      return
    }

    this.setData({ loading: true })
    showLoading('登录中...')

    try {
      let result

      result = await services.auth.login(phone, password)

      hideLoading()

      if (result && result.need_approval) {
        showSuccess(result.message || '注册成功，待管理员审批')
        this.setData({
          isRegister: false,
          loginType: 'phone',
          password: '',
          smsCode: ''
        })
        return
      }

      this.setLoginData(result)
      showSuccess(isBinding ? '绑定成功' : (isRegister ? '注册成功' : '登录成功'))

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

  // 注册入口
  goToRegister() {
    this.setData({
      loginType: 'phone',
      isBinding: false,
      isResetPassword: false,
      isRegister: true,
      smsCode: ''
    })
  },

  // 忘记密码
  goToForgot() {
    this.setData({
      loginType: 'phone',
      isBinding: false,
      isRegister: false,
      isResetPassword: true,
      smsCode: '',
      password: ''
    })
  },

  ensurePoliciesAgreed(action = '') {
    if (this.data.agreedToPolicies) return Promise.resolve(true)
    showError('请先阅读并勾选同意《用户协议》与《隐私政策》')
    return Promise.resolve(false)
  },

  confirmPolicyAgreement() {
    const payload = {
      agreed: true,
      version: POLICY_VERSION,
      agreedAt: Date.now()
    }
    wx.setStorageSync(POLICY_AGREEMENT_KEY, payload)
    this.setData({
      agreedToPolicies: true
    })
  },

  togglePoliciesAgreement() {
    if (this.data.agreedToPolicies) {
      wx.removeStorageSync(POLICY_AGREEMENT_KEY)
      this.setData({ agreedToPolicies: false })
      return
    }
    this.confirmPolicyAgreement()
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
