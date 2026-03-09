import request from './request'
import type { ApiResponse, PageData, Organization, OrgQuery, CreateOrgRequest, UpdateOrgRequest } from '@/types'

export const orgApi = {
  list: (params: OrgQuery) =>
    request.get<ApiResponse<PageData<Organization>>>('/organizations', { params }),

  getTree: () =>
    request.get<ApiResponse<Organization[]>>('/organizations/tree'),

  getById: (id: string) =>
    request.get<ApiResponse<Organization>>(`/organizations/${id}`),

  getChildren: (id: string) =>
    request.get<ApiResponse<Organization[]>>(`/organizations/${id}/children`),

  getPath: (id: string) =>
    request.get<ApiResponse<Organization[]>>(`/organizations/${id}/path`),

  create: (data: CreateOrgRequest) =>
    request.post<ApiResponse<Organization>>('/organizations', data),

  update: (id: string, data: UpdateOrgRequest) =>
    request.put<ApiResponse<Organization>>(`/organizations/${id}`, data),

  delete: (id: string) =>
    request.delete<ApiResponse<null>>(`/organizations/${id}`),

  move: (id: string, new_parent_id: string) =>
    request.put<ApiResponse<null>>(`/organizations/${id}/move`, { new_parent_id }),
}
