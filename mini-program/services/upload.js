const { get, del, uploadFile, refreshToken, API_BASE_URL } = require('../utils/request')

/**
 * 上传相关服务（refresh 与 request 单飞共享）
 */
module.exports = {
  upload(filePath, formData = {}) {
    return uploadFile('/upload', filePath, 'file', formData)
  },

  uploadBatch(filePaths, formData = {}) {
    // 串行上传，避免并发打爆弱网
    const run = async () => {
      const out = []
      for (let i = 0; i < (filePaths || []).length; i += 1) {
        // eslint-disable-next-line no-await-in-loop
        out.push(await uploadFile('/upload', filePaths[i], 'file', formData))
      }
      return out
    }
    return run()
  },

  getById(id) {
    return get(`/upload/${id}`)
  },

  download(id) {
    return new Promise((resolve, reject) => {
      const doDownload = (token, allowRetry) => {
        wx.downloadFile({
          url: `${API_BASE_URL}/upload/${id}/download`,
          header: {
            Authorization: token ? `Bearer ${token}` : ''
          },
          success: async (res) => {
            if (res.statusCode === 200) {
              resolve(res.tempFilePath)
              return
            }
            if (res.statusCode === 401 && allowRetry) {
              try {
                const newToken = await refreshToken()
                if (newToken) {
                  doDownload(newToken, false)
                  return
                }
              } catch (e) {
                // fall through
              }
            }
            reject(new Error(`Download failed: ${res.statusCode}`))
          },
          fail: reject
        })
      }

      const token = wx.getStorageSync('token') || ''
      doDownload(token, true)
    })
  },

  delete(id) {
    return del(`/upload/${id}`)
  },

  getByEntity(entityType, entityId) {
    return get(`/upload/entity/${entityType}/${entityId}`)
  },

  bind(fileId, entityType, entityId) {
    return require('../utils/request').put(`/upload/${fileId}/bind`, {
      entity_type: entityType,
      entity_id: entityId
    })
  },

  getStats() {
    return get('/upload/stats')
  }
}
