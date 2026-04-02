import { http } from '@/lib/request'
import type { Organization, Paginated } from '@/types/api'

export const organizationService = {
  list(params: Record<string, string | number | undefined> = {}) {
    return http<Paginated<Organization>>('/organizations', { query: params })
  },
  tree() {
    return http<Organization[]>('/organizations/tree')
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
  create(data: Record<string, any>) {
    return http<Organization>('/organizations', { method: 'POST', body: data })
  },
  update(id: string, data: Record<string, any>) {
    return http<Organization>(`/organizations/${id}`, { method: 'PUT', body: data })
  },
  remove(id: string) {
    return http<null>(`/organizations/${id}`, { method: 'DELETE' })
  },
  move(id: string, parentId: string | null) {
    return http<Organization>(`/organizations/${id}/move`, {
      method: 'PUT',
      body: { parent_id: parentId },
    })
  },
}
