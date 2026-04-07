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
   * 注册账号（待审批）
   * @param {String} phone 手机号
   * @param {String} password 密码
   * @param {String} code 验证码
   * @param {String} nickname 昵称（可选）
   */
  register(phone, password, code, nickname = '', wechatCode = '') {
    return post('/auth/register', { phone, password, code, nickname, wechat_code: wechatCode })
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
   * 解绑微信账号
   */
  unbindWechat() {
    return post('/auth/unbind-wechat')
  },

  /**
   * 发送验证码
   * @param {String} phone 手机号
   * @param {String} scene 场景：verify/reset_password/change_phone
   */
  sendVerifyCode(phone, scene = 'verify') {
    return post('/auth/send-code', { phone, scene })
  },

  /**
   * 重置密码
   * @param {String} phone 手机号
   * @param {String} code 验证码
   * @param {String} newPassword 新密码
   */
  resetPassword(phone, code, newPassword, wechatCode = '') {
    return post('/auth/reset-password', { phone, code, new_password: newPassword, wechat_code: wechatCode })
  }
}
