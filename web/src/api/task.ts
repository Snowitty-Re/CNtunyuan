import request from './request'
import type { ApiResponse, PageData, Task, TaskLog, TaskQuery, TaskStats, CreateTaskRequest, UpdateTaskRequest } from '@/types'

export const taskApi = {
  list: (params: TaskQuery) =>
    request.get<ApiResponse<PageData<Task>>>('/tasks', { params }),

  myTasks: (params: TaskQuery) =>
    request.get<ApiResponse<PageData<Task>>>('/tasks/my', { params }),

  pending: (params: TaskQuery) =>
    request.get<ApiResponse<PageData<Task>>>('/tasks/pending', { params }),

  overdue: (params: TaskQuery) =>
    request.get<ApiResponse<PageData<Task>>>('/tasks/overdue', { params }),

  getStats: () =>
    request.get<ApiResponse<TaskStats>>('/tasks/stats'),

  getById: (id: string) =>
    request.get<ApiResponse<Task>>(`/tasks/${id}`),

  create: (data: CreateTaskRequest) =>
    request.post<ApiResponse<Task>>('/tasks', data),

  update: (id: string, data: UpdateTaskRequest) =>
    request.put<ApiResponse<Task>>(`/tasks/${id}`, data),

  delete: (id: string) =>
    request.delete<ApiResponse<null>>(`/tasks/${id}`),

  assign: (id: string, assignee_id: string) =>
    request.post<ApiResponse<null>>(`/tasks/${id}/assign`, { assignee_id }),

  start: (id: string) =>
    request.post<ApiResponse<null>>(`/tasks/${id}/start`),

  complete: (id: string, result: string) =>
    request.post<ApiResponse<null>>(`/tasks/${id}/complete`, { result }),

  cancel: (id: string, reason: string) =>
    request.post<ApiResponse<null>>(`/tasks/${id}/cancel`, { reason }),

  updateProgress: (id: string, progress: number) =>
    request.put<ApiResponse<null>>(`/tasks/${id}/progress`, { progress }),

  getLogs: (id: string) =>
    request.get<ApiResponse<TaskLog[]>>(`/tasks/${id}/logs`),
}
