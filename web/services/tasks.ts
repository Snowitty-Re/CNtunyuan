import { http } from '@/lib/request'
import type { Paginated, Task, TaskFollowUp } from '@/types/api'

export const taskService = {
  list(params: Record<string, string | number | undefined> = {}) {
    return http<Paginated<Task>>('/tasks', { query: params })
  },
  my(params: Record<string, string | number | undefined> = {}) {
    return http<Paginated<Task>>('/tasks/my', { query: params })
  },
  stats() {
    return http<Record<string, any>>('/tasks/stats')
  },
  byId(id: string) {
    return http<Task>(`/tasks/${id}`)
  },
  create(data: Record<string, any>) {
    return http<Task>('/tasks', { method: 'POST', body: data })
  },
  update(id: string, data: Record<string, any>) {
    return http<Task>(`/tasks/${id}`, { method: 'PUT', body: data })
  },
  remove(id: string) {
    return http<null>(`/tasks/${id}`, { method: 'DELETE' })
  },
  assign(id: string, assigneeId: string) {
    return http<Task>(`/tasks/${id}/assign`, {
      method: 'POST',
      body: { assignee_id: assigneeId },
    })
  },
  start(id: string) {
    return http<Task>(`/tasks/${id}/start`, { method: 'POST' })
  },
  complete(id: string, data: Record<string, any>) {
    return http<Task>(`/tasks/${id}/complete`, { method: 'POST', body: data })
  },
  cancel(id: string, reason?: string) {
    return http<Task>(`/tasks/${id}/cancel`, { method: 'POST', body: { reason: reason || '' } })
  },
  followUps(id: string, params: Record<string, string | number | undefined> = {}) {
    return http<Paginated<TaskFollowUp> | TaskFollowUp[]>(`/tasks/${id}/follow-ups`, { query: params })
  },
  followUpById(taskId: string, followUpId: string) {
    return http<TaskFollowUp>(`/tasks/${taskId}/follow-ups/${followUpId}`)
  },
  followUpComments(taskId: string, followUpId: string, params: Record<string, string | number | undefined> = {}) {
    return http<Paginated<any> | any[]>(`/tasks/${taskId}/follow-ups/${followUpId}/comments`, { query: params })
  },
  addFollowUpComment(taskId: string, followUpId: string, content: string) {
    return http<any>(`/tasks/${taskId}/follow-ups/${followUpId}/comments`, {
      method: 'POST',
      body: { content },
    })
  },
  createFollowUp(id: string, data: Record<string, any>) {
    return http<TaskFollowUp>(`/tasks/${id}/follow-ups`, { method: 'POST', body: data })
  },
  reviewFollowUp(taskId: string, followUpId: string, data: Record<string, any>) {
    return http<TaskFollowUp>(`/tasks/${taskId}/follow-ups/${followUpId}/review`, {
      method: 'POST',
      body: data,
    })
  },
  logs(id: string, params: Record<string, string | number | undefined> = {}) {
    return http<Paginated<any> | any[]>(`/tasks/${id}/logs`, { query: params })
  },
}
