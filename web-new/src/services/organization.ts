import { http } from '@/utils/request';
import type { Organization } from '@/types';

export interface CreateOrgRequest {
  name: string;
  code: string;
  type: string;
  parent_id?: string;
  province?: string;
  city?: string;
  district?: string;
  address?: string;
  contact?: string;
  phone?: string;
  email?: string;
  description?: string;
}

export const organizationApi = {
  // 获取组织列表
  getOrganizations: (params?: { 
    page?: number; 
    page_size?: number; 
    parent_id?: string;
    type?: string;
  }) => http.get<{ list: Organization[]; total: number }>('/organizations', { params }),
  
  // 获取组织树
  getOrgTree: (root_id?: string) => 
    http.get<Organization[]>('/organizations/tree', { params: { root_id } }),
  
  // 获取组织详情
  getOrgById: (id: string) => http.get<Organization>(`/organizations/${id}`),
  
  // 获取子组织
  getOrgChildren: (id: string) => http.get<Organization[]>(`/organizations/${id}/children`),
  
  // 获取组织路径
  getOrgPath: (id: string) => http.get<Organization[]>(`/organizations/${id}/path`),
  
  // 创建组织
  createOrg: (data: CreateOrgRequest) => http.post<Organization>('/organizations', data),
  
  // 更新组织
  updateOrg: (id: string, data: Partial<CreateOrgRequest>) => 
    http.put<Organization>(`/organizations/${id}`, data),
  
  // 删除组织
  deleteOrg: (id: string) => http.delete(`/organizations/${id}`),
  
  // 移动组织
  moveOrg: (id: string, parent_id: string) => 
    http.put(`/organizations/${id}/move`, { parent_id }),
};
