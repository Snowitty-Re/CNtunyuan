const formatTime = date => {
  const year = date.getFullYear()
  const month = date.getMonth() + 1
  const day = date.getDate()
  const hour = date.getHours()
  const minute = date.getMinutes()
  const second = date.getSeconds()

  return `${[year, month, day].map(formatNumber).join('/')} ${[hour, minute, second].map(formatNumber).join(':')}`
}

const formatNumber = n => {
  n = n.toString()
  return n[1] ? n : `0${n}`
}

// 格式化日期
const formatDate = (dateStr) => {
  if (!dateStr) return ''
  const date = new Date(dateStr)
  return `${date.getFullYear()}-${formatNumber(date.getMonth() + 1)}-${formatNumber(date.getDate())}`
}

// 防抖
const debounce = (fn, delay = 500) => {
  let timer = null
  return function (...args) {
    if (timer) clearTimeout(timer)
    timer = setTimeout(() => {
      fn.apply(this, args)
    }, delay)
  }
}

// 节流
const throttle = (fn, interval = 500) => {
  let lastTime = 0
  return function (...args) {
    const now = Date.now()
    if (now - lastTime >= interval) {
      lastTime = now
      fn.apply(this, args)
    }
  }
}

// 显示加载
const showLoading = (title = '加载中...') => {
  wx.showLoading({ title, mask: true })
}

// 隐藏加载
const hideLoading = () => {
  wx.hideLoading()
}

// 显示成功提示
const showSuccess = (title = '成功') => {
  wx.showToast({ title, icon: 'success' })
}

// 显示错误提示
const showError = (title = '失败') => {
  wx.showToast({ title, icon: 'none' })
}

// 确认对话框
const showConfirm = (title, content) => {
  return new Promise((resolve) => {
    wx.showModal({
      title,
      content,
      success: (res) => {
        resolve(res.confirm)
      }
    })
  })
}

// 获取当前位置
const getLocation = () => {
  return new Promise((resolve, reject) => {
    wx.getLocation({
      type: 'gcj02',
      success: resolve,
      fail: reject
    })
  })
}

// 选择图片
const chooseImage = (count = 1) => {
  return new Promise((resolve, reject) => {
    wx.chooseMedia({
      count,
      mediaType: ['image'],
      sourceType: ['album', 'camera'],
      success: resolve,
      fail: reject
    })
  })
}

// 上传文件
const uploadFile = (url, filePath, name = 'file') => {
  return new Promise((resolve, reject) => {
    const token = wx.getStorageSync('token')
    wx.uploadFile({
      url: `${getApp().globalData.apiBaseUrl}${url}`,
      filePath,
      name,
      header: {
        'Authorization': token ? `Bearer ${token}` : ''
      },
      success: (res) => {
        const data = JSON.parse(res.data)
        if (data.code === 200) {
          resolve(data.data)
        } else {
          reject(new Error(data.message))
        }
      },
      fail: reject
    })
  })
}

module.exports = {
  formatTime,
  formatDate,
  formatNumber,
  debounce,
  throttle,
  showLoading,
  hideLoading,
  showSuccess,
  showError,
  showConfirm,
  getLocation,
  chooseImage,
  uploadFile
}
