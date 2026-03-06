const { get, post } = require('../utils/request')

/**
 * 认证相关服务
 */
module.exports = {
  /**
   * 微信登录
   * @param {String} code 微信登录码
   * @param {Object} userInfo 用户信息（可选）
   */
  wechatLogin(code, userInfo = null) {
    const data = { code }
    if (userInfo) {
      data.nickname = userInfo.nickName
      data.avatar = userInfo.avatarUrl
    }
    return post('/auth/wechat-login', data)
  },

  /**
   * 账号密码登录
   * @param {String} username 用户名/手机号
   * @param {String} password 密码
   */
  login(username, password) {
    return post('/auth/login', { username, password })
  },

  /**
   * 获取当前用户信息
   */
  getCurrentUser() {
    return get('/auth/me')
  },

  /**
   * 刷新 Token
   * @param {String} refreshToken 刷新令牌
   */
  refreshToken(refreshToken) {
    return post('/auth/refresh', { refresh_token: refreshToken })
  },

  /**
   * 退出登录
   */
  logout() {
    return post('/auth/logout')
  },

  /**
   * 绑定手机号
   * @param {String} phone 手机号
   * @param {String} code 验证码（可选，测试阶段可传空跳过）
   */
  bindPhone(phone, code) {
    const data = { phone }
    // 有验证码时传入，测试阶段可跳过
    if (code) {
      data.code = code
    }
    return post('/auth/bind-phone', data)
  },

  /**
   * 发送验证码
   * @param {String} phone 手机号
   */
  sendVerifyCode(phone) {
    return post('/auth/send-code', { phone })
  }
}
