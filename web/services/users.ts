import { http } from '@/lib/request'
import type { Paginated, User } from '@/types/api'

export const userService = {
  list(params: Record<string, string | number | undefined> = {}) {
    return http<Paginated<User>>('/users', { query: params })
  },
  byId(id: string) {
    return http<User>(`/users/${id}`)
  },
  create(data: Record<string, any>) {
    return http<User>('/users', { method: 'POST', body: data })
  },
  update(id: string, data: Record<string, any>) {
    return http<User>(`/users/${id}`, { method: 'PUT', body: data })
  },
  remove(id: string) {
    return http<null>(`/users/${id}`, { method: 'DELETE' })
  },
  updateStatus(id: string, status: string) {
    return http<User>(`/users/${id}/status`, { method: 'PUT', body: { status } })
  },
  updateRole(id: string, role: string) {
    return http<User>(`/users/${id}/role`, { method: 'PUT', body: { role } })
  },
}
