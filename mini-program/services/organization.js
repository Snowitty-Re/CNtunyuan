const { get, post, put, del } = require('../utils/request')

/**
 * 组织相关服务
 */
module.exports = {
  /**
   * 获取组织列表
   * @param {Object} params 查询参数
   */
  getList(params = {}) {
    return get('/organizations', params)
  },

  /**
   * 获取组织详情
   * @param {String} id 组织ID
   */
  getById(id) {
    return get(`/organizations/${id}`)
  },

  /**
   * 创建组织
   * @param {Object} data 组织数据
   */
  create(data) {
    return post('/organizations', data)
  },

  /**
   * 更新组织
   * @param {String} id 组织ID
   * @param {Object} data 组织数据
   */
  update(id, data) {
    return put(`/organizations/${id}`, data)
  },

  /**
   * 删除组织
   * @param {String} id 组织ID
   */
  delete(id) {
    return del(`/organizations/${id}`)
  },

  /**
   * 获取组织树
   */
  getTree() {
    return get('/organizations/tree')
  },

  /**
   * 获取当前用户的组织
   */
  getMyOrganization() {
    return get('/organizations/my')
  },

  /**
   * 获取组织成员
   * @param {String} id 组织ID
   * @param {Object} params 分页参数
   */
  getMembers(id, params = {}) {
    return get(`/organizations/${id}/members`, params)
  },

  /**
   * 获取组织统计
   * @param {String} id 组织ID
   */
  getStats(id) {
    return get(`/organizations/${id}/stats`)
  }
}
