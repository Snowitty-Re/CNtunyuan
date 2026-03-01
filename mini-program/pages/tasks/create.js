const { get, post } = require('../../utils/request')
const { formatDate, showSuccess, showToast } = require('../../utils/util')

Page({
  data: {
    form: {
      title: '',
      description: '',
      task_type: 'search',
      priority: 'normal',
      deadline: '',
      address: '',
      longitude: '',
      latitude: '',
      missing_person_id: ''
    },
    taskTypes: [
      { key: 'search', label: '实地寻访' },
      { key: 'call', label: '电话核实' },
      { key: 'info_collect', label: '信息收集' },
      { key: 'dialect_record', label: '方言录制' },
      { key: 'coordination', label: '协调沟通' },
      { key: 'other', label: '其他' }
    ],
    priorities: [
      { key: 'urgent', label: '紧急' },
      { key: 'high', label: '高' },
      { key: 'normal', label: '普通' },
      { key: 'low', label: '低' }
    ],
    typeIndex: 0,
    priorityIndex: 2,
    caseIndex: -1,
    typeName: '实地寻访',
    priorityName: '普通',
    caseName: '',
    cases: [],
    loading: false
  },

  onLoad() {
    this.loadCases()
  },

  async loadCases() {
    try {
      const result = await get('/missing-persons', { page_size: 100 })
      this.setData({ cases: result.list || [] })
    } catch (error) {
      console.error('加载案件失败:', error)
    }
  },

  onInput(e) {
    const { field } = e.currentTarget.dataset
    this.setData({ [`form.${field}`]: e.detail.value })
  },

  onTypeChange(e) {
    const index = e.detail.value
    const item = this.data.taskTypes[index]
    this.setData({ 
      'form.task_type': item.key,
      typeIndex: index,
      typeName: item.label
    })
  },

  onPriorityChange(e) {
    const index = e.detail.value
    const item = this.data.priorities[index]
    this.setData({ 
      'form.priority': item.key,
      priorityIndex: index,
      priorityName: item.label
    })
  },

  onCaseChange(e) {
    const index = e.detail.value
    const caseItem = this.data.cases[index]
    this.setData({ 
      'form.missing_person_id': caseItem.id,
      'form.address': caseItem.last_seen_location || '',
      caseIndex: index,
      caseName: caseItem.name
    })
  },

  onDateChange(e) {
    this.setData({ 'form.deadline': e.detail.value })
  },

  chooseLocation() {
    wx.chooseLocation({
      success: (res) => {
        this.setData({
          'form.address': res.address,
          'form.longitude': String(res.longitude),
          'form.latitude': String(res.latitude)
        })
      }
    })
  },

  async submit() {
    const { form } = this.data
    
    if (!form.title.trim()) {
      showToast('请输入任务标题')
      return
    }

    this.setData({ loading: true })

    try {
      const data = { ...form }
      if (data.deadline) {
        data.deadline = data.deadline + 'T23:59:59Z'
      }
      
      await post('/tasks', data)
      showSuccess('创建成功')
      wx.navigateBack()
    } catch (error) {
      console.error('创建失败:', error)
      showToast('创建失败')
    } finally {
      this.setData({ loading: false })
    }
  }
})
