// ===== API Response =====
export interface ApiResponse<T = unknown> {
  code: number
  data: T
  message: string
}

export interface PageData<T> {
  list: T[]
  total: number
  page: number
  page_size: number
  total_pages: number
}

export interface PageParams {
  page?: number
  page_size?: number
}

// ===== Auth =====
export interface LoginRequest {
  username: string
  password: string
}

export interface LoginResponse {
  access_token: string
  refresh_token: string
  expires_in: number
  token_type: string
  user: User
}

export interface WechatLoginRequest {
  code: string
  nickname?: string
  avatar?: string
}

export interface WechatLoginResponse {
  need_bind_phone: boolean
  access_token?: string
  refresh_token?: string
  expires_in?: number
  user?: User
}

export interface SendCodeRequest {
  phone: string
}

export interface BindPhoneRequest {
  phone: string
  code: string
}

// ===== User =====
export type UserRole = 'super_admin' | 'admin' | 'manager' | 'volunteer'
export type UserStatus = 'active' | 'inactive' | 'banned'

export interface User {
  id: string
  nickname: string
  phone: string
  email: string
  role: UserRole
  status: UserStatus
  avatar: string
  org_id: string
  org_name: string
  last_login_at: string | null
  created_at: string
}

export interface UserProfile extends User {
  real_name: string
  id_card: string
  gender: string
  address: string
  emergency: string
  emergency_tel: string
  introduction: string
}

export interface CreateUserRequest {
  nickname: string
  phone: string
  email?: string
  password: string
  role: UserRole
  org_id: string
}

export interface UpdateUserRequest {
  nickname?: string
  email?: string
  role?: UserRole
  org_id?: string
  status?: UserStatus
}

export interface UpdateProfileRequest {
  nickname?: string
  email?: string
  real_name?: string
  id_card?: string
  gender?: string
  address?: string
  emergency?: string
  emergency_tel?: string
  introduction?: string
}

export interface ChangePasswordRequest {
  old_password: string
  new_password: string
}

export interface UserQuery extends PageParams {
  keyword?: string
  role?: UserRole
  status?: UserStatus
  org_id?: string
}

// ===== Organization =====
export type OrgType = 'root' | 'province' | 'city' | 'district' | 'street' | 'community' | 'team'
export type OrgStatus = 'active' | 'inactive'

export interface Organization {
  id: string
  name: string
  code: string
  type: OrgType
  level: number
  parent_id: string | null
  description: string
  address: string
  contact_name: string
  contact_phone: string
  status: OrgStatus
  sort_order: number
  created_at: string
  children?: Organization[]
}

export interface CreateOrgRequest {
  name: string
  code: string
  type: OrgType
  parent_id?: string
  description?: string
  address?: string
  contact_name?: string
  contact_phone?: string
  sort_order?: number
}

export interface UpdateOrgRequest {
  name?: string
  code?: string
  description?: string
  address?: string
  contact_name?: string
  contact_phone?: string
  status?: OrgStatus
  sort_order?: number
}

export interface OrgQuery extends PageParams {
  keyword?: string
  type?: OrgType
  status?: OrgStatus
  parent_id?: string
}

// ===== Missing Person =====
export type MissingStatus = 'missing' | 'searching' | 'found' | 'reunited' | 'closed'
export type UrgencyLevel = 'low' | 'medium' | 'high' | 'urgent'

export interface MissingPerson {
  id: string
  case_no: string
  name: string
  gender: string
  birth_date: string | null
  age: number
  height: number
  weight: number
  description: string
  photo_url: string
  missing_time: string
  province: string
  city: string
  district: string
  address: string
  clothes: string
  features: string
  contact_name: string
  contact_phone: string
  contact_rel: string
  alt_contact: string
  status: MissingStatus
  urgency: UrgencyLevel
  views: number
  share_count: number
  reporter_id: string
  org_id: string
  assigned_to: string | null
  found_time: string | null
  found_location: string
  found_note: string
  reporter?: User
  assignee?: User
  photos?: MissingPhoto[]
  created_at: string
}

export interface MissingPhoto {
  id: string
  url: string
  type: string
  is_primary: boolean
}

export interface MissingPersonTrack {
  id: string
  missing_person_id: string
  reporter_id: string
  location: string
  province: string
  city: string
  district: string
  address: string
  time: string
  description: string
  is_key_point: boolean
  lat: number
  lng: number
  status: string
  reporter?: User
  created_at: string
}

export interface CreateMissingPersonRequest {
  name: string
  gender: string
  birth_date?: string
  age?: number
  height?: number
  weight?: number
  description?: string
  photo_url?: string
  missing_time: string
  province?: string
  city?: string
  district?: string
  address?: string
  clothes?: string
  features?: string
  contact_name: string
  contact_phone: string
  contact_rel?: string
  alt_contact?: string
  urgency_level?: string
}

export interface UpdateMissingPersonRequest extends Partial<CreateMissingPersonRequest> {}

export interface MissingPersonQuery extends PageParams {
  keyword?: string
  status?: MissingStatus
  gender?: string
  age_min?: number
  age_max?: number
  province?: string
  city?: string
  district?: string
  urgency_level?: UrgencyLevel
}

export interface CreateTrackRequest {
  location?: string
  province?: string
  city?: string
  district?: string
  address?: string
  time: string
  description: string
  is_key_point?: boolean
  lat?: number
  lng?: number
}

export interface MarkFoundRequest {
  location?: string
  note?: string
}

export interface MissingPersonStats {
  total: number
  missing: number
  searching: number
  found: number
  reunited: number
  closed: number
  today_new: number
  week_new: number
  month_new: number
}

// ===== Task =====
export type TaskStatus = 'draft' | 'pending' | 'assigned' | 'processing' | 'completed' | 'cancelled' | 'overdue'
export type TaskType = 'search' | 'verify' | 'assist' | 'follow' | 'interview' | 'other'
export type TaskPriority = 'low' | 'medium' | 'high' | 'urgent'

export interface Task {
  id: string
  title: string
  description: string
  type: TaskType
  priority: TaskPriority
  status: TaskStatus
  deadline: string | null
  started_at: string | null
  completed_at: string | null
  creator_id: string
  assignee_id: string | null
  org_id: string
  missing_person_id: string | null
  location: string
  province: string
  city: string
  district: string
  address: string
  lat: number
  lng: number
  result: string
  feedback: string
  progress: number
  view_count: number
  creator?: User
  assignee?: User
  missing_person?: MissingPerson
  created_at: string
}

export interface TaskLog {
  id: string
  task_id: string
  user_id: string
  action: string
  old_status: string
  new_status: string
  content: string
  user?: User
  created_at: string
}

export interface CreateTaskRequest {
  title: string
  description?: string
  type: TaskType
  priority?: TaskPriority
  deadline?: string
  missing_person_id?: string
  location?: string
  province?: string
  city?: string
  district?: string
  address?: string
  lat?: number
  lng?: number
}

export interface UpdateTaskRequest extends Partial<CreateTaskRequest> {}

export interface TaskQuery extends PageParams {
  keyword?: string
  type?: TaskType
  status?: TaskStatus
  priority?: TaskPriority
  assignee_id?: string
  missing_person_id?: string
  is_overdue?: boolean
}

export interface TaskStats {
  total: number
  draft: number
  pending: number
  assigned: number
  processing: number
  completed: number
  cancelled: number
  overdue: number
  my_tasks: number
  my_pending: number
  my_completed: number
}

// ===== Dialect =====
export type DialectStatus = 'active' | 'inactive' | 'pending'
export type DialectType = 'phrase' | 'story' | 'song' | 'daily' | 'other'

export interface Dialect {
  id: string
  title: string
  content: string
  region: string
  province: string
  city: string
  dialect_type: DialectType
  audio_url: string
  duration: number
  file_size: number
  format: string
  status: DialectStatus
  is_featured: boolean
  play_count: number
  like_count: number
  comment_count: number
  tags: string
  description: string
  uploader_id: string
  org_id: string
  uploader?: User
  created_at: string
}

export interface DialectComment {
  id: string
  dialect_id: string
  user_id: string
  content: string
  parent_id: string | null
  reply_count: number
  like_count: number
  user?: User
  created_at: string
}

export interface CreateDialectRequest {
  title: string
  content?: string
  region: string
  province?: string
  city?: string
  dialect_type?: DialectType
  audio_url: string
  duration?: number
  file_size?: number
  format?: string
  tags?: string
  description?: string
}

export interface UpdateDialectRequest extends Partial<CreateDialectRequest> {}

export interface DialectQuery extends PageParams {
  keyword?: string
  region?: string
  province?: string
  city?: string
  type?: DialectType
  status?: DialectStatus
  sort_by?: string
  sort_order?: string
}

export interface DialectStats {
  total: number
  active: number
  pending: number
  featured: number
  total_plays: number
  total_likes: number
}

// ===== File =====
export type FileType = 'image' | 'audio' | 'video' | 'document'

export interface FileInfo {
  id: string
  file_name: string
  original_name: string
  file_type: FileType
  mime_type: string
  size: number
  size_readable: string
  url: string
  storage_type: string
  uploader_id: string
  entity_type: string
  entity_id: string
  description: string
  created_at: string
}

// ===== Dashboard =====
export interface DashboardStats {
  users: { total: number; active: number; new_today: number; new_week: number; new_month: number }
  organizations: { total: number; provinces: number; cities: number; districts: number }
  missing_persons: { total: number; missing: number; searching: number; found: number; reunited: number; new_today: number; new_week: number }
  tasks: { total: number; pending: number; processing: number; completed: number; overdue: number }
  dialects: { total: number; featured: number; plays: number; likes: number }
  files: { total_count: number; total_size: number }
  recent_activity?: Activity[]
}

export interface Activity {
  id: string
  type: string
  title: string
  user_id: string
  user_name: string
  entity_type: string
  entity_id: string
  created_at: string
}

export interface TrendData {
  date: string
  new_cases: number
  resolved_cases: number
  new_tasks: number
  completed_tasks: number
  new_users: number
}

// ===== Audit =====
export type AuditLogType = 'login' | 'logout' | 'create' | 'update' | 'delete' | 'query' | 'export' | 'upload' | 'download' | 'other'

export interface AuditLog {
  id: string
  user_id: string
  username: string
  user_role: string
  module: string
  action: string
  type: AuditLogType
  description: string
  request_method: string
  request_url: string
  request_ip: string
  request_body: string
  response_code: number
  status: string
  error_message: string
  duration: number
  user_agent: string
  trace_id: string
  created_at: string
}

export interface AuditQuery extends PageParams {
  user_id?: string
  module?: string
  action?: string
  type?: AuditLogType
  status?: string
  start_time?: string
  end_time?: string
  keyword?: string
  request_ip?: string
}
