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
   * @param {String} code 短信验证码
   */
  bindPhone(phone, code) {
    return post('/auth/bind-phone', { phone, code })
  },

  /**
   * 微信一键绑定手机号
   * @param {String} wechatCode 微信 getPhoneNumber 返回的 code
   */
  bindPhoneByWechatCode(wechatCode) {
    return post('/auth/bind-phone', { wechat_code: wechatCode })
  },

  /**
   * 绑定微信账号（手机号登录用户）
   * @param {String} code wx.login 返回的 code
   */
  bindWechat(code) {
    return post('/auth/bind-wechat', { code })
  },

  /**
   * 发送验证码
   * @param {String} phone 手机号
   */
  sendVerifyCode(phone) {
    return post('/auth/send-code', { phone })
  }
}
