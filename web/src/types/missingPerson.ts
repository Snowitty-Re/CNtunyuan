import type { PaginationParams } from './api';
import type { UserResponse } from './user';

export interface MissingPersonPhoto {
  id: string;
  url: string;
  type: string;
  is_primary: boolean;
}

export interface MissingPersonResponse {
  id: string;
  case_no: string;
  name: string;
  gender: string;
  birth_date?: string;
  age: number;
  height: number;
  weight: number;
  description: string;
  photo_url: string;
  missing_time: string;
  province: string;
  city: string;
  district: string;
  address: string;
  clothes: string;
  features: string;
  contact_name: string;
  contact_phone: string;
  contact_rel: string;
  alt_contact: string;
  status: string;
  urgency: string;
  views: number;
  share_count: number;
  reporter_id: string;
  org_id: string;
  assigned_to?: string;
  found_time?: string;
  found_location: string;
  found_note: string;
  reporter?: UserResponse;
  assignee?: UserResponse;
  photos?: MissingPersonPhoto[];
  created_at: string;
}

export interface CreateMissingPersonRequest {
  name: string;
  gender: string;
  birth_date?: string;
  age?: number;
  height?: number;
  weight?: number;
  description?: string;
  photo_url?: string;
  missing_time: string;
  province?: string;
  city?: string;
  district?: string;
  address?: string;
  clothes?: string;
  features?: string;
  contact_name: string;
  contact_phone: string;
  contact_rel?: string;
  alt_contact?: string;
  urgency_level?: string;
}

export interface UpdateMissingPersonRequest {
  name?: string;
  gender?: string;
  birth_date?: string;
  age?: number;
  height?: number;
  weight?: number;
  description?: string;
  photo_url?: string;
  missing_time?: string;
  province?: string;
  city?: string;
  district?: string;
  address?: string;
  clothes?: string;
  features?: string;
  contact_name?: string;
  contact_phone?: string;
  contact_rel?: string;
  alt_contact?: string;
  urgency_level?: string;
}

export interface MissingPersonListRequest extends PaginationParams {
  keyword?: string;
  status?: string;
  gender?: string;
  age_min?: number;
  age_max?: number;
  province?: string;
  city?: string;
  district?: string;
  urgency_level?: string;
}

export interface MissingPersonTrackResponse {
  id: string;
  missing_person_id: string;
  reporter_id: string;
  location: string;
  province: string;
  city: string;
  district: string;
  address: string;
  time: string;
  description: string;
  is_key_point: boolean;
  lat: number;
  lng: number;
  status: string;
  reporter?: UserResponse;
  created_at: string;
}

export interface CreateMissingPersonTrackRequest {
  location?: string;
  province?: string;
  city?: string;
  district?: string;
  address?: string;
  time: string;
  description?: string;
  is_key_point?: boolean;
  lat?: number;
  lng?: number;
}

export interface MarkFoundRequest {
  location?: string;
  note?: string;
}

export interface MissingPersonStatsResponse {
  total: number;
  missing: number;
  searching: number;
  found: number;
  reunited: number;
  closed: number;
  today_new: number;
  week_new: number;
  month_new: number;
}
