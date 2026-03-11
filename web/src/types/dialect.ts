import type { PaginationParams } from './api';
import type { UserResponse } from './user';

export interface DialectResponse {
  id: string;
  title: string;
  content: string;
  region: string;
  province: string;
  city: string;
  dialect_type: string;
  audio_url: string;
  duration: number;
  file_size: number;
  format: string;
  status: string;
  is_featured: boolean;
  play_count: number;
  like_count: number;
  comment_count: number;
  tags: string;
  description: string;
  uploader_id: string;
  org_id: string;
  uploader?: UserResponse;
  created_at: string;
}

export interface CreateDialectRequest {
  title: string;
  content?: string;
  region: string;
  province?: string;
  city?: string;
  dialect_type?: string;
  audio_url: string;
  duration?: number;
  file_size?: number;
  format?: string;
  tags?: string;
  description?: string;
}

export interface UpdateDialectRequest {
  title?: string;
  content?: string;
  region?: string;
  province?: string;
  city?: string;
  dialect_type?: string;
  tags?: string;
  description?: string;
}

export interface DialectListRequest extends PaginationParams {
  keyword?: string;
  region?: string;
  province?: string;
  city?: string;
  type?: string;
  status?: string;
  sort_by?: string;
  sort_order?: string;
}

export interface DialectCommentResponse {
  id: string;
  dialect_id: string;
  user_id: string;
  content: string;
  parent_id?: string;
  reply_count: number;
  like_count: number;
  user?: UserResponse;
  created_at: string;
}

export interface CreateDialectCommentRequest {
  content: string;
  parent_id?: string;
}

export interface DialectStatsResponse {
  total: number;
  active: number;
  pending: number;
  featured: number;
  total_plays: number;
  total_likes: number;
}
