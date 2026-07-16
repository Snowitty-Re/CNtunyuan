const dashboardService = require('../../services/dashboard')
const missingPersonService = require('../../services/missingPerson')
const dialectService = require('../../services/dialect')
const userService = require('../../services/user')
const taskService = require('../../services/task')
const { formatDate, formatTimeAgo, showError, showLoading, hideLoading, joinLocation, normalizeMediaUrl, normalizeAge } = require('../../utils/util')
const { ACTIONS } = require('../../utils/permission')
const app = getApp()
const CASES_STATUS_FILTER_KEY = 'cases_status_filter'
const CASES_LIST_DIRTY_KEY = 'cases_list_dirty'
const DIALECT_LIST_DIRTY_KEY = 'dialect_list_dirty'

Page({
  data: {
    // 加载状态
    isLoading: false,
    hasError: false,
    errorMessage: '',

    // 统计数据（原始）
    stats: {
      totalCases: 0,
      resolvedCases: 0,
      volunteers: 0,
      dialects: 0,
      activeCases: 0,
      pendingTasks: 0,
      totalTasks: 0
    },
    statCards: [
      { type: 'cases', value: 0, label: '走失人员', icon: 'icon-people', bgClass: 'orange-bg' },
      { type: 'resolved', value: 0, label: '已找到', icon: 'icon-success', bgClass: 'green-bg' },
      { type: 'volunteers', value: 0, label: '志愿者', icon: 'icon-volunteer', bgClass: 'blue-bg' },
      { type: 'dialects', value: 0, label: '方言录音', icon: 'icon-audio', bgClass: 'purple-bg' }
    ],

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
    greeting: '',
    canCreateTask: false
  },

  onLoad() {
    if (!app.ensureAuth || !app.ensureAuth()) return
    this.updateGreeting()
    this.loadData()
  },

  onShow() {
    if (!app.ensureAuth || !app.ensureAuth()) return
    this.updateGreeting()
    const listDirty = wx.getStorageSync(CASES_LIST_DIRTY_KEY)
    const dialectDirty = wx.getStorageSync(DIALECT_LIST_DIRTY_KEY)
    if (listDirty) {
      wx.removeStorageSync(CASES_LIST_DIRTY_KEY)
      this.loadRecentCases()
    }
    if (dialectDirty) {
      wx.removeStorageSync(DIALECT_LIST_DIRTY_KEY)
      this.loadRecentDialects()
    }
    if (listDirty || dialectDirty) {
      return
    }
    // Throttle: skip if loaded less than 30s ago
    const now = Date.now()
    if (this._lastLoadTime && now - this._lastLoadTime < 30000) return
    this._lastLoadTime = now
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
      const userInfo = app.globalData.userInfo || wx.getStorageSync('userInfo') || {}
      const isManagerRole = app.canAnyPermission([
        ACTIONS.TASK_MANAGE,
        ACTIONS.MISSING_MANAGE,
        ACTIONS.DIALECT_MANAGE,
        ACTIONS.USER_VIEW
      ], userInfo)
      const canCreateTask = app.hasPermission(ACTIONS.TASK_MANAGE, userInfo)

      // 初始化统计数据
      let stats = {
        totalCases: 0,
        resolvedCases: 0,
        volunteers: 0,
        dialects: 0,
        activeCases: 0,
        pendingTasks: 0,
        totalTasks: 0
      }

      if (isManagerRole) {
        // 管理角色：全局统计
        try {
          const dashboardStats = await dashboardService.getStats()

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
          if (this.isPermissionDenied(e)) {
            console.log('仪表盘统计无权限，跳过管理端统计接口')
          } else {
            console.log('仪表盘统计获取失败:', e)
          }
        }
        // 如果仪表盘数据不完整，尝试单独获取
        const promises = []

        if (stats.totalCases === 0) {
          promises.push(
            missingPersonService.getStats().then(res => {
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
            }).catch((err) => {
              console.warn('missingPerson stats fallback failed', err)
            })
          )
        }

        if (stats.dialects === 0) {
          promises.push(
            dialectService.getStats().then(res => {
              if (res) {
                if (res.dialects) {
                  stats.dialects = res.dialects.total || 0
                } else {
                  stats.dialects = res.total || res.total_dialects || 0
                }
              }
            }).catch((err) => {
              if (!this.isPermissionDenied(err)) {
                console.warn('dialect stats fallback failed', err)
              }
            })
          )
        }

        // 尝试获取概览数据
        if (stats.volunteers === 0 || stats.totalCases === 0) {
          promises.push(
            dashboardService.getOverview().then(res => {
              if (res) {
                stats.volunteers = stats.volunteers || res.total_users || 0
                stats.totalCases = stats.totalCases || res.total_cases || 0
                stats.resolvedCases = stats.resolvedCases || res.resolved_cases || 0
              }
            }).catch((err) => {
              if (!this.isPermissionDenied(err)) {
                console.warn('dashboard overview fallback failed', err)
              }
            })
          )
        }

        await Promise.all(promises)
      } else {
        // 志愿者：个人维度统计
        const [profileStats, taskStats] = await Promise.all([
          userService.getStats().catch(() => null),
          taskService.getStats().catch(() => null)
        ])

        stats.totalCases = profileStats?.totalCases || 0
        stats.resolvedCases = profileStats?.completedCases || 0
        stats.activeCases = profileStats?.activeCases || 0
        stats.totalTasks = profileStats?.totalTasks || taskStats?.my_tasks || 0
        stats.pendingTasks = profileStats?.pendingTasks || taskStats?.my_pending || 0
      }

      this.setData({
        stats,
        canCreateTask,
        statCards: this.buildStatCards(userInfo, stats)
      })
    } catch (error) {
      console.error('加载统计数据失败:', error)
      // 统计数据加载失败不影响其他功能
    }
  },

  isPermissionDenied(err) {
    const message = (err && err.message ? err.message : '').toLowerCase()
    return message.includes('permission denied') || message.includes('forbidden')
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

      const list = result.list || []

      const cases = list.map(item => ({
        id: item.id,
        name: item.name || '未知',
        status: item.status || 'missing',
        photoUrl: this.getPhotoUrl(item),
        missingLocation: joinLocation(item, '未知地点'),
        missingTime: item.missing_time ? formatTimeAgo(item.missing_time) : '未知时间',
        age: normalizeAge(item.age),
        gender: item.gender === 'male' ? '男' : item.gender === 'female' ? '女' : '未知'
      }))

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
      // 优先获取精选方言；若为空则回退普通列表
      let result = null
      try {
        result = await dialectService.getFeatured({ page: 1, page_size: 5 })
      } catch (e) {
        result = null
      }

      let list = this.getDialectList(result)
      if (!list.length) {
        const fallbackResult = await dialectService.getList({ page: 1, page_size: 5 })
        list = this.getDialectList(fallbackResult)
      }
      const safeList = this.filterVisibleDialects(list)

      const dialects = safeList.map(item => ({
        id: item.id,
        title: item.title || item.content || '方言录音',
        province: item.province || '',
        city: item.city || '',
        playCount: this.formatCount(item.play_count || 0),
        likeCount: this.formatCount(item.like_count || 0),
        durationText: this._formatDuration(item.duration),
        createdAt: formatTimeAgo(item.created_at)
      }))

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
      return normalizeMediaUrl(item.photos[0].url || item.photos[0]) || '/assets/images/default-avatar.png'
    }
    if (item.photo_url) return normalizeMediaUrl(item.photo_url) || '/assets/images/default-avatar.png'
    if (item.avatar) return normalizeMediaUrl(item.avatar) || '/assets/images/default-avatar.png'
    if (item.image) return normalizeMediaUrl(item.image) || '/assets/images/default-avatar.png'
    if (item.cover) return normalizeMediaUrl(item.cover) || '/assets/images/default-avatar.png'
    if (item.cover_url) return normalizeMediaUrl(item.cover_url) || '/assets/images/default-avatar.png'
    // 默认头像
    return '/assets/images/default-avatar.png'
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
      return defaultValue
    }
    return String(value)
  },

  // 格式化秒数为 m:ss 字符串
  _formatDuration(seconds) {
    const s = parseInt(seconds) || 0
    return `${Math.floor(s / 60)}:${String(s % 60).padStart(2, '0')}`
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
    wx.navigateTo({ url: '/pages/map/index' })
  },

  /**
   * 我的任务
   */
  onMyTasks() {
    if (!app.ensurePhoneBound || !app.ensurePhoneBound()) return
    wx.navigateTo({ url: '/pages/tasks/my' })
  },

  onCreateTask() {
    if (!app.ensurePhoneBound || !app.ensurePhoneBound()) return
    if (!app.hasPermission(ACTIONS.TASK_MANAGE)) {
      wx.showToast({
        title: '仅管理员可创建任务',
        icon: 'none'
      })
      return
    }
    wx.navigateTo({ url: '/pages/tasks/create' })
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
        wx.setStorageSync(CASES_STATUS_FILTER_KEY, 'found')
        wx.switchTab({ url: '/pages/cases/list' })
        break
      case 'volunteers':
        if (app.hasPermission(ACTIONS.USER_VIEW)) {
          wx.navigateTo({ url: '/pages/admin/user-manage?role=volunteer' })
        } else {
          wx.showToast({
            title: '仅管理端可查看志愿者列表',
            icon: 'none'
          })
        }
        break
      case 'dialects':
        this.goToDialects()
        break
      case 'my_cases':
        this.goToCases()
        break
      case 'active_cases':
        wx.setStorageSync(CASES_STATUS_FILTER_KEY, 'searching')
        wx.switchTab({ url: '/pages/cases/list' })
        break
      case 'my_tasks':
        this.onMyTasks()
        break
      case 'pending_tasks':
        if (!app.ensurePhoneBound || !app.ensurePhoneBound()) return
        wx.navigateTo({ url: '/pages/tasks/my?status=pending' })
        break
    }
  },

  buildStatCards(userInfo, stats) {
    const isManagerRole = app.canAnyPermission([
      ACTIONS.TASK_MANAGE,
      ACTIONS.MISSING_MANAGE,
      ACTIONS.DIALECT_MANAGE,
      ACTIONS.USER_VIEW
    ], userInfo)
    if (isManagerRole) {
      return [
        { type: 'cases', value: stats.totalCases || 0, label: '走失人员', icon: 'icon-people', bgClass: 'orange-bg' },
        { type: 'resolved', value: stats.resolvedCases || 0, label: '已找到', icon: 'icon-success', bgClass: 'green-bg' },
        { type: 'volunteers', value: stats.volunteers || 0, label: '志愿者', icon: 'icon-volunteer', bgClass: 'blue-bg' },
        { type: 'dialects', value: stats.dialects || 0, label: '方言录音', icon: 'icon-audio', bgClass: 'purple-bg' }
      ]
    }

    return [
      { type: 'my_cases', value: stats.totalCases || 0, label: '我相关案件', icon: 'icon-people', bgClass: 'orange-bg' },
      { type: 'active_cases', value: stats.activeCases || 0, label: '进行中案件', icon: 'icon-success', bgClass: 'green-bg' },
      { type: 'my_tasks', value: stats.totalTasks || 0, label: '我的任务', icon: 'icon-task', bgClass: 'blue-bg' },
      { type: 'pending_tasks', value: stats.pendingTasks || 0, label: '待处理任务', icon: 'icon-notification', bgClass: 'purple-bg' }
    ]
  },

  filterVisibleDialects(list) {
    const userInfo = wx.getStorageSync('userInfo') || {}
    if (app.hasPermission(ACTIONS.DIALECT_MANAGE, userInfo)) return list
    return (list || []).filter(item => item.status !== 'pending' && item.status !== 'inactive')
  },

  getDialectList(result) {
    if (!result) return []
    if (Array.isArray(result)) return result
    if (Array.isArray(result.list)) return result.list
    if (Array.isArray(result.data)) return result.data
    if (result.data && Array.isArray(result.data.list)) return result.data.list
    return []
  }
})
