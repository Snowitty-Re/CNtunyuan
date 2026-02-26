import request from '../utils/request'
import type { Organization, PageData } from '../types'

export const orgApi = {
  // 获取组织列表
  getOrgs: (params: { page?: number; page_size?: number; type?: string; parent_id?: string }) =>
    request.get<PageData<Organization>>('/organizations', { params }),

  // 获取组织架构树
  getOrgTree: (parentId?: string) =>
    request.get<Organization[]>('/organizations/tree', { params: { parent_id: parentId } }),

  // 获取组织详情
  getOrg: (id: string) => request.get<Organization>(`/organizations/${id}`),

  // 创建组织
  createOrg: (data: Partial<Organization>) => request.post('/organizations', data),

  // 更新组织
  updateOrg: (id: string, data: Partial<Organization>) => request.put(`/organizations/${id}`, data),

  // 删除组织
  deleteOrg: (id: string) => request.delete(`/organizations/${id}`),
}
