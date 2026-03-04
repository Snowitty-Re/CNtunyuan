const { get, post, put, del } = require('../utils/request')

/**
 * 走失人员相关服务
 */
module.exports = {
  /**
   * 获取走失人员列表
   * @param {Object} params 查询参数 {page, page_size, status, keyword}
   */
  getList(params = {}) {
    return get('/missing-persons', params)
  },

  /**
   * 获取走失人员详情
   * @param {String} id 走失人员ID
   */
  getById(id) {
    return get(`/missing-persons/${id}`)
  },

  /**
   * 创建走失人员
   * @param {Object} data 走失人员数据
   */
  create(data) {
    return post('/missing-persons', data)
  },

  /**
   * 更新走失人员
   * @param {String} id 走失人员ID
   * @param {Object} data 走失人员数据
   */
  update(id, data) {
    return put(`/missing-persons/${id}`, data)
  },

  /**
   * 删除走失人员
   * @param {String} id 走失人员ID
   */
  delete(id) {
    return del(`/missing-persons/${id}`)
  },

  /**
   * 搜索走失人员
   * @param {String} keyword 关键词
   * @param {Object} params 其他参数
   */
  search(keyword, params = {}) {
    return get('/missing-persons/search', { keyword, ...params })
  },

  /**
   * 更新状态
   * @param {String} id 走失人员ID
   * @param {String} status 状态
   */
  updateStatus(id, status) {
    return put(`/missing-persons/${id}/status`, { status })
  },

  /**
   * 标记已找到
   * @param {String} id 走失人员ID
   * @param {Object} data 找到信息 {found_location, found_time, description}
   */
  markFound(id, data) {
    return post(`/missing-persons/${id}/found`, data)
  },

  /**
   * 标记已团圆
   * @param {String} id 走失人员ID
   */
  markReunited(id) {
    return post(`/missing-persons/${id}/reunited`)
  },

  /**
   * 获取轨迹记录
   * @param {String} id 走失人员ID
   */
  getTracks(id) {
    return get(`/missing-persons/${id}/tracks`)
  },

  /**
   * 添加轨迹记录
   * @param {String} id 走失人员ID
   * @param {Object} data 轨迹数据 {location, latitude, longitude, description, photos}
   */
  addTrack(id, data) {
    return post(`/missing-persons/${id}/tracks`, data)
  },

  /**
   * 获取统计数据
   */
  getStats() {
    return get('/missing-persons/stats')
  }
}
