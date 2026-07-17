import { http } from '@/lib/request'
import type { Organization, Paginated } from '@/types/api'

function normalizeTree(data: unknown): Organization[] {
  if (!data) return []
  if (Array.isArray(data)) return data as Organization[]
  if (typeof data === 'object') return [data as Organization]
  return []
}

export type OrgCreateBody = {
  name: string
  code: string
  type: string
  parent_id?: string
  description?: string
  address?: string
  contact_name?: string
  contact_phone?: string
  sort_order?: number
}

export type OrgUpdateBody = {
  name?: string
  code?: string
  description?: string
  address?: string
  contact_name?: string
  contact_phone?: string
  status?: string
  sort_order?: number
}

export const organizationService = {
  list(params: Record<string, string | number | undefined> = {}) {
    return http<Paginated<Organization>>('/organizations', { query: params })
  },
  async tree(rootId?: string) {
    const data = await http<Organization | Organization[]>('/organizations/tree', {
      query: rootId ? { root_id: rootId } : undefined,
    })
    return normalizeTree(data)
  },
  byId(id: string) {
    return http<Organization>(`/organizations/${id}`)
  },
  children(id: string) {
    return http<Organization[]>(`/organizations/${id}/children`)
  },
  path(id: string) {
    return http<Organization[]>(`/organizations/${id}/path`)
  },
  create(data: OrgCreateBody) {
    return http<Organization>('/organizations', { method: 'POST', body: data })
  },
  update(id: string, data: OrgUpdateBody) {
    return http<Organization>(`/organizations/${id}`, { method: 'PUT', body: data })
  },
  remove(id: string) {
    return http<null>(`/organizations/${id}`, { method: 'DELETE' })
  },
  /** Backend expects new_parent_id; empty string for super_admin root move */
  move(id: string, newParentId: string) {
    return http<Organization>(`/organizations/${id}/move`, {
      method: 'PUT',
      body: { new_parent_id: newParentId || '' },
    })
  },
}
