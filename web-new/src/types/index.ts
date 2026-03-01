// 用户相关
export interface User {
  id: string;
  nickname: string;
  real_name: string;
  avatar?: string;
  phone: string;
  email?: string;
  role: 'super_admin' | 'admin' | 'manager' | 'volunteer';
  status: 'active' | 'inactive' | 'banned';
  org_id?: string;
  org?: Organization;
  last_login?: string;
  created_at: string;
}

export interface UserProfile {
  id: string;
  user_id: string;
  gender?: string;
  birth_date?: string;
  address?: string;
  emergency_contact?: string;
  emergency_phone?: string;
  skills?: string;
  experience?: string;
  certificate_no?: string;
  join_date?: string;
}

// 组织相关
export interface Organization {
  id: string;
  name: string;
  code: string;
  type: 'root' | 'province' | 'city' | 'district' | 'street';
  level: number;
  parent_id?: string;
  parent?: Organization;
  children?: Organization[];
  leader_id?: string;
  leader?: User;
  province?: string;
  city?: string;
  district?: string;
  address?: string;
  contact?: string;
  phone?: string;
  email?: string;
  description?: string;
  status: 'active' | 'inactive';
  volunteer_count: number;
  case_count: number;
  created_at: string;
}

// 走失人员相关
export interface MissingPerson {
  id: string;
  case_no?: string;
  name: string;
  gender: 'male' | 'female' | 'other';
  age?: number;
  height?: number;
  weight?: number;
  birth_date?: string;
  case_type: 'elderly' | 'child' | 'adult' | 'disability' | 'other';
  status: 'missing' | 'searching' | 'found' | 'reunited' | 'closed';
  
  // 外貌特征
  appearance?: string;
  clothing?: string;
  special_features?: string;
  mental_status?: string;
  physical_status?: string;
  
  // 走失信息
  missing_time: string;
  missing_location: string;
  missing_longitude?: number;
  missing_latitude?: number;
  missing_detail?: string;
  possible_location?: string;
  
  // 联系人信息
  contact_name?: string;
  contact_phone?: string;
  contact_relation?: string;
  contact_address?: string;
  family_description?: string;
  
  // 关联
  reporter_id: string;
  reporter?: User;
  org_id: string;
  org?: Organization;
  photos?: MissingPhoto[];
  dialects?: Dialect[];
  
  // 结果
  found_time?: string;
  found_location?: string;
  found_detail?: string;
  
  created_at: string;
  updated_at: string;
}

export interface MissingPhoto {
  id: string;
  missing_person_id: string;
  url: string;
  type: 'normal' | 'simulated' | 'feature';
  description?: string;
  is_primary: boolean;
  created_at: string;
}

export interface MissingPersonTrack {
  id: string;
  missing_person_id: string;
  location: string;
  longitude?: number;
  latitude?: number;
  description?: string;
  reporter_id: string;
  reporter?: User;
  created_at: string;
}

// 方言相关
export interface Dialect {
  id: string;
  title: string;
  description?: string;
  audio_url: string;
  duration: number;
  file_size?: number;
  format?: string;
  
  // 地区
  province?: string;
  city?: string;
  district?: string;
  address?: string;
  longitude?: number;
  latitude?: number;
  
  // 采集人
  collector_id: string;
  collector?: User;
  org_id: string;
  org?: Organization;
  
  // 统计
  play_count: number;
  like_count: number;
  
  status: 'active' | 'inactive';
  created_at: string;
}

// 任务相关
export interface Task {
  id: string;
  task_no?: string;
  title: string;
  description?: string;
  type: 'search' | 'call' | 'info_collect' | 'dialect_record' | 'coordination' | 'other';
  priority: 'urgent' | 'high' | 'normal' | 'low';
  status: 'draft' | 'pending' | 'assigned' | 'processing' | 'completed' | 'cancelled' | 'timeout';
  
  // 关联
  missing_person_id?: string;
  missing_person?: MissingPerson;
  creator_id: string;
  creator?: User;
  assignee_id?: string;
  assignee?: User;
  org_id: string;
  org?: Organization;
  workflow_id?: string;
  
  // 时间
  deadline?: string;
  completed_at?: string;
  
  // 进度
  progress: number;
  result?: string;
  
  created_at: string;
  updated_at: string;
}

// 工作流相关
export interface Workflow {
  id: string;
  name: string;
  code: string;
  description?: string;
  type?: string;
  status: 'draft' | 'active' | 'inactive';
  version: number;
  is_default: boolean;
  creator_id: string;
  creator?: User;
  steps?: WorkflowStep[];
  created_at: string;
}

export interface WorkflowStep {
  id: string;
  workflow_id: string;
  name: string;
  description?: string;
  step_order: number;
  step_type?: string;
  assignee_type?: string;
  assignee_role?: string;
  duration?: number;
  form_config?: string;
  next_step_id?: string;
  prev_step_id?: string;
}

export interface WorkflowInstance {
  id: string;
  workflow_id: string;
  workflow?: Workflow;
  business_id?: string;
  business_type?: string;
  current_step_id?: string;
  current_step?: WorkflowStep;
  starter_id: string;
  starter?: User;
  status: 'running' | 'completed' | 'cancelled' | 'rejected';
  created_at: string;
  completed_at?: string;
}

// 统计相关
export interface DashboardStats {
  total_volunteers: number;
  new_volunteers: number;
  active_volunteers: number;
  total_cases: number;
  new_cases: number;
  resolved_cases: number;
  pending_cases: number;
  total_tasks: number;
  completed_tasks: number;
  overdue_tasks: number;
  total_dialects: number;
}

// 通知相关
export interface Notification {
  id: string;
  title: string;
  content?: string;
  type: string;
  priority: 'low' | 'normal' | 'high' | 'urgent';
  sender_id?: string;
  sender?: User;
  receiver_id: string;
  receiver?: User;
  is_read: boolean;
  read_time?: string;
  created_at: string;
}

// API 响应
export interface ApiResponse<T> {
  code: number;
  message: string;
  data: T;
}

export interface PaginatedData<T> {
  list: T[];
  total: number;
  page: number;
  size: number;
}

// 路由配置
export interface RouteConfig {
  path: string;
  element: React.ReactNode;
  title?: string;
  icon?: React.ReactNode;
  children?: RouteConfig[];
  roles?: string[];
  hideInMenu?: boolean;
}
