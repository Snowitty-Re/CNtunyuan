/**
 * 工具函数集合
 */

/**
 * 格式化日期
 * @param {String|Date} date 日期
 * @param {String} format 格式
 */
const formatDate = (date, format = 'YYYY-MM-DD') => {
  if (!date) return ''
  
  const d = new Date(date)
  if (isNaN(d.getTime())) return ''

  const year = d.getFullYear()
  const month = String(d.getMonth() + 1).padStart(2, '0')
  const day = String(d.getDate()).padStart(2, '0')
  const hour = String(d.getHours()).padStart(2, '0')
  const minute = String(d.getMinutes()).padStart(2, '0')
  const second = String(d.getSeconds()).padStart(2, '0')

  return format
    .replace('YYYY', year)
    .replace('MM', month)
    .replace('DD', day)
    .replace('HH', hour)
    .replace('mm', minute)
    .replace('ss', second)
}

/**
 * 格式化日期时间（相对时间）
 * @param {String|Date} date 日期
 */
const formatTimeAgo = (date) => {
  if (!date) return ''
  
  const d = new Date(date)
  const now = new Date()
  const diff = now.getTime() - d.getTime()
  
  const minute = 60 * 1000
  const hour = 60 * minute
  const day = 24 * hour
  const week = 7 * day
  const month = 30 * day
  
  if (diff < minute) {
    return '刚刚'
  } else if (diff < hour) {
    return Math.floor(diff / minute) + '分钟前'
  } else if (diff < day) {
    return Math.floor(diff / hour) + '小时前'
  } else if (diff < week) {
    return Math.floor(diff / day) + '天前'
  } else if (diff < month) {
    return Math.floor(diff / week) + '周前'
  } else {
    return formatDate(date)
  }
}

/**
 * 显示成功提示
 * @param {String} title 提示文字
 * @param {Function} callback 回调函数
 */
const showSuccess = (title = '操作成功', callback) => {
  wx.showToast({
    title,
    icon: 'success',
    duration: 1500,
    success: callback
  })
}

/**
 * 显示错误提示
 * @param {String} title 提示文字
 */
const showError = (title = '操作失败') => {
  wx.showToast({
    title,
    icon: 'error',
    duration: 2000
  })
}

/**
 * 显示普通提示
 * @param {String} title 提示文字
 */
const showToast = (title, icon = 'none') => {
  wx.showToast({
    title,
    icon,
    duration: 2000
  })
}

/**
 * 显示确认对话框
 * @param {String} title 标题
 * @param {String} content 内容
 * @returns {Promise<Boolean>}
 */
const showConfirm = (title = '提示', content = '') => {
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

/**
 * 显示加载中
 * @param {String} title 提示文字
 * @param {Number} timeout 超时时间
 */
const showLoading = (title = '加载中...', timeout = 10000) => {
  wx.showLoading({ title, mask: true })
  
  // 自动关闭
  setTimeout(() => {
    wx.hideLoading()
  }, timeout)
}

/**
 * 隐藏加载中
 */
const hideLoading = () => {
  wx.hideLoading()
}

/**
 * 防抖函数
 * @param {Function} fn 函数
 * @param {Number} delay 延迟时间
 */
const debounce = (fn, delay = 300) => {
  let timer = null
  return function(...args) {
    if (timer) clearTimeout(timer)
    timer = setTimeout(() => {
      fn.apply(this, args)
    }, delay)
  }
}

/**
 * 节流函数
 * @param {Function} fn 函数
 * @param {Number} interval 间隔时间
 */
const throttle = (fn, interval = 300) => {
  let lastTime = 0
  return function(...args) {
    const now = Date.now()
    if (now - lastTime >= interval) {
      lastTime = now
      fn.apply(this, args)
    }
  }
}

/**
 * 深拷贝
 * @param {Object} obj 对象
 */
const deepClone = (obj) => {
  if (obj === null || typeof obj !== 'object') return obj
  if (obj instanceof Date) return new Date(obj)
  if (obj instanceof Array) return obj.map(item => deepClone(item))
  if (obj instanceof Object) {
    const copy = {}
    Object.keys(obj).forEach(key => {
      copy[key] = deepClone(obj[key])
    })
    return copy
  }
  return obj
}

/**
 * 验证手机号
 * @param {String} phone 手机号
 */
const validatePhone = (phone) => {
  const reg = /^1[3-9]\d{9}$/
  return reg.test(phone)
}

/**
 * 验证身份证号
 * @param {String} idCard 身份证号
 */
const validateIdCard = (idCard) => {
  const reg = /^\d{17}[\dXx]$/
  return reg.test(idCard)
}

/**
 * 验证邮箱
 * @param {String} email 邮箱
 */
const validateEmail = (email) => {
  const reg = /^[\w-]+(\.[\w-]+)*@[\w-]+(\.[\w-]+)+$/
  return reg.test(email)
}

/**
 * 格式化数字（千分位）
 * @param {Number} num 数字
 */
const formatNumber = (num) => {
  if (num === null || num === undefined) return '0'
  return num.toString().replace(/\B(?=(\d{3})+(?!\d))/g, ',')
}

/**
 * 格式化文件大小
 * @param {Number} bytes 字节数
 */
const formatFileSize = (bytes) => {
  if (bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
}

/**
 * 生成UUID
 */
const generateUUID = () => {
  return 'xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx'.replace(/[xy]/g, function(c) {
    const r = Math.random() * 16 | 0
    const v = c === 'x' ? r : (r & 0x3 | 0x8)
    return v.toString(16)
  })
}

/**
 * 截取字符串
 * @param {String} str 字符串
 * @param {Number} length 长度
 * @param {String} suffix 后缀
 */
const truncate = (str, length = 50, suffix = '...') => {
  if (!str) return ''
  if (str.length <= length) return str
  return str.substring(0, length) + suffix
}

/**
 * 计算两个经纬度之间的距离（米）
 * @param {Number} lat1 纬度1
 * @param {Number} lng1 经度1
 * @param {Number} lat2 纬度2
 * @param {Number} lng2 经度2
 */
const calculateDistance = (lat1, lng1, lat2, lng2) => {
  const radLat1 = lat1 * Math.PI / 180.0
  const radLat2 = lat2 * Math.PI / 180.0
  const a = radLat1 - radLat2
  const b = lng1 * Math.PI / 180.0 - lng2 * Math.PI / 180.0
  let s = 2 * Math.asin(Math.sqrt(Math.pow(Math.sin(a / 2), 2) +
    Math.cos(radLat1) * Math.cos(radLat2) * Math.pow(Math.sin(b / 2), 2)))
  s = s * 6378.137 // EARTH_RADIUS
  s = Math.round(s * 10000) / 10 // 转换为米并保留一位小数
  return s
}

module.exports = {
  formatDate,
  formatTimeAgo,
  showSuccess,
  showError,
  showToast,
  showConfirm,
  showLoading,
  hideLoading,
  debounce,
  throttle,
  deepClone,
  validatePhone,
  validateIdCard,
  validateEmail,
  formatNumber,
  formatFileSize,
  generateUUID,
  truncate,
  calculateDistance
}
