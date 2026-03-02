import { http } from '@/utils/request';
import type { Dialect } from '@/types';

export interface CreateDialectRequest {
  title: string;
  description?: string;
  audio_url: string;
  duration: number;
  province?: string;
  city?: string;
  district?: string;
  address?: string;
}

export const dialectApi = {
  // 获取方言列表
  getDialects: (params?: { 
    page?: number; 
    page_size?: number; 
    status?: string;
    province?: string;
    city?: string;
  }) => http.get<{ list: Dialect[]; total: number }>('/dialects', { params }),
  
  // 获取精选方言
  getFeaturedDialects: (params?: { page?: number; page_size?: number }) => 
    http.get<{ list: Dialect[]; total: number }>('/dialects/featured', { params }),
  
  // 获取方言详情
  getDialectById: (id: string) => http.get<Dialect>(`/dialects/${id}`),
  
  // 创建方言
  createDialect: (data: CreateDialectRequest) => http.post<Dialect>('/dialects', data),
  
  // 更新方言
  updateDialect: (id: string, data: Partial<CreateDialectRequest>) => 
    http.put<Dialect>(`/dialects/${id}`, data),
  
  // 删除方言
  deleteDialect: (id: string) => http.delete(`/dialects/${id}`),
  
  // 更新方言状态
  updateDialectStatus: (id: string, status: string) => 
    http.put(`/dialects/${id}/status`, { status }),
  
  // 设为精选
  markAsFeatured: (id: string) => http.post(`/dialects/${id}/feature`),
  
  // 取消精选
  unmarkAsFeatured: (id: string) => http.delete(`/dialects/${id}/feature`),
  
  // 播放记录
  recordPlay: (id: string) => http.post(`/dialects/${id}/play`),
  
  // 点赞
  like: (id: string) => http.post(`/dialects/${id}/like`),
  
  // 取消点赞
  unlike: (id: string) => http.delete(`/dialects/${id}/like`),
  
  // 获取评论
  getComments: (id: string, params?: { page?: number; page_size?: number }) => 
    http.get(`/dialects/${id}/comments`, { params }),
  
  // 添加评论
  addComment: (id: string, content: string, parent_id?: string) => 
    http.post(`/dialects/${id}/comments`, { content, parent_id }),
  
  // 获取方言统计
  getDialectStats: () => http.get<{
    total: number;
    active: number;
    pending: number;
    featured: number;
    total_plays: number;
    total_likes: number;
    total_comments: number;
  }>('/dialects/stats'),
};
