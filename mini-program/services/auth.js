const { get, post } = require('../utils/request')

/**
 * 认证相关服务
 */
module.exports = {
  /**
   * 微信登录
   * @param {String} code 微信登录码
   */
  wechatLogin(code) {
    return post('/auth/wechat-login', { code })
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
   * @param {String} code 验证码
   */
  bindPhone(phone, code) {
    return post('/auth/bind-phone', { phone, code })
  },

  /**
   * 发送验证码
   * @param {String} phone 手机号
   */
  sendVerifyCode(phone) {
    return post('/auth/send-code', { phone })
  }
}
