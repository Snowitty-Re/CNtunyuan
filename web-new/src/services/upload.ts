import { http } from '@/utils/request';

export interface UploadResponse {
  id: string;
  url: string;
  filename: string;
  size: number;
  mime_type: string;
}

export const uploadApi = {
  // 单文件上传
  uploadFile: (file: File) => {
    const formData = new FormData();
    formData.append('file', file);
    return http.post<UploadResponse>('/upload', formData, {
      headers: {
        'Content-Type': 'multipart/form-data',
      },
    });
  },
  
  // 批量上传
  uploadFiles: (files: File[]) => {
    const formData = new FormData();
    files.forEach(file => formData.append('files', file));
    return http.post<UploadResponse[]>('/upload/batch', formData, {
      headers: {
        'Content-Type': 'multipart/form-data',
      },
    });
  },
  
  // 获取文件信息
  getFileInfo: (id: string) => http.get<UploadResponse>(`/upload/${id}`),
  
  // 下载文件
  downloadFile: (id: string) => http.get(`/upload/${id}/download`, { responseType: 'blob' }),
  
  // 获取实体关联的文件
  getEntityFiles: (entityType: string, entityId: string) => 
    http.get<UploadResponse[]>(`/upload/entity/${entityType}/${entityId}`),
  
  // 删除文件
  deleteFile: (id: string) => http.delete(`/upload/${id}`),
  
  // 绑定文件到实体
  bindFile: (id: string, entityType: string, entityId: string) => 
    http.put(`/upload/${id}/bind`, { entity_type: entityType, entity_id: entityId }),
};
