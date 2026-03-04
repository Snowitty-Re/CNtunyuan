const { get } = require('../utils/request')

/**
 * 仪表盘相关服务
 */
module.exports = {
  /**
   * 获取仪表盘统计数据
   */
  getStats() {
    return get('/dashboard/stats')
  },

  /**
   * 获取概览数据
   */
  getOverview() {
    return get('/dashboard/overview')
  },

  /**
   * 获取趋势数据
   * @param {Number} days 天数，默认7天
   */
  getTrend(days = 7) {
    return get('/dashboard/trend', { days })
  }
}
