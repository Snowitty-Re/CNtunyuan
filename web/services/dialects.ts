import { http } from '@/lib/request'
import type { Dialect, Paginated } from '@/types/api'

export const dialectService = {
  list(params: Record<string, string | number | undefined> = {}) {
    return http<Paginated<Dialect>>('/dialects', { query: params })
  },
  byId(id: string) {
    return http<Dialect>(`/dialects/${id}`)
  },
  create(data: Record<string, any>) {
    return http<Dialect>('/dialects', { method: 'POST', body: data })
  },
  update(id: string, data: Record<string, any>) {
    return http<Dialect>(`/dialects/${id}`, { method: 'PUT', body: data })
  },
  remove(id: string) {
    return http<null>(`/dialects/${id}`, { method: 'DELETE' })
  },
  updateStatus(id: string, status: string) {
    return http<Dialect>(`/dialects/${id}/status`, { method: 'PUT', body: { status } })
  },
  feature(id: string) {
    return http<Dialect>(`/dialects/${id}/feature`, { method: 'POST' })
  },
  comments(id: string, params: Record<string, string | number | undefined> = {}) {
    return http<Paginated<any> | any[]>(`/dialects/${id}/comments`, { query: params })
  },
  addComment(id: string, content: string) {
    return http<any>(`/dialects/${id}/comments`, {
      method: 'POST',
      body: { content },
    })
  },
}
