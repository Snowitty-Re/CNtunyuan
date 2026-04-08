import { http } from '@/lib/request'
import type { Dialect, DialectCard, DialectCardGroup, Paginated } from '@/types/api'

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
  createBatch(data: Record<string, any>) {
    return http<{ batch_id: string; total: number; items: Dialect[] }>('/dialects/batch', { method: 'POST', body: data })
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
  cardTemplate(includeInactive = false) {
    return http<{ groups: DialectCardGroup[] }>('/dialect-cards/template', {
      query: includeInactive ? { include_inactive: true } : undefined,
    })
  },
  cardGroups() {
    return http<{ groups: DialectCardGroup[] }>('/dialect-cards/groups')
  },
  createCardGroup(data: { name: string; description?: string; sort_order?: number; status?: string }) {
    return http<DialectCardGroup>('/dialect-cards/groups', { method: 'POST', body: data })
  },
  updateCardGroup(id: string, data: { name?: string; description?: string; sort_order?: number; status?: string }) {
    return http<DialectCardGroup>(`/dialect-cards/groups/${id}`, { method: 'PUT', body: data })
  },
  removeCardGroup(id: string) {
    return http<null>(`/dialect-cards/groups/${id}`, { method: 'DELETE' })
  },
  cards(params: Record<string, string | number | boolean | undefined> = {}) {
    return http<{ groups: DialectCardGroup[]; list?: DialectCard[] }>('/dialect-cards', { query: params })
  },
  createCard(data: { group_id: string; content: string; image_url: string; sort_order?: number; required?: boolean; status?: string }) {
    return http<DialectCard>('/dialect-cards', { method: 'POST', body: data })
  },
  updateCard(id: string, data: { group_id?: string; content?: string; image_url?: string; sort_order?: number; required?: boolean; status?: string }) {
    return http<DialectCard>(`/dialect-cards/${id}`, { method: 'PUT', body: data })
  },
  removeCard(id: string) {
    return http<null>(`/dialect-cards/${id}`, { method: 'DELETE' })
  },
}
