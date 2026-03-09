// ===== Status Maps =====
export const statusMap: Record<string, { label: string; color: string }> = {
  missing: { label: '走失', color: 'red' },
  searching: { label: '搜寻中', color: 'orange' },
  found: { label: '已找到', color: 'blue' },
  reunited: { label: '已团圆', color: 'green' },
  closed: { label: '已关闭', color: 'default' },
}

export const urgencyMap: Record<string, { label: string; color: string }> = {
  low: { label: '普通', color: 'default' },
  medium: { label: '中等', color: 'blue' },
  high: { label: '紧急', color: 'orange' },
  urgent: { label: '特急', color: 'red' },
}

export const taskStatusMap: Record<string, { label: string; color: string }> = {
  draft: { label: '草稿', color: 'default' },
  pending: { label: '待分配', color: 'gold' },
  assigned: { label: '已分配', color: 'cyan' },
  processing: { label: '进行中', color: 'blue' },
  completed: { label: '已完成', color: 'green' },
  cancelled: { label: '已取消', color: 'red' },
  overdue: { label: '已逾期', color: 'magenta' },
}

export const taskTypeMap: Record<string, string> = {
  search: '搜索', verify: '核实', assist: '协助', follow: '跟进', interview: '访谈', other: '其他',
}

export const taskPriorityMap: Record<string, { label: string; color: string }> = {
  low: { label: '低', color: 'default' },
  medium: { label: '中', color: 'blue' },
  high: { label: '高', color: 'orange' },
  urgent: { label: '紧急', color: 'red' },
}

export const roleMap: Record<string, { label: string; color: string }> = {
  super_admin: { label: '超级管理员', color: 'purple' },
  admin: { label: '管理员', color: 'blue' },
  manager: { label: '管理者', color: 'cyan' },
  volunteer: { label: '志愿者', color: 'green' },
}

export const userStatusMap: Record<string, { text: string; status: 'success' | 'default' | 'error' }> = {
  active: { text: '正常', status: 'success' },
  inactive: { text: '禁用', status: 'default' },
  banned: { text: '封禁', status: 'error' },
}

export const dialectTypeMap: Record<string, string> = {
  phrase: '短语', story: '故事', song: '歌曲', daily: '日常', other: '其他',
}

export const dialectStatusMap: Record<string, { label: string; color: string }> = {
  active: { label: '已发布', color: 'green' },
  inactive: { label: '已下架', color: 'default' },
  pending: { label: '待审核', color: 'orange' },
}

export const orgTypeMap: Record<string, { label: string; color: string }> = {
  root: { label: '总部', color: 'red' },
  province: { label: '省级', color: 'orange' },
  city: { label: '市级', color: 'blue' },
  district: { label: '区级', color: 'cyan' },
  street: { label: '街道', color: 'green' },
  community: { label: '社区', color: 'lime' },
  team: { label: '小组', color: 'default' },
}

export const auditTypeMap: Record<string, string> = {
  login: '登录', logout: '登出', create: '创建', update: '更新',
  delete: '删除', query: '查询', upload: '上传', download: '下载',
  export: '导出', other: '其他',
}

export const httpMethodColor: Record<string, string> = {
  GET: 'green', POST: 'blue', PUT: 'orange', DELETE: 'red', PATCH: 'cyan',
}

// ===== Validation Patterns =====
export const PHONE_PATTERN = /^1[3-9]\d{9}$/
export const PHONE_RULE = { pattern: PHONE_PATTERN, message: '请输入正确的手机号' }
