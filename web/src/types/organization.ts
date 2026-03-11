import type { PaginationParams } from './api';

export interface OrganizationResponse {
  id: string;
  name: string;
  code: string;
  type: string;
  level: number;
  parent_id?: string | null;
  description: string;
  address: string;
  contact_name: string;
  contact_phone: string;
  status: string;
  sort_order: number;
  created_at: string;
  children?: OrganizationResponse[];
}

export interface OrganizationTreeResponse extends OrganizationResponse {
  children?: OrganizationTreeResponse[];
}

export interface CreateOrganizationRequest {
  name: string;
  code: string;
  type: string;
  parent_id?: string;
  description?: string;
  address?: string;
  contact_name?: string;
  contact_phone?: string;
  sort_order?: number;
}

export interface UpdateOrganizationRequest {
  name?: string;
  code?: string;
  description?: string;
  address?: string;
  contact_name?: string;
  contact_phone?: string;
  status?: string;
  sort_order?: number;
}

export interface OrganizationListRequest extends PaginationParams {
  keyword?: string;
  type?: string;
  status?: string;
  parent_id?: string;
}

export interface MoveOrganizationRequest {
  new_parent_id: string;
}
