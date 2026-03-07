/**
 * 团圆寻亲小程序 - 主应用文件
 * 
 * 环境配置说明：
 * 1. 修改 API_CONFIG 中的 production 配置为你的实际域名
 * 2. 修改 initEnvironment 中的 FORCE_ENV 来切换环境
 * 3. 每次修改后需要重新上传代码才能生效
 * 
 * 常见问题：
 * Q: 为什么修改配置后还是请求开发环境？
 * A: 需要重新上传代码，且清除小程序缓存（开发者工具 → 详情 → 清除缓存）
 * 
 * Q: 体验版如何切换到生产环境？
 * A: 将 FORCE_ENV 设为 'production' 并重新上传
 */

const { request, uploadFile } = require('./utils/request')

// API 基础配置
// 【重要】生产环境域名必须正确配置！
const API_CONFIG = {
  development: {
    baseUrl: 'http://localhost:8080/api/v1',
    wsUrl: 'ws://localhost:8080/ws'
  },
  production: {
    baseUrl: 'https://cntuanyuan.com/api/v1',
    wsUrl: 'wss://cntuanyuan.com/ws'
  }
}

App({
  globalData: {
    userInfo: null,
    token: null,
    refreshToken: null,
    systemInfo: null,
    apiBaseUrl: '',
    wsUrl: '',
    isLogin: false
  },

  onLaunch() {
    // 初始化环境
    this.initEnvironment()
    
    // 获取系统信息
    this.getSystemInfo()
    
    // 检查登录状态
    this.checkLoginStatus()
  },

  // 初始化环境配置
  initEnvironment() {
    const accountInfo = wx.getAccountInfoSync()
    
    // envVersion: develop(开发版), trial(体验版), release(正式版)
    const envVersion = accountInfo.miniProgram.envVersion
    
    // 强制使用生产环境（临时方案，方便测试）
    // 如需切换环境，修改此行：
    // 'production' - 生产环境
    // 'development' - 开发环境
    // 'auto' - 自动根据 envVersion 判断
    const FORCE_ENV = 'production'  // <-- 修改这里切换环境
    
    let env
    if (FORCE_ENV === 'auto') {
      // 自动判断：release 为生产环境，其他为开发环境
      env = envVersion === 'release' ? 'production' : 'development'
    } else {
      env = FORCE_ENV
    }
    
    this.globalData.apiBaseUrl = API_CONFIG[env].baseUrl
    this.globalData.wsUrl = API_CONFIG[env].wsUrl
    
    console.log(`[App] 当前环境: ${env}, envVersion: ${envVersion}, API: ${this.globalData.apiBaseUrl}`)
  },

  // 获取系统信息
  getSystemInfo() {
    try {
      // 使用新的 API 替代废弃的 wx.getSystemInfoSync
      const windowInfo = wx.getWindowInfo ? wx.getWindowInfo() : {}
      const deviceInfo = wx.getDeviceInfo ? wx.getDeviceInfo() : {}
      const appBaseInfo = wx.getAppBaseInfo ? wx.getAppBaseInfo() : {}
      
      // 合并信息保持兼容性
      const systemInfo = {
        ...windowInfo,
        ...deviceInfo,
        ...appBaseInfo
      }
      
      this.globalData.systemInfo = systemInfo
      
      // 设置导航栏适配
      if (windowInfo.statusBarHeight) {
        this.globalData.statusBarHeight = windowInfo.statusBarHeight
        this.globalData.navBarHeight = windowInfo.statusBarHeight + 44
      }
    } catch (e) {
      console.error('[App] 获取系统信息失败:', e)
    }
  },

  // 检查登录状态
  checkLoginStatus() {
    const token = wx.getStorageSync('token')
    const refreshToken = wx.getStorageSync('refresh_token')
    
    if (token && refreshToken) {
      this.globalData.token = token
      this.globalData.refreshToken = refreshToken
      this.globalData.isLogin = true
      
      // 获取用户信息
      this.getUserInfo().catch(() => {
        // 获取失败，清除登录状态
        this.clearLoginData()
      })
    }
  },

  // ==================== 认证相关 ====================

  // 微信登录
  wxLogin() {
    return new Promise((resolve, reject) => {
      wx.login({
        success: (res) => {
          if (res.code) {
            // 调用后端微信登录接口
            request({
              url: '/auth/wechat-login',
              method: 'POST',
              data: { code: res.code }
            }).then((result) => {
              this.setLoginData(result)
              resolve(result)
            }).catch(reject)
          } else {
            reject(new Error(res.errMsg || '微信登录失败'))
          }
        },
        fail: reject
      })
    })
  },

  // 手机号密码登录
  passwordLogin(phone, password) {
    return request({
      url: '/auth/login',
      method: 'POST',
      data: { 
        username: phone,
        password: password 
      }
    }).then((result) => {
      this.setLoginData(result)
      return result
    })
  },

  // 绑定手机号
  bindPhone(phone, code) {
    return request({
      url: '/auth/bind-phone',
      method: 'POST',
      data: { phone, code }
    })
  },

  // 退出登录
  logout() {
    return request({
      url: '/auth/logout',
      method: 'POST'
    }).finally(() => {
      this.clearLoginData()
    })
  },

  // 设置登录数据
  setLoginData(data) {
    const { access_token, refresh_token, user } = data
    
    if (access_token) {
      wx.setStorageSync('token', access_token)
      this.globalData.token = access_token
    }
    
    if (refresh_token) {
      wx.setStorageSync('refresh_token', refresh_token)
      this.globalData.refreshToken = refresh_token
    }
    
    if (user) {
      wx.setStorageSync('userInfo', user)
      this.globalData.userInfo = user
    }
    
    this.globalData.isLogin = true
  },

  // 清除登录数据
  clearLoginData() {
    wx.removeStorageSync('token')
    wx.removeStorageSync('refresh_token')
    wx.removeStorageSync('userInfo')
    
    this.globalData.token = null
    this.globalData.refreshToken = null
    this.globalData.userInfo = null
    this.globalData.isLogin = false
  },

  // 获取当前用户信息
  getUserInfo() {
    return request({
      url: '/auth/me',
      method: 'GET'
    }).then((user) => {
      this.globalData.userInfo = user
      wx.setStorageSync('userInfo', user)
      return user
    })
  },

  // 刷新用户信息
  refreshUserInfo() {
    return this.getUserInfo()
  },

  // ==================== 全局请求方法 ====================

  // 封装请求（兼容旧代码）
  request(options) {
    return request(options)
  },

  // 上传文件
  upload(options) {
    return uploadFile(options.url, options.filePath, options.name || 'file', options.formData)
  },

  // ==================== 全局工具方法 ====================

  // 显示加载提示
  showLoading(title = '加载中...') {
    wx.showLoading({ title, mask: true })
  },

  // 隐藏加载提示
  hideLoading() {
    wx.hideLoading()
  },

  // 显示成功提示
  showSuccess(title = '操作成功') {
    wx.showToast({ title, icon: 'success' })
  },

  // 显示错误提示
  showError(title = '操作失败') {
    wx.showToast({ title, icon: 'error' })
  },

  // 显示普通提示
  showToast(title, icon = 'none') {
    wx.showToast({ title, icon })
  },

  // 显示确认对话框
  showModal(title, content) {
    return new Promise((resolve) => {
      wx.showModal({
        title,
        content,
        success: (res) => resolve(res.confirm)
      })
    })
  },

  // 页面跳转 - 保留当前页面
  navigateTo(url) {
    wx.navigateTo({ url })
  },

  // 页面跳转 - 关闭当前页面
  redirectTo(url) {
    wx.redirectTo({ url })
  },

  // 页面跳转 - 跳转到 tabBar 页面
  switchTab(url) {
    wx.switchTab({ url })
  },

  // 页面跳转 - 关闭所有页面
  reLaunch(url) {
    wx.reLaunch({ url })
  },

  // 返回上一页
  navigateBack(delta = 1) {
    wx.navigateBack({ delta })
  },

  // ==================== 权限检查 ====================

  // 检查是否已登录
  checkAuth() {
    if (!this.globalData.isLogin) {
      wx.navigateTo({ url: '/pages/login/index' })
      return false
    }
    return true
  },

  // 检查角色权限
  checkRole(roles) {
    const userInfo = this.globalData.userInfo
    if (!userInfo) return false
    
    if (Array.isArray(roles)) {
      return roles.includes(userInfo.role)
    }
    return userInfo.role === roles
  },

  // 是否为管理员
  isAdmin() {
    return this.checkRole(['super_admin', 'admin'])
  },

  // 是否为管理者及以上
  isManager() {
    return this.checkRole(['super_admin', 'admin', 'manager'])
  }
})
