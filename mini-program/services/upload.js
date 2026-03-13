const { get, del } = require('../utils/request')
const { uploadFile } = require('../utils/request')

/**
 * 上传相关服务
 */
module.exports = {
  /**
   * 单文件上传
   * @param {String} filePath 文件路径
   * @param {Object} formData 附加数据
   */
  upload(filePath, formData = {}) {
    return uploadFile('/upload', filePath, 'file', formData)
  },

  /**
   * 批量上传（逐个调用单文件上传接口）
   * @param {Array} filePaths 文件路径数组
   * @param {Object} formData 附加数据
   */
  uploadBatch(filePaths, formData = {}) {
    const uploadPromises = filePaths.map(path =>
      uploadFile('/upload', path, 'file', formData)
    )
    return Promise.all(uploadPromises)
  },

  /**
   * 获取文件信息
   * @param {String} id 文件ID
   */
  getById(id) {
    return get(`/upload/${id}`)
  },

  /**
   * 下载文件
   * @param {String} id 文件ID
   */
  download(id) {
    return new Promise((resolve, reject) => {
      const token = wx.getStorageSync('token') || ''
      const API_BASE_URL = 'https://cntuanyuan.com/api/v1'
      wx.downloadFile({
        url: `${API_BASE_URL}/upload/${id}/download`,
        header: {
          'Authorization': token ? `Bearer ${token}` : ''
        },
        success: (res) => {
          if (res.statusCode === 200) {
            resolve(res.tempFilePath)
          } else {
            reject(new Error(`Download failed: ${res.statusCode}`))
          }
        },
        fail: reject
      })
    })
  },

  /**
   * 删除文件
   * @param {String} id 文件ID
   */
  delete(id) {
    return del(`/upload/${id}`)
  },

  /**
   * 获取实体的文件列表
   * @param {String} entityType 实体类型
   * @param {String} entityId 实体ID
   */
  getFilesByEntity(entityType, entityId) {
    return get(`/upload/entity/${entityType}/${entityId}`)
  },

  /**
   * 绑定文件到实体
   * @param {String} fileId 文件ID
   * @param {String} entityType 实体类型
   * @param {String} entityId 实体ID
   */
  bindToEntity(fileId, entityType, entityId) {
    return require('../utils/request').put(`/upload/${fileId}/bind`, {
      entity_type: entityType,
      entity_id: entityId
    })
  },

  /**
   * 获取上传统计
   */
  getStats() {
    return get('/upload/stats')
  }
}
