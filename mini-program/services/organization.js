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
   * 获取子组织
   * @param {String} id 组织ID
   */
  getChildren(id) {
    return get(`/organizations/${id}/children`)
  },

  /**
   * 获取组织路径
   * @param {String} id 组织ID
   */
  getPath(id) {
    return get(`/organizations/${id}/path`)
  },

  /**
   * 移动组织
   * @param {String} id 组织ID
   * @param {String} newParentId 新父组织 ID（后端字段 new_parent_id，必填；不支持空父级）
   */
  move(id, newParentId) {
    return put(`/organizations/${id}/move`, { new_parent_id: newParentId })
  }
}
