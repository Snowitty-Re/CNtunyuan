import { http } from '@/utils/request';
import type { MissingPerson, MissingPersonTrack } from '@/types';

export interface CreateCaseRequest {
  name: string;
  gender: string;
  age?: number;
  case_type: string;
  missing_time: string;
  missing_location: string;
  contact_name?: string;
  contact_phone?: string;
  description?: string;
}

export const caseApi = {
  // 获取案件列表
  getCases: (params?: { 
    page?: number; 
    page_size?: number; 
    status?: string; 
    keyword?: string;
    org_id?: string;
  }) => http.get<{ list: MissingPerson[]; total: number }>('/missing-persons', { params }),
  
  // 搜索案件
  searchCases: (keyword: string, params?: { page?: number; page_size?: number }) => 
    http.get<{ list: MissingPerson[]; total: number }>('/missing-persons/search', { 
      params: { ...params, keyword } 
    }),
  
  // 获取案件详情
  getCaseById: (id: string) => http.get<MissingPerson>(`/missing-persons/${id}`),
  
  // 创建案件
  createCase: (data: CreateCaseRequest) => http.post<MissingPerson>('/missing-persons', data),
  
  // 更新案件
  updateCase: (id: string, data: Partial<CreateCaseRequest>) => 
    http.put<MissingPerson>(`/missing-persons/${id}`, data),
  
  // 删除案件
  deleteCase: (id: string) => http.delete(`/missing-persons/${id}`),
  
  // 更新案件状态
  updateCaseStatus: (id: string, status: string) => 
    http.put(`/missing-persons/${id}/status`, { status }),
  
  // 标记为已找到
  markAsFound: (id: string, data: { location?: string; note?: string }) => 
    http.post(`/missing-persons/${id}/found`, data),
  
  // 标记为已团聚
  markAsReunited: (id: string) => http.post(`/missing-persons/${id}/reunited`),
  
  // 获取案件轨迹
  getCaseTracks: (id: string) => http.get<MissingPersonTrack[]>(`/missing-persons/${id}/tracks`),
  
  // 添加轨迹
  addCaseTrack: (id: string, data: Partial<MissingPersonTrack>) => 
    http.post<MissingPersonTrack>(`/missing-persons/${id}/tracks`, data),
  
  // 获取案件统计
  getCaseStats: () => http.get<{
    total: number;
    missing: number;
    searching: number;
    found: number;
    reunited: number;
    closed: number;
  }>('/missing-persons/stats'),
};
