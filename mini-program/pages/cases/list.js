const missingPersonService = require('../../services/missingPerson')
const { formatDate, showLoading, hideLoading, debounce } = require('../../utils/util')

// 状态映射
const STATUS_MAP = {
  '': { label: '全部', color: '' },
  'missing': { label: '失踪中', color: '#ff4d4f' },
  'searching': { label: '寻找中', color: '#1890ff' },
  'found': { label: '已找到', color: '#52c41a' },
  'reunited': { label: '已团圆', color: '#52c41a' }
}

// 案件类型映射
const CASE_TYPE_MAP = {
  'elderly': '老人',
  'child': '儿童',
  'adult': '成人',
  'disability': '残障',
  'other': '其他'
}

Page({
  data: {
    // 列表数据
    cases: [],
    
    // 分页参数
    page: 1,
    pageSize: 10,
    loading: false,
    noMore: false,
    
    // 搜索和筛选
    keyword: '',
    currentStatus: '',
    
    // 状态选项
    statusTabs: [
      { value: '', label: '全部' },
      { value: 'missing', label: '失踪中' },
      { value: 'searching', label: '寻找中' },
      { value: 'found', label: '已找到' },
      { value: 'reunited', label: '已团圆' }
    ],
    
    // 映射常量
    statusMap: STATUS_MAP,
    caseTypeMap: CASE_TYPE_MAP
  },

  onLoad() {
    this.loadCases()
  },

  onPullDownRefresh() {
    this.setData({ page: 1, noMore: false })
    this.loadCases().finally(() => {
      wx.stopPullDownRefresh()
    })
  },

  onReachBottom() {
    if (!this.data.noMore && !this.data.loading) {
      this.loadMore()
    }
  },

  /**
   * 加载案件列表
   */
  async loadCases() {
    if (this.data.loading) return

    this.setData({ loading: true })
    if (this.data.page === 1) {
      showLoading()
    }

    try {
      const params = {
        page: this.data.page,
        page_size: this.data.pageSize,
        status: this.data.currentStatus,
        keyword: this.data.keyword
      }

      const result = await missingPersonService.getList(params)
      const list = result.list || result || []
      
      const cases = list.map(item => ({
        ...item,
        missing_time: formatDate(item.missing_time),
        photoUrl: this.getFirstPhoto(item.photos)
      }))

      this.setData({
        cases: this.data.page === 1 ? cases : [...this.data.cases, ...cases],
        noMore: cases.length < this.data.pageSize
      })
    } catch (error) {
      console.error('加载案件失败:', error)
      wx.showToast({
        title: '加载失败，请重试',
        icon: 'none'
      })
    } finally {
      hideLoading()
      this.setData({ loading: false })
    }
  },

  /**
   * 加载更多
   */
  loadMore() {
    this.setData({ page: this.data.page + 1 })
    this.loadCases()
  },

  /**
   * 获取第一张图片
   */
  getFirstPhoto(photos) {
    if (photos && photos.length > 0) {
      return photos[0].url || photos[0]
    }
    return '/assets/images/default-avatar.png'
  },

  /**
   * 搜索输入（防抖）
   */
  onSearchInput: debounce(function(e) {
    const keyword = e.detail.value
    this.setData({ keyword, page: 1, cases: [] })
    this.loadCases()
  }, 500),

  /**
   * 搜索确认
   */
  onSearch() {
    this.setData({ page: 1, cases: [] })
    this.loadCases()
  },

  /**
   * 切换状态筛选
   */
  switchStatus(e) {
    const status = e.currentTarget.dataset.status
    if (status === this.data.currentStatus) return
    
    this.setData({ 
      currentStatus: status, 
      page: 1, 
      cases: [],
      noMore: false
    })
    this.loadCases()
  },

  /**
   * 跳转到详情页
   */
  goToDetail(e) {
    const id = e.currentTarget.dataset.id
    wx.navigateTo({ 
      url: `/pages/cases/detail?id=${id}` 
    })
  },

  /**
   * 跳转到创建页
   */
  goToCreate() {
    wx.navigateTo({ 
      url: '/pages/cases/create' 
    })
  }
})
