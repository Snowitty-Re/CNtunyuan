Component({
  properties: {
    status: { type: String, value: '' },
    type: { type: String, value: 'case' } // case, task
  },

  data: {
    label: '',
    colorClass: ''
  },

  observers: {
    'status, type': function(status, type) {
      const maps = {
        case: {
          missing: { label: '失踪中', colorClass: 'danger' },
          searching: { label: '寻找中', colorClass: 'info' },
          found: { label: '已找到', colorClass: 'success' },
          reunited: { label: '已团圆', colorClass: 'success' },
          closed: { label: '已结案', colorClass: 'default' }
        },
        task: {
          draft: { label: '草稿', colorClass: 'default' },
          pending: { label: '待处理', colorClass: 'warning' },
          assigned: { label: '已分配', colorClass: 'info' },
          processing: { label: '进行中', colorClass: 'primary' },
          completed: { label: '已完成', colorClass: 'success' },
          cancelled: { label: '已取消', colorClass: 'default' }
        }
      }
      const map = maps[type] || maps.case
      const info = map[status] || { label: status || '未知', colorClass: 'default' }
      this.setData({ label: info.label, colorClass: info.colorClass })
    }
  }
})
