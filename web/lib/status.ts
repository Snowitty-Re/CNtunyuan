const CASE_STATUS: Record<string, string> = {
  missing: '失踪中',
  searching: '寻访中',
  found: '已找到',
  reunited: '已团圆',
  closed: '已关闭',
}

const TASK_STATUS: Record<string, string> = {
  draft: '草稿',
  pending: '待处理',
  assigned: '已分配',
  processing: '进行中',
  in_progress: '进行中',
  completed: '已完成',
  cancelled: '已取消',
  overdue: '已逾期',
}

const USER_STATUS: Record<string, string> = {
  active: '正常',
  inactive: '待审核',
  banned: '禁用',
}

export function caseStatusLabel(status?: string | null): string {
  const key = String(status || '').trim()
  return CASE_STATUS[key] || key || '-'
}

export function taskStatusLabel(status?: string | null): string {
  const key = String(status || '').trim()
  return TASK_STATUS[key] || key || '-'
}

export function userStatusLabel(status?: string | null): string {
  const key = String(status || '').trim()
  return USER_STATUS[key] || key || '-'
}

export function taskStatusColor(status?: string | null): string {
  const key = String(status || '').trim()
  const map: Record<string, string> = {
    draft: 'default',
    pending: 'warning',
    assigned: 'processing',
    processing: 'processing',
    in_progress: 'processing',
    completed: 'success',
    cancelled: 'default',
    overdue: 'error',
  }
  return map[key] || 'default'
}
