const dashboardService = require('../../services/dashboard')
const missingPersonService = require('../../services/missingPerson')
const dialectService = require('../../services/dialect')
const { formatDate, formatTimeAgo, showError, showLoading, hideLoading } = require('../../utils/util')

Page({
  data: {
    // 加载状态
    isLoading: false,
    hasError: false,
    errorMessage: '',

    // 统计数据
    stats: {
      totalCases: 0,
      resolvedCases: 0,
      volunteers: 0,
      dialects: 0
    },

    // 最新案件列表
    recentCases: [],
    casesLoading: false,
    casesError: false,

    // 最新方言列表
    recentDialects: [],
    dialectsLoading: false,
    dialectsError: false,

    // 状态文本映射
    statusText: {
      missing: '失踪中',
      searching: '寻找中',
      found: '已找到',
      reunited: '已团圆',
      closed: '已结案'
    },

    // 问候语
    greeting: ''
  },

  onLoad() {
    this.updateGreeting()
    this.loadData()
  },

  onShow() {
    this.updateGreeting()
    // 每次显示页面时刷新数据
    this.loadData()
  },

  onPullDownRefresh() {
    this.loadData().finally(() => {
      wx.stopPullDownRefresh()
    })
  },

  /**
   * 更新问候语
   */
  updateGreeting() {
    const hour = new Date().getHours()
    let greeting = '你好'
    if (hour < 6) {
      greeting = '夜深了'
    } else if (hour < 9) {
      greeting = '早上好'
    } else if (hour < 12) {
      greeting = '上午好'
    } else if (hour < 14) {
      greeting = '中午好'
    } else if (hour < 18) {
      greeting = '下午好'
    } else {
      greeting = '晚上好'
    }
    this.setData({ greeting })
  },

  /**
   * 加载所有数据
   */
  async loadData() {
    this.setData({ isLoading: true, hasError: false })

    try {
      // 并行加载所有数据
      await Promise.all([
        this.loadStats(),
        this.loadRecentCases(),
        this.loadRecentDialects()
      ])
    } catch (error) {
      console.error('加载数据失败:', error)
      this.setData({ hasError: true, errorMessage: '加载失败，请下拉刷新重试' })
    } finally {
      this.setData({ isLoading: false })
    }
  },

  /**
   * 加载统计数据
   */
  async loadStats() {
    try {
      // 初始化统计数据
      let stats = {
        totalCases: 0,
        resolvedCases: 0,
        volunteers: 0,
        dialects: 0
      }

      // 获取仪表盘统计数据
      try {
        const dashboardStats = await dashboardService.getStats()
        console.log('仪表盘统计数据:', dashboardStats)
        
        if (dashboardStats) {
          // 处理嵌套结构：missing_persons.total, users.total, dialects.total
          if (dashboardStats.missing_persons) {
            stats.totalCases = dashboardStats.missing_persons.total || 0
            stats.resolvedCases = (dashboardStats.missing_persons.found || 0) + (dashboardStats.missing_persons.reunited || 0)
          }
          
          if (dashboardStats.users) {
            stats.volunteers = dashboardStats.users.total || 0
          }
          
          if (dashboardStats.dialects) {
            stats.dialects = dashboardStats.dialects.total || 0
          }
          
          // 也尝试平铺结构的兼容
          stats.totalCases = stats.totalCases || dashboardStats.total_cases || 0
          stats.resolvedCases = stats.resolvedCases || dashboardStats.resolved_cases || 0
          stats.volunteers = stats.volunteers || dashboardStats.total_users || 0
          stats.dialects = stats.dialects || dashboardStats.total_dialects || 0
        }
      } catch (e) {
        console.log('仪表盘统计获取失败:', e)
      }

      // 如果仪表盘数据不完整，尝试单独获取
      const promises = []

      if (stats.totalCases === 0) {
        promises.push(
          missingPersonService.getStats().then(res => {
            console.log('案件统计:', res)
            if (res) {
              // 处理嵌套或平铺结构
              if (res.missing_persons) {
                stats.totalCases = res.missing_persons.total || 0
                stats.resolvedCases = (res.missing_persons.found || 0) + (res.missing_persons.reunited || 0)
              } else {
                stats.totalCases = res.total || res.total_cases || 0
                stats.resolvedCases = res.found || res.resolved || 0
              }
            }
          }).catch(() => {})
        )
      }

      if (stats.dialects === 0) {
        promises.push(
          dialectService.getStats().then(res => {
            console.log('方言统计:', res)
            if (res) {
              if (res.dialects) {
                stats.dialects = res.dialects.total || 0
              } else {
                stats.dialects = res.total || res.total_dialects || 0
              }
            }
          }).catch(() => {})
        )
      }

      // 尝试获取概览数据
      if (stats.volunteers === 0 || stats.totalCases === 0) {
        promises.push(
          dashboardService.getOverview().then(res => {
            console.log('概览数据:', res)
            if (res) {
              stats.volunteers = stats.volunteers || res.total_users || 0
              stats.totalCases = stats.totalCases || res.total_cases || 0
              stats.resolvedCases = stats.resolvedCases || res.resolved_cases || 0
            }
          }).catch(() => {})
        )
      }

      await Promise.all(promises)
      
      console.log('最终统计数据:', stats)
      this.setData({ stats })
    } catch (error) {
      console.error('加载统计数据失败:', error)
      // 统计数据加载失败不影响其他功能
    }
  },

  /**
   * 加载最新案件列表
   */
  async loadRecentCases() {
    this.setData({ casesLoading: true, casesError: false })

    try {
      const result = await missingPersonService.getList({ 
        page: 1, 
        page_size: 5 
      })

      console.log('案件列表原始数据:', result)
      
      const list = result.list || result.data || result || []
      
      const cases = list.map(item => {
        // 组合地址：province + city + district + address
        const locationParts = []
        if (item.province) locationParts.push(item.province)
        if (item.city) locationParts.push(item.city)
        if (item.district) locationParts.push(item.district)
        if (item.address) locationParts.push(item.address)
        
        const missingLocation = locationParts.length > 0 
          ? locationParts.join(' ') 
          : '未知地点'
        
        // 处理时间字段 - 后端返回的是 ISO 格式字符串
        const missingTime = item.missing_time 
          ? formatTimeAgo(item.missing_time) 
          : '未知时间'
        
        return {
          id: item.id,
          name: item.name || '未知',
          status: item.status || 'missing',
          photoUrl: this.getPhotoUrl(item),
          missingLocation: missingLocation,
          missingTime: missingTime,
          age: this.calculateAge(item.birth_date),
          gender: item.gender === 'male' ? '男' : item.gender === 'female' ? '女' : '未知'
        }
      })

      this.setData({ 
        recentCases: cases,
        casesLoading: false 
      })
    } catch (error) {
      console.error('加载案件列表失败:', error)
      this.setData({ 
        casesLoading: false, 
        casesError: true 
      })
    }
  },

  /**
   * 加载最新方言列表
   */
  async loadRecentDialects() {
    this.setData({ dialectsLoading: true, dialectsError: false })

    try {
      // 优先获取精选方言，如果没有则获取普通列表
      let result
      try {
        result = await dialectService.getFeatured({ page: 1, page_size: 5 })
      } catch (e) {
        result = await dialectService.getList({ page: 1, page_size: 5 })
      }

      console.log('方言列表原始数据:', result)
      
      const list = result.list || result.data || result || []

      const dialects = list.map(item => {
        // 确保所有字段都是字符串
        const title = item.title || item.content || '方言录音'
        const province = item.province || ''
        const city = item.city || ''
        
        return {
          id: item.id,
          title: String(title),
          province: String(province),
          city: String(city),
          playCount: this.formatCount(item.play_count || item.playCount || 0),
          likeCount: this.formatCount(item.like_count || item.likeCount || 0),
          duration: item.duration || '00:00',
          createdAt: formatTimeAgo(item.created_at || item.createdAt)
        }
      })

      this.setData({ 
        recentDialects: dialects,
        dialectsLoading: false 
      })
    } catch (error) {
      console.error('加载方言列表失败:', error)
      this.setData({ 
        dialectsLoading: false, 
        dialectsError: true 
      })
    }
  },

  /**
   * 获取照片URL
   */
  getPhotoUrl(item) {
    // 尝试多种可能的图片字段
    if (item.photos && item.photos.length > 0) {
      return item.photos[0].url || item.photos[0]
    }
    if (item.photo_url) return item.photo_url
    if (item.avatar) return item.avatar
    if (item.image) return item.image
    if (item.cover) return item.cover
    if (item.cover_url) return item.cover_url
    // 默认头像
    return '/assets/images/default-avatar.png'
  },

  /**
   * 计算年龄
   */
  calculateAge(birthDate) {
    if (!birthDate) return null
    try {
      const birth = new Date(birthDate)
      const now = new Date()
      let age = now.getFullYear() - birth.getFullYear()
      const monthDiff = now.getMonth() - birth.getMonth()
      if (monthDiff < 0 || (monthDiff === 0 && now.getDate() < birth.getDate())) {
        age--
      }
      return age > 0 ? age : null
    } catch (e) {
      return null
    }
  },

  /**
   * 格式化数字（超过1000显示为k）
   */
  formatCount(count) {
    const num = parseInt(count) || 0
    if (num >= 1000000) {
      return (num / 1000000).toFixed(1) + 'M'
    }
    if (num >= 1000) {
      return (num / 1000).toFixed(1) + 'k'
    }
    return num.toString()
  },

  /**
   * 安全转换为字符串（处理对象类型）
   */
  safeString(value, defaultValue = '') {
    if (value === null || value === undefined) {
      return defaultValue
    }
    if (typeof value === 'string') {
      return value
    }
    if (typeof value === 'object') {
      // 如果是对象，尝试获取 name 或 title 字段
      if (value.name) return String(value.name)
      if (value.title) return String(value.title)
      if (value.label) return String(value.label)
      // 否则返回默认字符串
      console.warn('字段值为对象:', value)
      return defaultValue
    }
    return String(value)
  },

  // ========== 导航方法 ==========

  /**
   * 跳转到案件详情
   */
  goToCaseDetail(e) {
    const id = e.currentTarget.dataset.id
    if (!id) return
    wx.navigateTo({ url: `/pages/cases/detail?id=${id}` })
  },

  /**
   * 跳转到方言详情
   */
  goToDialectDetail(e) {
    const id = e.currentTarget.dataset.id
    if (!id) return
    wx.navigateTo({ url: `/pages/dialect/detail?id=${id}` })
  },

  /**
   * 跳转到案件列表
   */
  goToCases() {
    wx.switchTab({ url: '/pages/cases/list' })
  },

  /**
   * 跳转到方言列表
   */
  goToDialects() {
    wx.navigateTo({ url: '/pages/dialect/list' })
  },

  // ========== 快捷入口 ==========

  /**
   * 发布案件
   */
  onCreateCase() {
    wx.navigateTo({ url: '/pages/cases/create' })
  },

  /**
   * 录制方言
   */
  onRecordDialect() {
    wx.navigateTo({ url: '/pages/dialect/create' })
  },

  /**
   * 查看地图
   */
  onViewMap() {
    wx.switchTab({ url: '/pages/map/index' })
  },

  /**
   * 我的任务
   */
  onMyTasks() {
    wx.navigateTo({ url: '/pages/tasks/my' })
  },

  // ========== 重试方法 ==========

  /**
   * 重试加载统计数据
   */
  retryLoadStats() {
    this.loadStats()
  },

  /**
   * 重试加载案件列表
   */
  retryLoadCases() {
    this.loadRecentCases()
  },

  /**
   * 重试加载方言列表
   */
  retryLoadDialects() {
    this.loadRecentDialects()
  },

  /**
   * 点击统计卡片
   */
  onStatCardTap(e) {
    const type = e.currentTarget.dataset.type
    switch (type) {
      case 'cases':
        this.goToCases()
        break
      case 'resolved':
        // 跳转到已找到的案件列表
        wx.navigateTo({ 
          url: '/pages/cases/list?status=found' 
        })
        break
      case 'volunteers':
        // 跳转到志愿者页面
        wx.navigateTo({ url: '/pages/volunteer/profile' })
        break
      case 'dialects':
        this.goToDialects()
        break
    }
  }
})
