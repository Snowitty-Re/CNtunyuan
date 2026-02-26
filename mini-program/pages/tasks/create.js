const { post, get } = require('../../utils/request')
const { showLoading, hideLoading, showSuccess, showError } = require('../../utils/util')

Page({
  data: {
    form: {
      title: '',
      typeIndex: 0,
      priorityIndex: 2,
      caseId: '',
      assigneeId: '',
      location: '',
      latitude: '',
      longitude: '',
      deadline: '',
      description: '',
      requirements: ''
    },
    typeOptions: [
      { value: 'search', label: '实地寻访' },
      { value: 'call', label: '电话核实' },
      { value: 'info_collect', label: '信息收集' },
      { value: 'dialect_record', label: '方言录制' },
      { value: 'coordination', label: '协调沟通' },
      { value: 'other', label: '其他' }
    ],
    priorityOptions: [
      { value: 'urgent', label: '紧急' },
      { value: 'high', label: '高' },
      { value: 'normal', label: '普通' },
      { value: 'low', label: '低' }
    ],
    caseOptions: [{ name: '不关联' }],
    userOptions: [{ nickname: '暂不分配' }],
    selectedCase: null,
    selectedUser: null,
    dateTimeArray: [],
    dateTime: [0, 0, 0, 0, 0],
    submitting: false
  },

  onLoad(options) {
    this.initDateTimePicker()
    this.loadCases()
    this.loadUsers()

    if (options.caseId) {
      this.setData({ 'form.caseId': options.caseId })
    }
  },

  initDateTimePicker() {
    const now = new Date()
    const years = []
    const months = []
    const days = []
    const hours = []
    const minutes = []

    for (let i = now.getFullYear(); i <= now.getFullYear() + 1; i++) {
      years.push(i + '年')
    }
    for (let i = 1; i <= 12; i++) {
      months.push(i + '月')
    }
    for (let i = 1; i <= 31; i++) {
      days.push(i + '日')
    }
    for (let i = 0; i < 24; i++) {
      hours.push(i + '时')
    }
    for (let i = 0; i < 60; i += 5) {
      minutes.push(i + '分')
    }

    this.setData({
      dateTimeArray: [years, months, days, hours, minutes]
    })
  },

  async loadCases() {
    try {
      const result = await get('/missing-persons', { status: 'searching', page: 1, page_size: 50 })
      this.setData({
        caseOptions: [{ name: '不关联案件' }, ...result.list]
      })
    } catch (error) {
      console.error('加载案件失败:', error)
    }
  },

  async loadUsers() {
    try {
      const result = await get('/users', { page: 1, page_size: 50 })
      this.setData({
        userOptions: [{ nickname: '暂不分配' }, ...result.list]
      })
    } catch (error) {
      console.error('加载用户失败:', error)
    }
  },

  onInput(e) {
    const { field } = e.currentTarget.dataset
    const { value } = e.detail
    this.setData({ [`form.${field}`]: value })
  },

  onTypeChange(e) {
    this.setData({ 'form.typeIndex': e.detail.value })
  },

  onPriorityChange(e) {
    this.setData({ 'form.priorityIndex': e.detail.value })
  },

  onCaseChange(e) {
    const index = e.detail.value
    const selectedCase = this.data.caseOptions[index]
    this.setData({
      selectedCase,
      'form.caseId': index > 0 ? selectedCase.id : ''
    })
  },

  onUserChange(e) {
    const index = e.detail.value
    const selectedUser = this.data.userOptions[index]
    this.setData({
      selectedUser,
      'form.assigneeId': index > 0 ? selectedUser.id : ''
    })
  },

  onDateTimeChange(e) {
    const { value } = e.detail
    const { dateTimeArray } = this.data
    const dateStr = `${dateTimeArray[0][value[0]]}${dateTimeArray[1][value[1]]}${dateTimeArray[2][value[2]]} ${dateTimeArray[3][value[3]]}${dateTimeArray[4][value[4]]}`
    this.setData({
      'form.deadline': dateStr,
      dateTime: value
    })
  },

  chooseLocation() {
    wx.chooseLocation({
      success: (res) => {
        this.setData({
          'form.location': res.name,
          'form.latitude': res.latitude,
          'form.longitude': res.longitude
        })
      }
    })
  },

  async submit() {
    const { form, typeOptions, priorityOptions } = this.data

    if (!form.title) {
      showError('请输入任务标题')
      return
    }

    this.setData({ submitting: true })
    showLoading('创建中...')

    try {
      const data = {
        title: form.title,
        type: typeOptions[form.typeIndex]?.value,
        priority: priorityOptions[form.priorityIndex]?.value,
        missing_person_id: form.caseId || undefined,
        assignee_id: form.assigneeId || undefined,
        location: form.location,
        latitude: form.latitude,
        longitude: form.longitude,
        description: form.description,
        requirements: form.requirements
      }

      await post('/tasks', data)
      hideLoading()
      showSuccess('创建成功')
      setTimeout(() => {
        wx.navigateBack()
      }, 1500)
    } catch (error) {
      hideLoading()
      this.setData({ submitting: false })
      showError('创建失败')
    }
  }
})
