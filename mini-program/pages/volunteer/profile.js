const userService = require('../../services/user')
const organizationService = require('../../services/organization')
const authService = require('../../services/auth')
const { showConfirm, showSuccess, showToast } = require('../../utils/util')
const { ROLE_MAP } = require('../../utils/constants')
const app = getApp()
const CASES_STATUS_FILTER_KEY = 'cases_status_filter'

Page({
  data: {
    // 用户信息
    userInfo: {
      id: '',
      avatar: '',
      nickname: '',
      realName: '',
      role: 'volunteer',
      orgName: ''  // 使用 org_name 而非 org 对象
    },
    
    roleMap: ROLE_MAP,
    
    // 角色等级颜色
    roleColorMap: {
      super_admin: '#E74C3C',
      admin: '#E67E22',
      manager: '#3498DB',
      volunteer: '#27AE60'
    },
    
    // 统计数据（使用后端实际返回的字段）
    stats: {
      totalTasks: 0,      // 总任务数 (total_tasks)
      totalCases: 0,      // 总案件数 (total_cases)
      activeCases: 0,     // 进行中案件 (active_cases)
      completedCases: 0   // 已完成案件 (completed_cases)
    },
    
    // 功能菜单
    menuList: [
      { icon: 'edit', text: '编辑资料', url: '/pages/volunteer/edit-profile', type: 'navigate' },
      { icon: 'user', text: '绑定微信', type: 'action', action: 'bindWechat' },
      { icon: 'task', text: '我的任务', url: '/pages/tasks/my', type: 'navigate' },
      { icon: 'notification', text: '消息通知', url: '/pages/notification/list', type: 'navigate', badge: 0 },
      { icon: 'certificate', text: '志愿者证书', url: '', type: 'toast' },
      { icon: 'settings', text: '设置', url: '/pages/settings/index', type: 'navigate' }
    ]
  },

  onLoad() {
    if (!app.ensureAuth || !app.ensureAuth()) return
    this.loadData()
  },

  onShow() {
    if (!app.ensureAuth || !app.ensureAuth()) return
    this.loadData()
  },

  onPullDownRefresh() {
    this.loadData().finally(() => {
      wx.stopPullDownRefresh()
    })
  },

  // 加载所有数据
  async loadData() {
    try {
      await Promise.all([
        this.loadUserInfo(),
        this.loadStats()
      ])
    } catch (error) {
      console.error('加载数据失败:', error)
    }
  },

  // 加载用户信息
  async loadUserInfo() {
    try {
      const userInfo = await app.getUserInfo() || wx.getStorageSync('userInfo') || {}
      const profile = await userService.getProfile().catch(() => ({}))

      const resolveOrgName = (src) => {
        if (!src) return ''
        return src.org_name ||
          src.orgName ||
          src.organization_name ||
          (src.org && src.org.name) ||
          (src.organization && src.organization.name) ||
          ''
      }

      let orgName = resolveOrgName(userInfo) || resolveOrgName(profile)
      const orgID = userInfo.org_id || userInfo.orgId || profile.org_id || profile.orgId || ''
      if (!orgName && orgID) {
        const org = await organizationService.getById(orgID).catch(() => null)
        orgName = (org && org.name) || ''
      }
      
      const mergedUserInfo = {
        ...userInfo,
        ...profile,
        id: userInfo.id || profile.id || '',
        avatar: userInfo.avatar || profile.avatar || '/assets/images/avatar-default.png',
        nickname: userInfo.nickname || profile.nickname || '志愿者',
        realName: userInfo.real_name || profile.real_name || '',
        role: userInfo.role || profile.role || 'volunteer',
        // 兼容 org_name / organization_name / org.name，最后兜底按 org_id 回查
        orgName: orgName || '未知组织',
        wxBound: !!(userInfo.wx_bound || profile.wx_bound)
      }

      const menuList = (this.data.menuList || []).map((item) => {
        if (item.action === 'bindWechat') {
          return {
            ...item,
            text: mergedUserInfo.wxBound ? '已绑定微信' : '绑定微信'
          }
        }
        return item
      })

      this.setData({ userInfo: mergedUserInfo, menuList })
      wx.setStorageSync('userInfo', mergedUserInfo)
    } catch (error) {
      console.error('加载用户信息失败:', error)
    }
  },

  // 加载统计数据
  async loadStats() {
    try {
      const userStats = await userService.getStats().catch(() => ({}))
      
      this.setData({
        stats: {
          totalTasks:     userStats.totalTasks     || 0,
          totalCases:     userStats.totalCases     || 0,
          activeCases:    userStats.activeCases    || 0,
          completedCases: userStats.completedCases || 0
        }
      })
    } catch (error) {
      console.error('加载统计失败:', error)
    }
  },

  // 点击统计卡片
  onStatTap(e) {
    const { type } = e.currentTarget.dataset
    switch (type) {
      case 'task':
        wx.navigateTo({ url: '/pages/tasks/my' })
        break
      case 'case':
        wx.switchTab({ url: '/pages/cases/list' })
        break
      case 'activeCase':
        wx.setStorageSync(CASES_STATUS_FILTER_KEY, 'searching')
        wx.switchTab({ url: '/pages/cases/list' })
        break
      case 'completed':
        wx.setStorageSync(CASES_STATUS_FILTER_KEY, 'found')
        wx.switchTab({ url: '/pages/cases/list' })
        break
    }
  },

  // 菜单点击
  onMenuTap(e) {
    const { index } = e.currentTarget.dataset
    const menu = this.data.menuList[index]
    
    if (!menu) return
    
    switch (menu.type) {
      case 'navigate':
        if (menu.url) {
          wx.navigateTo({ url: menu.url })
        }
        break
      case 'toast':
        showToast('功能开发中，敬请期待')
        break
      case 'switchTab':
        if (menu.url) {
          wx.switchTab({ url: menu.url })
        }
        break
      case 'action':
        if (menu.action === 'bindWechat') {
          this.bindWechat()
        }
        break
    }
  },

  async bindWechat() {
    try {
      if (this.data.userInfo.wxBound) {
        showToast('当前账号已绑定微信')
        return
      }
      const loginRes = await wx.login()
      if (!loginRes.code) {
        showToast('获取微信授权失败')
        return
      }
      await authService.bindWechat(loginRes.code)

      const userInfo = {
        ...(this.data.userInfo || {}),
        wxBound: true,
        wx_bound: true
      }
      this.setData({ userInfo })
      wx.setStorageSync('userInfo', userInfo)
      showSuccess('微信绑定成功')
    } catch (error) {
      console.error('绑定微信失败:', error)
      showToast(error.message || '绑定失败')
    }
  },

  // 复制用户ID
  copyUserId() {
    const { id } = this.data.userInfo
    if (!id) {
      showToast('用户ID获取失败')
      return
    }
    
    wx.setClipboardData({
      data: id,
      success: () => {
        showSuccess('已复制用户ID')
      }
    })
  },

  // 退出登录
  async logout() {
    const confirm = await showConfirm('确认退出', '退出后需要重新登录')
    if (!confirm) return
    
    try {
      await app.logout()
    } catch (error) {
      console.error('退出登录接口调用失败:', error)
      // 即使API调用失败，也清除本地数据
      app.clearLoginData()
    }

    // 跳转到登录页
    wx.reLaunch({ url: '/pages/login/index' })
  }
})
