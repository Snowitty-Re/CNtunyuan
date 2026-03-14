const services = require('../../services/index')
const { showError, showToast } = require('../../utils/util')

Page({
  data: {
    // 案件列表
    cases: [],
    
    // 当前选中的案件
    selectedCase: null,
    
    // 页面状态
    loading: false,
    showCaseList: false,
    
    // 筛选条件
    status: '',
    keyword: ''
  },

  onLoad(options) {
    // 加载案件数据
    this.loadCases()
  },

  // 加载案件数据
  async loadCases() {
    this.setData({ loading: true })
    
    try {
      const params = {
        page: 1,
        page_size: 100,
        status: this.data.status || undefined
      }
      
      const result = await services.missingPerson.getList(params)
      const cases = (result.list || []).map(item => ({
        ...item,
        // 拼接地址信息用于显示
        displayLocation: [item.province, item.city, item.district, item.address].filter(Boolean).join(' ') || '未知位置'
      }))
      
      this.setData({
        cases,
        loading: false
      })
    } catch (error) {
      console.error('加载案件失败:', error)
      this.setData({ loading: false })
      showError('加载失败')
    }
  },

  // 获取状态标签颜色
  getStatusColor(status) {
    const colorMap = {
      'missing': '#ff4d4f',
      'searching': '#faad14',
      'found': '#52c41a',
      'reunited': '#52c41a'
    }
    return colorMap[status] || '#999'
  },

  // 显示案件列表
  showCaseList() {
    this.setData({ showCaseList: true })
  },

  // 隐藏案件列表
  hideCaseList() {
    this.setData({ showCaseList: false, selectedCase: null })
  },

  // 选择案件
  selectCase(e) {
    const id = e.currentTarget.dataset.id
    const caseItem = this.data.cases.find(c => c.id === id)
    if (caseItem) {
      this.setData({
        selectedCase: caseItem,
        showCaseList: true
      })
    }
  },

  // 筛选状态改变
  onStatusChange(e) {
    const status = e.currentTarget.dataset.status
    this.setData({ status }, () => {
      this.loadCases()
    })
  },

  // 搜索输入
  onSearchInput(e) {
    this.setData({ keyword: e.detail.value })
  },

  // 搜索
  onSearch() {
    const { keyword, cases } = this.data
    
    if (!keyword) {
      this.loadCases()
      return
    }
    
    // 本地筛选
    const filtered = cases.filter(item => 
      item.name.includes(keyword) ||
      (item.displayLocation && item.displayLocation.includes(keyword))
    )
    
    this.setData({
      cases: filtered,
      showCaseList: true
    })
  },

  // 跳转到案件详情
  goToCaseDetail(e) {
    const id = e.currentTarget.dataset.id
    wx.navigateTo({
      url: `/pages/cases/detail?id=${id}`
    })
  },

  // 跳转到创建案件
  goToCreateCase() {
    wx.navigateTo({
      url: '/pages/cases/create'
    })
  },

  // 刷新数据
  onRefresh() {
    this.loadCases()
  },

  // 拨打电话
  makePhoneCall(e) {
    const phone = e.currentTarget.dataset.phone
    if (!phone) {
      showToast('暂无联系电话')
      return
    }
    wx.makePhoneCall({ phoneNumber: phone })
  }
})
