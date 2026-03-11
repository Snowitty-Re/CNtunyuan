export const USER_ROLE_MAP: Record<string, string> = {
  super_admin: '超级管理员',
  admin: '管理员',
  manager: '管理者',
  volunteer: '志愿者',
};

export const USER_STATUS_MAP: Record<string, string> = {
  active: '正常',
  inactive: '未激活',
  banned: '已禁用',
};

export const MISSING_PERSON_STATUS_MAP: Record<string, string> = {
  missing: '走失',
  searching: '搜寻中',
  found: '已找到',
  reunited: '已团圆',
  closed: '已关闭',
};

export const URGENCY_MAP: Record<string, string> = {
  urgent: '紧急',
  high: '较高',
  normal: '一般',
  low: '较低',
};

export const TASK_STATUS_MAP: Record<string, string> = {
  draft: '草稿',
  pending: '待处理',
  assigned: '已分配',
  processing: '进行中',
  completed: '已完成',
  cancelled: '已取消',
};

export const TASK_PRIORITY_MAP: Record<string, string> = {
  urgent: '紧急',
  high: '高',
  normal: '普通',
  low: '低',
};

export const DIALECT_STATUS_MAP: Record<string, string> = {
  pending: '待审核',
  active: '已通过',
  rejected: '已拒绝',
};

export const FILE_TYPE_MAP: Record<string, string> = {
  image: '图片',
  audio: '音频',
  video: '视频',
  document: '文档',
  other: '其他',
};

export const AUDIT_TYPE_MAP: Record<string, string> = {
  login: '登录',
  logout: '登出',
  create: '创建',
  update: '更新',
  delete: '删除',
  query: '查询',
  export: '导出',
  upload: '上传',
  download: '下载',
  other: '其他',
};

export const AUDIT_STATUS_MAP: Record<string, string> = {
  success: '成功',
  failure: '失败',
};

export const ORG_TYPE_MAP: Record<string, string> = {
  province: '省级',
  city: '市级',
  district: '区县级',
  station: '站点',
  team: '团队',
};

export const GENDER_MAP: Record<string, string> = {
  male: '男',
  female: '女',
  unknown: '未知',
};

export const ERROR_CODE_MAP: Record<number, string> = {
  1000: '用户不存在',
  1001: '用户已存在',
  1002: '密码错误',
  1003: '无效的认证令牌',
  1004: '认证令牌已过期',
  1005: '验证码错误',
  1006: '账号已禁用',
  1007: '账号已锁定',
  2000: '组织不存在',
  2001: '组织已存在',
  2002: '该组织下有子组织',
  2003: '该组织下有用户',
  3000: '案件不存在',
  3001: '案件已存在',
  3002: '案件已关闭',
  4000: '任务不存在',
  4001: '任务已存在',
  4002: '任务已分配',
  4003: '任务已完成',
  5000: '方言不存在',
  5001: '方言已存在',
  6000: '文件不存在',
  6001: '文件过大',
  6002: '不支持的文件类型',
  6003: '上传失败',
};
