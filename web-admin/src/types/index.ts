export interface User {
  id: string
  open_id: string
  union_id: string
  nickname: string
  avatar: string
  phone: string
  email: string
  real_name: string
  id_card: string
  role: string
  status: string
  org_id?: string
  org?: Organization
  last_login?: string
  created_at: string
  updated_at: string
}

export interface Organization {
  id: string
  name: string
  code: string
  type: string
  level: number
  parent_id?: string
  parent?: Organization
  children?: Organization[]
  leader_id?: string
  leader?: User
  province: string
  city: string
  district: string
  street: string
  address: string
  contact: string
  phone: string
  volunteer_count: number
  case_count: number
  status: string
  created_at: string
}

export interface MissingPerson {
  id: string
  case_no: string
  status: string
  case_type: string
  name: string
  gender: string
  birth_date?: string
  age: number
  height: number
  weight: number
  appearance: string
  clothing: string
  special_features: string
  mental_status: string
  physical_status: string
  missing_time: string
  missing_location: string
  missing_longitude: number
  missing_latitude: number
  missing_detail: string
  possible_location: string
  photos: MissingPhoto[]
  contact_name: string
  contact_phone: string
  contact_relation: string
  contact_address: string
  family_description: string
  reporter_id: string
  reporter: User
  org_id: string
  org: Organization
  found_time?: string
  found_location: string
  found_detail: string
  view_count: number
  share_count: number
  task_count: number
  created_at: string
  updated_at: string
}

export interface MissingPhoto {
  id: string
  url: string
  thumbnail_url: string
  type: string
  description: string
  sort: number
}

export interface Dialect {
  id: string
  title: string
  description: string
  audio_url: string
  duration: number
  file_size: number
  format: string
  province: string
  city: string
  district: string
  town: string
  village: string
  longitude: number
  latitude: number
  address: string
  collector_id: string
  collector: User
  org_id: string
  org: Organization
  play_count: number
  like_count: number
  status: string
  created_at: string
  updated_at: string
}

export interface Task {
  id: string
  task_no: string
  title: string
  description: string
  type: string
  priority: string
  status: string
  missing_person_id?: string
  missing_person?: MissingPerson
  creator_id: string
  creator: User
  assignee_id?: string
  assignee?: User
  org_id: string
  org: Organization
  start_time?: string
  deadline?: string
  completed_time?: string
  location: string
  longitude: number
  latitude: number
  progress: number
  created_at: string
  updated_at: string
}

export interface PageData<T> {
  list: T[]
  total: number
  page: number
  page_size: number
  total_pages: number
}

export interface ApiResponse<T> {
  code: number
  message: string
  data: T
}
