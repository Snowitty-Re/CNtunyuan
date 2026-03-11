import request from './request';
import type { PaginatedData, OrganizationResponse, OrganizationTreeResponse, OrganizationListRequest, CreateOrganizationRequest, UpdateOrganizationRequest, MoveOrganizationRequest } from '@/types';

export function getOrganizations(params: OrganizationListRequest) {
  return request.get<never, PaginatedData<OrganizationResponse>>('/organizations', { params });
}

export function getOrganizationTree() {
  return request.get<never, OrganizationTreeResponse[]>('/organizations/tree');
}

export function getOrganization(id: string) {
  return request.get<never, OrganizationResponse>(`/organizations/${id}`);
}

export function getOrganizationChildren(id: string) {
  return request.get<never, OrganizationResponse[]>(`/organizations/${id}/children`);
}

export function getOrganizationPath(id: string) {
  return request.get<never, OrganizationResponse[]>(`/organizations/${id}/path`);
}

export function createOrganization(data: CreateOrganizationRequest) {
  return request.post<never, OrganizationResponse>('/organizations', data);
}

export function updateOrganization(id: string, data: UpdateOrganizationRequest) {
  return request.put<never, OrganizationResponse>(`/organizations/${id}`, data);
}

export function deleteOrganization(id: string) {
  return request.delete<never, null>(`/organizations/${id}`);
}

export function moveOrganization(id: string, data: MoveOrganizationRequest) {
  return request.put<never, null>(`/organizations/${id}/move`, data);
}
