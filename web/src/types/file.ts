import type { PaginationParams } from './api';

export interface FileResponse {
  id: string;
  file_name: string;
  original_name: string;
  file_type: string;
  mime_type: string;
  size: number;
  size_readable: string;
  url: string;
  storage_type: string;
  uploader_id: string;
  entity_type: string;
  entity_id: string;
  description: string;
  created_at: string;
}

export interface FileListRequest extends PaginationParams {
  keyword?: string;
  file_type?: string;
  uploader_id?: string;
}

export interface FileStatsResponse {
  total_count: number;
  total_size: number;
  image_count: number;
  image_size: number;
  audio_count: number;
  audio_size: number;
  video_count: number;
  video_size: number;
  doc_count: number;
  doc_size: number;
}

export interface BindFileRequest {
  entity_type: string;
  entity_id: string;
}

export interface UploadResponse {
  files: FileResponse[];
}
