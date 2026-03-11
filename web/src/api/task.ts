import request from './request';
import type {
  PaginatedData,
  TaskResponse,
  TaskListRequest,
  CreateTaskRequest,
  UpdateTaskRequest,
  AssignTaskRequest,
  CompleteTaskRequest,
  CancelTaskRequest,
  UpdateTaskProgressRequest,
  TaskLogResponse,
  TaskStatsResponse,
} from '@/types';

export function getTasks(params: TaskListRequest) {
  return request.get<never, PaginatedData<TaskResponse>>('/tasks', { params });
}

export function getMyTasks(params: TaskListRequest) {
  return request.get<never, PaginatedData<TaskResponse>>('/tasks/my', { params });
}

export function getPendingTasks(params: TaskListRequest) {
  return request.get<never, PaginatedData<TaskResponse>>('/tasks/pending', { params });
}

export function getOverdueTasks(params: TaskListRequest) {
  return request.get<never, PaginatedData<TaskResponse>>('/tasks/overdue', { params });
}

export function getTaskStats() {
  return request.get<never, TaskStatsResponse>('/tasks/stats');
}

export function getTask(id: string) {
  return request.get<never, TaskResponse>(`/tasks/${id}`);
}

export function getTaskLogs(id: string) {
  return request.get<never, TaskLogResponse[]>(`/tasks/${id}/logs`);
}

export function createTask(data: CreateTaskRequest) {
  return request.post<never, TaskResponse>('/tasks', data);
}

export function updateTask(id: string, data: UpdateTaskRequest) {
  return request.put<never, TaskResponse>(`/tasks/${id}`, data);
}

export function deleteTask(id: string) {
  return request.delete<never, null>(`/tasks/${id}`);
}

export function assignTask(id: string, data: AssignTaskRequest) {
  return request.post<never, null>(`/tasks/${id}/assign`, data);
}

export function startTask(id: string) {
  return request.post<never, null>(`/tasks/${id}/start`);
}

export function completeTask(id: string, data?: CompleteTaskRequest) {
  return request.post<never, null>(`/tasks/${id}/complete`, data);
}

export function cancelTask(id: string, data?: CancelTaskRequest) {
  return request.post<never, null>(`/tasks/${id}/cancel`, data);
}

export function updateTaskProgress(id: string, data: UpdateTaskProgressRequest) {
  return request.put<never, null>(`/tasks/${id}/progress`, data);
}
