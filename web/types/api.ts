export type ApiEnvelope<T> = {
  code: number
  message: string
  data: T
}

export type Paginated<T> = {
  list: T[]
  total: number
  page: number
  page_size: number
}

export type User = {
  id: string
  nickname?: string
  username?: string
  phone?: string
  email?: string
  avatar?: string
  role: 'super_admin' | 'admin' | 'manager' | 'volunteer' | string
  status?: string
  wx_bound?: boolean
  org_id?: string | null
  org_name?: string
  organization?: { id: string; name: string } | null
  created_at?: string
}

export type AuthLoginResponse = {
  access_token: string
  refresh_token: string
  expires_in?: number
  token_type?: string
  user: User
}

export type WechatLoginResponse = {
  need_bind_phone: boolean
  access_token: string
  refresh_token: string
  expires_in?: number
  token_type?: string
  user: User
}

export type MissingPerson = {
  id: string
  name: string
  gender: string
  age?: number
  case_type?: string
  status?: string
  province?: string
  city?: string
  district?: string
  address?: string
  missing_time?: string
  missing_latitude?: number | null
  missing_longitude?: number | null
  description?: string
  contact_name?: string
  contact_phone?: string
  photo_url?: string
  photos?: Array<{ id?: string; url?: string } | string>
  created_at?: string
}

export type MissingTrack = {
  id: string
  time?: string
  location?: string
  province?: string
  city?: string
  district?: string
  address?: string
  lat?: number | null
  lng?: number | null
  description?: string
  is_key_point?: boolean
  created_at?: string
}

export type Task = {
  id: string
  title: string
  description?: string
  type?: string
  priority?: string
  status?: string
  progress?: number
  deadline?: string
  assignee_id?: string
  assignee?: { id: string; nickname?: string; phone?: string } | null
  missing_person_id?: string
  missing_person?: { id: string; name?: string } | null
  created_at?: string
}

export type TaskFollowUp = {
  id: string
  content: string
  progress?: number
  attachments?: string[]
  review_status?: string
  review_remark?: string
  reviewer?: { id: string; nickname?: string } | null
  created_at?: string
}

export type Dialect = {
  id: string
  title: string
  content?: string
  region?: string
  province?: string
  city?: string
  collect_address?: string
  collect_latitude?: number
  collect_longitude?: number
  status?: string
  audio_url?: string
  featured?: boolean
  play_count?: number
  like_count?: number
  created_at?: string
}

export type Organization = {
  id: string
  name: string
  code: string
  type?: string
  parent_id?: string | null
  description?: string
  address?: string
  contact_name?: string
  contact_phone?: string
  sort_order?: number
  children?: Organization[]
  created_at?: string
}
