const { get, post, put, del } = require('../utils/request')

/**
 * 任务相关服务
 */
module.exports = {
  /**
   * 获取任务列表
   * @param {Object} params 查询参数 {page, page_size, status, priority}
   */
  getList(params = {}) {
    return get('/tasks', params)
  },

  /**
   * 获取任务详情
   * @param {String} id 任务ID
   */
  getById(id) {
    return get(`/tasks/${id}`)
  },

  /**
   * 创建任务
   * @param {Object} data 任务数据
   */
  create(data) {
    return post('/tasks', data)
  },

  /**
   * 更新任务
   * @param {String} id 任务ID
   * @param {Object} data 任务数据
   */
  update(id, data) {
    return put(`/tasks/${id}`, data)
  },

  /**
   * 删除任务
   * @param {String} id 任务ID
   */
  delete(id) {
    return del(`/tasks/${id}`)
  },

  /**
   * 分配任务
   * @param {String} id 任务ID
   * @param {String} assigneeId 执行人ID
   */
  assign(id, assigneeId) {
    return post(`/tasks/${id}/assign`, { assignee_id: assigneeId })
  },

  /**
   * 开始任务
   * @param {String} id 任务ID
   */
  start(id) {
    return post(`/tasks/${id}/start`)
  },

  /**
   * 完成任务
   * @param {String} id 任务ID
   * @param {Object} data 完成数据 {result, feedback}
   */
  complete(id, data) {
    return post(`/tasks/${id}/complete`, data)
  },

  /**
   * 取消任务
   * @param {String} id 任务ID
   * @param {String} reason 取消原因
   */
  cancel(id, reason) {
    return post(`/tasks/${id}/cancel`, { reason })
  },

  /**
   * 更新进度
   * @param {String} id 任务ID
   * @param {Number} progress 进度 0-100
   * @param {String} remark 备注
   */
  updateProgress(id, progress, remark) {
    return put(`/tasks/${id}/progress`, { progress, remark })
  },

  /**
   * 获取我的任务
   * @param {Object} params 分页参数
   */
  getMyTasks(params = {}) {
    return get('/tasks/my', params)
  },

  /**
   * 获取待分配任务
   * @param {Object} params 分页参数
   */
  getPendingTasks(params = {}) {
    return get('/tasks/pending', params)
  },

  /**
   * 获取逾期任务
   * @param {Object} params 分页参数
   */
  getOverdueTasks(params = {}) {
    return get('/tasks/overdue', params)
  },

  /**
   * 获取任务日志
   * @param {String} id 任务ID
   */
  getLogs(id) {
    return get(`/tasks/${id}/logs`)
  },

  /**
   * 获取统计数据
   */
  getStats() {
    return get('/tasks/stats')
  }
}
