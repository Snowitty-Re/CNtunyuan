import { http } from '@/utils/request';
import type { Task } from '@/types';

export interface CreateTaskRequest {
  title: string;
  description?: string;
  type: string;
  priority?: string;
  missing_person_id?: string;
  deadline?: string;
}

export const taskApi = {
  // 获取任务列表
  getTasks: (params?: { 
    page?: number; 
    page_size?: number; 
    status?: string;
    type?: string;
  }) => http.get<{ list: Task[]; total: number }>('/tasks', { params }),
  
  // 获取我的任务
  getMyTasks: (params?: { page?: number; page_size?: number }) => 
    http.get<{ list: Task[]; total: number }>('/tasks/my', { params }),
  
  // 获取待处理任务
  getPendingTasks: (params?: { page?: number; page_size?: number }) => 
    http.get<{ list: Task[]; total: number }>('/tasks/pending', { params }),
  
  // 获取逾期任务
  getOverdueTasks: (params?: { page?: number; page_size?: number }) => 
    http.get<{ list: Task[]; total: number }>('/tasks/overdue', { params }),
  
  // 获取任务详情
  getTaskById: (id: string) => http.get<Task>(`/tasks/${id}`),
  
  // 创建任务
  createTask: (data: CreateTaskRequest) => http.post<Task>('/tasks', data),
  
  // 更新任务
  updateTask: (id: string, data: Partial<CreateTaskRequest>) => http.put<Task>(`/tasks/${id}`, data),
  
  // 删除任务
  deleteTask: (id: string) => http.delete(`/tasks/${id}`),
  
  // 分配任务
  assignTask: (id: string, assignee_id: string) => http.post(`/tasks/${id}/assign`, { assignee_id }),
  
  // 开始任务
  startTask: (id: string) => http.post(`/tasks/${id}/start`),
  
  // 完成任务
  completeTask: (id: string, result?: string) => http.post(`/tasks/${id}/complete`, { result }),
  
  // 取消任务
  cancelTask: (id: string, reason?: string) => http.post(`/tasks/${id}/cancel`, { reason }),
  
  // 更新进度
  updateProgress: (id: string, progress: number) => http.put(`/tasks/${id}/progress`, { progress }),
  
  // 获取任务日志
  getTaskLogs: (id: string) => http.get(`/tasks/${id}/logs`),
  
  // 获取任务统计
  getTaskStats: () => http.get<{
    total: number;
    draft: number;
    pending: number;
    assigned: number;
    processing: number;
    completed: number;
    cancelled: number;
    overdue: number;
  }>('/tasks/stats'),
};
