import request from '../utils/request'
import type { Task, TaskLog, TaskComment } from '../types'

export interface TaskListParams {
  status?: string
  assignee_id?: string
  creator_id?: string
  org_id?: string
  type?: string
  priority?: string
  keyword?: string
  page?: number
  page_size?: number
}

export interface CreateTaskRequest {
  title: string
  description?: string
  type: string
  priority?: string
  missing_person_id?: string
  assignee_id?: string
  org_id: string
  start_time?: string
  deadline?: string
  estimated_hours?: number
  location?: string
  longitude?: number
  latitude?: number
  address?: string
  requirements?: string
  materials?: string[]
  notes?: string
}

export interface UpdateTaskRequest {
  title?: string
  description?: string
  type?: string
  priority?: string
  missing_person_id?: string | null
  deadline?: string
  estimated_hours?: number
  location?: string
  longitude?: number
  latitude?: number
  address?: string
  requirements?: string
  materials?: string[]
  notes?: string
}

export interface AssignTaskRequest {
  assignee_id: string
  comment?: string
}

export interface TransferTaskRequest {
  to_user_id: string
  reason?: string
}

export interface CompleteTaskRequest {
  feedback: string
  result?: string
  attachments?: string[]
  actual_hours?: number
}

export interface AddCommentRequest {
  content: string
  attachments?: string[]
  parent_id?: string
}

export const taskApi = {
  // 获取任务列表
  getTasks: (params?: TaskListParams) => 
    request.get('/tasks', { params }),

  // 获取任务详情
  getTask: (id: string) => 
    request.get(`/tasks/${id}`),

  // 创建任务
  createTask: (data: CreateTaskRequest) => 
    request.post('/tasks', data),

  // 更新任务
  updateTask: (id: string, data: UpdateTaskRequest) => 
    request.put(`/tasks/${id}`, data),

  // 删除任务
  deleteTask: (id: string) => 
    request.delete(`/tasks/${id}`),

  // 分配任务
  assignTask: (id: string, data: AssignTaskRequest) => 
    request.post(`/tasks/${id}/assign`, data),

  // 取消分配
  unassignTask: (id: string, reason?: string) => 
    request.post(`/tasks/${id}/unassign`, { reason }),

  // 转派任务
  transferTask: (id: string, data: TransferTaskRequest) => 
    request.post(`/tasks/${id}/transfer`, data),

  // 完成任务
  completeTask: (id: string, data: CompleteTaskRequest) => 
    request.post(`/tasks/${id}/complete`, data),

  // 取消任务
  cancelTask: (id: string, reason?: string) => 
    request.post(`/tasks/${id}/cancel`, { reason }),

  // 更新进度
  updateProgress: (id: string, progress: number) => 
    request.put(`/tasks/${id}/progress`, { progress }),

  // 获取我的任务
  getMyTasks: (status?: string) => 
    request.get('/tasks/my', { params: { status } }),

  // 获取我创建的任务
  getCreatedTasks: () => 
    request.get('/tasks/created'),

  // 获取任务统计
  getStatistics: (orgId?: string) => 
    request.get('/tasks/statistics', { params: { org_id: orgId } }),

  // 获取任务日志
  getTaskLogs: (id: string) => 
    request.get(`/tasks/${id}/logs`),

  // 添加评论
  addComment: (id: string, data: AddCommentRequest) => 
    request.post(`/tasks/${id}/comments`, data),

  // 获取评论列表
  getComments: (id: string) => 
    request.get(`/tasks/${id}/comments`),

  // 批量分配任务
  batchAssign: (taskIds: string[], assigneeId: string, comment?: string) => 
    request.post('/tasks/batch-assign', { task_ids: taskIds, assignee_id: assigneeId, comment }),

  // 自动分配任务
  autoAssign: (orgId?: string, limit?: number) => 
    request.post('/tasks/auto-assign', null, { params: { org_id: orgId, limit } }),
}
