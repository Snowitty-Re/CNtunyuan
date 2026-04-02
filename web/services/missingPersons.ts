import { http } from '@/lib/request'
import type { MissingPerson, MissingTrack, Paginated } from '@/types/api'

export const missingPersonService = {
  list(params: Record<string, string | number | undefined> = {}) {
    return http<Paginated<MissingPerson>>('/missing-persons', { query: params })
  },
  byId(id: string) {
    return http<MissingPerson>(`/missing-persons/${id}`)
  },
  create(data: Record<string, any>) {
    return http<MissingPerson>('/missing-persons', { method: 'POST', body: data })
  },
  update(id: string, data: Record<string, any>) {
    return http<MissingPerson>(`/missing-persons/${id}`, { method: 'PUT', body: data })
  },
  remove(id: string) {
    return http<null>(`/missing-persons/${id}`, { method: 'DELETE' })
  },
  updateStatus(id: string, status: string) {
    return http<MissingPerson>(`/missing-persons/${id}/status`, {
      method: 'PUT',
      body: { status },
    })
  },
  markFound(id: string, data: Record<string, any>) {
    return http<MissingPerson>(`/missing-persons/${id}/found`, {
      method: 'POST',
      body: data,
    })
  },
  markReunited(id: string) {
    return http<MissingPerson>(`/missing-persons/${id}/reunited`, { method: 'POST' })
  },
  tracks(id: string) {
    return http<Paginated<MissingTrack> | MissingTrack[]>(`/missing-persons/${id}/tracks`)
  },
  addTrack(id: string, data: Record<string, any>) {
    return http<MissingTrack>(`/missing-persons/${id}/tracks`, {
      method: 'POST',
      body: data,
    })
  },
}
