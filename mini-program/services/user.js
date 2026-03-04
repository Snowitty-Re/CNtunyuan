const { get, post, put, del } = require('../utils/request')

/**
 * 用户相关服务
 */
module.exports = {
  /**
   * 获取用户列表
   * @param {Object} params 查询参数
   */
  getList(params = {}) {
    return get('/users', params)
  },

  /**
   * 获取用户详情
   * @param {String} id 用户ID
   */
  getById(id) {
    return get(`/users/${id}`)
  },

  /**
   * 创建用户
   * @param {Object} data 用户数据
   */
  create(data) {
    return post('/users', data)
  },

  /**
   * 更新用户
   * @param {String} id 用户ID
   * @param {Object} data 用户数据
   */
  update(id, data) {
    return put(`/users/${id}`, data)
  },

  /**
   * 删除用户
   * @param {String} id 用户ID
   */
  delete(id) {
    return del(`/users/${id}`)
  },

  /**
   * 更新用户状态
   * @param {String} id 用户ID
   * @param {String} status 状态
   */
  updateStatus(id, status) {
    return put(`/users/${id}/status`, { status })
  },

  /**
   * 更新用户角色
   * @param {String} id 用户ID
   * @param {String} role 角色
   */
  updateRole(id, role) {
    return put(`/users/${id}/role`, { role })
  },

  // ==================== 个人资料 ====================

  /**
   * 获取个人资料
   */
  getProfile() {
    return get('/profile')
  },

  /**
   * 更新个人资料
   * @param {Object} data 资料数据
   */
  updateProfile(data) {
    return put('/profile', data)
  },

  /**
   * 修改密码
   * @param {String} oldPassword 旧密码
   * @param {String} newPassword 新密码
   */
  changePassword(oldPassword, newPassword) {
    return put('/profile/password', {
      old_password: oldPassword,
      new_password: newPassword
    })
  },

  /**
   * 获取个人统计
   */
  getStats() {
    return get('/profile/stats')
  }
}
