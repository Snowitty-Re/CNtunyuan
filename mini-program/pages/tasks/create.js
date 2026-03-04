const taskService = require('../../services/task')
const missingPersonService = require('../../services/missingPerson')
const userService = require('../../services/user')
const { formatDate, showSuccess, showToast } = require('../../utils/util')

Page({
  data: {
    form: {
      title: '',
      description: '',
      priority: 'normal',
      deadline: '',
      missing_person_id: '',
      assignee_id: ''
    },
    priorities: [
      { key: 'urgent', label: '紧急' },
      { key: 'high', label: '高' },
      { key: 'normal', label: '普通' },
      { key: 'low', label: '低' }
    ],
    priorityIndex: 2,
    cases: [],
    caseIndex: -1,
    users: [],
    userIndex: -1,
    loading: false,
    submitting: false,
    minDate: ''
  },

  onLoad() {
    // 检查权限
    const userInfo = wx.getStorageSync('userInfo') || {}
    if (!['super_admin', 'admin', 'manager'].includes(userInfo.role)) {
      showToast('无权限创建任务')
      wx.navigateBack()
      return
    }

    // 设置最小日期为今天
    const today = new Date()
    this.setData({
      minDate: formatDate(today)
    })

    this.loadCases()
    this.loadUsers()
  },

  // 加载案件列表
  async loadCases() {
    this.setData({ loading: true })
    try {
      const result = await missingPersonService.getList({ 
        page: 1, 
        page_size: 100,
        status: 'searching'
      })
      const cases = result.list || result || []
      this.setData({ 
        cases,
        loading: false 
      })
    } catch (error) {
      console.error('加载案件失败:', error)
      showToast('加载案件失败')
      this.setData({ loading: false })
    }
  },

  // 加载用户列表（用于分配）
  async loadUsers() {
    try {
      const result = await userService.getList({ 
        page: 1, 
        page_size: 100,
        status: 'active'
      })
      const users = result.list || result || []
      this.setData({ users })
    } catch (error) {
      console.error('加载用户失败:', error)
    }
  },

  // 输入框变化
  onInput(e) {
    const { field } = e.currentTarget.dataset
    const { value } = e.detail
    this.setData({ [`form.${field}`]: value })
  },

  // 优先级选择
  onPriorityChange(e) {
    const index = parseInt(e.detail.value)
    const item = this.data.priorities[index]
    this.setData({
      priorityIndex: index,
      'form.priority': item.key
    })
  },

  // 日期选择
  onDateChange(e) {
    this.setData({ 'form.deadline': e.detail.value })
  },

  // 案件选择
  onCaseChange(e) {
    const index = parseInt(e.detail.value)
    const caseItem = this.data.cases[index]
    this.setData({
      caseIndex: index,
      'form.missing_person_id': caseItem.id
    })
  },

  // 执行人选择
  onUserChange(e) {
    const index = parseInt(e.detail.value)
    const user = this.data.users[index]
    this.setData({
      userIndex: index,
      'form.assignee_id': user.id
    })
  },

  // 验证表单
  validateForm() {
    const { form } = this.data
    
    if (!form.title.trim()) {
      showToast('请输入任务标题')
      return false
    }
    
    if (form.title.trim().length < 2) {
      showToast('任务标题至少2个字')
      return false
    }

    if (!form.description.trim()) {
      showToast('请输入任务描述')
      return false
    }

    if (!form.deadline) {
      showToast('请选择截止日期')
      return false
    }

    return true
  },

  // 提交创建
  async submit() {
    if (!this.validateForm()) return

    this.setData({ submitting: true })

    try {
      const data = {
        ...this.data.form,
        title: this.data.form.title.trim(),
        description: this.data.form.description.trim()
      }

      // 格式化截止日期
      if (data.deadline) {
        data.deadline = data.deadline + 'T23:59:59+08:00'
      }

      await taskService.create(data)
      showSuccess('创建成功')
      
      // 返回上一页并刷新
      const pages = getCurrentPages()
      const prevPage = pages[pages.length - 2]
      if (prevPage && prevPage.loadTasks) {
        prevPage.loadTasks()
      }
      
      wx.navigateBack()
    } catch (error) {
      console.error('创建任务失败:', error)
      showToast('创建失败')
    } finally {
      this.setData({ submitting: false })
    }
  }
})
