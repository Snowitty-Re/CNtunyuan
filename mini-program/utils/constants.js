/**
 * 全局常量 — 状态映射、优先级映射等
 * 各页面从此处引入，避免重复定义
 */

const TASK_STATUS_MAP = {
  draft:      '草稿',
  pending:    '待分配',
  assigned:   '已分配',
  processing: '进行中',
  completed:  '已完成',
  cancelled:  '已取消'
}

const TASK_PRIORITY_MAP = {
  urgent: '紧急',
  high:   '高',
  medium: '普通',
  normal: '普通',
  low:    '低'
}

const TASK_PRIORITY_COLOR = {
  urgent: '#ff4d4f',
  high:   '#faad14',
  medium: '#1890ff',
  normal: '#1890ff',
  low:    '#52c41a'
}

const TASK_TYPE_MAP = {
  search:    '搜寻',
  verify:    '核实',
  assist:    '协助',
  follow:    '跟进',
  interview: '寻访',
  other:     '其他'
}

const CASE_STATUS_MAP = {
  missing:   '失踪中',
  searching: '寻找中',
  found:     '已找到',
  reunited:  '已团圆',
  closed:    '已结案'
}

const CASE_STATUS_COLOR = {
  missing:   '#ff4d4f',
  searching: '#faad14',
  found:     '#52c41a',
  reunited:  '#52c41a',
  closed:    '#999999'
}

const GENDER_MAP = {
  male:   '男',
  female: '女',
  other:  '其他'
}

const ROLE_MAP = {
  super_admin: '超级管理员',
  admin:       '管理员',
  manager:     '负责人',
  volunteer:   '志愿者'
}

module.exports = {
  TASK_STATUS_MAP,
  TASK_PRIORITY_MAP,
  TASK_PRIORITY_COLOR,
  TASK_TYPE_MAP,
  CASE_STATUS_MAP,
  CASE_STATUS_COLOR,
  GENDER_MAP,
  ROLE_MAP
}
