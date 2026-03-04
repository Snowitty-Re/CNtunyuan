const { get, post, put, del } = require('../utils/request')

/**
 * 方言相关服务
 */
module.exports = {
  /**
   * 获取方言列表
   * @param {Object} params 查询参数 {page, page_size, region, tags}
   */
  getList(params = {}) {
    return get('/dialects', params)
  },

  /**
   * 获取方言详情
   * @param {String} id 方言ID
   */
  getById(id) {
    return get(`/dialects/${id}`)
  },

  /**
   * 创建方言
   * @param {Object} data 方言数据
   */
  create(data) {
    return post('/dialects', data)
  },

  /**
   * 更新方言
   * @param {String} id 方言ID
   * @param {Object} data 方言数据
   */
  update(id, data) {
    return put(`/dialects/${id}`, data)
  },

  /**
   * 删除方言
   * @param {String} id 方言ID
   */
  delete(id) {
    return del(`/dialects/${id}`)
  },

  /**
   * 更新状态
   * @param {String} id 方言ID
   * @param {String} status 状态
   */
  updateStatus(id, status) {
    return put(`/dialects/${id}/status`, { status })
  },

  /**
   * 设为精选
   * @param {String} id 方言ID
   */
  feature(id) {
    return post(`/dialects/${id}/feature`)
  },

  /**
   * 取消精选
   * @param {String} id 方言ID
   */
  unfeature(id) {
    return del(`/dialects/${id}/feature`)
  },

  /**
   * 记录播放
   * @param {String} id 方言ID
   */
  recordPlay(id) {
    return post(`/dialects/${id}/play`)
  },

  /**
   * 点赞
   * @param {String} id 方言ID
   */
  like(id) {
    return post(`/dialects/${id}/like`)
  },

  /**
   * 取消点赞
   * @param {String} id 方言ID
   */
  unlike(id) {
    return del(`/dialects/${id}/like`)
  },

  /**
   * 添加评论
   * @param {String} id 方言ID
   * @param {Object} data 评论数据 {content}
   */
  addComment(id, data) {
    return post(`/dialects/${id}/comments`, data)
  },

  /**
   * 获取评论列表
   * @param {String} id 方言ID
   * @param {Object} params 分页参数
   */
  getComments(id, params = {}) {
    return get(`/dialects/${id}/comments`, params)
  },

  /**
   * 获取精选方言
   * @param {Object} params 分页参数
   */
  getFeatured(params = {}) {
    return get('/dialects/featured', params)
  },

  /**
   * 获取统计数据
   */
  getStats() {
    return get('/dialects/stats')
  }
}
