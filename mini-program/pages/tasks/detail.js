const { get, post } = require('../../utils/request')
const { formatDate, showLoading, hideLoading, showSuccess, showConfirm } = require('../../utils/util')

Page({
  data: {
    task: {},
    feedbacks: [],
    markers: [],
    statusMap: {
      draft: '草稿',
      pending: '待分配',
      assigned: '已分配',
      processing: '进行中',
      completed: '已完成',
      cancelled: '已取消'
    },
    priorityMap: {
      urgent: '紧急',
      high: '高',
      normal: '普通',
      low: '低'
    },
    taskTypeMap: {
      search: '实地寻访',
      call: '电话核实',
      info_collect: '信息收集',
      dialect_record: '方言录制',
      coordination: '协调沟通',
      other: '其他'
    },
    actionText: {
      draft: '编辑任务',
      pending: '领取任务',
      assigned: '开始执行',
      processing: '更新进度',
      completed: '已完成',
      cancelled: '重新激活'
    }
  },

  onLoad(options) {
    const { id } = options
    if (id) {
      this.loadTaskDetail(id)
    }
  },

  async loadTaskDetail(id) {
    showLoading()
    try {
      const data = await get(`/tasks/${id}`)
      data.deadline = data.deadline ? formatDate(data.deadline) : null

      const markers = []
      if (data.latitude && data.longitude) {
        markers.push({
          id: 1,
          latitude: data.latitude,
          longitude: data.longitude,
          title: '任务地点'
        })
      }

      this.setData({
        task: data,
        markers
      })

      this.loadFeedbacks(id)
    } catch (error) {
      console.error('加载任务详情失败:', error)
    } finally {
      hideLoading()
    }
  },

  async loadFeedbacks(taskId) {
    try {
      // 这里应该调用反馈列表接口
      this.setData({
        feedbacks: [
          {
            id: 1,
            user: { nickname: '志愿者A', avatar: '' },
            content: '已经联系到家属，正在核实信息',
            images: [],
            created_at: '2024-02-26 10:30'
          }
        ]
      })
    } catch (error) {
      console.error('加载反馈失败:', error)
    }
  },

  goToCase() {
    const id = this.data.task.missing_person?.id
    if (id) {
      wx.navigateTo({ url: `/pages/cases/detail?id=${id}` })
    }
  },

  async handleAction() {
    const { task } = this.data
    const statusActions = {
      pending: async () => {
        await post(`/tasks/${task.id}/assign`, { assignee_id: 'current_user_id' })
        showSuccess('领取成功')
        this.loadTaskDetail(task.id)
      },
      assigned: async () => {
        await post(`/tasks/${task.id}/start`)
        showSuccess('开始执行')
        this.loadTaskDetail(task.id)
      },
      processing: () => {
        wx.showModal({
          title: '更新进度',
          editable: true,
          placeholderText: '输入进度百分比(0-100)',
          success: (res) => {
            if (res.confirm && res.content) {
              const progress = parseInt(res.content)
              if (progress >= 0 && progress <= 100) {
                post(`/tasks/${task.id}/progress`, { progress }).then(() => {
                  showSuccess('进度更新成功')
                  this.loadTaskDetail(task.id)
                })
              }
            }
          }
        })
      },
      completed: () => {
        showSuccess('任务已完成')
      }
    }

    const action = statusActions[task.status]
    if (action) {
      action()
    }
  },

  addFeedback() {
    wx.showModal({
      title: '添加反馈',
      editable: true,
      placeholderText: '请输入反馈内容',
      success: (res) => {
        if (res.confirm && res.content) {
          showSuccess('反馈已添加')
        }
      }
    })
  },

  previewImage(e) {
    const { url, urls } = e.currentTarget.dataset
    wx.previewImage({
      current: url,
      urls
    })
  }
})
