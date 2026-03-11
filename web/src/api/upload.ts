import request from './request';
import type { PaginatedData, FileResponse, FileListRequest, FileStatsResponse, BindFileRequest } from '@/types';

export function uploadFile(file: File, onProgress?: (percent: number) => void) {
  const formData = new FormData();
  formData.append('file', file);
  return request.post<never, FileResponse>('/upload', formData, {
    headers: { 'Content-Type': 'multipart/form-data' },
    onUploadProgress: (e) => {
      if (onProgress && e.total) {
        onProgress(Math.round((e.loaded / e.total) * 100));
      }
    },
  });
}

export function uploadFiles(files: File[]) {
  const formData = new FormData();
  files.forEach((f) => formData.append('files', f));
  return request.post<never, { files: FileResponse[] }>('/upload/batch', formData, {
    headers: { 'Content-Type': 'multipart/form-data' },
  });
}

export function getFileInfo(id: string) {
  return request.get<never, FileResponse>(`/upload/${id}`);
}

export function getFilesByEntity(type: string, id: string) {
  return request.get<never, FileResponse[]>(`/upload/entity/${type}/${id}`);
}

export function deleteFile(id: string) {
  return request.delete<never, null>(`/upload/${id}`);
}

export function bindFile(id: string, data: BindFileRequest) {
  return request.put<never, null>(`/upload/${id}/bind`, data);
}

export function getFileStats() {
  return request.get<never, FileStatsResponse>('/upload/stats');
}

export function getFiles(params: FileListRequest) {
  return request.get<never, PaginatedData<FileResponse>>('/upload', { params });
}
